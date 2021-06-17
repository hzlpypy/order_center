package center

import (
	"encoding/json"
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
		//exchangeName := "order_exchange"
		//routingKey := "create_order"
		//contentType := "application/json"
		//queueName := "order_queue"
		//correlationId := utils.NewUUID()
		// 入库消息存入本地消息库 message
		messageID := utils.NewUUID()
		message := &model.Message{
			ID:         messageID,
			Created:    int(time.Now().Unix()),
			State:      1,
			Content:    req.Content,
			OrderID:    orderID,
			RetryCount: 1,
			//ExchangeName:  exchangeName,
			//RoutingKey:    routingKey,
			//ContentType:   contentType,
			//QueueName:     queueName,
			//CorrelationId: correlationId,
		}
		err = o.db.Model(model.Message{}).Create(message).Error
		if err != nil {
			o.l.WithField("CreateOrder", "create message error").Error(err)
			return err
		}
		// 入库信息发送mq
		err = o.SendMsg(rbConn, c, order, message)
		if err != nil {
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

func (o *OrderCenter) SendMsg(rbConn *amqp.Connection, c *gin.Context, order *model.Order, message *model.Message) error {
	ch, err := rbConn.Channel()
	if err != nil {
		o.l.WithField("sendMsg", "amqp channel error").Error(err)
		return err
	}
	exchangeName := "order_exchange"
	routingKey := "create_order"
	contentType := "application/json"
	queueName := "order_queue"
	orderMsg := &model.OrderMsg{
		Order: order,
		Msg:   message,
	}
	orderByte, _ := json.Marshal(orderMsg)
	ti, err := topic.NewTopicReq(c, &topic.TopicReq{
		Conn:         rbConn,
		Ch:           ch,
		ExchangeName: exchangeName,
		ExchangeType: "topic",
		Durable:      true,
		Msg:          string(orderByte),
		ContentType:  contentType,
		RoutingKey:   routingKey,
		Queue: &topic.Queue{
			QueueName:       queueName,
			QueueDeclareMap: map[string]interface{}{"x-max-length": 10},
		},
	})
	if err != nil {
		o.l.WithField("sendMsg", "amqp NewTopicReq error").Error(err)
		return err
	}
	err = ti.CreateExchange()
	if err != nil {
		o.l.WithField("sendMsg", "amqp CreateExchange error").Error(err)
		return err
	}
	err = ti.QueueDeclareAndBindRoutingKey()
	if err != nil {
		o.l.WithField("sendMsg", "amqp QueueDeclareAndBindRoutingKey error").Error(err)
		return err
	}
	//correlationId := utils.NewUUID()
	err = ch.Publish(exchangeName, routingKey, false, false, amqp.Publishing{
		ContentType: contentType,
		Body:        orderByte,
		//CorrelationId: correlationId,
		//ReplyTo:       queueName,
	})
	if err != nil {
		o.l.WithField("sendMsg", "amqp Publish error").Error(err)
		return err
	}
	// 获取回调信息
	confirmation := ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	go o.getRmpCallback(confirmation, orderMsg)
	return nil
}

func (o *OrderCenter) getRmpCallback(confirmation <-chan amqp.Confirmation, orderMsg *model.OrderMsg) {
	select {
	case c := <-confirmation:
		if c.Ack {
			err := o.db.Model(&model.Message{}).Updates(
				&model.Message{Updated: int(time.Now().Unix()), State: 2}).Where("id", orderMsg.Msg.ID).Error
			if err != nil {
				o.l.Errorf("GetCallback error, msgID=%s,err=%v", orderMsg.Msg.ID, err)
			}
			return
		}
	}
}
