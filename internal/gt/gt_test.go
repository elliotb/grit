package gt

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type mockExecutor struct {
	output     string
	err        error
	calledName string
	calledArgs []string
}

func (m *mockExecutor) Execute(ctx context.Context, name string, args ...string) (string, error) {
	m.calledName = name
	m.calledArgs = args
	return m.output, m.err
}

func TestLogShort_Success(t *testing.T) {
	want := "◉ main\n├── feature-a\n└── feature-b\n"
	mock := &mockExecutor{output: want}
	client := New(mock)

	got, err := client.LogShort(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
	if mock.calledName != "gt" {
		t.Errorf("called %q, want %q", mock.calledName, "gt")
	}
	wantArgs := []string{"log", "short", "--no-interactive"}
	if len(mock.calledArgs) != len(wantArgs) {
		t.Fatalf("got args %v, want %v", mock.calledArgs, wantArgs)
	}
	for i, arg := range wantArgs {
		if mock.calledArgs[i] != arg {
			t.Errorf("arg[%d] = %q, want %q", i, mock.calledArgs[i], arg)
		}
	}
}

func TestLogShort_Error(t *testing.T) {
	mock := &mockExecutor{err: errors.New("gt not found")}
	client := New(mock)

	_, err := client.LogShort(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func assertArgs(t *testing.T, mock *mockExecutor, wantArgs []string) {
	t.Helper()
	if mock.calledName != "gt" {
		t.Errorf("called %q, want %q", mock.calledName, "gt")
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

func TestCheckout_Success(t *testing.T) {
	mock := &mockExecutor{}
	client := New(mock)

	err := client.Checkout(context.Background(), "feature-a")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertArgs(t, mock, []string{"checkout", "feature-a", "--no-interactive"})
}

func TestCheckout_Error(t *testing.T) {
	mock := &mockExecutor{err: errors.New("branch not found")}
	client := New(mock)

	err := client.Checkout(context.Background(), "nonexistent")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestStackSubmit_Success(t *testing.T) {
	mock := &mockExecutor{}
	client := New(mock)

	err := client.StackSubmit(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertArgs(t, mock, []string{"stack", "submit", "--no-interactive"})
}

func TestStackSubmit_Error(t *testing.T) {
	mock := &mockExecutor{err: errors.New("submit failed")}
	client := New(mock)

	err := client.StackSubmit(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDownstackSubmit_Success(t *testing.T) {
	mock := &mockExecutor{}
	client := New(mock)

	err := client.DownstackSubmit(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertArgs(t, mock, []string{"downstack", "submit", "--no-interactive"})
}

func TestDownstackSubmit_Error(t *testing.T) {
	mock := &mockExecutor{err: errors.New("submit failed")}
	client := New(mock)

	err := client.DownstackSubmit(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestStackRestack_Success(t *testing.T) {
	mock := &mockExecutor{}
	client := New(mock)

	err := client.StackRestack(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertArgs(t, mock, []string{"stack", "restack", "--no-interactive"})
}

func TestStackRestack_Error(t *testing.T) {
	mock := &mockExecutor{err: errors.New("restack failed")}
	client := New(mock)

	err := client.StackRestack(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestRepoSync_Success(t *testing.T) {
	mock := &mockExecutor{}
	client := New(mock)

	err := client.RepoSync(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertArgs(t, mock, []string{"repo", "sync", "--no-interactive"})
}

func TestRepoSync_Error(t *testing.T) {
	mock := &mockExecutor{err: errors.New("sync failed")}
	client := New(mock)

	err := client.RepoSync(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestOpenPR_Success(t *testing.T) {
	mock := &mockExecutor{}
	client := New(mock)

	err := client.OpenPR(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	assertArgs(t, mock, []string{"pr"})
}

func TestOpenPR_Error(t *testing.T) {
	mock := &mockExecutor{err: errors.New("no PR found")}
	client := New(mock)

	err := client.OpenPR(context.Background())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestExecCommandExecutor_Echo(t *testing.T) {
	exec := &ExecCommandExecutor{}
	got, err := exec.Execute(context.Background(), "echo", "hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != "hello\n" {
		t.Errorf("got %q, want %q", got, "hello\n")
	}
}

func TestExecCommandExecutor_Failure(t *testing.T) {
	exec := &ExecCommandExecutor{}
	_, err := exec.Execute(context.Background(), "false")
	if err == nil {
		t.Fatal("expected error from 'false' command, got nil")
	}
}

func TestExecCommandExecutor_StderrInError(t *testing.T) {
	exec := &ExecCommandExecutor{}
	// bash -c 'echo error message >&2; exit 1' writes to stderr and exits with 1
	_, err := exec.Execute(context.Background(), "bash", "-c", "echo error message >&2; exit 1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "error message") {
		t.Errorf("error should contain stderr output, got: %q", err.Error())
	}
}
