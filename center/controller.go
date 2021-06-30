package center

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
)

func (o *OrderCenter) CreateOrderController(c *gin.Context) {
	// todo：add validate
	reqBody := c.Request.Body
	bytes, _ := ioutil.ReadAll(reqBody)
	req := &OrderReq{}
	err := json.Unmarshal(bytes, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, "bad request")
		return
	}
	c.Set("user_id", req.UserID)
	err = o.CreateOrder(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, `{"code":200,"msg":"ok"}`)
	return
}

func (o *OrderCenter) ListOrderController(c *gin.Context) {
	reqBody := c.Request.Body
	bytes, _ := ioutil.ReadAll(reqBody)
	req := &OrderListReq{}
	err := json.Unmarshal(bytes, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, "bad request")
		return
	}
	orderList, err := o.ListOrder(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, orderList)
	return
}
