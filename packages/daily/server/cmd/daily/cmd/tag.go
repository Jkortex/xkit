package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// tagCmd 表示 tag 子命令树
var tagCmd = &cobra.Command{
	Use:   "tag",
	Short: "Manage tags",
}

var tagListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all tags with memo counts",
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := getApp()
		if err != nil {
			return err
		}

		tags, err := app.TagCtrl.ListTags(cmd.Context(), app.AdminID)
		if err != nil {
			return fmt.Errorf("list tags: %w", err)
		}

		return printJSON(tags)
	},
}

var tagRenameCmd = &cobra.Command{
	Use:   "rename [from] [to]",
	Short: "Rename a tag (all memos updated)",
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := getApp()
		if err != nil {
			return err
		}

		res, err := app.TagCtrl.RenameTag(cmd.Context(), app.AdminID, args[0], args[1])
		if err != nil {
			return fmt.Errorf("rename tag: %w", err)
		}

		return printJSON(res)
	},
}

var tagMergeCmd = &cobra.Command{
	Use:   "merge [target] [sources...]",
	Short: "Merge multiple tags into one",
	Args:  cobra.MinimumNArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := getApp()
		if err != nil {
			return err
		}

		target := args[0]
		sources := args[1:]

		res, err := app.TagCtrl.MergeTags(cmd.Context(), app.AdminID, sources, target)
		if err != nil {
			return fmt.Errorf("merge tags: %w", err)
		}

		return printJSON(res)
	},
}

func init() {
	tagCmd.AddCommand(tagListCmd)
	tagCmd.AddCommand(tagRenameCmd)
	tagCmd.AddCommand(tagMergeCmd)
	rootCmd.AddCommand(tagCmd)
}
