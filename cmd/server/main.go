package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/bytedance/ImageSearch/internal/api"
	"github.com/bytedance/ImageSearch/internal/config"
	"github.com/bytedance/ImageSearch/internal/repository"
	"github.com/bytedance/ImageSearch/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	// 加载配置
	cfg := config.LoadConfig()

	// 设置日志
	config.SetupLogger(&cfg.Log)

	logrus.Info("正在启动 ImageSearch 服务...")

	// 连接数据库
	db, err := repository.NewDatabase(cfg.Database.DSN)
	if err != nil {
		logrus.Fatalf("连接数据库失败: %v", err)
	}

	// 自动迁移数据库表结构
	if err := db.AutoMigrate(); err != nil {
		logrus.Fatalf("自动迁移数据库表结构失败: %v", err)
	}

	// 初始化仓库
	imageRepo := repository.NewImageRepository(db)

	// 初始化服务
	imageService := service.NewImageService(imageRepo, cfg.Storage.ImageDir)

	// 初始化API处理器
	handler := api.NewHandler(imageService, cfg.Storage.ImageDir)

	// 设置Gin模式
	if cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建Gin路由
	router := gin.Default()

	// 注册路由
	handler.RegisterRoutes(router)

	// 启动服务器
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	logrus.Infof("服务器正在启动，监听地址: %s", serverAddr)

	// 在goroutine中启动服务器
	go func() {
		if err := router.Run(serverAddr); err != nil {
			logrus.Fatalf("启动服务器失败: %v", err)
		}
	}()

	// 等待中断信号优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logrus.Info("正在关闭服务器...")

	// 关闭数据库连接
	sqlDB, err := db.DB.DB()
	if err != nil {
		logrus.Errorf("获取数据库连接失败: %v", err)
	} else {
		if err := sqlDB.Close(); err != nil {
			logrus.Errorf("关闭数据库连接失败: %v", err)
		}
	}

	logrus.Info("服务器已关闭")
}