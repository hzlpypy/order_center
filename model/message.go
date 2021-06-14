package model

// 订单发送到 mq 的本地消息表
type Message struct {
	// 消息编号
	ID string `gorm:"type:varchar(32);primary_key;"`
	// UnixTime
	// 创建时间
	Created int `gorm:"type:int(10);not null"`
	Updated int `gorm:"type:int(10);default:0;autoUpdateTime;"`
	Deleted int `gorm:"type:int(10);default:0;autoCreateTime"`
	// 1 未发送 2 已发送（初始为1,通过MQ ACK 回执更新状态）
	State int `gorm:"type:tinyint(1);default:1"`
	// 消息内容
	Content string `gorm:"type:varchar(512);not null"`
}