package manager

import (
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"

	"servicemanager/pkg/global"
	pkglog "servicemanager/pkg/log"
)

func newServer(sm *serviceManager) (*gin.Engine, error) {
	r := gin.New()
	r.Use(gin.LoggerWithWriter(pkglog.GetLogWriter()))
	r.Use(gin.RecoveryWithWriter(pkglog.GetLogWriter()))

	r.LoadHTMLGlob(filepath.Join(global.StaticDir, "*.html"))

	r.GET("/", func(c *gin.Context) {
		ts := sm.getTaskList()
		c.HTML(http.StatusOK, "home.html", ts)
	})
	return r, nil
}
