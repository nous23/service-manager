package manager

import (
	"github.com/gin-gonic/gin"
	"net/http"
	pkglog "servicemanager/pkg/log"
)

func newServer(sm *serviceManager) (*gin.Engine, error) {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.LoggerWithWriter(pkglog.GetLogWriter()))
	r.Use(gin.RecoveryWithWriter(pkglog.GetLogWriter()))

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "service manager running")
	})
	return r, nil
}
