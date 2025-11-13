package config

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

// Config 应用程序配置
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Storage  StorageConfig
	Log      LogConfig
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port int
	Host string
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	DSN string
}

// StorageConfig 存储配置
type StorageConfig struct {
	ImageDir string
}

// LogConfig 日志配置
type LogConfig struct {
	Level string
}

// LoadConfig 加载配置
func LoadConfig() *Config {
	port, _ := strconv.Atoi(getEnv("SERVER_PORT", "8080"))

	return &Config{
		Server: ServerConfig{
			Port: port,
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
		},
		Database: DatabaseConfig{
			DSN: getEnv("DATABASE_DSN", "./imagesearch.db"),
		},
		Storage: StorageConfig{
			ImageDir: getEnv("STORAGE_IMAGE_DIR", "./assets/images"),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "info"),
		},
	}
}

// SetupLogger 设置日志
func SetupLogger(config *LogConfig) {
	level, err := logrus.ParseLevel(config.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})
}

// getEnv 获取环境变量，如果不存在则返回默认值
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}