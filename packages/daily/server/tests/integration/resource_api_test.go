package integration

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"daily/internal/application/dto"
	"daily/internal/application/usecase/resource"
	"daily/internal/infrastructure/api/handler"
	"daily/internal/infrastructure/api/middleware"
	api_presenter "daily/internal/infrastructure/api/presenter"
	persistence "daily/internal/infrastructure/persistence/sqlite"
	"daily/internal/infrastructure/storage"
	"daily/internal/interfaces/controller"

	"github.com/gin-gonic/gin"
)

func TestResourceAPI_UploadAndGet(t *testing.T) {
	// 1. 设置测试环境
	dbConn := persistence.SetupTestDB(t)
	defer dbConn.Close()

	memoRepo := persistence.NewSqliteMemoRepository(dbConn)
	resRepo := persistence.NewSqliteResourceRepository(dbConn)

	tmpDir, _ := os.MkdirTemp("", "daily-test-*")
	defer os.RemoveAll(tmpDir)
	store, _ := storage.NewLocalBlobStore(tmpDir)

	uploadUC := resource.NewUploadResourceUseCase(resRepo, store)
	getUC := resource.NewGetResourceUseCase(resRepo, store)
	exportUC := resource.NewExportDataUseCase(memoRepo, resRepo, store)
	importUC := resource.NewImportDataUseCase(memoRepo, resRepo, store)

	// 组装新架构链路
	ctrl := controller.NewResourceController(uploadUC, getUC, exportUC, importUC)
	presenter := api_presenter.NewJsonPresenter()
	l := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	resHandler := handler.NewResourceHandler(ctrl, presenter, l)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		middleware.SetCurrentUser(c, &dto.UserResponse{
			ID:       1,
			Username: "test",
			Role:     "admin",
			Status:   "active",
		})
		c.Next()
	})
	r.POST("/api/v1/resources", resHandler.Upload)
	r.GET("/api/v1/resources/:id", resHandler.Get)
	r.POST("/api/v1/system/import", resHandler.Import)

	var resourceID string

	t.Run("Upload file", func(t *testing.T) {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "hello.txt")
		part.Write([]byte("hello daily content"))
		writer.Close()

		req, _ := http.NewRequest("POST", "/api/v1/resources", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d, body: %s", w.Code, w.Body.String())
		}

		var res struct {
			ID string `json:"id"`
		}
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		resourceID = res.ID
	})

	t.Run("Get file content", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/resources/"+resourceID, nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Get failed, got code %d", w.Code)
		}
		if w.Body.String() != "hello daily content" {
			t.Errorf("Content mismatch, got %s", w.Body.String())
		}
	})

	t.Run("Get missing resource should return 404", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/resources/not-exist", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("Expected 404, got %d", w.Code)
		}
	})

	t.Run("Import returns structured report", func(t *testing.T) {
		var zipBuf bytes.Buffer
		zw := zip.NewWriter(&zipBuf)

		resMetaFile, _ := zw.Create("resources.json")
		_, _ = resMetaFile.Write([]byte("[]"))
		memoMetaFile, _ := zw.Create("memos.json")
		_, _ = memoMetaFile.Write([]byte("[]"))
		_ = zw.Close()

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "import.zip")
		_, _ = part.Write(zipBuf.Bytes())
		_ = writer.Close()

		req, _ := http.NewRequest("POST", "/api/v1/system/import", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d, body=%s", w.Code, w.Body.String())
		}

		var payload map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &payload); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if _, ok := payload["report"]; !ok {
			t.Fatalf("Expected report in response, got: %v", payload)
		}
	})
}
