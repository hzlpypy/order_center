package center

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type OrderCenter struct {
	db *gorm.DB
	l  *logrus.Logger
}

func NewOrderCenter(db *gorm.DB, errorLog *logrus.Logger) *OrderCenter {
	return &OrderCenter{
		db: db,
		l:  errorLog,
	}
}

func (o *OrderCenter) CreateOrder(c *gin.Context, order *Order) error {
	// todo
	return nil
}
