package api

import (
	"fmt"
	"net/http"

	"daily/internal/infrastructure/api/handler"
	"daily/internal/infrastructure/config"
)

// NewServer 创建并配置 HTTP Server
func NewServer(cfg *config.Config, handlers *handler.Handlers) *http.Server {
	router := NewRouter(cfg, handlers)

	return &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: router,
	}
}
