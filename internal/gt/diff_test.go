package gt

import (
	"context"
	"errors"
	"testing"
)

func assertCommand(t *testing.T, mock *mockExecutor, wantName string, wantArgs []string) {
	t.Helper()
	if mock.calledName != wantName {
		t.Errorf("called %q, want %q", mock.calledName, wantName)
	}
	if len(mock.calledArgs) != len(wantArgs) {
		t.Fatalf("got args %v, want %v", mock.calledArgs, wantArgs)
	}
	for i, arg := range wantArgs {
		if mock.calledArgs[i] != arg {
			t.Errorf("arg[%d] = %q, want %q", i, mock.calledArgs[i], arg)
		}
	}
}

func TestDiffStat_Success(t *testing.T) {
	want := " file.go | 5 +++--\n 1 file changed\n"
	mock := &mockExecutor{output: want}
	client := New(mock)

	got, err := client.DiffStat(context.Background(), "main", "feature-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	assertCommand(t, mock, "git", []string{"diff", "--stat", "main...feature-a"})
}

func TestDiffStat_Error(t *testing.T) {
	mock := &mockExecutor{err: errors.New("diff failed")}
	client := New(mock)

	_, err := client.DiffStat(context.Background(), "main", "feature-a")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDiffFile_Success(t *testing.T) {
	want := "diff --git a/file.go b/file.go\n+added line\n"
	mock := &mockExecutor{output: want}
	client := New(mock)

	got, err := client.DiffFile(context.Background(), "main", "feature-a", "file.go")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	assertCommand(t, mock, "git", []string{"diff", "--color=always", "main...feature-a", "--", "file.go"})
}

func TestDiffFile_Error(t *testing.T) {
	mock := &mockExecutor{err: errors.New("diff failed")}
	client := New(mock)

	_, err := client.DiffFile(context.Background(), "main", "feature-a", "file.go")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestFindParent_DirectChild(t *testing.T) {
	branches := []*Branch{
		{Name: "main", Children: []*Branch{
			{Name: "feature-a"},
		}},
	}
	parent, ok := FindParent(branches, "feature-a")
	if !ok {
		t.Fatal("expected to find parent")
	}
	if parent != "main" {
		t.Errorf("got %q, want %q", parent, "main")
	}
}

func TestFindParent_NestedChild(t *testing.T) {
	branches := []*Branch{
		{Name: "main", Children: []*Branch{
			{Name: "feature-a", Children: []*Branch{
				{Name: "feature-b"},
			}},
		}},
	}
	parent, ok := FindParent(branches, "feature-b")
	if !ok {
		t.Fatal("expected to find parent")
	}
	if parent != "feature-a" {
		t.Errorf("got %q, want %q", parent, "feature-a")
	}
}

func TestFindParent_RootBranch(t *testing.T) {
	branches := []*Branch{
		{Name: "main", Children: []*Branch{
			{Name: "feature-a"},
		}},
	}
	_, ok := FindParent(branches, "main")
	if ok {
		t.Error("root branch should have no parent")
	}
}

func TestFindParent_NotFound(t *testing.T) {
	branches := []*Branch{
		{Name: "main", Children: []*Branch{
			{Name: "feature-a"},
		}},
	}
	_, ok := FindParent(branches, "nonexistent")
	if ok {
		t.Error("nonexistent branch should not be found")
	}
}

func TestFindParent_EmptyTree(t *testing.T) {
	_, ok := FindParent(nil, "feature-a")
	if ok {
		t.Error("should not find parent in empty tree")
	}
}

func TestFindParent_MultipleBranches(t *testing.T) {
	branches := []*Branch{
		{Name: "main", Children: []*Branch{
			{Name: "feature-a", Children: []*Branch{
				{Name: "feature-a2"},
			}},
			{Name: "feature-b"},
		}},
	}

	// feature-b's parent should be main
	parent, ok := FindParent(branches, "feature-b")
	if !ok {
		t.Fatal("expected to find parent for feature-b")
	}
	if parent != "main" {
		t.Errorf("got %q, want %q", parent, "main")
	}

	// feature-a2's parent should be feature-a
	parent, ok = FindParent(branches, "feature-a2")
	if !ok {
		t.Fatal("expected to find parent for feature-a2")
	}
	if parent != "feature-a" {
		t.Errorf("got %q, want %q", parent, "feature-a")
	}
}
