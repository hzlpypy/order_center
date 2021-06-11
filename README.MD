# 分布式事务测试
组件：
    1.order_center 订单系统(go,rabbitmq,mysql,gin,logrus)
    // 2.notify 通知系统(go,rabbitmq,grpc,logrus)
    2.waybill_center 运单系统(go,rabbitmq,mysql,grpc,logrus)
    3.nameservice 名字服务(etcd,go,logrus,grpc)：注册/检测第三方组件状态（例如mysql，rabbitmq），更新 etcd服务信息

