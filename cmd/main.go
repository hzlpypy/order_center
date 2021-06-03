package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/hzlpypy/common/clog"
	hm "github.com/hzlpypy/common/databases/mysql"
	"github.com/hzlpypy/common/databases/mysql/v2"
	"gorm.io/gorm/logger"
	"log"
	"order_center/center"
	"order_center/cmd/config"
	"order_center/middleware"
	"os"
	"time"
)

func main() {
	cf := config.NewConfig()
	e := gin.Default()
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
	// middleware
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
	// add server
	orderCenter := center.NewOrderCenter(db, &errorLog)
	orderCenter.RegisterOrderCenter(e)

	// listen
	_ = e.Run(fmt.Sprintf("%s:%d", cf.Server.Ip, cf.Server.Port))
}
