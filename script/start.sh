#!/bin/bash

# 设置Go环境变量
export GO111MODULE=on

# 进入项目根目录
cd "$(dirname "$0")/.."

# 确保图片存储目录存在
mkdir -p ./assets/images

# 下载依赖
echo "正在下载依赖..."
go mod tidy

# 编译项目
echo "正在编译项目..."
go build -o ./bin/server ./cmd/server

# 检查编译是否成功
if [ $? -eq 0 ]; then
    echo "编译成功！"
    # 运行服务
    echo "正在启动服务..."
    ./bin/server
else
    echo "编译失败！"
    exit 1
fi