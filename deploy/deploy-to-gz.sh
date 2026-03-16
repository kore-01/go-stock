#!/bin/bash
# 部署 Go-Stock MCP Server 到 GZ 服务器

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# 服务器配置
SERVER="root@gz"
REMOTE_DIR="/opt/go-stock-mcp"
PORT="28080"
APP_NAME="go-stock-mcp"

print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

# 显示菜单
show_menu() {
    echo "=========================================="
    echo "Go-Stock MCP Server 部署工具"
    echo "=========================================="
    echo ""
    echo "服务器: $SERVER"
    echo "远程目录: $REMOTE_DIR"
    echo "端口: $PORT"
    echo ""
    echo "请选择操作:"
    echo "  1) 完整部署 (上传代码 + 编译 + 启动)"
    echo "  2) 更新代码并重启"
    echo "  3) 仅重启服务"
    echo "  4) 查看服务状态"
    echo "  5) 查看日志"
    echo "  6) 停止服务"
    echo "  0) 退出"
    echo ""
    read -p "请输入选项 [0-6]: " choice
}

# 检查 SSH 连接
check_ssh() {
    print_info "检查 SSH 连接..."
    if ! ssh -o ConnectTimeout=5 "$SERVER" "echo 'OK'" > /dev/null 2>&1; then
        print_error "无法连接到 $SERVER"
        print_info "请确保:"
        echo "  1. SSH 密钥已配置 (~/.ssh/id_rsa)"
        echo "  2. 服务器可访问"
        echo "  3. ~/.ssh/config 中配置了 gz 主机名"
        exit 1
    fi
    print_success "SSH 连接正常"
}

# 上传代码
upload_code() {
    print_info "上传代码到服务器..."

    # 创建远程目录
    ssh "$SERVER" "mkdir -p $REMOTE_DIR"

    # 上传文件
    rsync -avz --exclude '.git' \
              --exclude '*.exe' \
              --exclude 'deploy/' \
              --exclude '*.md' \
              ./ "$SERVER:$REMOTE_DIR/"

    print_success "代码上传完成"
}

# 编译并启动
deploy() {
    print_info "在服务器上编译和部署..."

    ssh "$SERVER" << EOF
        cd $REMOTE_DIR

        # 检查 Go 安装
        if ! command -v go &> /dev/null; then
            echo "安装 Go..."
            wget -q https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
            tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
            export PATH=\$PATH:/usr/local/go/bin
            echo 'export PATH=\$PATH:/usr/local/go/bin' > /etc/profile.d/go.sh
        fi

        export PATH=\$PATH:/usr/local/go/bin
        export GOPROXY=https://proxy.golang.com.cn,direct

        # 编译
        echo "编译中..."
        go mod download
        CGO_ENABLED=0 go build -ldflags="-s -w" -o $APP_NAME main.go sse_server.go

        # 停止旧服务
        pkill -f "$APP_NAME" || true
        sleep 1

        # 启动新服务
        echo "启动服务..."
        nohup ./$APP_NAME -mode=sse > /dev/null 2>&1 &
        sleep 2

        # 检查是否运行
        if pgrep -f "$APP_NAME" > /dev/null; then
            echo "服务启动成功"
        else
            echo "服务启动失败"
            exit 1
        fi
EOF

    print_success "部署完成"
}

# 更新代码
update_code() {
    print_info "更新代码..."

    ssh "$SERVER" << EOF
        cd $REMOTE_DIR

        # 如果存在 .git 目录则 pull，否则重新克隆
        if [[ -d ".git" ]]; then
            git pull
        else
            rm -rf *
            git clone https://github.com/kore-01/go-stock-mcp.git /tmp/go-stock-mcp
            cp -r /tmp/go-stock-mcp/* .
            rm -rf /tmp/go-stock-mcp
        fi

        export PATH=\$PATH:/usr/local/go/bin

        # 重新编译
        go mod download
        CGO_ENABLED=0 go build -ldflags="-s -w" -o $APP_NAME main.go sse_server.go

        # 重启服务
        pkill -f "$APP_NAME" || true
        sleep 1
        nohup ./$APP_NAME -mode=sse > /dev/null 2>&1 &
        sleep 2

        if pgrep -f "$APP_NAME" > /dev/null; then
            echo "服务更新成功"
        else
            echo "服务更新失败"
            exit 1
        fi
EOF

    print_success "代码更新完成"
}

# 重启服务
restart_service() {
    print_info "重启服务..."

    ssh "$SERVER" << EOF
        cd $REMOTE_DIR
        pkill -f "$APP_NAME" || true
        sleep 1
        nohup ./$APP_NAME -mode=sse > /dev/null 2>&1 &
        sleep 2

        if pgrep -f "$APP_NAME" > /dev/null; then
            echo "服务重启成功"
        else
            echo "服务重启失败"
            exit 1
        fi
EOF

    print_success "服务重启完成"
}

# 查看状态
check_status() {
    print_info "查看服务状态..."

    ssh "$SERVER" << EOF
        cd $REMOTE_DIR

        echo "=== 进程状态 ==="
        ps aux | grep "$APP_NAME" | grep -v grep || echo "服务未运行"

        echo ""
        echo "=== 端口监听 ==="
        netstat -tlnp | grep "$PORT" || ss -tlnp | grep "$PORT" || echo "端口未监听"

        echo ""
        echo "=== 健康检查 ==="
        curl -s http://localhost:$PORT/health || echo "健康检查失败"
EOF
}

# 查看日志
view_logs() {
    print_info "查看日志..."

    ssh "$SERVER" "cd $REMOTE_DIR && tail -f nohup.out 2>/dev/null || echo '日志文件不存在'"
}

# 停止服务
stop_service() {
    print_info "停止服务..."

    ssh "$SERVER" "pkill -f \"$APP_NAME\" || true"

    print_success "服务已停止"
}

# 主函数
main() {
    # 检查是否在正确目录
    if [[ ! -f "main.go" ]]; then
        print_error "请在 go-stock-mcp 项目目录中运行此脚本"
        exit 1
    fi

    check_ssh

    while true; do
        show_menu

        case $choice in
            1)
                upload_code
                deploy
                print_completion
                ;;
            2)
                update_code
                print_completion
                ;;
            3)
                restart_service
                ;;
            4)
                check_status
                ;;
            5)
                view_logs
                ;;
            6)
                stop_service
                ;;
            0)
                echo "退出"
                exit 0
                ;;
            *)
                print_error "无效选项"
                ;;
        esac

        echo ""
        read -p "按回车键继续..."
    done
}

# 打印完成信息
print_completion() {
    echo ""
    echo "=========================================="
    echo -e "${GREEN}部署完成！${NC}"
    echo "=========================================="
    echo ""
    echo "🔗 访问地址:"
    echo "   • SSE 端点: http://10.1.20.3:$PORT/sse"
    echo "   • 健康检查: http://10.1.20.3:$PORT/health"
    echo ""
    echo "📋 常用命令:"
    echo "   ssh $SERVER 'ps aux | grep $APP_NAME'"
    echo "   ssh $SERVER 'curl http://localhost:$PORT/health'"
    echo ""
    echo "=========================================="
}

# 运行主函数
main "$@"
