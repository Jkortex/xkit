package notify

import (
	"os"
	"path/filepath"
	"time"
)

var dsn string

func Init(sqliteDSN string) {
	dsn = sqliteDSN
}

func Touch() {
	if dsn == "" {
		return
	}
	absDB, err := filepath.Abs(dsn)
	if err != nil {
		return
	}
	eventPath := filepath.Join(filepath.Dir(absDB), ".daily_event")
	_ = os.MkdirAll(filepath.Dir(eventPath), 0755)
	f, err := os.OpenFile(eventPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	_, _ = f.WriteString(time.Now().Format(time.RFC3339Nano) + "\n")
}
