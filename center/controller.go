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
	req := &Order{}
	err := json.Unmarshal(bytes, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, "bad request")
		return
	}
	err = o.CreateOrder(c, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, nil)
	return
}
