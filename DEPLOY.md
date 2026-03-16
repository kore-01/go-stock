# Go-Stock MCP Server 部署指南

本文档详细介绍如何在各种环境中部署 Go-Stock MCP Server。

## 📍 已部署服务

**GZ 内网服务器**：
- **IP**: `10.1.20.3`
- **端口**: `28080` (原 8080/18080 被占用)
- **SSE 端点**: `http://10.1.20.3:28080/sse`
- **健康检查**: `http://10.1.20.3:28080/health`
- **状态**: ✅ 运行中

---

## 目录

1. [快速部署](#快速部署)
2. [宝塔面板部署](#宝塔面板部署)
3. [Docker 部署](#docker-部署)
4. [手动部署](#手动部署)
5. [配置说明](#配置说明)
6. [故障排查](#故障排查)

---

## 快速部署

### 一键脚本安装（推荐）

```bash
# 下载并运行安装脚本
curl -fsSL https://raw.githubusercontent.com/kore-01/go-stock-mcp/main/deploy/install.sh | bash

# 或使用 wget
wget -qO- https://raw.githubusercontent.com/kore-01/go-stock-mcp/main/deploy/install.sh | bash
```

安装完成后，服务会自动启动，并配置为开机自启。

---

## 宝塔面板部署

### 方式一：Docker 部署（推荐）

#### 步骤 1: 安装 Docker

在宝塔面板中：
1. 进入 **软件商店**
2. 搜索 **Docker**
3. 安装 **Docker管理器**

#### 步骤 2: 创建容器

```bash
# SSH 连接到服务器
ssh root@your-server

# 拉取镜像
docker pull ghcr.io/kore-01/go-stock-mcp:latest

# 运行容器
docker run -d \
  --name go-stock-mcp \
  --restart unless-stopped \
  -p 8080:8080 \
  -e MCP_MODE=sse \
  -e PORT=8080 \
  ghcr.io/kore-01/go-stock-mcp:latest
```

#### 步骤 3: 配置防火墙

在宝塔面板中：
1. 进入 **安全**
2. 添加端口规则：**8080**
3. 备注：**Go-Stock MCP Server**

### 方式二：通过宝塔 Supervisor 管理

#### 步骤 1: 安装依赖

```bash
# 安装 Go（如果未安装）
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
source /etc/profile
```

#### 步骤 2: 下载并编译

```bash
cd /www
git clone https://github.com/kore-01/go-stock-mcp.git
cd go-stock-mcp
go build -ldflags="-s -w" -o go-stock-mcp main.go sse_server.go
```

#### 步骤 3: 配置 Supervisor

在宝塔面板中：
1. 进入 **软件商店**
2. 安装 **Supervisor管理器**
3. 点击 **添加守护进程**

填写配置：
- **名称**: go-stock-mcp
- **启动用户**: root
- **运行目录**: `/www/go-stock-mcp`
- **启动命令**: `./go-stock-mcp -mode=sse`
- **环境变量**:
  ```
  MCP_MODE=sse
  PORT=8080
  ```

4. 点击 **确定**

### 方式三：宝塔 Nginx 反向代理

如果你希望通过域名访问，可以使用 Nginx 反向代理：

#### 步骤 1: 创建网站

在宝塔面板中：
1. 进入 **网站**
2. 点击 **添加站点**
3. 填写域名（如：`mcp.yourdomain.com`）
4. 选择 **纯静态**

#### 步骤 2: 配置反向代理

1. 点击网站设置
2. 进入 **反向代理**
3. 添加反向代理：
   - **代理名称**: go-stock-mcp
   - **目标 URL**: `http://127.0.0.1:8080`
   - **发送域名**: `$host`

#### 步骤 3: 申请 SSL（可选但推荐）

1. 进入 **SSL**
2. 点击 **Let's Encrypt**
3. 申请证书

#### 步骤 4: 修改 MCP 配置

```json
{
  "mcpServers": {
    "go-stock": {
      "url": "https://mcp.yourdomain.com/sse"
    }
  }
}
```

---

## Docker 部署

### 使用 Docker Compose

创建 `docker-compose.yml`:

```yaml
version: '3.8'

services:
  go-stock-mcp:
    image: ghcr.io/kore-01/go-stock-mcp:latest
    container_name: go-stock-mcp
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - MCP_MODE=sse
      - PORT=8080
      - LOG_LEVEL=info
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:8080/health"]
      interval: 30s
      timeout: 3s
      retries: 3
      start_period: 5s
```

启动：

```bash
docker-compose up -d
```

### 使用 Docker 命令

```bash
# 拉取镜像
docker pull ghcr.io/kore-01/go-stock-mcp:latest

# 运行
docker run -d \
  --name go-stock-mcp \
  --restart unless-stopped \
  -p 8080:8080 \
  -e MCP_MODE=sse \
  -e PORT=8080 \
  ghcr.io/kore-01/go-stock-mcp:latest

# 查看日志
docker logs -f go-stock-mcp

# 停止
docker stop go-stock-mcp

# 删除
docker rm go-stock-mcp
```

---

## 手动部署

### 步骤 1: 连接到服务器

```bash
ssh root@your-server
```

### 步骤 2: 安装 Go

```bash
# 下载 Go
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz

# 解压
tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz

# 配置环境变量
echo 'export PATH=$PATH:/usr/local/go/bin' >> /etc/profile
source /etc/profile

# 验证
go version
```

### 步骤 3: 下载源码

```bash
cd /opt
git clone https://github.com/kore-01/go-stock-mcp.git
cd go-stock-mcp
```

### 步骤 4: 编译

```bash
go mod download
go build -ldflags="-s -w" -o go-stock-mcp main.go sse_server.go
```

### 步骤 5: 创建 Systemd 服务

```bash
cat > /etc/systemd/system/go-stock-mcp.service <<EOF
[Unit]
Description=Go-Stock MCP Server
After=network.target

[Service]
Type=simple
User=root
WorkingDirectory=/opt/go-stock-mcp
ExecStart=/opt/go-stock-mcp/go-stock-mcp -mode=sse
Restart=always
RestartSec=5
Environment="MCP_MODE=sse"
Environment="PORT=8080"

[Install]
WantedBy=multi-user.target
EOF

# 重载配置
systemctl daemon-reload

# 启动服务
systemctl enable go-stock-mcp
systemctl start go-stock-mcp

# 查看状态
systemctl status go-stock-mcp
```

### 步骤 6: 配置防火墙

```bash
# 开放端口 8080
ufw allow 8080/tcp

# 或使用 iptables
iptables -A INPUT -p tcp --dport 8080 -j ACCEPT
```

---

## 配置说明

### 环境变量

编辑配置文件：

```bash
# 如果使用 Systemd
systemctl edit go-stock-mcp

# 或编辑环境文件
vim /etc/go-stock-mcp/env
```

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `MCP_MODE` | 运行模式 | `sse` |
| `PORT` | 监听端口 | `8080` |
| `BASE_URL` | 服务基础 URL | `http://0.0.0.0:8080` |
| `LOG_LEVEL` | 日志级别 | `info` |

### 修改配置后重启

```bash
# Systemd
systemctl restart go-stock-mcp

# Docker
docker restart go-stock-mcp

# Docker Compose
docker-compose restart
```

---

## 验证部署

### 已部署服务（GZ 内网）

**服务地址**：`http://10.1.20.3:28080`

```bash
# 健康检查
curl http://10.1.20.3:28080/health

# 应返回：
# {"status":"ok","version":"1.0.0","time":"2026-03-16T12:00:00Z"}
```

### 测试你自己的部署

```bash
# 健康检查
curl http://your-server:8080/health

# 应返回：
# {"status":"ok","version":"1.0.0","time":"2026-03-16T12:00:00Z"}

# 查看服务信息
curl http://your-server:8080/
```

### MCP 客户端配置

**已部署服务（GZ 内网）**：
```json
{
  "mcpServers": {
    "go-stock": {
      "url": "http://10.1.20.3:28080/sse",
      "description": "Go-Stock 股票数据服务 (GZ内网)"
    }
  }
}
```

**自定义服务器**：
```json
{
  "mcpServers": {
    "go-stock": {
      "url": "http://your-server:8080/sse",
      "description": "Go-Stock 股票数据服务"
    }
  }
}
```

---

## 故障排查

### 服务无法启动

```bash
# 查看日志
journalctl -u go-stock-mcp -f

# 或 Docker
docker logs go-stock-mcp
```

### 端口被占用

```bash
# 检查端口占用
netstat -tlnp | grep 8080
lsof -i :8080

# 修改端口
export PORT=28080

# 或在 systemd 服务文件中修改
systemctl edit go-stock-mcp
# 添加：
# [Service]
# Environment="PORT=28080"
```

### 防火墙问题

```bash
# 检查防火墙状态
ufw status

# 临时关闭防火墙测试
ufw disable

# 开放端口
ufw allow 8080/tcp
```

### 测试连接

```bash
# 本地测试
curl http://localhost:8080/health

# 远程测试（GZ 服务器已部署服务）
curl http://10.1.20.3:28080/health
```

---

## 更新部署

### 使用一键脚本

```bash
# 重新运行安装脚本
curl -fsSL https://raw.githubusercontent.com/kore-01/go-stock-mcp/main/deploy/install.sh | bash
```

### 手动更新

```bash
# 进入目录
cd /opt/go-stock-mcp

# 拉取最新代码
git pull

# 重新编译
go build -ldflags="-s -w" -o go-stock-mcp main.go sse_server.go

# 重启服务
systemctl restart go-stock-mcp
```

### Docker 更新

```bash
# 拉取最新镜像
docker pull ghcr.io/kore-01/go-stock-mcp:latest

# 停止并删除旧容器
docker stop go-stock-mcp
docker rm go-stock-mcp

# 运行新容器
docker run -d \
  --name go-stock-mcp \
  --restart unless-stopped \
  -p 8080:8080 \
  ghcr.io/kore-01/go-stock-mcp:latest
```

---

## 安全建议

1. **使用 HTTPS**: 生产环境配置 SSL 证书
2. **限制访问**: 配置防火墙，只允许特定 IP 访问
3. **定期更新**: 及时更新到最新版本
4. **监控日志**: 定期检查日志文件
5. **备份配置**: 定期备份配置文件

---

## 获取帮助

- GitHub Issues: https://github.com/kore-01/go-stock-mcp/issues
- 文档: https://github.com/kore-01/go-stock-mcp/blob/main/README.md
