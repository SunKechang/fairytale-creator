package database

import (
	"errors"
	"fairytale-creator/flag"
	"fairytale-creator/logger"
	"fmt"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	logger2 "gorm.io/gorm/logger"
)

var (
	gormDB          *gorm.DB
	InterError      = errors.New("服务器报错，请稍候再试")
	RequestError    = errors.New("请求参数有误")
	FileFormatError = errors.New("文件格式错误")
)

func Init() error {
	if gormDB != nil {
		return errors.New("connection already exists")
	}
	host := flag.MysqlHost
	port := flag.MysqlPort
	database := flag.MysqlDatabase
	username := flag.MysqlUsername
	password := flag.MysqlPassword
	charset := "utf8"
	db_host := os.Getenv("DB_HOST")
	if db_host != "" {
		host = db_host
	}
	address := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=%s&parseTime=true",
		username,
		password,
		host,
		port,
		database,
		charset,
	)
	var err error
	for i := 0; i < 40; i++ {
		gormDB, err = gorm.Open(mysql.Open(address), &gorm.Config{
			Logger: logger2.Default.LogMode(logger2.Info),
		})
		if err != nil {
			time.Sleep(1 * time.Second)
		} else {
			break
		}
	}
	if err != nil {
		logger.Error("连接数据库失败：", err.Error())
		return err
	}
	gormDB.AutoMigrate(&Story{}, &Chapter{})
	return nil
}

type BaseDao struct {
	Engine *gorm.DB
}

func GetDB() *gorm.DB {
	// return gormDB.Debug()
	return gormDB
}

func (p *BaseDao) GetDB() *gorm.DB {
	return p.Engine
}

func (p *BaseDao) Transaction(db *gorm.DB) {
	p.Engine = db
}
