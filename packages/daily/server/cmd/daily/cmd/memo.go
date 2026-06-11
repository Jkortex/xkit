package cmd

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"daily/internal/infrastructure/notify"

	"daily/internal/application/dto"
	"daily/internal/application/port"
	"daily/internal/domain/entity"

	"github.com/spf13/cobra"
)

// memoCmd 表示 memo 子命令树 — 无子命令时显示帮助
var memoCmd = &cobra.Command{
	Use:   "memo",
	Short: "Manage memos (non-interactive, JSON output)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return cmd.Help()
	},
}

var memoCreateCmd = &cobra.Command{
	Use:   "create [content]",
	Short: "Create a new memo",
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := getApp()
		if err != nil {
			return err
		}

		var content string
		fileOption, _ := cmd.Flags().GetString("file")
		var absPath string
		var hash string

		if fileOption != "" {
			bytes, err := os.ReadFile(fileOption)
			if err != nil {
				return fmt.Errorf("read file %s: %w", fileOption, err)
			}
			content = string(bytes)

			absPath, _ = filepath.Abs(fileOption)
			hash = fmt.Sprintf("%x", md5.Sum(bytes))

			// Check database mapping for non-intrusive metadata
			var existingUUID, existingHash string
			err = app.DB.QueryRowContext(cmd.Context(), "SELECT memo_uuid, content_hash FROM workspace_file_sync WHERE file_path = ?", absPath).Scan(&existingUUID, &existingHash)
			if err == nil {
				// Mapping exists!
				if existingHash == hash {
					// Content unchanged: load existing memo and return
					m, err := app.MemoCtrl.Get(cmd.Context(), app.AdminID, existingUUID)
					if err == nil {
						return printJSON(m)
					}
				}

				// Content changed or memo got deleted: update it
				tags, _ := cmd.Flags().GetStringSlice("tag")
				input := dto.UpdateMemoRequest{
					Content: content,
					Tags:    tags,
				}
				m, err := app.MemoCtrl.Update(cmd.Context(), app.AdminID, existingUUID, input)
				if err != nil {
					// Fallthrough to create new if update failed (e.g. memo was physically deleted)
				} else {
					// Update the hash mapping
					_, _ = app.DB.ExecContext(cmd.Context(), "UPDATE workspace_file_sync SET content_hash = ?, updated_at = CURRENT_TIMESTAMP WHERE file_path = ?", hash, absPath)
					return printJSON(m)
				}
			}
		} else {
			if len(args) != 1 {
				return fmt.Errorf("content or --file is required")
			}
			content = args[0]
		}

		tags, _ := cmd.Flags().GetStringSlice("tag")
		ttl, _ := cmd.Flags().GetString("ttl")
		resourceIDs, _ := cmd.Flags().GetStringSlice("resource")

		memo, err := app.MemoCtrl.Create(
			cmd.Context(),
			app.AdminID,
			content,
			tags,
			resourceIDs,
			ttl,
		)
		if err != nil {
			return fmt.Errorf("create memo: %w", err)
		}

		// Save mapping if file was provided
		if fileOption != "" {
			_, _ = app.DB.ExecContext(cmd.Context(), "INSERT OR REPLACE INTO workspace_file_sync (file_path, memo_uuid, content_hash) VALUES (?, ?, ?)", absPath, memo.UUID, hash)
		}

		touchEventFile()
		return printJSON(memo)
	},
}

var memoListCmd = &cobra.Command{
	Use:   "list",
	Short: "List memos with optional filters",
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := getApp()
		if err != nil {
			return err
		}

		filter := buildFilter(cmd)

		memos, err := app.MemoCtrl.List(cmd.Context(), app.AdminID, filter)
		if err != nil {
			return fmt.Errorf("list memos: %w", err)
		}

		return printJSON(memos)
	},
}

var memoGetCmd = &cobra.Command{
	Use:   "get [uuid]",
	Short: "Get a memo by UUID",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := getApp()
		if err != nil {
			return err
		}

		m, err := app.MemoCtrl.Get(cmd.Context(), app.AdminID, args[0])
		if err != nil {
			return fmt.Errorf("get memo: %w", err)
		}

		return printJSON(m)
	},
}

var memoUpdateCmd = &cobra.Command{
	Use:   "update [uuid] [content]",
	Short: "Update a memo's content, tags, or TTL",
	Args:  cobra.RangeArgs(1, 2),
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := getApp()
		if err != nil {
			return err
		}

		uuid := args[0]
		var content string
		fileOption, _ := cmd.Flags().GetString("file")
		if fileOption != "" {
			bytes, err := os.ReadFile(fileOption)
			if err != nil {
				return fmt.Errorf("read file %s: %w", fileOption, err)
			}
			content = string(bytes)
		} else {
			if len(args) != 2 {
				return fmt.Errorf("content or --file is required")
			}
			content = args[1]
		}

		tags, _ := cmd.Flags().GetStringSlice("tag")
		ttl, _ := cmd.Flags().GetString("ttl")

		input := dto.UpdateMemoRequest{
			Content:    content,
			Tags:       tags,
			TimeToLive: ttl,
		}

		m, err := app.MemoCtrl.Update(cmd.Context(), app.AdminID, uuid, input)
		if err != nil {
			return fmt.Errorf("update memo: %w", err)
		}

		touchEventFile()
		return printJSON(m)
	},
}

var memoDeleteCmd = &cobra.Command{
	Use:   "delete [uuid]",
	Short: "Delete a memo permanently",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := getApp()
		if err != nil {
			return err
		}

		if err := app.MemoCtrl.Delete(cmd.Context(), app.AdminID, args[0]); err != nil {
			return fmt.Errorf("delete memo: %w", err)
		}

		touchEventFile()
		return printJSON(map[string]string{"status": "deleted", "uuid": args[0]})
	},
}

var memoArchiveCmd = &cobra.Command{
	Use:   "archive [uuid]",
	Short: "Archive a memo (single-item batch)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := getApp()
		if err != nil {
			return err
		}

		res, err := app.MemoCtrl.BatchArchive(cmd.Context(), app.AdminID, []string{args[0]})
		if err != nil {
			return fmt.Errorf("archive memo: %w", err)
		}

		return printJSON(res)
	},
}

func init() {
	// create flags
	memoCreateCmd.Flags().StringSlice("tag", nil, "Tags to attach (can specify multiple: --tag a --tag b)")
	memoCreateCmd.Flags().String("ttl", "", "Time-to-live (e.g. 3d, 1w)")
	memoCreateCmd.Flags().StringSlice("resource", nil, "Resource IDs to attach")
	memoCreateCmd.Flags().StringP("file", "f", "", "Path to a file containing the memo content")

	// update flags
	memoUpdateCmd.Flags().StringSlice("tag", nil, "Replace tags (can specify multiple)")
	memoUpdateCmd.Flags().String("ttl", "", "Time-to-live (e.g. 3d, 1w)")
	memoUpdateCmd.Flags().StringP("file", "f", "", "Path to a file containing the updated content")

	// list flags
	memoListCmd.Flags().String("search", "", "Full-text search query")
	memoListCmd.Flags().String("tag", "", "Filter by a single tag")
	memoListCmd.Flags().String("from", "", "Start date (YYYY-MM-DD)")
	memoListCmd.Flags().String("to", "", "End date (YYYY-MM-DD)")
	memoListCmd.Flags().String("sort", "created_at", "Sort field: created_at, updated_at")
	memoListCmd.Flags().Int("limit", 50, "Max results")
	memoListCmd.Flags().Int("offset", 0, "Result offset")
	memoListCmd.Flags().String("status", "normal", "Row status: normal, archived")
	memoListCmd.Flags().Bool("include-resources", false, "Include resource metadata")

	memoCmd.AddCommand(memoCreateCmd)
	memoCmd.AddCommand(memoListCmd)
	memoCmd.AddCommand(memoGetCmd)
	memoCmd.AddCommand(memoUpdateCmd)
	memoCmd.AddCommand(memoDeleteCmd)
	memoCmd.AddCommand(memoArchiveCmd)

	rootCmd.AddCommand(memoCmd)
}

func getEffectiveDBPath() string {
	effectiveDB := dbPath
	if v, ok := os.LookupEnv("DAILY_SQLITE_DSN"); ok && effectiveDB == "~/.daily/daily.db" {
		effectiveDB = v
	}
	return expandHome(effectiveDB)
}

func getEventFilePath() string {
	absDB, err := filepath.Abs(getEffectiveDBPath())
	if err != nil {
		return ""
	}
	return filepath.Join(filepath.Dir(absDB), ".daily_event")
}

func touchEventFile() {
	notify.Init(getEffectiveDBPath())
	notify.Touch()
}

// buildFilter 从 CLI flags 构建 MemoFilter
func buildFilter(cmd *cobra.Command) port.MemoFilter {
	f := port.MemoFilter{
		Sort:   "created_at",
		Limit:  50,
		Offset: 0,
	}

	if v, _ := cmd.Flags().GetString("search"); v != "" {
		f.Search = &v
	}
	if v, _ := cmd.Flags().GetString("tag"); v != "" {
		f.Tag = &v
	}
	if v, _ := cmd.Flags().GetString("from"); v != "" {
		f.FromDate = &v
	}
	if v, _ := cmd.Flags().GetString("to"); v != "" {
		f.ToDate = &v
	}
	if v, _ := cmd.Flags().GetString("sort"); v != "" {
		f.Sort = v
	}
	if v, _ := cmd.Flags().GetInt("limit"); v > 0 {
		f.Limit = v
	}
	if v, _ := cmd.Flags().GetInt("offset"); v > 0 {
		f.Offset = v
	}
	if v, _ := cmd.Flags().GetBool("include-resources"); v {
		f.IncludeResources = true
	}
	if v, _ := cmd.Flags().GetString("status"); v != "" {
		s := entity.RowStatus(strings.ToLower(v))
		f.RowStatus = &s
	}

	return f
}

// printJSON 将值格式化为 JSON 并输出到 stdout
func printJSON(v any) error {
	out, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}
	fmt.Println(string(out))
	return nil
}
