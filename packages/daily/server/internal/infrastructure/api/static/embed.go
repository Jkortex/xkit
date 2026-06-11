package static

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

//go:embed all:dist
var distFS embed.FS

var subFS fs.FS

func init() {
	var err error
	subFS, err = fs.Sub(distFS, "dist")
	if err != nil {
		panic(err)
	}
}

func GetGinHandler() gin.HandlerFunc {
	indexData, err := fs.ReadFile(subFS, "index.html")
	if err != nil {
		panic(err)
	}

	fileServer := http.FileServer(http.FS(subFS))
	return func(c *gin.Context) {
		p := strings.TrimPrefix(c.Request.URL.Path, "/")
		if p == "" {
			c.Data(http.StatusOK, "text/html; charset=utf-8", indexData)
			return
		}
		if _, err := subFS.Open(p); err != nil {
			c.Data(http.StatusOK, "text/html; charset=utf-8", indexData)
			return
		}
		fileServer.ServeHTTP(c.Writer, c.Request)
	}
}
