package center

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"strconv"
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
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", "1"))
	if err != nil {
		c.JSON(http.StatusBadRequest, err.Error())
		return
	}
	req := &OrderListReq{
		Page:       page,
		PageSize:   pageSize,
		SearchInfo: c.DefaultQuery("search_info", ""),
	}
	orderList, err := o.ListOrder(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, orderList)
	return
}
