package dao

import (
	"fmt"
	"log"
	"time"

	"k8s.io/sample-controller/cfg"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func InitMysql(cfg *cfg.MysqlConfig) error {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.Username,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Database)
	dao, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger:                 logger.Default.LogMode(logger.Warn),
		PrepareStmt:            true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		log.Println(fmt.Sprintf("connect mysql error:%v", err))
		return err
	}
	sqlDB, _ := dao.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	err = dao.AutoMigrate(&Service{})
	if err != nil {
		log.Println(fmt.Sprintf("auto migrate error:%v", err))
		return err
	}
	DB = dao
	return nil
}
