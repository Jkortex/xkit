package presenter

import "daily/internal/application/dto"

// IMemoPresenter 定义了业务结果的呈现契约
type IMemoPresenter interface {
	PresentMemo(data *dto.MemoResponse) interface{}
	PresentMemos(data []*dto.MemoResponse) interface{}
	PresentStats(data *dto.StatsResponse) interface{}
	PresentTags(data []dto.TagStatResponse) interface{}
	PresentTagSetGroup(data *dto.TagSetGroupResponse) interface{}
	PresentTagSetGroups(data []*dto.TagSetGroupResponse) interface{}
	PresentTagSet(data *dto.TagSetResponse) interface{}
	PresentTagSets(data []*dto.TagSetResponse) interface{}
	PresentBatchResult(data *dto.BatchResult) interface{}
	PresentError(err error) interface{}
}

// IResourcePresenter 定义了资源的呈现契约
type IResourcePresenter interface {
	PresentResource(data *dto.ResourceResponse) interface{}
	PresentError(err error) interface{}
}
