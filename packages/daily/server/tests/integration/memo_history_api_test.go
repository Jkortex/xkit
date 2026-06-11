package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"daily/internal/application/dto"
	"daily/internal/application/usecase/memo"
	"daily/internal/infrastructure/api/handler"
	"daily/internal/infrastructure/api/middleware"
	api_presenter "daily/internal/infrastructure/api/presenter"
	persistence "daily/internal/infrastructure/persistence/sqlite"
	"daily/internal/infrastructure/tokenizer"
	"daily/internal/interfaces/controller"
	"github.com/gin-gonic/gin"
)

func TestMemoHistoryAPI_Workflow(t *testing.T) {
	dbConn := persistence.SetupTestDB(t)
	defer dbConn.Close()

	gseTokenizer, _ := tokenizer.NewGseTokenizer()
	repo := persistence.NewSqliteMemoRepository(dbConn)
	resRepo := persistence.NewSqliteResourceRepository(dbConn)

	memoSvc := memo.NewMemoService(repo, resRepo, repo, gseTokenizer)
	memoCtrl := controller.NewMemoController(memoSvc)
	historyCtrl := controller.NewMemoHistoryController(memoSvc)

	l := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	presenter := api_presenter.NewJsonPresenter()
	memoHandler := handler.NewMemoHandler(memoCtrl, presenter, l)
	historyHandler := handler.NewMemoHistoryHandler(historyCtrl, presenter, l)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		middleware.SetCurrentUser(c, &dto.UserResponse{ID: 1, Role: "member"})
		c.Next()
	})
	r.POST("/api/v1/memos", memoHandler.Create)
	r.PATCH("/api/v1/memos/:uuid", memoHandler.Update)
	r.GET("/api/v1/memos/:uuid/history", historyHandler.ListHistory)
	r.POST("/api/v1/memos/:uuid/rollback/:hid", historyHandler.Rollback)

	// 1. Create
	createResp := performRequest(r, "POST", "/api/v1/memos", `{"content":"original content"}`)
	var created dto.MemoResponse
	json.Unmarshal(createResp.Body.Bytes(), &created)

	// 2. Update (should trigger history)
	performRequest(r, "PATCH", fmt.Sprintf("/api/v1/memos/%s", created.UUID), `{"content":"updated content"}`)

	// 3. List History
	historyResp := performRequest(r, "GET", fmt.Sprintf("/api/v1/memos/%s/history", created.UUID), "")
	var histories []*dto.MemoHistoryResponse
	json.Unmarshal(historyResp.Body.Bytes(), &histories)

	if len(histories) != 1 {
		t.Fatalf("Expected 1 history record, got %d", len(histories))
	}
	if histories[0].Content != "original content" {
		t.Errorf("History content mismatch: %s", histories[0].Content)
	}

	// 4. Rollback
	rollbackResp := performRequest(r, "POST", fmt.Sprintf("/api/v1/memos/%s/rollback/%s", created.UUID, histories[0].ID), "")
	if rollbackResp.Code != http.StatusOK {
		t.Fatalf("Rollback failed: %d", rollbackResp.Code)
	}

	var rolledBack dto.MemoResponse
	json.Unmarshal(rollbackResp.Body.Bytes(), &rolledBack)
	if rolledBack.Content != "original content" {
		t.Errorf("Rollback failed to restore content: %s", rolledBack.Content)
	}
}

func performRequest(r http.Handler, method, path, body string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}
