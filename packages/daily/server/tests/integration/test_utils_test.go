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

	"daily/internal/application/port"
	authuc "daily/internal/application/usecase/auth"
	"daily/internal/infrastructure/api"
	"daily/internal/infrastructure/api/handler"
	"daily/internal/infrastructure/config"
	"daily/internal/infrastructure/container"
	"daily/internal/infrastructure/persistence/sqlite"
	"daily/internal/infrastructure/storage"
)

type TestApp struct {
	Router    http.Handler
	Container *container.Container
}

func setupTestApp(t *testing.T) (*TestApp, *TestEnv, func()) {
	env, cleanup := SetupTestEnv(t)

	// Determine DB engine and tag set repos
	var tagSetGroupRepo port.TagSetGroupRepository
	var tagSetRepo port.TagSetRepository
	tagSetGroupRepo = sqlite.NewSqliteTagSetGroupRepository(env.SLConn)
	tagSetRepo = sqlite.NewSqliteTagSetRepository(env.SLConn)

	tmpDir := t.TempDir()
	blobStore, err := storage.NewLocalBlobStore(tmpDir)
	if err != nil {
		t.Fatalf("failed to init blob store: %v", err)
	}

	c := &container.Container{
		MemoRepo:        env.MemoRepo,
		ResRepo:         env.ResRepo,
		UserRepo:        env.UserRepo,
		TagSetGroupRepo: tagSetGroupRepo,
		TagSetRepo:      tagSetRepo,
		Tokenizer:       env.Tokenizer,
		BlobStore:       blobStore,
	}

	// Create bootstrap admin user
	ctx := context.Background()
	identitySvc := authuc.NewIdentityService(c.UserRepo)
	if err := identitySvc.EnsureBootstrapAdmin(ctx, "admin", "password123"); err != nil {
		t.Fatalf("failed to ensure bootstrap admin: %v", err)
	}

	cfg := &config.Config{
		LogLevel:               "debug",
		BootstrapAdminUsername: "admin",
		BootstrapAdminPassword: "password123",
	}

	l := slog.New(slog.NewTextHandler(os.Stdout, nil))
	handlers := handler.NewHandlers(c, l)
	router := api.NewRouter(cfg, handlers)

	app := &TestApp{
		Router:    router,
		Container: c,
	}

	return app, env, cleanup
}

type testClient struct {
	t      *testing.T
	router http.Handler
}

func newTestClient(t *testing.T, router http.Handler) *testClient {
	return &testClient{t: t, router: router}
}

func (c *testClient) post(path string, payload interface{}, target interface{}) {
	reqBody, _ := json.Marshal(payload)
	req, _ := http.NewRequest("POST", path, bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	c.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK && w.Code != http.StatusCreated {
		c.t.Fatalf("POST %s failed: status=%d body=%s", path, w.Code, w.Body.String())
	}
	if target != nil {
		if err := json.Unmarshal(w.Body.Bytes(), target); err != nil {
			c.t.Fatalf("failed to unmarshal POST response: %v", err)
		}
	}
}

func (c *testClient) get(path string, target interface{}) {
	req, _ := http.NewRequest("GET", path, nil)
	w := httptest.NewRecorder()
	c.router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		c.t.Fatalf("GET %s failed: status=%d body=%s", path, w.Code, w.Body.String())
	}
	if target != nil {
		if err := json.Unmarshal(w.Body.Bytes(), target); err != nil {
			c.t.Fatalf("failed to unmarshal GET response: %v", err)
		}
	}
}
