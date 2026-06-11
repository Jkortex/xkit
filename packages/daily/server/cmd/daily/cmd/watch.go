package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"daily/internal/interfaces/tui"
	"github.com/fsnotify/fsnotify"
	"github.com/spf13/cobra"
)

// watchCmd 实时监听新创建的 memos（非终端模式，适合管道/工具调用）
var watchCmd = &cobra.Command{
	Use:   "watch",
	Short: "Watch for new memos in real-time",
	Long: `Poll for new memos and output them as plain text as they arrive.

Useful for piping into other tools, notifications, or lightweight monitoring.
Non-interactive mode — runs until interrupted.

  daily watch --tag work
  daily watch | while read line; do notify-send "$line"; done
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := getApp()
		if err != nil {
			return err
		}

		tagFilter, _ := cmd.Flags().GetString("tag")

		eventPath := getEventFilePath()
		if eventPath == "" {
			return fmt.Errorf("failed to resolve SQLite database event path")
		}

		_ = os.MkdirAll(filepath.Dir(eventPath), 0755)
		if _, err := os.Stat(eventPath); os.IsNotExist(err) {
			_ = os.WriteFile(eventPath, []byte("init"), 0644)
		}

		// 捞取已有最近 5 条作为初始显示
		var uiMemos []tui.UIMemo
		var lastTime string

		var initQuery string
		var initArgs []any
		if tagFilter != "" {
			initQuery = `
				SELECT m.memo_uuid, m.content, m.created_at 
				FROM memo m
				JOIN memo_tag mt ON m.memo_uuid = mt.memo_uuid
				WHERE m.row_status = 'normal' AND mt.tag_name = ?
				ORDER BY m.created_at DESC LIMIT 5
			`
			initArgs = append(initArgs, tagFilter)
		} else {
			initQuery = `
				SELECT memo_uuid, content, created_at 
				FROM memo 
				WHERE row_status = 'normal'
				ORDER BY created_at DESC LIMIT 5
			`
		}

		rows, err := app.DB.QueryContext(cmd.Context(), initQuery, initArgs...)
		if err == nil {
			defer rows.Close()
			var list []tui.UIMemo
			for rows.Next() {
				var muuid, content, createdAt string
				if err := rows.Scan(&muuid, &content, &createdAt); err == nil {
					list = append(list, tui.UIMemo{
						UUID:      muuid,
						Content:   content,
						CreatedAt: createdAt,
					})
				}
			}
			for i := len(list) - 1; i >= 0; i-- {
				uiMemos = append(uiMemos, list[i])
			}
		}

		err = app.DB.QueryRowContext(cmd.Context(), "SELECT COALESCE(MAX(created_at), CURRENT_TIMESTAMP) FROM memo").Scan(&lastTime)
		if err != nil {
			lastTime = time.Now().UTC().Format("2006-01-02 15:04:05")
		}

		// 非终端模式：简单轮询输出
		watcher, err := fsnotify.NewWatcher()
		if err != nil {
			return fmt.Errorf("create watcher: %w", err)
		}
		defer watcher.Close()
		_ = watcher.Add(eventPath)

		if len(uiMemos) > 0 {
			latest := uiMemos[len(uiMemos)-1]
			fmt.Printf("Status: Listening for new memos... | Received Memo [%s] (%s): %s\n", tui.SimplifyTime(latest.CreatedAt), latest.UUID[:8], tui.GetShortSummary(latest.Content))
		} else {
			fmt.Println("Status: Listening for new memos...")
		}

		for {
			select {
			case <-cmd.Context().Done():
				return nil
			case ev, ok := <-watcher.Events:
				if !ok {
					return nil
				}
				if ev.Has(fsnotify.Write) {
					var query string
					var qargs []any
					qargs = append(qargs, lastTime)
					if tagFilter != "" {
						query = `SELECT m.memo_uuid, m.content, m.created_at FROM memo m JOIN memo_tag mt ON m.memo_uuid = mt.memo_uuid WHERE m.created_at > ? AND m.row_status = 'normal' AND mt.tag_name = ? ORDER BY m.created_at ASC`
						qargs = append(qargs, tagFilter)
					} else {
						query = `SELECT memo_uuid, content, created_at FROM memo WHERE created_at > ? AND row_status = 'normal' ORDER BY created_at ASC`
					}
					r, err := app.DB.QueryContext(cmd.Context(), query, qargs...)
					if err == nil {
						for r.Next() {
							var muuid, content, createdAt string
							if err := r.Scan(&muuid, &content, &createdAt); err == nil {
								lastTime = createdAt
								fmt.Printf("Status: New memo | Received [%s] (%s): %s\n", tui.SimplifyTime(createdAt), muuid[:8], tui.GetShortSummary(content))
							}
						}
						r.Close()
					}
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return nil
				}
				return err
			}
		}
	},
}

func init() {
	watchCmd.Flags().String("tag", "", "Filter by a single tag to watch")
	rootCmd.AddCommand(watchCmd)
}
