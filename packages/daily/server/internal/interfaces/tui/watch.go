package tui

import (
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	tea "github.com/charmbracelet/bubbletea"
)

type EventMsg struct{}
type ErrMsg struct{ Err error }

func WatchFile(eventPath string) tea.Cmd {
	return func() tea.Msg {
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return ErrMsg{err}
		}
		defer watcher.Close()
		if err := watcher.Add(eventPath); err != nil {
			return ErrMsg{err}
		}
		for {
			select {
			case ev, ok := <-watcher.Events:
				if !ok {
					return nil
				}
				if ev.Has(fsnotify.Write) {
					return EventMsg{}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return nil
				}
				return ErrMsg{err}
			}
		}
	}
}

func EnsureEventPath(path string) error {
	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.WriteFile(path, []byte("init"), 0644)
	}
	return nil
}
