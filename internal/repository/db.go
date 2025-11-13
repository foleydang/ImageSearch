package repository

import (
	"time"

	"github.com/bytedance/ImageSearch/internal/model"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Database 数据库连接管理器
type Database struct {
	DB *gorm.DB
}

// NewDatabase 创建数据库连接
func NewDatabase(dsn string) (*Database, error) {
	logrus.Info("正在连接数据库...")
	// 连接数据库
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		logrus.Errorf("连接数据库失败: %v", err)
		return nil, err
	}

	// 获取底层SQL DB
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// 设置连接池参数
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)

	logrus.Info("数据库连接成功")
	return &Database{DB: db}, nil
}

// AutoMigrate 自动迁移数据库表结构
func (d *Database) AutoMigrate() error {
	logrus.Info("正在自动迁移数据库表结构...")
	err := d.DB.AutoMigrate(
		&model.Image{},
		&model.ImageEmbedding{},
	)
	if err != nil {
		logrus.Errorf("自动迁移数据库表结构失败: %v", err)
		return err
	}

	logrus.Info("数据库表结构迁移成功")
	return nil
}

// Close 关闭数据库连接
func (d *Database) Close() error {
	sqlDB, err := d.DB.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}