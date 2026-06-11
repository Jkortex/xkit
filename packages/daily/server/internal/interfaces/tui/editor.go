package tui

import (
	"io"
	"os"
	"os/exec"
	"runtime"
	"sync"
)

var (
	cachedEditor string
	editorOnce   sync.Once
)

type ExecCommandWrapper struct {
	*exec.Cmd
}

func (w ExecCommandWrapper) SetStdin(r io.Reader) {
	w.Cmd.Stdin = r
}

func (w ExecCommandWrapper) SetStdout(wr io.Writer) {
	w.Cmd.Stdout = wr
}

func (w ExecCommandWrapper) SetStderr(wr io.Writer) {
	w.Cmd.Stderr = wr
}

type EditFinishedMsg struct {
	Err      error
	Content  string
	MemoUUID string
	TempPath string
}

func GetEditor() string {
	editorOnce.Do(func() {
		editor := os.Getenv("EDITOR")
		if editor != "" {
			cachedEditor = editor
			return
		}
		editors := []string{"nvim", "vim", "code"}
		if runtime.GOOS == "windows" {
			editors = append(editors, "notepad.exe")
		} else {
			editors = append(editors, "nano", "vi")
		}
		for _, e := range editors {
			if _, err := exec.LookPath(e); err == nil {
				cachedEditor = e
				return
			}
		}
		if runtime.GOOS == "windows" {
			cachedEditor = "notepad.exe"
			return
		}
		cachedEditor = "vi"
	})
	return cachedEditor
}
