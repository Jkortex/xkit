package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"daily/internal/application/dto"
	"daily/internal/application/port"
	"daily/internal/application/usecase/memo"
	"daily/internal/domain/entity"
	"daily/internal/infrastructure/api/handler"
	"daily/internal/infrastructure/api/middleware"
	api_presenter "daily/internal/infrastructure/api/presenter"
	persistence "daily/internal/infrastructure/persistence/sqlite"
	"daily/internal/infrastructure/tokenizer"
	"daily/internal/interfaces/controller"

	"github.com/gin-gonic/gin"
)

func TestEphemeralMemo(t *testing.T) {
	// 1. 设置测试环境
	dbConn := persistence.SetupTestDB(t)
	defer dbConn.Close()

	gseTokenizer, _ := tokenizer.NewGseTokenizer()
	repo := persistence.NewSqliteMemoRepository(dbConn)
	resRepo := persistence.NewSqliteResourceRepository(dbConn)

	memoSvc := memo.NewMemoService(repo, resRepo, repo, gseTokenizer)
	archiveUC := memo.NewArchiveExpiredMemosUseCase(repo)

	// 组装控制器与 Handler
	ctrl := controller.NewMemoController(memoSvc)
	presenter := api_presenter.NewJsonPresenter()
	l := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	memoHandler := handler.NewMemoHandler(ctrl, presenter, l)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		middleware.SetCurrentUser(c, &dto.UserResponse{ID: 1, Username: "test", Role: "admin", Status: "active"})
		c.Next()
	})
	r.POST("/api/v1/memos", memoHandler.Create)
	r.GET("/api/v1/memos", memoHandler.List)
	r.PATCH("/api/v1/memos/:uuid", memoHandler.Update)

	t.Run("Create memo with explicit TTL", func(t *testing.T) {
		body := `{"content": "TTL 1h note", "ttl": "1h"}`
		req, _ := http.NewRequest("POST", "/api/v1/memos", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
		}
		var resp dto.MemoResponse
		json.Unmarshal(w.Body.Bytes(), &resp)

		if resp.ExpiresAt == nil {
			t.Errorf("expected ExpiresAt to be set")
		} else {
			expected := time.Now().Add(time.Hour)
			diff := resp.ExpiresAt.Sub(expected)
			if diff < 0 {
				diff = -diff
			}
			if diff > 10*time.Second {
				t.Errorf("ExpiresAt %v far from expected %v", resp.ExpiresAt, expected)
			}
		}

	})

	t.Run("Create memo with #temp tag (default 3d TTL)", func(t *testing.T) {
		body := `{"content": "Temp note\n#temp"}`
		req, _ := http.NewRequest("POST", "/api/v1/memos", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("expected status 201, got %d", w.Code)
		}
		var resp dto.MemoResponse
		json.Unmarshal(w.Body.Bytes(), &resp)

		if resp.ExpiresAt == nil {
			t.Errorf("expected ExpiresAt to be set for #temp tag")
		} else {
			expected := time.Now().Add(3 * 24 * time.Hour)
			diff := resp.ExpiresAt.Sub(expected)
			if diff < 0 {
				diff = -diff
			}
			if diff > 10*time.Second {
				t.Errorf("ExpiresAt %v far from expected %v", resp.ExpiresAt, expected)
			}
		}

	})

	t.Run("Update memo to be ephemeral", func(t *testing.T) {
		// 1. 先创建一个普通笔记
		body := `{"content": "Normal note"}`
		req, _ := http.NewRequest("POST", "/api/v1/memos", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		var resp dto.MemoResponse
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.ExpiresAt != nil {
			t.Errorf("expected nil ExpiresAt for normal note")
		}

		// 2. 更新它，增加 #temp 标签
		updateBody := `{"content": "Now it is temp\n#temp"}`
		req, _ = http.NewRequest("PATCH", "/api/v1/memos/"+resp.UUID, bytes.NewBufferString(updateBody))
		req.Header.Set("Content-Type", "application/json")
		w = httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status 200, got %d", w.Code)
		}
		json.Unmarshal(w.Body.Bytes(), &resp)
		if resp.ExpiresAt == nil {
			t.Errorf("expected ExpiresAt to be set after update")
		}
	})

	t.Run("Archive expired memos", func(t *testing.T) {
		// 手动插入一个已过期的笔记
		expiredAt := time.Now().Add(-1 * time.Hour)
		m := entity.NewMemo("Expired note")
		m.ExpiresAt = &expiredAt
		err := repo.Create(context.Background(), m, 1, nil, nil, port.SearchIndex{IsEphemeral: true})
		if err != nil {
			t.Fatalf("failed to create expired memo: %v", err)
		}

		// 运行归档任务
		err = archiveUC.Execute(context.Background())
		if err != nil {
			t.Errorf("archive task failed: %v", err)
		}

		// 验证笔记状态已变为 archived
		dbMemo, _ := repo.GetByUUID(context.Background(), m.UUID, 1)
		if string(dbMemo.RowStatus) != "archived" {
			t.Errorf("expected row_status archived, got %s", dbMemo.RowStatus)
		}
	})
}
