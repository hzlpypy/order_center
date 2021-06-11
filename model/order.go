package model

// 订单
type Order struct {
	// 订单编号
	ID string `gorm:"type:varchar(32);primary_key;uniqueIndex:index_uid_oid;uniqueIndex:index_touid_oid;"`
	// UnixTime
	// 创建时间/下单时间
	Created  int    `gorm:"type:int(10);not null"`
	Updated  int    `gorm:"type:int(10);default:0;autoUpdateTime;"`
	Deleted  int    `gorm:"type:int(10);default:0;autoCreateTime"`
	Handlers string `gorm:"type:varchar(32);default:''"`
	// 下单用户ID
	UserID string `gorm:"type:varchar(32);uniqueIndex:index_uid_oid;not null"`
	// 配送外卖员ID(todo:可以考虑放到 运单中心)
	TakeOutUserID string `gorm:"type:varchar(32);uniqueIndex:index_touid_oid;not null"`
	// 被下单商铺ID
	StoreID string `gorm:"type:varchar(32);not null"`
	// 订单内容
	Content string `gorm:"type:varchar(512);not null"`
	// 备注
	Comment string `gorm:"type:varchar(128);not null"`
	// 1 未支付 2 已支付 默认未支付，由对接的支付平台，支付成功后回调项目对应接口更新该状态
	State int `gorm:"type:tinyint(1);default:1"`
	// 接单时间 (todo:可以考虑放到 运单中心)
	OrderReceiveTime int `gorm:"type:int(10);not null"`
	// 预计到达时间
	ExpectArriveTime int `gorm:"type:int(10);not null"`
	// 送达时间 (todo:可以考虑放到 运单中心)
	DeliveryTime int `gorm:"type:int(10);not null"`
	// 是否被删除 1：否 2：是
	FakeDelete int `gorm:"type:tinyint(1);default:1"`
	// ...
}
