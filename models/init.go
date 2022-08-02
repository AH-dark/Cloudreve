package model

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"github.com/cloudreve/Cloudreve/v3/pkg/conf"
	"github.com/cloudreve/Cloudreve/v3/pkg/util"
	"github.com/gin-gonic/gin"
)

// DB 数据库链接单例
var DB *gorm.DB

// Init 初始化 MySQL 链接
func Init() {
	util.Log().Info("初始化数据库连接")

	// Debug模式下，输出所有 SQL 日志
	logLevel := logger.Silent
	if conf.SystemConfig.Debug {
		logLevel = logger.Info
	}

	gormConfig := &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   conf.DatabaseConfig.TablePrefix, // 处理表前缀
			SingularTable: true,                            // 单一数据库
		},
		Logger: logger.New(log.New(os.Stdout, "\r\n", log.LstdFlags), logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logLevel,
			IgnoreRecordNotFoundError: false,
			Colorful:                  true,
		}),
	}

	var dialector gorm.Dialector

	if gin.Mode() == gin.TestMode {
		// 测试模式下，使用内存数据库
		dialector = sqlite.Open("file::memory:?cache=shared")
	} else {
		switch conf.DatabaseConfig.Type {
		case "UNSET", "sqlite", "sqlite3":
			// 未指定数据库或者明确指定为 sqlite 时，使用 SQLite3 数据库
			dialector = sqlite.Open(util.RelativePath(conf.DatabaseConfig.DBFile))
		case "postgres":
			dialector = postgres.Open(fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
				conf.DatabaseConfig.Host,
				conf.DatabaseConfig.User,
				conf.DatabaseConfig.Password,
				conf.DatabaseConfig.Name,
				conf.DatabaseConfig.Port))
		case "mysql":
			dialector = mysql.Open(fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
				conf.DatabaseConfig.User,
				conf.DatabaseConfig.Password,
				conf.DatabaseConfig.Host,
				conf.DatabaseConfig.Port,
				conf.DatabaseConfig.Name,
				conf.DatabaseConfig.Charset))
		case "mssql", "sqlserver":
			dialector = sqlserver.Open(fmt.Sprintf("%s:%s@(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
				conf.DatabaseConfig.User,
				conf.DatabaseConfig.Password,
				conf.DatabaseConfig.Host,
				conf.DatabaseConfig.Port,
				conf.DatabaseConfig.Name,
				conf.DatabaseConfig.Charset))
		default:
			util.Log().Panic("不支持数据库类型: %s", conf.DatabaseConfig.Type)
		}
	}

	var err error
	DB, err = gorm.Open(dialector, gormConfig)

	//db.SetLogger(util.Log())
	if err != nil {
		util.Log().Panic("连接数据库不成功, %s", err)
	}

	instance, err := DB.DB()
	if err != nil {
		util.Log().Panic("获取数据库实例不成功, %s", err)
	}

	//设置连接池
	instance.SetMaxIdleConns(50)
	if conf.DatabaseConfig.Type == "sqlite" || conf.DatabaseConfig.Type == "sqlite3" || conf.DatabaseConfig.Type == "UNSET" {
		instance.SetMaxOpenConns(1)
	} else {
		instance.SetMaxOpenConns(100)
	}

	//超时
	instance.SetConnMaxLifetime(time.Second * 30)

	//执行迁移
	migration()
}
