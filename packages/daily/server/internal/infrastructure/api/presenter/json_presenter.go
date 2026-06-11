package api_presenter

import (
	"daily/internal/application/apperr"
	"daily/internal/application/dto"
	"errors"
)

type JsonPresenter struct{}

func NewJsonPresenter() *JsonPresenter {
	return &JsonPresenter{}
}

func (p *JsonPresenter) PresentMemo(data *dto.MemoResponse) interface{} {
	return data
}

func (p *JsonPresenter) PresentMemos(data []*dto.MemoResponse) interface{} {
	return data
}

func (p *JsonPresenter) PresentStats(data *dto.StatsResponse) interface{} {
	return data
}

func (p *JsonPresenter) PresentTags(data []dto.TagStatResponse) interface{} {
	return data
}

func (p *JsonPresenter) PresentResource(data *dto.ResourceResponse) interface{} {
	return data
}

func (p *JsonPresenter) PresentTagSetGroup(data *dto.TagSetGroupResponse) interface{} {
	return data
}

func (p *JsonPresenter) PresentTagSetGroups(data []*dto.TagSetGroupResponse) interface{} {
	return data
}

func (p *JsonPresenter) PresentTagSet(data *dto.TagSetResponse) interface{} {
	return data
}

func (p *JsonPresenter) PresentTagSets(data []*dto.TagSetResponse) interface{} {
	return data
}

func (p *JsonPresenter) PresentBatchResult(data *dto.BatchResult) interface{} {
	return data
}

func (p *JsonPresenter) PresentError(err error) interface{} {
	return map[string]string{
		"error": err.Error(),
		"code":  errorCodeFromError(err),
	}
}

func errorCodeFromError(err error) string {
	switch {
	case errors.Is(err, apperr.ErrInvalidInput):
		return "INVALID_INPUT"
	case errors.Is(err, apperr.ErrUnauthorized):
		return "UNAUTHORIZED"
	case errors.Is(err, apperr.ErrForbidden):
		return "FORBIDDEN"
	case errors.Is(err, apperr.ErrNotFound):
		return "NOT_FOUND"
	case errors.Is(err, apperr.ErrConflict):
		return "CONFLICT"
	default:
		return "INTERNAL_ERROR"
	}
}
