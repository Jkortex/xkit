package dto

type UserResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
	Status   string `json:"status"`
}
