#!/bin/bash

# 数据库迁移脚本
# 使用方法: 
# ./scripts/migrate.sh up      # 执行迁移
# ./scripts/migrate.sh down 002 # 回滚到版本002
# ./scripts/migrate.sh status  # 查看迁移状态

set -e

ACTION=${1:-up}
VERSION=${2:-""}

echo "开始执行数据库迁移操作: $ACTION"

case $ACTION in
    "up")
        echo "执行所有未应用的迁移..."
        go run cmd/migrate/main.go -action=up
        ;;
    "down")
        if [ -z "$VERSION" ]; then
            echo "错误: 回滚操作需要指定版本号"
            echo "使用方法: ./scripts/migrate.sh down 002"
            exit 1
        fi
        echo "回滚迁移到版本: $VERSION"
        go run cmd/migrate/main.go -action=down -version=$VERSION
        ;;
    "status")
        echo "检查迁移状态..."
        go run cmd/migrate/main.go -action=status
        ;;
    *)
        echo "未知操作: $ACTION"
        echo "支持的操作: up, down, status"
        exit 1
        ;;
esac

echo "迁移操作完成!"