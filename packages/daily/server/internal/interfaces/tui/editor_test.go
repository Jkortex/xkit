package tui

import (
	"bytes"
	"errors"
	"os/exec"
	"testing"
)

func TestGetEditor(t *testing.T) {
	editor := GetEditor()
	if editor == "" {
		t.Fatal("expected non-empty editor (should fall back to vi)")
	}
}

func TestExecCommandWrapper(t *testing.T) {
	cmd := exec.Command("echo", "hello")
	w := ExecCommandWrapper{Cmd: cmd}

	var stdout bytes.Buffer
	w.SetStdout(&stdout)
	w.SetStderr(&stdout)

	var stdin bytes.Buffer
	stdin.WriteString("test")
	w.SetStdin(&stdin)

	if w.Cmd.Stdout != &stdout {
		t.Error("Stdout not set correctly")
	}
}

func TestEditFinishedMsg(t *testing.T) {
	msg := EditFinishedMsg{
		Err:      nil,
		Content:  "test content",
		MemoUUID: "uuid-1",
		TempPath: "/tmp/test",
	}
	if msg.Content != "test content" {
		t.Errorf("Content = %q, want %q", msg.Content, "test content")
	}
	if msg.MemoUUID != "uuid-1" {
		t.Errorf("MemoUUID = %q, want %q", msg.MemoUUID, "uuid-1")
	}
	if msg.TempPath != "/tmp/test" {
		t.Errorf("TempPath = %q, want %q", msg.TempPath, "/tmp/test")
	}
	if msg.Err != nil {
		t.Errorf("Err = %v, want nil", msg.Err)
	}

	// Test with error
	myErr := errors.New("test error")
	errMsg := EditFinishedMsg{Err: myErr}
	if errMsg.Err != myErr {
		t.Errorf("Err = %v, want %v", errMsg.Err, myErr)
	}
}
