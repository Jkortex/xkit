package dto

// DailyStat 每日记录统计
type DailyStat struct {
	Date  string `json:"date"` // YYYY-MM-DD
	Count int    `json:"count"`
}

// StatsResponse 资产概览响应
type StatsResponse struct {
	MemosTotal     int64       `json:"memos_total"`
	TagsTotal      int64       `json:"tags_total"`
	ResourcesTotal int64       `json:"resources_total"`
	Heatmap        []DailyStat `json:"heatmap"`
}
