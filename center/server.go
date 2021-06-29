package center

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hzlpypy/common/utils"
	"github.com/hzlpypy/order_center/model"
	protos "github.com/hzlpypy/waybill_center/proto_info/protos"
	"github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"time"
)

type OrderCenter struct {
	db          *gorm.DB
	l           *logrus.Logger
	rbConnPools []*amqp.Connection
	client      protos.WaybillCenterClient
}

func NewOrderCenter(db *gorm.DB, errorLog *logrus.Logger, rbConnPools []*amqp.Connection, client protos.WaybillCenterClient) *OrderCenter {
	return &OrderCenter{
		db:          db,
		l:           errorLog,
		rbConnPools: rbConnPools,
		client:      client,
	}
}

// CreateOrder: 创建订单
func (o *OrderCenter) CreateOrder(c *gin.Context, req *OrderReq) error {
	userID := c.Value("user_id").(string)
	orderID := utils.NewUUID()
	// 随机获取rabbitmq conn
	rbConn := o.rbConnPools[0]
	if len(o.rbConnPools) >= 2 {
		rbConn = o.rbConnPools[rand.Intn(len(o.rbConnPools)-1)]
	}
	err := o.db.Transaction(func(tx *gorm.DB) error {
		// 订单入库
		order := &model.Order{
			ID:         orderID,
			Created:    int(time.Now().Unix()),
			UserID:     userID,
			StoreID:    req.StoreID,
			Content:    req.Content,
			Comment:    req.Comment,
			State:      1,
			FakeDelete: 1,
		}
		err := o.db.Model(model.Order{}).Create(order).Error
		if err != nil {
			o.l.WithField("CreateOrder", "create order error").Error(err)
			return err
		}
		// 入库消息存入本地消息库 message
		messageID := utils.NewUUID()
		message := &model.Message{
			ID:         messageID,
			Created:    int(time.Now().Unix()),
			State:      1,
			Content:    req.Content,
			OrderID:    orderID,
			RetryCount: 1,
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

func (o *OrderCenter) SendMsg(rbConn *amqp.Connection, c context.Context, order *model.Order, message *model.Message) error {
	ch, err := rbConn.Channel()
	if err != nil {
		o.l.WithField("sendMsg", "amqp channel error").Error(err)
		return err
	}
	exchangeName := "order_exchange"
	routingKey := "login.order"
	contentType := "application/json"
	//queueName := "order_queue"
	orderMsg := &model.OrderMsg{
		Order: order,
		Msg:   message,
	}
	orderByte, _ := json.Marshal(orderMsg)
	err = ch.Confirm(false)
	if err != nil {
		log.Println("this.Channel.Confirm  ", err)
	}
	// send
	err = ch.Publish(exchangeName, routingKey, false, false, amqp.Publishing{
		ContentType:  contentType,
		Body:         orderByte,
		DeliveryMode: amqp.Persistent,
	})
	// 获取回调信息
	if err != nil {
		o.l.WithField("sendMsg", "amqp Publish error").Error(err)
		return err
	}
	confirm := make(chan amqp.Confirmation, 1)
	ch.NotifyPublish(confirm)
	go o.getRmpCallback(confirm, orderMsg, ch)
	return nil
}

func (o *OrderCenter) getRmpCallback(confirmation <-chan amqp.Confirmation, orderMsg *model.OrderMsg, ch *amqp.Channel) {
	for {
		select {
		case c := <-confirmation:
			if c.Ack {
				err := o.db.Model(&model.Message{}).Where("id", orderMsg.Msg.ID).Updates(
					&model.Message{Updated: int(time.Now().Unix()), State: 2}).Error
				if err != nil {
					o.l.Errorf("GetCallback error, msgID=%s,err=%v", orderMsg.Msg.ID, err)
				}
				log.Println("ack true")
				_ = ch.Close()
				return
			}
		case <-time.After(1 * time.Minute):
			o.l.Error("GetCallback time out")
			_ = ch.Close()
			//ch.Confirm()
			return
		}
	}

}

func (o *OrderCenter) ListOrder(c *gin.Context, req *OrderListReq) (*OrderList, error) {
	// get orders
	searchKeys := []string{"id", "store_name", "content"}
	db := o.db
	for _, searchKey := range searchKeys {
		db = o.db.Model(model.Order{}).Or(fmt.Sprintf("%s like ?%", searchKey), req.SearchInfo)
	}
	orders := []*model.Order{}
	err := db.Find(orders).Error
	if err != nil {
		o.l.Errorf("Find orders error,err=%s", err)
		return nil, err
	}
	orderIds := make([]string, len(orders))
	for i, order := range orders {
		orderIds[i] = order.ID
	}
	// 对接 waybill_center(运单中心) 实时获取订单状态
	wb, err := o.client.ListWaybill(c, &protos.ListWaybillReq{
		Page:     int32(req.Page),
		PageSize: int32(req.PageSize),
		OrderIds: orderIds,
	})
	if err != nil {
		o.l.Errorf("ListWaybill error,err=%s", err)
		return nil, err
	}
	// 数据整理
	var wrMap = make(map[string]*protos.Waybill)
	for _, wr := range wb.Waybills {
		wrMap[wr.Id] = wr
	}
	var resOrders = make([]*Order, len(orders))
	for i, order := range orders {
		wr := wrMap[order.ID]
		resOrders[i] = &Order{
			ID:               order.ID,
			Created:          order.Created,
			UserID:           order.UserID,
			TakeOutUserID:    wr.TakeOutUserId,
			TakeOutUsername:  wr.TakeOutUserName,
			StoreID:          order.StoreID,
			StoreName:        order.StoreName,
			Content:          order.Content,
			Comment:          order.Comment,
			State:            order.State,
			OrderReceiveTime: 0,
			ExpectArriveTime: 0,
			DeliveryTime:     int(wr.DeliveryTime),
			FakeDelete:       order.FakeDelete,
		}
	}
	return nil, nil
}
