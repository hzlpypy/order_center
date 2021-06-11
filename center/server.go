package center

import (
	"github.com/gin-gonic/gin"
	"github.com/hzlpypy/common/rabbitmq/topic"
	"github.com/hzlpypy/common/utils"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
	"math/rand"
	"order_center/model"
	"time"
)

type OrderCenter struct {
	db          *gorm.DB
	l           *logrus.Logger
	rbConnPools []*amqp.Connection
}

func NewOrderCenter(db *gorm.DB, errorLog *logrus.Logger, rbConnPools []*amqp.Connection) *OrderCenter {
	return &OrderCenter{
		db:          db,
		l:           errorLog,
		rbConnPools: rbConnPools,
	}
}

// CreateOrder: 创建订单
func (o *OrderCenter) CreateOrder(c *gin.Context, req *Order) error {
	userID := c.Value("user_id").(string)
	orderID := utils.NewUUID()
	// 随机获取rabbitmq conn
	rbConn := o.rbConnPools[rand.Intn(len(o.rbConnPools)-1)]
	err := o.db.Transaction(func(tx *gorm.DB) error {
		// 订单入库
		order := &model.Order{
			ID:      orderID,
			Created: int(time.Now().Unix()),
			UserID:  userID,
			StoreID: req.StoreID,
			Content: req.Content,
			Comment: req.Comment,
			State:   1,
			// 理论上需要通过商铺和买家距离计算得出到达时间，这里简化直接当前时间 + 30min
			ExpectArriveTime: int(time.Now().Unix() + int64(30*time.Minute)),
			FakeDelete:       1,
		}
		err := o.db.Model(model.Order{}).Create(order).Error
		if err != nil {
			o.l.WithField("CreateOrder", "create order error").Error(err)
			return err
		}
		// 入库消息存入本地消息库 message
		messageID := utils.NewUUID()
		message := &model.Message{
			ID:      messageID,
			Created: int(time.Now().Unix()),
			State:   1,
			Content: req.Content,
		}
		err = o.db.Model(model.Message{}).Create(message).Error
		if err != nil {
			o.l.WithField("CreateOrder", "create message error").Error(err)
			return err
		}
		// 入库信息发送mq todo 对接 notify
		ch, err := rbConn.Channel()
		if err != nil {
			o.l.WithField("Channel", "amqp channel error").Error(err)
			return err
		}

		return nil
	})
	if err != nil {
		o.l.WithField("CreateOrder", "create order error").Error(err)
		return err
	}
	return nil
}
