package ticker

import (
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
	"math/rand"
	"order_center/center"
	"order_center/model"
	"time"
)

// 若 消费者回调失败，此时Ack或NAck
// choose1：Ack，订单信息表更新失败 state=1,仍为未发送，其实已经发送成功
//
//
func TimeTicker(db *gorm.DB, rbConnPools []*amqp.Connection, l *logrus.Logger, interval int) {
	tChannel := time.After(time.Duration(interval) * time.Minute)
	for {
		select {
		case <-tChannel:
			var messages []*model.Message
			err := db.Model(&model.Message{}).Where("state", 1).Find(messages).Error
			if err != nil {
				l.Error(err)
				continue
			}
			rbConn := rbConnPools[rand.Intn(len(rbConnPools)-1)]
			o := center.NewOrderCenter(db, l, rbConnPools)
			//ctx := context.Background()
			ctx := gin.Context{}
			o.SendMsg(rbConn, ctx)
		}
	}
}
