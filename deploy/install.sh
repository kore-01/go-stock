#!/bin/bash
# Go-Stock MCP Server 一键安装脚本
# 支持 Debian/Ubuntu/CentOS/RHEL

set -e

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 配置
APP_NAME="go-stock-mcp"
APP_DIR="/opt/go-stock-mcp"
GITHUB_REPO="kore-01/go-stock-mcp"
GO_VERSION="1.24.0"
PORT="${PORT:-8080}"

# 打印带颜色的信息
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

# 检查是否为 root 用户
check_root() {
    if [[ $EUID -ne 0 ]]; then
        print_error "请使用 root 用户运行此脚本"
        exit 1
    fi
}

# 检测系统类型
detect_os() {
    if [[ -f /etc/os-release ]]; then
        . /etc/os-release
        OS=$NAME
        VERSION=$VERSION_ID
    else
        print_error "无法检测操作系统类型"
        exit 1
    fi
    print_info "检测到操作系统: $OS $VERSION"
}

# 安装基础依赖
install_dependencies() {
    print_info "正在安装基础依赖..."

    if [[ "$OS" == *"Ubuntu"* ]] || [[ "$OS" == *"Debian"* ]]; then
        apt-get update -qq
        apt-get install -y -qq git curl wget
    elif [[ "$OS" == *"CentOS"* ]] || [[ "$OS" == *"Red Hat"* ]]; then
        yum install -y -q git curl wget
    else
        print_warning "未知的操作系统，尝试使用 apt-get..."
        apt-get update -qq
        apt-get install -y -qq git curl wget
    fi

    print_success "基础依赖安装完成"
}

# 安装 Go
install_go() {
    if command -v go &> /dev/null && go version | grep -q "go1.24"; then
        print_info "Go 1.24 已安装，跳过安装步骤"
        return
    fi

    print_info "正在安装 Go $GO_VERSION..."

    # 下载并安装 Go
    cd /tmp
    wget -q "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz"
    rm -rf /usr/local/go
    tar -C /usr/local -xzf "go${GO_VERSION}.linux-amd64.tar.gz"

    # 配置环境变量
    if ! grep -q "/usr/local/go/bin" /etc/profile; then
        echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
    fi

    export PATH=$PATH:/usr/local/go/bin

    # 验证安装
    if go version | grep -q "go$GO_VERSION"; then
        print_success "Go $GO_VERSION 安装成功"
    else
        print_error "Go 安装失败"
        exit 1
    fi
}

# 下载源码
download_source() {
    print_info "正在下载 Go-Stock MCP Server..."

    if [[ -d "$APP_DIR" ]]; then
        print_warning "目标目录已存在，备份旧版本..."
        mv "$APP_DIR" "${APP_DIR}.backup.$(date +%Y%m%d%H%M%S)"
    fi

    git clone -q "https://github.com/${GITHUB_REPO}.git" "$APP_DIR"
    cd "$APP_DIR"

    print_success "源码下载完成"
}

# 编译应用
build_app() {
    print_info "正在编译 Go-Stock MCP Server..."

    cd "$APP_DIR"

    # 设置 Go 环境
    export PATH=$PATH:/usr/local/go/bin
    export GOPROXY=https://proxy.golang.com.cn,direct

    # 下载依赖
    go mod download

    # 编译
    CGO_ENABLED=0 go build -ldflags="-s -w" -o "$APP_NAME" main.go sse_server.go

    if [[ -f "$APP_NAME" ]]; then
        print_success "编译成功: $APP_NAME"
    else
        print_error "编译失败"
        exit 1
    fi
}

# 创建 systemd 服务
create_service() {
    print_info "正在创建 systemd 服务..."

    cat > /etc/systemd/system/${APP_NAME}.service <<EOF
[Unit]
Description=Go-Stock MCP Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=${APP_DIR}
ExecStart=${APP_DIR}/${APP_NAME} -mode=sse
Restart=always
RestartSec=5
Environment="MCP_MODE=sse"
Environment="PORT=${PORT}"

[Install]
WantedBy=multi-user.target
EOF

    # 重载 systemd
    systemctl daemon-reload
    systemctl enable "$APP_NAME"

    print_success "systemd 服务创建成功"
}

# 配置防火墙
configure_firewall() {
    print_info "正在配置防火墙..."

    if command -v ufw &> /dev/null; then
        ufw allow "${PORT}/tcp" > /dev/null 2>&1 || true
        print_success "UFW 防火墙规则已添加"
    elif command -v firewall-cmd &> /dev/null; then
        firewall-cmd --permanent --add-port="${PORT}/tcp" > /dev/null 2>&1 || true
        firewall-cmd --reload > /dev/null 2>&1 || true
        print_success "Firewalld 防火墙规则已添加"
    elif command -v iptables &> /dev/null; then
        iptables -I INPUT -p tcp --dport "$PORT" -j ACCEPT
        print_success "iptables 规则已添加"
    else
        print_warning "未检测到防火墙工具，请手动开放端口 $PORT"
    fi
}

# 启动服务
start_service() {
    print_info "正在启动服务..."

    systemctl restart "$APP_NAME"
    sleep 2

    if systemctl is-active --quiet "$APP_NAME"; then
        print_success "服务启动成功"
    else
        print_error "服务启动失败，请检查日志: journalctl -u $APP_NAME"
        exit 1
    fi
}

# 打印完成信息
print_completion() {
    local ip
    ip=$(hostname -I | awk '{print $1}')

    echo ""
    echo "=========================================="
    echo -e "${GREEN}Go-Stock MCP Server 安装完成！${NC}"
    echo "=========================================="
    echo ""
    echo "📍 服务信息:"
    echo "   • 安装目录: $APP_DIR"
    echo "   • 服务名称: $APP_NAME"
    echo "   • 监听端口: $PORT"
    echo ""
    echo "🔗 访问地址:"
    echo "   • SSE 端点: http://${ip}:${PORT}/sse"
    echo "   • 健康检查: http://${ip}:${PORT}/health"
    echo ""
    echo "📋 常用命令:"
    echo "   • 查看状态: systemctl status $APP_NAME"
    echo "   • 查看日志: journalctl -u $APP_NAME -f"
    echo "   • 重启服务: systemctl restart $APP_NAME"
    echo "   • 停止服务: systemctl stop $APP_NAME"
    echo ""
    echo "=========================================="
}

# 主函数
main() {
    echo "=========================================="
    echo "Go-Stock MCP Server 一键安装脚本"
    echo "=========================================="
    echo ""

    check_root
    detect_os
    install_dependencies
    install_go
    download_source
    build_app
    create_service
    configure_firewall
    start_service
    print_completion
}

# 运行主函数
main "$@"
