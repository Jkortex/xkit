package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBatchAPI(t *testing.T) {
	app, _, teardown := setupTestApp(t)
	defer teardown()

	router := app.Router
	client := newTestClient(t, router)

	u1Client := client

	// Create memos for User 1
	var m1, m2, m3 struct {
		UUID string   `json:"uuid"`
		Tags []string `json:"tags"`
	}

	u1Client.post("/api/v1/memos", map[string]interface{}{"content": "m1 #t1"}, &m1)
	u1Client.post("/api/v1/memos", map[string]interface{}{"content": "m2 #t2"}, &m2)
	u1Client.post("/api/v1/memos", map[string]interface{}{"content": "m3 #t3"}, &m3)

	t.Run("BatchTag", func(t *testing.T) {
		req := map[string]interface{}{
			"uuids":  []string{m1.UUID, m2.UUID},
			"add":    []string{"newTag"},
			"remove": []string{"t1"},
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/api/v1/memos/batch/tag", bytes.NewBuffer(body))
		router.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
		}

		var res struct {
			Succeeded []string `json:"succeeded"`
			Failed    []struct {
				UUID   string `json:"uuid"`
				Reason string `json:"reason"`
			} `json:"failed"`
		}
		json.Unmarshal(w.Body.Bytes(), &res)

		if len(res.Succeeded) != 2 {
			t.Errorf("expected 2 succeeded, got %d", len(res.Succeeded))
		}

		// Verify tags
		u1Client.get("/api/v1/memos/"+m1.UUID, &m1)
		if len(m1.Tags) != 1 || m1.Tags[0] != "newTag" {
			t.Errorf("expected [newTag] on m1, got %v", m1.Tags)
		}
	})

	t.Run("BatchArchive", func(t *testing.T) {
		req := map[string]interface{}{
			"uuids": []string{m1.UUID, "invalid-uuid-format"},
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/api/v1/memos/batch/archive", bytes.NewBuffer(body))
		router.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
		}

		var res struct {
			Succeeded []string `json:"succeeded"`
			Failed    []struct {
				UUID   string `json:"uuid"`
				Reason string `json:"reason"`
			} `json:"failed"`
		}
		json.Unmarshal(w.Body.Bytes(), &res)

		if len(res.Succeeded) != 1 || res.Succeeded[0] != m1.UUID {
			t.Errorf("expected 1 succeeded, got %v", res.Succeeded)
		}
		if len(res.Failed) != 1 {
			t.Errorf("expected 1 failed, got %v", res.Failed)
		}
	})

	t.Run("BatchDelete", func(t *testing.T) {
		req := map[string]interface{}{
			"uuids": []string{m1.UUID, m2.UUID},
		}

		body, _ := json.Marshal(req)
		w := httptest.NewRecorder()
		r, _ := http.NewRequest("POST", "/api/v1/memos/batch/delete", bytes.NewBuffer(body))
		router.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d", w.Code)
		}

		var res struct {
			Succeeded []string      `json:"succeeded"`
			Failed    []interface{} `json:"failed"`
		}
		json.Unmarshal(w.Body.Bytes(), &res)

		if len(res.Succeeded) != 2 {
			t.Errorf("expected 2 succeeded, got %v", res.Succeeded)
		}
	})
}
