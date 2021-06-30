package ticker

import (
	"context"
	"fmt"
	"github.com/hzlpypy/order_center/center"
	"github.com/hzlpypy/order_center/model"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"time"
)

// 若 mq回调失败，进行重试并记录重试次数，超过一定重试次数，通知管理员人工介入

func TimeTicker(db *gorm.DB, rbConnPools []*amqp.Connection, l *logrus.Logger, interval int, ctx context.Context) {
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	for {
		select {
		case <-ticker.C:
			// 需要重试的message
			log.Print("ticker      start")
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
			rbConn := rbConnPools[0]
			if len(rbConnPools) >= 2 {
				rbConn = rbConnPools[rand.Intn(len(rbConnPools)-1)]
			}
			o := center.NewOrderCenter(db, l, rbConnPools, nil)
			ctx := context.Background()
			for _, message := range sMessages {
				var order *model.Order
				err = db.Model(&model.Order{}).Where("id", message.OrderID).Find(&order).Error
				if err != nil {
					l.Error(err)
					continue
				}
				message.RetryCount += 1
				err = db.Model(&model.Message{}).Where("id", message.ID).Updates(&model.Message{Updated: int(time.Now().Unix()), RetryCount: message.RetryCount}).Error
				if err != nil {
					l.Error(err)
					continue
				}
				_, err := rbConn.Channel()
				err = o.SendMsg(rbConn, ctx, order, message)
				if err != nil {
					l.Error(err)
					continue
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
