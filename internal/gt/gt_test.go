package gt

import (
	"context"
	"errors"
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
