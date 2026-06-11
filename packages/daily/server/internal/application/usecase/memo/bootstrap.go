package memo

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"strings"
	"time"

	"daily/internal/application/port"
	"daily/internal/domain/entity"
)

type seedMemo struct {
	MemoUUID string   `json:"memo_uuid"`
	Content  string   `json:"content"`
	Tags     []string `json:"tags"`
}

func BootstrapDemoMemos(ctx context.Context, l *slog.Logger, memoRepo port.MemoRepository, userRepo port.UserRepository, adminUsername, path string) {
	if adminUsername == "" {
		return
	}

	admin, err := userRepo.GetByUsername(ctx, adminUsername)
	if err != nil {
		l.Warn("failed to fetch admin user for demo memos", "error", err)
		return
	}

	existing, err := memoRepo.ListAll(ctx, admin.ID)
	if err != nil {
		l.Warn("failed to check existing memos", "error", err)
		return
	}
	if len(existing) > 0 {
		return
	}

	data, err := os.ReadFile(path)
	if err != nil {
		l.Warn("failed to read seed file", "path", path, "error", err)
		return
	}

	var memos []seedMemo
	if err := json.Unmarshal(data, &memos); err != nil {
		l.Warn("failed to parse seed file", "error", err)
		return
	}

	created := 0
	for _, sm := range memos {
		bodyTokens := sm.Content
		tagsTokens := strings.Join(sm.Tags, " ")
		now := time.Now().UTC()
		memo := &entity.Memo{
			UUID:      sm.MemoUUID,
			Content:   sm.Content,
			RowStatus: entity.RowStatusNormal,
			Tags:      sm.Tags,
			CreatedAt: now,
			UpdatedAt: now,
		}
		si := port.SearchIndex{
			TagsTokens:  tagsTokens,
			FilesTokens: "",
			BodyTokens:  bodyTokens,
			IsEphemeral: false,
		}
		if err := memoRepo.Create(ctx, memo, admin.ID, sm.Tags, nil, si); err != nil {
			l.Warn("failed to create seed memo", "uuid", sm.MemoUUID, "error", err)
			continue
		}
		created++
	}

	l.Info("demo memos bootstrapped", "admin", adminUsername, "count", created)
}
