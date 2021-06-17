package ticker

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
	"math/rand"
	"order_center/center"
	"order_center/model"
	"time"
)

// 若 mq回调失败，进行重试并记录重试次数，超过一定重试次数，通知管理员人工介入

func TimeTicker(db *gorm.DB, rbConnPools []*amqp.Connection, l *logrus.Logger, interval int) {
	tChannel := time.After(time.Duration(interval) * time.Minute)
	for {
		select {
		case <-tChannel:
			// 需要重试的message
			var sMessages []*model.Message
			err := db.Model(&model.Message{}).Where("state", 1).Where(fmt.Sprintf("retry_count<=%d", 5)).Find(&sMessages).Error
			if err != nil {
				l.Error(err)
				continue
			}
			// 超过5次重试仍然失败，需要通知管理员的message
			go func() {
				var nMessages []*model.Message
				err = db.Model(&model.Message{}).Where("state", 1).Where(fmt.Sprintf("retry_count<=%d", 5)).Find(&nMessages).Error
				if err != nil {
					l.Error(err)
					return
				}
				// todo notify admin
			}()
			rbConn := rbConnPools[rand.Intn(len(rbConnPools)-1)]
			o := center.NewOrderCenter(db, l, rbConnPools)
			//ctx := context.Background()
			ctx := &gin.Context{}
			for _, message := range sMessages {
				var order *model.Order
				err = db.Model(&model.Order{}).Where("id", message.OrderID).Find(&order).Error
				if err != nil {
					l.Error(err)
					continue
				}
				message.RetryCount += 1
				err = db.Model(&model.Message{}).Updates(&model.Message{Updated: int(time.Now().Unix()), RetryCount: message.RetryCount}).Where("id", message.ID).Error
				if err != nil {
					l.Error(err)
					continue
				}
				err = o.SendMsg(rbConn, ctx, order, message)
				if err != nil {
					l.Error(err)
					continue
				}
			}
		}
	}
}
