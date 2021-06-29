package center

type OrderReq struct {
	// 订单内容
	Content string `json:"content"`
	// 备注
	Comment string `json:"comment"`
	// 下单用户ID
	UserID string `json:"user_id"`
	// 被下单商铺ID
	StoreID string `json:"store_id"`
	// 1 未支付 2 已支付
	State int `json:"state"`
}

type OrderListReq struct {
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	SearchInfo string `json:"search_info"` // 模糊匹配的搜索信息
}

type Order struct {
	// 订单编号
	ID string `json:"id"`
	// UnixTime
	// 创建时间/下单时间
	Created int `json:"created"`
	Updated int `json:"updated"`
	Deleted int `json:"deleted"`
	// 下单用户ID
	UserID string `json:"user_id"`
	// 配送外卖员ID
	TakeOutUserID string `json:"take_out_user_id"`
	// 配送外卖员名称
	TakeOutUsername string `json:"take_out_username"`
	// 被下单商铺ID
	StoreID string `json:"store_id"`
	// 被下单商铺名称
	StoreName string `json:"store_name"`

	// 订单内容
	Content string `json:"content"`
	// 备注
	Comment string `json:"comment"`
	// 1 未支付 2 已支付 默认未支付，由对接的支付平台，支付成功后回调项目对应接口更新该状态
	State int `json:"state"`
	// 接单时间
	OrderReceiveTime int `json:"order_receive_time"`
	// 预计到达时间
	ExpectArriveTime int `json:"expect_arrive_time"`
	// 送达时间
	DeliveryTime int `json:"delivery_time"`
	// 是否被删除 1：否 2：是
	FakeDelete int `json:"fake_delete"`
}

type OrderList struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
	Total    int `json:"total"`
	Orders   []*Order
}
