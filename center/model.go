package center

type Order struct {
	// 订单内容
	Content string `json:"content"`
	// 备注
	Comment string `json:"comment"`
	// 下单用户ID
	UserID string `json:"user_id"`
	// 配送外卖员ID
	TakeOutUserID string `json:"take_out_user_id"`
	// 被下单商铺ID
	StoreID string `json:"store_id"`
	// 1 未支付 2 已支付
	State int `json:"state"`
}
