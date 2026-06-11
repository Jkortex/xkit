package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"daily/internal/application/port"

	"github.com/spf13/cobra"
)

// ============================================================
// tree 命令 — 通用树协议
//
// 一棵任意深度的 key-value 树，展平为 tagged memos。
// 关系通过 tag 表达，重建时按 tag 还原树结构。
//
// 输入格式:
//
//	{
//	  "root-key": {
//	    "content": "标题",
//	    "key": "plan",
//	    "tags": ["sprint-24"],
//	    "children": [
//	      {"child-key": {"content": "...", "key": "task", "tags": ["est:4h"]}}
//	    ]
//	  }
//	}
//
// 展平为 memos:
//   root: id="root-key", parent="", root="root-key", key="plan", ...
//   leaf: id="root-key.child-key", parent="root-key", root="root-key", key="task", ...
// ============================================================

var treeCmd = &cobra.Command{
	Use:   "tree",
	Short: "Manage tree documents (protocol: arbitrary key-value trees → flat memos)",
	Long: `Tree documents: arbitrary-depth key-value trees flattened into tagged memos.

  Protocol:
    - Each node becomes a memo with tags: id:<path>, parent:<parent-path>, root:<root-id>, key:<type>
    - User tags from the node are added directly as tag values
    - Content is freeform text (supporting markdown, ADR, issue reports, etc.)
    - Tree is reconstructed from tags for display

  Examples:
    daily-cli tree put < input.json
    daily-cli tree get user-auth
    daily-cli tree list
    daily-cli tree delete user-auth
`,
}

// TreeNode 表示树中的一个节点（接收输入时使用原始 JSON 结构）
type TreeNode struct {
	Content  string                `json:"content"`
	Key      string                `json:"key"`
	Tags     []string              `json:"tags"`
	Children []map[string]TreeNode `json:"children,omitempty"`
}

// FlatNode 展平后的节点（用于输出和 memos）
type FlatNode struct {
	UUID     string     `json:"uuid,omitempty"`
	ID       string     `json:"id"`
	Parent   string     `json:"parent"`
	Root     string     `json:"root"`
	Key      string     `json:"key"`
	Content  string     `json:"content"`
	Tags     []string   `json:"tags"`
	Children []FlatNode `json:"children,omitempty"`
}

// nodeEntry 解析输入时的中间表示
type nodeEntry struct {
	Key      string
	Parent   string
	Content  string
	NodeKey  string
	Tags     []string
	Children []map[string]TreeNode
}

// ============================================================
// tree put
// ============================================================

var treePutCmd = &cobra.Command{
	Use:   "put",
	Short: "Create or update a tree document from JSON input",
	Long: `Read a JSON tree from stdin or --file and flatten it into tagged memos.

  Input is a JSON object: each top-level key is a root node.
  Multiple roots can be specified at once.

  # From stdin
  cat plan.json | daily-cli tree put

  # From file
  daily-cli tree put --file plan.json
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := getApp()
		if err != nil {
			return err
		}

		// Read input
		rawJSON, err := readInput(cmd)
		if err != nil {
			return fmt.Errorf("read input: %w", err)
		}

		// Parse: top-level object, keys are root node keys
		var input map[string]TreeNode
		if err := json.Unmarshal(rawJSON, &input); err != nil {
			return fmt.Errorf("parse JSON: %w", err)
		}

		// Flatten tree → flat nodes
		flattenIdx = 0
		var flat []nodeEntry
		for rootKey, node := range input {
			// rootKey is the node's path identifier; node.Key is its type key
			flattenTree(rootKey, "", rootKey, node, &flat)
		}

		// Create memos
		type createdNode struct {
			Key     string   `json:"key"`
			ID      string   `json:"id"`
			UUID    string   `json:"uuid"`
			Content string   `json:"content"`
			Tags    []string `json:"tags"`
		}
		created := make([]createdNode, 0, len(flat))
		for _, entry := range flat {
			memoTags := buildMemoTags(entry)
			// Guard content against TagExtractor misinterpreting #-prefixed lines as tags
			safeContent := makeTagSafe(entry.Content)
			m, err := app.MemoCtrl.Create(cmd.Context(), app.AdminID, safeContent, memoTags, nil, "")
			if err != nil {
				return fmt.Errorf("create memo for %q: %w", entry.Key, err)
			}
			created = append(created, createdNode{
				Key:     entry.Key,
				ID:      entry.NodeKey,
				UUID:    m.UUID,
				Content: entry.Content, // original (un-guarded) content in response
				Tags:    m.Tags,
			})
		}

		return printJSON(map[string]interface{}{
			"status":  "ok",
			"count":   len(created),
			"created": created,
		})
	},
}

// flattenTree 递归展平树形结构为扁平的 nodeEntry 列表
var flattenIdx int

func flattenTree(nodeKey, parentKey, rootKey string, node TreeNode, out *[]nodeEntry) {
	flattenIdx++
	// 当前节点
	*out = append(*out, nodeEntry{
		Key:      node.Key, // type key: "plan", "task", "reference"
		Parent:   parentKey,
		Content:  node.Content,
		NodeKey:  nodeKey, // path-based ID: "user-auth.login"
		Tags:     append(node.Tags, fmt.Sprintf("idx:%d", flattenIdx)),
		Children: node.Children,
	})

	// 递归子节点
	for _, child := range node.Children {
		for childKey, childNode := range child {
			// childKey is the path identifier; childNode.Key is its type key
			childID := nodeKey + "." + childKey
			flattenTree(childID, nodeKey, rootKey, childNode, out)
		}
	}
}

// buildMemoTags 从展平后的节点条目构建 memo tags
func buildMemoTags(entry nodeEntry) []string {
	tags := []string{
		"id:" + entry.NodeKey,
		"key:" + entry.Key,
		"root:" + extractRootID(entry.NodeKey),
	}
	if entry.Parent != "" {
		tags = append(tags, "parent:"+entry.Parent)
	}
	// 用户自定义 tags
	for _, t := range entry.Tags {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

// extractRootID 从 nodeKey 中提取根 ID（第一个"."之前的部分）
func extractRootID(nodeKey string) string {
	if idx := strings.Index(nodeKey, "."); idx > 0 {
		return nodeKey[:idx]
	}
	return nodeKey
}

// ============================================================
// tree get
// ============================================================

var treeGetCmd = &cobra.Command{
	Use:   "get [root-id]",
	Short: "Reconstruct and display a tree document",
	Long: `Query all memos belonging to root-id, rebuild the tree structure.

  daily-cli tree get user-auth
  daily-cli tree get xkit-main.adr-001
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := getApp()
		if err != nil {
			return err
		}

		rootID := args[0]

		memos, err := app.MemoCtrl.List(cmd.Context(), app.AdminID, port.MemoFilter{
			TagsAll: []string{"root:" + rootID},
			Limit:   500,
			Sort:    "created_at",
		})
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		if len(memos) == 0 {
			return printJSON(map[string]string{"status": "not found", "root": rootID})
		}

		// Debug: count
		// Parse flat memos into flat nodes
		nodes := make([]FlatNode, 0, len(memos))
		for _, m := range memos {
			flat := FlatNode{
				UUID:    m.UUID,
				Content: stripTagGuard(m.Content),
			}
			for _, t := range m.Tags {
				switch {
				case strings.HasPrefix(t, "id:"):
					flat.ID = strings.TrimPrefix(t, "id:")
				case strings.HasPrefix(t, "parent:"):
					flat.Parent = strings.TrimPrefix(t, "parent:")
				case strings.HasPrefix(t, "root:"):
					flat.Root = strings.TrimPrefix(t, "root:")
				case strings.HasPrefix(t, "key:"):
					flat.Key = strings.TrimPrefix(t, "key:")
				default:
					flat.Tags = append(flat.Tags, t)
				}
			}
			nodes = append(nodes, flat)
		}

		// Build tree
		tree := buildTree(nodes)

		// Output
		format, _ := cmd.Flags().GetString("format")
		if format == "flat" {
			return printJSON(nodes)
		}
		return printJSON(tree)
	},
}

// buildTree 从扁平的 nodes 列表中按 parent 关系重建树
func buildTree(nodes []FlatNode) []FlatNode {
	if len(nodes) == 0 {
		return nil
	}

	// Build lookup with pointers into the slice
	byID := make(map[string]*FlatNode, len(nodes))
	for i := range nodes {
		byID[nodes[i].ID] = &nodes[i]
	}

	// Collect root pointers (no parent), attach children to their parents via pointer
	var rootPtrs []*FlatNode
	for i := range nodes {
		node := &nodes[i]
		if node.Parent == "" {
			rootPtrs = append(rootPtrs, node)
		} else {
			if parent, ok := byID[node.Parent]; ok {
				parent.Children = append(parent.Children, *node)
			}
		}
	}

	// Build result slice (copy roots, each already has populated children)
	roots := make([]FlatNode, 0, len(rootPtrs))
	for _, r := range rootPtrs {
		// Sort children by ID
		var sortChildren func(parent *FlatNode)
		sortChildren = func(parent *FlatNode) {
			sort.Slice(parent.Children, func(i, j int) bool {
				return parent.Children[i].ID < parent.Children[j].ID
			})
			for i := range parent.Children {
				sortChildren(&parent.Children[i])
			}
		}
		sortChildren(r)
		roots = append(roots, *r)
	}

	if roots == nil {
		return []FlatNode{}
	}
	return roots
}

// buildTreeOrdered 按创建顺序重建树（不排序 children，保留 flatten 顺序）
func buildTreeOrdered(nodes []FlatNode) []FlatNode {
	if len(nodes) == 0 {
		return nil
	}

	byID := make(map[string]*FlatNode, len(nodes))
	for i := range nodes {
		byID[nodes[i].ID] = &nodes[i]
	}

	var rootPtrs []*FlatNode
	for i := range nodes {
		node := &nodes[i]
		if node.Parent == "" {
			rootPtrs = append(rootPtrs, node)
		} else {
			if parent, ok := byID[node.Parent]; ok {
				parent.Children = append(parent.Children, *node)
			}
		}
	}

	roots := make([]FlatNode, 0, len(rootPtrs))
	for _, r := range rootPtrs {
		roots = append(roots, *r)
	}

	if roots == nil {
		return []FlatNode{}
	}
	return roots
}

// ============================================================
// tree list
// ============================================================

var treeListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all root tree documents",
	Long: `List all root tree documents (memos with no "parent:" tag but have "root:" and "key:" tags).

  daily-cli tree list
  daily-cli tree list --key plan
`,
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := getApp()
		if err != nil {
			return err
		}

		keyFilter, _ := cmd.Flags().GetString("key")

		// Get all memos with "root:" tag — these are tree nodes
		filter := port.MemoFilter{
			Limit: 1000,
			Sort:  "created_at",
		}

		// We need to get all memos and filter client-side for "root:" tag
		// since the tag filter might not support regex matching
		allMemos, err := app.MemoCtrl.List(cmd.Context(), app.AdminID, filter)
		if err != nil {
			return fmt.Errorf("list memos: %w", err)
		}

		// Collect unique roots
		rootSet := make(map[string]*FlatNode)
		for _, m := range allMemos {
			var hasRoot bool
			var rootID, nodeKey string
			for _, t := range m.Tags {
				if strings.HasPrefix(t, "root:") {
					rootID = strings.TrimPrefix(t, "root:")
					hasRoot = true
				}
				if strings.HasPrefix(t, "key:") {
					nodeKey = strings.TrimPrefix(t, "key:")
				}
			}
			if !hasRoot || rootID == "" {
				continue
			}
			// Only root nodes (no parent tag)
			hasParent := false
			for _, t := range m.Tags {
				if strings.HasPrefix(t, "parent:") {
					hasParent = true
					break
				}
			}
			if hasParent {
				continue
			}

			// Apply key filter
			if keyFilter != "" && nodeKey != keyFilter {
				continue
			}

			if _, exists := rootSet[rootID]; !exists {
				rootSet[rootID] = &FlatNode{
					ID:      rootID,
					Key:     nodeKey,
					Content: truncateContent(m.Content, 80),
					UUID:    m.UUID,
				}
			}
		}

		// Sort roots by ID
		roots := make([]FlatNode, 0, len(rootSet))
		for _, r := range rootSet {
			roots = append(roots, *r)
		}
		sort.Slice(roots, func(i, j int) bool {
			return roots[i].ID < roots[j].ID
		})

		if roots == nil {
			roots = []FlatNode{}
		}

		return printJSON(roots)
	},
}

// ============================================================
// tree delete
// ============================================================

var treeDeleteCmd = &cobra.Command{
	Use:   "delete [root-id]",
	Short: "Delete an entire tree document",
	Long: `Delete all memos belonging to root-id (all nodes in the tree).

  daily-cli tree delete user-auth
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := getApp()
		if err != nil {
			return err
		}

		rootID := args[0]

		memos, err := app.MemoCtrl.List(cmd.Context(), app.AdminID, port.MemoFilter{
			TagsAll: []string{"root:" + rootID},
			Limit:   500,
			Sort:    "created_at",
		})
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		if len(memos) == 0 {
			return printJSON(map[string]string{"status": "not found", "root": rootID})
		}

		uuids := make([]string, len(memos))
		for i, m := range memos {
			uuids[i] = m.UUID
		}

		result, err := app.MemoCtrl.BatchDelete(cmd.Context(), app.AdminID, uuids)
		if err != nil {
			return fmt.Errorf("delete: %w", err)
		}

		return printJSON(map[string]interface{}{
			"status":  "ok",
			"root":    rootID,
			"deleted": len(uuids),
			"result":  result,
		})
	},
}

// ============================================================
// 辅助函数
// ============================================================

// readInput 从 --file 或 stdin 读取 JSON
func readInput(cmd *cobra.Command) ([]byte, error) {
	filePath, _ := cmd.Flags().GetString("file")
	if filePath != "" {
		return readFileBytes(filePath)
	}
	return readStdin()
}

func truncateContent(content string, maxLen int) string {
	// Take first line, truncated
	firstLine := content
	if idx := strings.Index(content, "\n"); idx > 0 {
		firstLine = content[:idx]
	}
	runes := []rune(firstLine)
	if len(runes) > maxLen {
		return string(runes[:maxLen]) + "..."
	}
	return firstLine
}

// makeTagSafe ensures content doesn't end with lines starting with #,
// which the TagExtractor would misinterpret as tags. Appends a guard
// suffix that tree_get strips on retrieval.
func makeTagSafe(content string) string {
	lines := strings.Split(content, "\n")
	// Find last non-empty line
	lastNonEmpty := -1
	for i := len(lines) - 1; i >= 0; i-- {
		if strings.TrimSpace(lines[i]) != "" {
			lastNonEmpty = i
			break
		}
	}
	if lastNonEmpty >= 0 && strings.HasPrefix(strings.TrimSpace(lines[lastNonEmpty]), "#") {
		// Last non-empty line starts with # → would be mis-parsed as tag
		// Append guard: newline + U+200B (zero-width space, not in unicode.IsSpace)
		return content + "\n\u200b"
	}
	return content
}

// stripTagGuard removes the tag-safety guard appended by makeTagSafe.
func stripTagGuard(content string) string {
	const guard = "\n\u200b"
	return strings.TrimSuffix(content, guard)
}

// readStdin 从标准输入读取所有字节
func readStdin() ([]byte, error) {
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return nil, fmt.Errorf("read stdin: %w", err)
	}
	return data, nil
}

// readFileBytes 读取文件的全部字节
func readFileBytes(path string) ([]byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read file %q: %w", path, err)
	}
	return data, nil
}

// ============================================================
// tree export
// ============================================================

var treeExportCmd = &cobra.Command{
	Use:   "export [root-id]",
	Short: "Render a tree as markdown",
	Long: `Fetch a tree by root-id and render it as human-readable markdown.

  Root content → H1
  Children grouped by key type → H2 sections with bullet lists

  daily-cli tree export sprint-25
  daily-cli tree export bug-43
`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		app, err := getApp()
		if err != nil {
			return err
		}

		rootID := args[0]

		memos, err := app.MemoCtrl.List(cmd.Context(), app.AdminID, port.MemoFilter{
			TagsAll: []string{"root:" + rootID},
			Limit:   500,
			Sort:    "created_at_asc",
		})
		if err != nil {
			return fmt.Errorf("list: %w", err)
		}
		if len(memos) == 0 {
			return fmt.Errorf("root %q not found", rootID)
		}

		// Parse into flat nodes preserving creation order (created_at ASC)
		// Root memo comes first, then children in DFS flatten order
		nodes := make([]FlatNode, 0, len(memos))
		for _, m := range memos {
			flat := FlatNode{UUID: m.UUID, Content: stripTagGuard(m.Content)}
			for _, t := range m.Tags {
				switch {
				case strings.HasPrefix(t, "id:"):
					flat.ID = strings.TrimPrefix(t, "id:")
				case strings.HasPrefix(t, "parent:"):
					flat.Parent = strings.TrimPrefix(t, "parent:")
				case strings.HasPrefix(t, "key:"):
					flat.Key = strings.TrimPrefix(t, "key:")
				case t == "done":
					flat.Tags = append(flat.Tags, t)
				default:
					flat.Tags = append(flat.Tags, t)
				}
			}
			nodes = append(nodes, flat)
		}

		// Build tree preserving creation order (don't use buildTree which sorts alphabetically)
		// Sort by idx tag to preserve DFS flatten order from tree_put
		sort.Slice(nodes, func(i, j int) bool {
			return nodeIdx(nodes[i]) < nodeIdx(nodes[j])
		})
		roots := buildTreeOrdered(nodes)
		if len(roots) == 0 {
			return fmt.Errorf("no root node found for %q", rootID)
		}
		// Render markdown
		var buf strings.Builder

		for _, root := range roots {
			// H1: root content
			for _, line := range strings.Split(root.Content, "\n") {
				line = strings.TrimSpace(line)
				if line != "" {
					buf.WriteString("# " + line + "\n")
					break
				}
			}
			if root.Content != "" {
				buf.WriteString("\n")
			}

			if len(root.Children) == 0 {
				continue
			}

			// Iterate children in creation order, no grouping
			// (grouping by key would reorder, which is wrong for narrative documents)
			for _, child := range root.Children {
				renderChild(&buf, child, child.Key)
			}
		}

		fmt.Print(buf.String())
		return nil
	},
}

// groupLabel returns a human-readable section heading for a key type.
func groupLabel(key string) string {
	switch key {
	case "task":
		return "任务"
	case "validate":
		return "验证"
	case "risk":
		return "风险"
	case "reference":
		return "参考"
	case "step":
		return "复现步骤"
	case "decision":
		return "决策"
	case "note":
		return "备注"
	case "paragraph":
		return "概述"
	case "quote":
		return "引用"
	case "code":
		return "代码"
	default:
		return key
	}
}

// renderChild renders a single child node as markdown, with recursion for nested children.
func renderChild(buf *strings.Builder, node FlatNode, key string) {
	isDone := false
	for _, t := range node.Tags {
		if t == "done" {
			isDone = true
			break
		}
	}

	prefix := "- "
	switch key {
	case "task":
		if isDone {
			prefix = "- [x] "
		} else {
			prefix = "- [ ] "
		}
	case "validate":
		prefix = "- [ ] "
	case "risk":
		prefix = "> ⚠️ "
	case "quote":
		prefix = "> "
	case "code":
		buf.WriteString("```\n" + strings.TrimSpace(node.Content) + "\n```\n")
		return
	case "reference":
		prefix = "- 📄 "
	}

	// content: first line as summary, rest indented
	lines := strings.SplitN(node.Content, "\n", 2)
	firstLine := strings.TrimSpace(lines[0])
	if firstLine != "" {
		buf.WriteString(prefix + firstLine + "\n")
	}
	if len(lines) > 1 && strings.TrimSpace(lines[1]) != "" {
		for _, line := range strings.Split(lines[1], "\n") {
			trimmed := strings.TrimRight(line, " \t")
			buf.WriteString("  " + trimmed + "\n")
		}
	}

	// Recursively render grandchildren
	for _, gc := range node.Children {
		gcContent := strings.TrimSpace(gc.Content)
		if gcContent != "" {
			buf.WriteString("    - " + firstLineOf(gcContent) + "\n")
		}
	}
}

func firstLineOf(s string) string {
	lines := strings.SplitN(s, "\n", 2)
	return strings.TrimSpace(lines[0])
}

// nodeIdx extracts the idx:N tag value for ordering, returns max int if not found.
func nodeIdx(n FlatNode) int {
	for _, t := range n.Tags {
		if strings.HasPrefix(t, "idx:") {
			var v int
			if _, err := fmt.Sscanf(t, "idx:%d", &v); err == nil {
				return v
			}
		}
	}
	return 999999
}

func groupPriority(k string) int {
	switch k {
	case "task":
		return 0
	case "validate":
		return 1
	case "risk":
		return 2
	case "reference":
		return 3
	case "note":
		return 4
	case "decision":
		return 5
	default:
		return 100
	}
}

func init() {
	treePutCmd.Flags().String("file", "", "Read JSON input from file instead of stdin")
	treeGetCmd.Flags().String("format", "tree", "Output format: tree (default) or flat")
	treeListCmd.Flags().String("key", "", "Filter by key type (e.g. plan, issue, adr, context)")
	treeListCmd.Flags().String("status", "", "Filter by status: active, archived")

	treeCmd.AddCommand(treePutCmd)
	treeCmd.AddCommand(treeGetCmd)
	treeCmd.AddCommand(treeListCmd)
	treeCmd.AddCommand(treeExportCmd)
	treeCmd.AddCommand(treeDeleteCmd)
	rootCmd.AddCommand(treeCmd)
}
