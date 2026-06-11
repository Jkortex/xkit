package tag_set

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"daily/internal/application/apperr"
	"daily/internal/application/dto"
	"daily/internal/application/port"
	"daily/internal/domain/entity"
	"github.com/google/uuid"
)

type Service struct {
	groupRepo port.TagSetGroupRepository
	setRepo   port.TagSetRepository
}

func NewService(groupRepo port.TagSetGroupRepository, setRepo port.TagSetRepository) *Service {
	return &Service{groupRepo: groupRepo, setRepo: setRepo}
}

// --- TagSetGroup ---

func (s *Service) ListGroups(ctx context.Context, userID int64) ([]*dto.TagSetGroupResponse, error) {
	groups, err := s.groupRepo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	res := make([]*dto.TagSetGroupResponse, 0, len(groups))
	for _, g := range groups {
		res = append(res, groupToResponse(g))
	}
	return res, nil
}

func (s *Service) CreateGroup(ctx context.Context, userID int64, req dto.CreateTagSetGroupRequest) (*dto.TagSetGroupResponse, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("%w: group name cannot be empty", apperr.ErrInvalidInput)
	}
	now := time.Now().UTC()
	g := &entity.TagSetGroup{
		ID:        uuid.New().String(),
		UserID:    userID,
		Name:      req.Name,
		Weight:    req.Weight,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.groupRepo.Create(ctx, g); err != nil {
		return nil, err
	}
	return groupToResponse(g), nil
}

func (s *Service) UpdateGroup(ctx context.Context, userID int64, id string, req dto.UpdateTagSetGroupRequest) (*dto.TagSetGroupResponse, error) {
	g, err := s.groupRepo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		if *req.Name == "" {
			return nil, fmt.Errorf("%w: group name cannot be empty", apperr.ErrInvalidInput)
		}
		g.Name = *req.Name
	}
	if req.Weight != nil {
		g.Weight = *req.Weight
	}
	g.UpdatedAt = time.Now().UTC()
	if err := s.groupRepo.Update(ctx, g); err != nil {
		return nil, err
	}
	return groupToResponse(g), nil
}

func (s *Service) DeleteGroup(ctx context.Context, userID int64, id string) error {
	return s.groupRepo.Delete(ctx, userID, id)
}

// --- TagSet ---

func (s *Service) ListTagSets(ctx context.Context, userID int64, groupID *string) ([]*dto.TagSetResponse, error) {
	sets, err := s.setRepo.ListByUser(ctx, userID, groupID)
	if err != nil {
		return nil, err
	}
	res := make([]*dto.TagSetResponse, 0, len(sets))
	for _, ts := range sets {
		r, err := setToResponse(ts)
		if err != nil {
			return nil, err
		}
		res = append(res, r)
	}
	return res, nil
}

func (s *Service) CreateTagSet(ctx context.Context, userID int64, req dto.CreateTagSetRequest) (*dto.TagSetResponse, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("%w: tag set name cannot be empty", apperr.ErrInvalidInput)
	}
	tagsAny, err := json.Marshal(req.TagsAny)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid tags_any", apperr.ErrInvalidInput)
	}
	tagsAll, err := json.Marshal(req.TagsAll)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid tags_all", apperr.ErrInvalidInput)
	}
	tagsExclude, err := json.Marshal(req.TagsExclude)
	if err != nil {
		return nil, fmt.Errorf("%w: invalid tags_exclude", apperr.ErrInvalidInput)
	}
	now := time.Now().UTC()
	ts := &entity.TagSet{
		ID:          uuid.New().String(),
		UserID:      userID,
		GroupID:     req.GroupID,
		Name:        req.Name,
		TagsAny:     string(tagsAny),
		TagsAll:     string(tagsAll),
		TagsExclude: string(tagsExclude),
		Weight:      req.Weight,
		LastUsedAt:  &now,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := s.setRepo.Create(ctx, ts); err != nil {
		return nil, err
	}
	return setToResponse(ts)
}

func (s *Service) GetTagSet(ctx context.Context, userID int64, id string) (*dto.TagSetResponse, error) {
	ts, err := s.setRepo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	return setToResponse(ts)
}

func (s *Service) UpdateTagSet(ctx context.Context, userID int64, id string, req dto.UpdateTagSetRequest) (*dto.TagSetResponse, error) {
	ts, err := s.setRepo.GetByID(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	if req.Name != nil {
		if *req.Name == "" {
			return nil, fmt.Errorf("%w: tag set name cannot be empty", apperr.ErrInvalidInput)
		}
		ts.Name = *req.Name
	}
	if req.GroupID != nil {
		ts.GroupID = *req.GroupID
	}
	if req.TagsAny != nil {
		b, _ := json.Marshal(req.TagsAny)
		ts.TagsAny = string(b)
	}
	if req.TagsAll != nil {
		b, _ := json.Marshal(req.TagsAll)
		ts.TagsAll = string(b)
	}
	if req.TagsExclude != nil {
		b, _ := json.Marshal(req.TagsExclude)
		ts.TagsExclude = string(b)
	}
	if req.Weight != nil {
		ts.Weight = *req.Weight
	}
	ts.UpdatedAt = time.Now().UTC()
	if err := s.setRepo.Update(ctx, ts); err != nil {
		return nil, err
	}
	return setToResponse(ts)
}

func (s *Service) DeleteTagSet(ctx context.Context, userID int64, id string) error {
	return s.setRepo.Delete(ctx, userID, id)
}

func (s *Service) TouchTagSet(ctx context.Context, userID int64, id string) error {
	return s.setRepo.TouchLastUsed(ctx, id, userID)
}

var defaultTagSetData = []struct {
	GroupName string
	Sets      []struct {
		Name    string
		TagsAny []string
	}
}{
	{
		GroupName: "工作流",
		Sets: []struct {
			Name    string
			TagsAny []string
		}{
			{Name: "待办事项", TagsAny: []string{"todo", "wip"}},
			{Name: "已完成", TagsAny: []string{"done", "completed"}},
			{Name: "阻塞中", TagsAny: []string{"blocked"}},
		},
	},
	{
		GroupName: "阅读管理",
		Sets: []struct {
			Name    string
			TagsAny []string
		}{
			{Name: "待阅读", TagsAny: []string{"to-read"}},
			{Name: "正在阅读", TagsAny: []string{"reading"}},
			{Name: "已读完", TagsAny: []string{"read"}},
		},
	},
	{
		GroupName: "项目追踪",
		Sets: []struct {
			Name    string
			TagsAny []string
		}{
			{Name: "需求", TagsAny: []string{"feature", "requirement"}},
			{Name: "缺陷", TagsAny: []string{"bug"}},
			{Name: "优化", TagsAny: []string{"optimization", "refactor"}},
		},
	},
}

func (s *Service) BootstrapDefaults(ctx context.Context, userID int64) error {
	existing, err := s.groupRepo.ListByUser(ctx, userID)
	if err != nil {
		return fmt.Errorf("check existing tag sets: %w", err)
	}
	if len(existing) > 0 {
		return nil
	}

	now := time.Now().UTC()
	for _, g := range defaultTagSetData {
		group := &entity.TagSetGroup{
			ID:        uuid.New().String(),
			UserID:    userID,
			Name:      g.GroupName,
			Weight:    0,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := s.groupRepo.Create(ctx, group); err != nil {
			return fmt.Errorf("create default group %q: %w", g.GroupName, err)
		}
		for _, set := range g.Sets {
			tagsAny, _ := json.Marshal(set.TagsAny)
			tagsAll, _ := json.Marshal([]string{})
			tagsExclude, _ := json.Marshal([]string{})
			ts := &entity.TagSet{
				ID:          uuid.New().String(),
				UserID:      userID,
				GroupID:     &group.ID,
				Name:        set.Name,
				TagsAny:     string(tagsAny),
				TagsAll:     string(tagsAll),
				TagsExclude: string(tagsExclude),
				Weight:      0,
				LastUsedAt:  &now,
				CreatedAt:   now,
				UpdatedAt:   now,
			}
			if err := s.setRepo.Create(ctx, ts); err != nil {
				return fmt.Errorf("create default set %q: %w", set.Name, err)
			}
		}
	}
	return nil
}

// --- helpers ---

func groupToResponse(g *entity.TagSetGroup) *dto.TagSetGroupResponse {
	return &dto.TagSetGroupResponse{
		ID:        g.ID,
		Name:      g.Name,
		Weight:    g.Weight,
		CreatedAt: g.CreatedAt,
		UpdatedAt: g.UpdatedAt,
	}
}

func setToResponse(ts *entity.TagSet) (*dto.TagSetResponse, error) {
	var tagsAny, tagsAll, tagsExclude []string
	if err := json.Unmarshal([]byte(ts.TagsAny), &tagsAny); err != nil {
		return nil, fmt.Errorf("corrupt tags_any: %w", err)
	}
	if err := json.Unmarshal([]byte(ts.TagsAll), &tagsAll); err != nil {
		return nil, fmt.Errorf("corrupt tags_all: %w", err)
	}
	if err := json.Unmarshal([]byte(ts.TagsExclude), &tagsExclude); err != nil {
		return nil, fmt.Errorf("corrupt tags_exclude: %w", err)
	}
	return &dto.TagSetResponse{
		ID:          ts.ID,
		Name:        ts.Name,
		GroupID:     ts.GroupID,
		TagsAny:     tagsAny,
		TagsAll:     tagsAll,
		TagsExclude: tagsExclude,
		Weight:      ts.Weight,
		LastUsedAt:  ts.LastUsedAt,
		CreatedAt:   ts.CreatedAt,
		UpdatedAt:   ts.UpdatedAt,
	}, nil
}
