package center

import "github.com/gin-gonic/gin"

func (o *OrderCenter) RegisterOrderCenter(e *gin.Engine) {
	// 发起订单
	e.POST("api/v1/order_center/create", o.CreateOrderController)
	e.GET("api/v1/order_center/list", o.ListOrderController)
}
