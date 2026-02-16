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

func TestParent_Success(t *testing.T) {
	mock := &mockExecutor{output: "main\n"}
	client := New(mock)

	got, err := client.Parent(context.Background(), "feature-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "main" {
		t.Errorf("got %q, want %q", got, "main")
	}
	assertCommand(t, mock, "gt", []string{"parent", "--branch", "feature-a", "--no-interactive"})
}

func TestParent_TrimsWhitespace(t *testing.T) {
	mock := &mockExecutor{output: "  main  \n"}
	client := New(mock)

	got, err := client.Parent(context.Background(), "feature-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "main" {
		t.Errorf("got %q, want %q", got, "main")
	}
}

func TestParent_Error(t *testing.T) {
	mock := &mockExecutor{err: errors.New("no parent")}
	client := New(mock)

	_, err := client.Parent(context.Background(), "main")
	if err == nil {
		t.Fatal("expected error, got nil")
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
