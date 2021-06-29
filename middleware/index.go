package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"time"
)

type M struct {
	L *logrus.Logger
}

func (m *M) Access(c *gin.Context) {
	now := time.Now().Unix()
	m.L.Infof("access time = %d", now)
	c.Next()
	end := time.Now().Unix()
	m.L.Infof("req time consuming = %ds", end-now)
}
