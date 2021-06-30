package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hzlpypy/common/clog"
	hm "github.com/hzlpypy/common/databases/mysql"
	"github.com/hzlpypy/common/databases/mysql/v2"
	"github.com/hzlpypy/common/name_service"
	"github.com/hzlpypy/common/utils"
	"github.com/hzlpypy/grpc-lb/balancer"
	"github.com/hzlpypy/grpc-lb/common"
	"github.com/hzlpypy/grpc-lb/registry"
	etcd3 "github.com/hzlpypy/grpc-lb/registry/etcd3"
	"github.com/hzlpypy/order_center/center"
	"github.com/hzlpypy/order_center/cmd/config"
	"github.com/hzlpypy/order_center/init_service"
	"github.com/hzlpypy/order_center/internal/ticker"
	"github.com/hzlpypy/order_center/middleware"
	protos "github.com/hzlpypy/waybill_center/proto_info/protos"
	"github.com/ozonru/etcd/v3/clientv3"
	"github.com/streadway/amqp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

var nodeID = flag.String("node", "node1", "node ID")

type CallbackGather struct {
	OrderMsg      chan amqp.Delivery
	CorrelationId string
}

func main() {
	flag.Parse()
	cf := config.NewConfig()
	e := gin.Default()
	// local ip
	ip, _ := utils.ExternalIP()
	// register to etcd
	nameServiceMap, err := name_service.GetNameService()
	if err != nil {
		log.Fatal(err)
	}
	serviceName := nameServiceMap[cf.Server.Name]
	lbConfig := &etcd3.Config{
		EtcdConfig: clientv3.Config{
			//Endpoints:   ds.Config.Etcd.EndPoints,
			Endpoints:   []string{"http://127.0.0.1:2379"},
			DialTimeout: 5 * time.Second,
		},
		RegistryDir: cf.Server.Name,
		Ttl:         time.Duration(cf.Etcd.Ttl) * time.Second,
	}
	register, _ := etcd3.NewRegistrar(lbConfig)
	service := &registry.ServiceInfo{
		InstanceId: *nodeID,
		Name:       serviceName,
		Version:    "1.0",
		Address:    fmt.Sprintf("%s:%d", ip.String(), cf.Server.Port),
		Metadata:   metadata.Pairs(common.WeightKey, "1"),
	}
	err = register.Register(service)
	if err != nil {
		log.Fatal(err)
	}
	// conn waybill_center
	etcdConfig := clientv3.Config{
		Endpoints: []string{fmt.Sprintf("http://%s:%d", cf.Etcd.Ip, cf.Etcd.Port)},
	}
	etcd3.RegisterResolver("etcd3", etcdConfig, cf.WaybillCenter.Name, "test", "1.0")

	c, err := grpc.Dial("etcd3:///", grpc.WithInsecure(), grpc.WithBalancerName(balancer.RoundRobin))
	if err != nil {
		log.Printf("grpc dial: %s", err)
		return
	}
	defer c.Close()
	client := protos.NewWaybillCenterClient(c)

	// log
	logPathMap := map[string]string{"access": cf.Log.AccessPath, "error": cf.Log.ErrorPath}
	logCfg := &clog.Cfg{}
	for name, path := range logPathMap {
		file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
		if err != nil {
			file, _ = os.Create(path)
		}
		defer file.Close()
		logCfg.CfgFiles = append(logCfg.CfgFiles, &clog.CfgFile{
			Name: name,
			File: file,
		})
	}
	l, err := clog.Init(logCfg)
	if err != nil {
		log.Fatal(err)
	}
	accessLog := *l.Access()
	errorLog := *l.Error()

	// conn rabbitmq
	w := etcd3.NewWatcher(cf.Rabbitmq.Name, register.Etcd3Client)
	addressList := w.GetAllAddresses()
	// rabbitMq 连接池
	var rbConnPools []*amqp.Connection
	for _, address := range addressList {
		addr := address.Addr
		conn, err := amqp.Dial(fmt.Sprintf("amqp://admin:admin@%s", addr))
		if err != nil {
			errorLog.Errorf("Failed to connect to RabbitMQ, err=%v", err)
			continue
		}
		defer conn.Close()
		rbConnPools = append(rbConnPools, conn)
	}
	if len(rbConnPools) == 0 {
		log.Fatal("len(rbConnPools) == 0")
	}
	// register middleware
	m := &middleware.M{
		L: &accessLog,
	}
	e.Use(m.Access)
	// db
	mysqlCf := cf.Mysql
	var debug bool
	logLevel := logger.Error
	if cf.Model == "debug" {
		debug = true
		logLevel = logger.Info
	}
	db, err := v2.NewDbConnection(&hm.Config{
		Username:                                 mysqlCf.User,
		Password:                                 mysqlCf.Pwd,
		DBName:                                   mysqlCf.DBName,
		Host:                                     mysqlCf.Host,
		Port:                                     mysqlCf.Port,
		Charset:                                  mysqlCf.Charset,
		Debug:                                    debug,
		ConnMaxLifetime:                          time.Duration(mysqlCf.ConnMaxLifetime),
		MaxIdleConns:                             mysqlCf.MaxIdleConns,
		MaxOpenConns:                             mysqlCf.MaxOpenConns,
		LogLevel:                                 logLevel,
		DisableForeignKeyConstraintWhenMigrating: mysqlCf.DisableForeignKeyConstraintWhenMigrating,
	})
	if err != nil {
		log.Fatal(err)
	}
	// init service
	i, err := init_service.NewInit(db, &errorLog)
	if err != nil {
		log.Fatal(err)
	}
	i.InitVCTable()
	// add server
	orderCenter := center.NewOrderCenter(db, &errorLog, rbConnPools, client)
	orderCenter.RegisterOrderCenter(e)
	// add ticker
	ctx := context.Background()
	defer ctx.Done()
	go ticker.TimeTicker(db, rbConnPools, &errorLog, cf.Server.Ticker, ctx)
	// run
	_ = e.Run(fmt.Sprintf("%s:%d", ip.String(), cf.Server.Port))
}
