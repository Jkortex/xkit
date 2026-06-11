package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"daily/internal/application/dto"
	"daily/internal/application/port"
	"daily/internal/application/usecase/memo"
	"daily/internal/domain/entity"
	"daily/internal/infrastructure/api/handler"
	"daily/internal/infrastructure/api/middleware"
	api_presenter "daily/internal/infrastructure/api/presenter"
	"daily/internal/interfaces/controller"

	"github.com/gin-gonic/gin"
)

func TestMemoAPI_CreateAndSearch(t *testing.T) {
	// 1. 设置测试环境
	env, cleanup := SetupTestEnv(t)
	defer cleanup()

	repo := env.MemoRepo
	resRepo := env.ResRepo

	memoSvc := memo.NewMemoService(repo, resRepo, repo, env.Tokenizer)
	tagSvc := memo.NewTagService(repo)

	// 组装新架构链路
	ctrl := controller.NewMemoController(memoSvc)
	tagCtrl := controller.NewTagController(tagSvc)
	historyCtrl := controller.NewMemoHistoryController(memoSvc)

	presenter := api_presenter.NewJsonPresenter()
	memoHandler := handler.NewMemoHandler(ctrl, presenter, env.Logger)
	tagHandler := handler.NewTagHandler(tagCtrl, presenter, env.Logger)
	historyHandler := handler.NewMemoHistoryHandler(historyCtrl, presenter, env.Logger)

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
	r.POST("/api/v1/memos", memoHandler.Create)
	r.GET("/api/v1/memos", memoHandler.List)
	r.PATCH("/api/v1/memos/:uuid", memoHandler.Update)
	r.DELETE("/api/v1/memos/:uuid", memoHandler.Delete)
	r.POST("/api/v1/tags/rename", tagHandler.RenameTag)
	r.POST("/api/v1/tags/merge", tagHandler.MergeTags)
	r.GET("/api/v1/tags/aliases", tagHandler.ListTagAliases)
	r.POST("/api/v1/tags/aliases", tagHandler.UpsertTagAlias)
	r.DELETE("/api/v1/tags/aliases/:alias", tagHandler.DeleteTagAlias)
	r.GET("/api/v1/tags/audits", tagHandler.TagAudits)
	r.GET("/api/v1/memos/:uuid/history", historyHandler.ListHistory)
	r.POST("/api/v1/memos/:uuid/rollback/:history_id", historyHandler.Rollback)

	t.Run("Create a memo", func(t *testing.T) {
		body := `{"content": "集成测试笔记\n#Integration"}`
		req, _ := http.NewRequest("POST", "/api/v1/memos", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected 201, got %d", w.Code)
		}

		var res dto.MemoResponse
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if res.UUID == "" || len(res.Tags) != 1 {
			t.Errorf("Invalid memo returned: %+v", res)
		}
	})

	t.Run("Search for the memo", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/memos?search=Integration", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d", w.Code)
		}

		var res []*dto.MemoResponse
		if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if len(res) == 0 {
			t.Fatal("Search should return the created memo")
		}
	})

	t.Run("Advanced filters should work", func(t *testing.T) {
		ctx := context.Background()
		m := entity.NewMemo("第二条 #Ops #Infra")
		m.Tags = []string{"Ops", "Infra"}
		if err := repo.Create(ctx, m, 1, []string{"Ops", "Infra"}, nil, port.SearchIndex{BodyTokens: m.Content}); err != nil {
			t.Fatalf("seed memo failed: %v", err)
		}
		res := &entity.Resource{
			ID:           "res-filter",
			FileName:     "f.png",
			Hash:         "hash-filter",
			Size:         1,
			MimeType:     "image/png",
			InternalPath: "2026/03/hash-filter.png",
		}
		if err := resRepo.Save(ctx, res, 1); err != nil {
			t.Fatalf("seed resource failed: %v", err)
		}
		if err := resRepo.LinkToMemo(ctx, "res-filter", m.UUID, 1); err != nil {
			t.Fatalf("seed resource link failed: %v", err)
		}

		req, _ := http.NewRequest(
			"GET",
			"/api/v1/memos?tags_all=Ops,Infra&has_resource=true&sort=updated_at_desc&include_resources=true",
			nil,
		)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Fatalf("Expected 200, got %d, body=%s", w.Code, w.Body.String())
		}

		var list []*dto.MemoResponse
		if err := json.Unmarshal(w.Body.Bytes(), &list); err != nil {
			t.Fatalf("Unmarshal failed: %v", err)
		}
		if len(list) != 1 || list[0].UUID != m.UUID {
			t.Fatalf("unexpected filtered list: %+v", list)
		}
		if len(list[0].Resources) != 1 || list[0].Resources[0].ID != "res-filter" {
			t.Fatalf("expected resources in list response, got %+v", list[0].Resources)
		}
	})

	t.Run("Update should replace memo resources", func(t *testing.T) {
		createBody := `{"content":"资源更新测试 #Res"}`
		createReq, _ := http.NewRequest("POST", "/api/v1/memos", bytes.NewBufferString(createBody))
		createReq.Header.Set("Content-Type", "application/json")
		createResp := httptest.NewRecorder()
		r.ServeHTTP(createResp, createReq)
		if createResp.Code != http.StatusCreated {
			t.Fatalf("create memo failed: code=%d body=%s", createResp.Code, createResp.Body.String())
		}
		var created dto.MemoResponse
		if err := json.Unmarshal(createResp.Body.Bytes(), &created); err != nil {
			t.Fatalf("unmarshal create response failed: %v", err)
		}

		ctx := context.Background()
		firstRes := &entity.Resource{
			ID:           "res-updated-1",
			FileName:     "u1.png",
			Hash:         "hash-updated-1",
			Size:         1,
			MimeType:     "image/png",
			InternalPath: "2026/03/hash-updated-1.png",
		}
		secondRes := &entity.Resource{
			ID:           "res-updated-2",
			FileName:     "u2.png",
			Hash:         "hash-updated-2",
			Size:         1,
			MimeType:     "image/png",
			InternalPath: "2026/03/hash-updated-2.png",
		}
		if err := resRepo.Save(ctx, firstRes, 1); err != nil {
			t.Fatalf("seed first resource failed: %v", err)
		}
		if err := resRepo.Save(ctx, secondRes, 1); err != nil {
			t.Fatalf("seed second resource failed: %v", err)
		}

		updateBody := `{"content":"资源更新测试\n#Res #Updated","resource_ids":["res-updated-1","res-updated-2"]}`
		updateReq, _ := http.NewRequest(
			"PATCH",
			"/api/v1/memos/"+created.UUID,
			bytes.NewBufferString(updateBody),
		)
		updateReq.Header.Set("Content-Type", "application/json")
		updateResp := httptest.NewRecorder()
		r.ServeHTTP(updateResp, updateReq)
		if updateResp.Code != http.StatusOK {
			t.Fatalf("update memo failed: code=%d body=%s", updateResp.Code, updateResp.Body.String())
		}

		listReq, _ := http.NewRequest(
			"GET",
			"/api/v1/memos?tag=Updated&include_resources=true",
			nil,
		)
		listResp := httptest.NewRecorder()
		r.ServeHTTP(listResp, listReq)
		if listResp.Code != http.StatusOK {
			t.Fatalf("list updated memo failed: code=%d body=%s", listResp.Code, listResp.Body.String())
		}
		var list []*dto.MemoResponse
		if err := json.Unmarshal(listResp.Body.Bytes(), &list); err != nil {
			t.Fatalf("unmarshal updated list failed: %v", err)
		}
		if len(list) == 0 {
			t.Fatal("expected updated memo in list")
		}
		if len(list[0].Resources) != 2 {
			t.Fatalf("expected 2 resources after update, got %d", len(list[0].Resources))
		}
	})

	t.Run("Invalid date filter should return 400", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/memos?from=2026-99-99", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("Expected 400, got %d", w.Code)
		}
	})

	t.Run("Invalid date range should return 400", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/memos?from=2026-04-01&to=2026-03-01", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("Expected 400, got %d", w.Code)
		}
	})

	t.Run("Invalid has_resource should return 400", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/memos?has_resource=unknown", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("Expected 400, got %d", w.Code)
		}
	})

	t.Run("Update with invalid memo id should return 400", func(t *testing.T) {
		body := `{"content":"updated"}`
		req, _ := http.NewRequest("PATCH", "/api/v1/memos/invalid-id", bytes.NewBufferString(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("Expected 400, got %d", w.Code)
		}
	})

	t.Run("Delete not found memo should return 404", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/v1/memos/00000000-0000-0000-0000-000000000000", nil)
		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Fatalf("Expected 404, got %d", w.Code)
		}
	})

	t.Run("Rename tag should merge to target tag", func(t *testing.T) {
		createBody := `{"content":"标签治理\n#Legacy #Ops"}`
		createReq, _ := http.NewRequest("POST", "/api/v1/memos", bytes.NewBufferString(createBody))
		createReq.Header.Set("Content-Type", "application/json")
		createResp := httptest.NewRecorder()
		r.ServeHTTP(createResp, createReq)
		if createResp.Code != http.StatusCreated {
			t.Fatalf("seed create failed, code=%d", createResp.Code)
		}

		renameBody := `{"from":"Legacy","to":"Ops"}`
		renameReq, _ := http.NewRequest("POST", "/api/v1/tags/rename", bytes.NewBufferString(renameBody))
		renameReq.Header.Set("Content-Type", "application/json")
		renameResp := httptest.NewRecorder()
		r.ServeHTTP(renameResp, renameReq)
		if renameResp.Code != http.StatusOK {
			t.Fatalf("rename tag failed, code=%d body=%s", renameResp.Code, renameResp.Body.String())
		}

		listReq, _ := http.NewRequest("GET", "/api/v1/memos?tag=Legacy", nil)
		listResp := httptest.NewRecorder()
		r.ServeHTTP(listResp, listReq)
		if listResp.Code != http.StatusOK {
			t.Fatalf("list by old tag failed, code=%d", listResp.Code)
		}
		var legacyList []*dto.MemoResponse
		if err := json.Unmarshal(listResp.Body.Bytes(), &legacyList); err != nil {
			t.Fatalf("unmarshal old tag list failed: %v", err)
		}
		if len(legacyList) != 0 {
			t.Fatalf("old tag should be empty after rename, got=%d", len(legacyList))
		}

		listReq2, _ := http.NewRequest("GET", "/api/v1/memos?tag=Ops", nil)
		listResp2 := httptest.NewRecorder()
		r.ServeHTTP(listResp2, listReq2)
		if listResp2.Code != http.StatusOK {
			t.Fatalf("list by target tag failed, code=%d", listResp2.Code)
		}
		var opsList []*dto.MemoResponse
		if err := json.Unmarshal(listResp2.Body.Bytes(), &opsList); err != nil {
			t.Fatalf("unmarshal target tag list failed: %v", err)
		}
		if len(opsList) == 0 {
			t.Fatal("target tag should return memo after rename")
		}
	})

	t.Run("Merge tags should merge multiple sources to target", func(t *testing.T) {
		createBody := `{"content":"批量治理\n#Infra #Platform #Ops"}`
		createReq, _ := http.NewRequest("POST", "/api/v1/memos", bytes.NewBufferString(createBody))
		createReq.Header.Set("Content-Type", "application/json")
		createResp := httptest.NewRecorder()
		r.ServeHTTP(createResp, createReq)
		if createResp.Code != http.StatusCreated {
			t.Fatalf("seed create failed, code=%d", createResp.Code)
		}

		mergeBody := `{"sources":["Infra","Platform","NotExist"],"target":"Ops"}`
		mergeReq, _ := http.NewRequest("POST", "/api/v1/tags/merge", bytes.NewBufferString(mergeBody))
		mergeReq.Header.Set("Content-Type", "application/json")
		mergeResp := httptest.NewRecorder()
		r.ServeHTTP(mergeResp, mergeReq)
		if mergeResp.Code != http.StatusOK {
			t.Fatalf("merge tags failed, code=%d body=%s", mergeResp.Code, mergeResp.Body.String())
		}

		infraReq, _ := http.NewRequest("GET", "/api/v1/memos?tag=Infra", nil)
		infraResp := httptest.NewRecorder()
		r.ServeHTTP(infraResp, infraReq)
		if infraResp.Code != http.StatusOK {
			t.Fatalf("list by infra tag failed, code=%d", infraResp.Code)
		}
		var infraList []*dto.MemoResponse
		if err := json.Unmarshal(infraResp.Body.Bytes(), &infraList); err != nil {
			t.Fatalf("unmarshal infra tag list failed: %v", err)
		}
		if len(infraList) != 0 {
			t.Fatalf("infra should be empty after merge, got=%d", len(infraList))
		}

		opsReq, _ := http.NewRequest("GET", "/api/v1/memos?tag=Ops", nil)
		opsResp := httptest.NewRecorder()
		r.ServeHTTP(opsResp, opsReq)
		if opsResp.Code != http.StatusOK {
			t.Fatalf("list by ops tag failed, code=%d", opsResp.Code)
		}
		var opsList []*dto.MemoResponse
		if err := json.Unmarshal(opsResp.Body.Bytes(), &opsList); err != nil {
			t.Fatalf("unmarshal ops tag list failed: %v", err)
		}
		if len(opsList) == 0 {
			t.Fatal("ops should contain memo after merge")
		}
	})

	t.Run("Tag alias should normalize create and search", func(t *testing.T) {
		aliasBody := `{"alias":"SRE","canonical":"Ops"}`
		aliasReq, _ := http.NewRequest("POST", "/api/v1/tags/aliases", bytes.NewBufferString(aliasBody))
		aliasReq.Header.Set("Content-Type", "application/json")
		aliasResp := httptest.NewRecorder()
		r.ServeHTTP(aliasResp, aliasReq)
		if aliasResp.Code != http.StatusOK {
			t.Fatalf("upsert alias failed, code=%d body=%s", aliasResp.Code, aliasResp.Body.String())
		}

		createBody := `{"content":"别名输入\n#SRE"}`
		createReq, _ := http.NewRequest("POST", "/api/v1/memos", bytes.NewBufferString(createBody))
		createReq.Header.Set("Content-Type", "application/json")
		createResp := httptest.NewRecorder()
		r.ServeHTTP(createResp, createReq)
		if createResp.Code != http.StatusCreated {
			t.Fatalf("create with alias tag failed, code=%d body=%s", createResp.Code, createResp.Body.String())
		}

		opsReq, _ := http.NewRequest("GET", "/api/v1/memos?tag=Ops", nil)
		opsResp := httptest.NewRecorder()
		r.ServeHTTP(opsResp, opsReq)
		if opsResp.Code != http.StatusOK {
			t.Fatalf("search by canonical failed, code=%d", opsResp.Code)
		}
		var opsList []*dto.MemoResponse
		if err := json.Unmarshal(opsResp.Body.Bytes(), &opsList); err != nil {
			t.Fatalf("unmarshal canonical list failed: %v", err)
		}
		if len(opsList) == 0 {
			t.Fatal("canonical tag should return memo created with alias")
		}

		aliasSearchReq, _ := http.NewRequest("GET", "/api/v1/memos?tag=SRE", nil)
		aliasSearchResp := httptest.NewRecorder()
		r.ServeHTTP(aliasSearchResp, aliasSearchReq)
		if aliasSearchResp.Code != http.StatusOK {
			t.Fatalf("search by alias failed, code=%d", aliasSearchResp.Code)
		}
		var aliasList []*dto.MemoResponse
		if err := json.Unmarshal(aliasSearchResp.Body.Bytes(), &aliasList); err != nil {
			t.Fatalf("unmarshal alias list failed: %v", err)
		}
		if len(aliasList) == 0 {
			t.Fatal("alias search should resolve to canonical and return memo")
		}

		listAliasReq, _ := http.NewRequest("GET", "/api/v1/tags/aliases", nil)
		listAliasResp := httptest.NewRecorder()
		r.ServeHTTP(listAliasResp, listAliasReq)
		if listAliasResp.Code != http.StatusOK {
			t.Fatalf("list alias failed, code=%d", listAliasResp.Code)
		}

		deleteAliasReq, _ := http.NewRequest("DELETE", "/api/v1/tags/aliases/SRE", nil)
		deleteAliasResp := httptest.NewRecorder()
		r.ServeHTTP(deleteAliasResp, deleteAliasReq)
		if deleteAliasResp.Code != http.StatusNoContent {
			t.Fatalf("delete alias failed, code=%d", deleteAliasResp.Code)
		}

		auditReq, _ := http.NewRequest("GET", "/api/v1/tags/audits?limit=5", nil)
		auditResp := httptest.NewRecorder()
		r.ServeHTTP(auditResp, auditReq)
		if auditResp.Code != http.StatusOK {
			t.Fatalf("list tag audits failed, code=%d body=%s", auditResp.Code, auditResp.Body.String())
		}
		var audits []dto.TagAuditResponse
		if err := json.Unmarshal(auditResp.Body.Bytes(), &audits); err != nil {
			t.Fatalf("unmarshal audits failed: %v", err)
		}
		if len(audits) == 0 {
			t.Fatal("expected at least one tag governance audit record")
		}

		filterAuditReq, _ := http.NewRequest(
			"GET",
			"/api/v1/tags/audits?limit=10&action=alias_delete",
			nil,
		)
		filterAuditResp := httptest.NewRecorder()
		r.ServeHTTP(filterAuditResp, filterAuditReq)
		if filterAuditResp.Code != http.StatusOK {
			t.Fatalf(
				"filtered tag audits failed, code=%d body=%s",
				filterAuditResp.Code,
				filterAuditResp.Body.String(),
			)
		}
		var filtered []dto.TagAuditResponse
		if err := json.Unmarshal(filterAuditResp.Body.Bytes(), &filtered); err != nil {
			t.Fatalf("unmarshal filtered audits failed: %v", err)
		}
		if len(filtered) == 0 {
			t.Fatal("expected filtered alias_delete audits")
		}
		for _, item := range filtered {
			if item.Action != "alias_delete" {
				t.Fatalf("expected alias_delete action, got %s", item.Action)
			}
		}
	})

	t.Run("Filename search should work", func(t *testing.T) {
		// 1. 先上传/模拟一个资源记录
		ctx := context.Background()
		res := &entity.Resource{
			ID:           "res-filename-search",
			FileName:     "会议纪要2026.docx",
			Hash:         "hash-filename",
			Size:         1,
			MimeType:     "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
			InternalPath: "path/to/res",
		}
		if err := resRepo.Save(ctx, res, 1); err != nil {
			t.Fatalf("seed resource failed: %v", err)
		}

		// 2. 创建一个关联该资源的笔记
		createBody := `{"content": "请查看附件", "resource_ids": ["res-filename-search"]}`
		req, _ := http.NewRequest("POST", "/api/v1/memos", bytes.NewBufferString(createBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)
		if w.Code != http.StatusCreated {
			t.Fatalf("create memo failed: %d", w.Code)
		}

		// 3. 搜索文件名中的关键词
		reqSearch, _ := http.NewRequest("GET", "/api/v1/memos?search=会议纪要", nil)
		wSearch := httptest.NewRecorder()
		r.ServeHTTP(wSearch, reqSearch)
		if wSearch.Code != http.StatusOK {
			t.Fatalf("search failed: %d", wSearch.Code)
		}

		var resList []*dto.MemoResponse
		if err := json.Unmarshal(wSearch.Body.Bytes(), &resList); err != nil {
			t.Fatalf("unmarshal failed: %v", err)
		}
		if len(resList) == 0 {
			t.Fatal("Filename search failed to return the associated memo")
		}
	})

	t.Run("Multi-tag DSL search should work", func(t *testing.T) {
		createBody := `{"content": "多标签测试\n#Tech #Draft"}`
		req, _ := http.NewRequest("POST", "/api/v1/memos", bytes.NewBufferString(createBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		// 搜索同时包含两个标签
		reqSearch, _ := http.NewRequest("GET", "/api/v1/memos?search=tag:Tech tag:Draft", nil)
		wSearch := httptest.NewRecorder()
		r.ServeHTTP(wSearch, reqSearch)

		var resList []*dto.MemoResponse
		errUnmarshal := json.Unmarshal(wSearch.Body.Bytes(), &resList)
		if errUnmarshal != nil {
			t.Fatalf("unmarshal failed: %v", errUnmarshal)
		}
		if len(resList) == 0 {
			t.Fatal("Multi-tag search failed")
		}
	})
}
