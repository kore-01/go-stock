# Go-Stock MCP Server 快速开始指南

本文档帮助你在 5 分钟内完成从代码到部署的全过程。

## 🚀 已部署服务（GZ 内网）

**服务已部署并运行中**：

| 项目 | 详情 |
|------|------|
| **服务器** | GZ 内网服务器 |
| **IP** | `10.1.20.3` |
| **端口** | `28080`（端口 8080/18080 被占用） |
| **SSE 端点** | `http://10.1.20.3:28080/sse` |
| **健康检查** | `http://10.1.20.3:28080/health` |
| **状态** | ✅ 运行中 |

**快速配置 OpenClaw**：
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

---

## 目录

1. [推送到 GitHub](#推送到-github)
2. [部署到服务器](#部署到服务器)
3. [配置 MCP 客户端](#配置-mcp-客户端)
4. [验证部署](#验证部署)

---

## 推送到 GitHub

### 步骤 1: 创建 GitHub 仓库

1. 访问 https://github.com/new
2. 仓库名称: `go-stock-mcp`
3. 选择 **Public** 或 **Private**
4. 不要初始化 README（已存在）
5. 点击 **Create repository**

### 步骤 2: 推送代码

```bash
# 进入项目目录
cd D:\ai\go-stock-mcp

# 初始化 git 仓库
git init

# 添加所有文件
git add .

# 提交
git commit -m "Initial commit: Go-Stock MCP Server v1.0.0"

# 添加远程仓库（替换为你的用户名）
git remote add origin https://github.com/kore-01/go-stock-mcp.git

# 推送
git push -u origin main
```

### 步骤 3: 验证推送

访问 `https://github.com/kore-01/go-stock-mcp` 确认代码已推送。

---

## 部署到服务器

### 方式一：一键自动部署（推荐）

在本地运行部署脚本：

```bash
# 进入项目目录
cd D:\ai\go-stock-mcp

# 运行部署脚本（需要配置 SSH）
bash deploy/deploy-to-gz.sh
```

按照提示选择：
- 输入 `1` 进行完整部署
- 脚本会自动连接服务器、安装依赖、编译、启动服务

### 方式二：手动部署到服务器

SSH 连接到服务器并执行：

```bash
# 连接服务器
ssh root@gz

# 运行一键安装脚本
curl -fsSL https://raw.githubusercontent.com/kore-01/go-stock-mcp/main/deploy/install.sh | bash
```

或者手动步骤：

```bash
# 连接服务器
ssh root@gz

# 安装 Go
wget https://go.dev/dl/go1.24.0.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.24.0.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin

# 克隆仓库
git clone https://github.com/kore-01/go-stock-mcp.git /opt/go-stock-mcp
cd /opt/go-stock-mcp

# 编译
go mod download
go build -ldflags="-s -w" -o go-stock-mcp main.go sse_server.go

# 创建服务
cat > /etc/systemd/system/go-stock-mcp.service <<'EOF'
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

[Install]
WantedBy=multi-user.target
EOF

# 启动服务
systemctl daemon-reload
systemctl enable go-stock-mcp
systemctl start go-stock-mcp

# 开放防火墙
ufw allow 8080/tcp
```

### 方式三：Docker 部署

```bash
ssh root@gz

# 安装 Docker
curl -fsSL https://get.docker.com | bash

# 运行容器
docker run -d \
  --name go-stock-mcp \
  --restart unless-stopped \
  -p 8080:8080 \
  ghcr.io/kore-01/go-stock-mcp:latest
```

---

## 配置 MCP 客户端

### 快速配置（使用已部署服务）

**OpenClaw 配置**（编辑 `~/.config/openclaw/mcp.json`）：
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

**Claude Desktop 配置**：
- Windows: `%APPDATA%\Claude\settings.json`
- macOS: `~/Library/Application Support/Claude/settings.json`

```json
{
  "mcpServers": {
    "go-stock": {
      "url": "http://10.1.20.3:28080/sse"
    }
  }
}
```

---

### OpenClaw（自定义服务器）

编辑配置文件（位置：`~/.config/openclaw/mcp.json`）：

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

重启 OpenClaw：

```bash
openclaw restart
```

---

### Claude Desktop（自定义服务器）

编辑配置文件：

- Windows: `%APPDATA%\Claude\settings.json`
- macOS: `~/Library/Application Support/Claude/settings.json`

```json
{
  "mcpServers": {
    "go-stock": {
      "url": "http://your-server:8080/sse"
    }
  }
}
```

**注意**：如果端口 8080 被占用，请使用 28080 或其他端口。

重启 Claude Desktop。

### 其他 MCP 客户端

通用配置格式：

```json
{
  "mcpServers": {
    "go-stock": {
      "url": "http://your-server:8080/sse"
    }
  }
}
```

---

## 验证部署

### 1. 检查已部署服务（GZ 内网）

```bash
# 健康检查
curl http://10.1.20.3:28080/health

# 预期返回
{"status":"ok","version":"1.0.0","time":"2026-03-16T12:00:00Z"}

# 查看服务信息
curl http://10.1.20.3:28080/
```

### 2. 检查自定义部署

```bash
# SSH 到服务器
ssh root@your-server

# 检查服务状态
systemctl status go-stock-mcp

# 本地健康检查
curl http://localhost:8080/health
```

### 3. 在 MCP 客户端中测试

向 OpenClaw 或 Claude 提问：

```
贵州茅台今天的股价是多少？
```

如果 MCP 客户端正确配置，它会自动调用 `get_stock_realtime` 工具。

**测试查询**：
- "查看今天的股票行情"
- "有什么最新财经新闻？"
- "查查茅台的最新研报"

---

## 管理命令

### 查看日志（GZ 服务器）

```bash
# 查看运行日志
ssh root@gz "cat /opt/go-stock-mcp/nohup.out"

# 或使用 ps 查看进程
ssh root@gz "ps aux | grep go-stock-mcp"
```

### 重启服务（GZ 服务器）

```bash
ssh root@gz "pkill go-stock-mcp; cd /opt/go-stock-mcp && nohup ./go-stock-mcp -mode=sse > /dev/null 2>&1 &"
```

### 更新到最新版本（自定义服务器）

```bash
# 方式 1: 使用部署脚本
cd D:\ai\go-stock-mcp
bash deploy/deploy-to-gz.sh
# 选择选项 2 (更新代码)

# 方式 2: 手动更新
ssh root@your-server << 'EOF'
cd /opt/go-stock-mcp
git pull
/usr/local/go/bin/go build -ldflags="-s -w" -o go-stock-mcp main.go sse_server.go
systemctl restart go-stock-mcp
EOF
```

---

## 故障排查

### 无法连接到 GZ 服务器

```bash
# 测试 SSH（使用 gz 主机名）
ssh root@gz echo "OK"

# 检查服务器是否在线
ssh root@gz "hostname -I"
```

### 服务无法启动

```bash
# 查看进程是否在运行
ssh root@gz "ps aux | grep go-stock-mcp"

# 检查端口监听
ssh root@gz "netstat -tlnp | grep 28080"

# 手动启动查看错误
ssh root@gz "cd /opt/go-stock-mcp && ./go-stock-mcp -mode=sse"
```

### 端口无法访问

```bash
# 检查防火墙
ssh root@gz "ufw status"

# 检查端口监听（注意端口是 28080）
ssh root@gz "netstat -tlnp | grep 28080"
# 或
ssh root@gz "ss -tlnp | grep 28080"
```

### MCP 客户端无法连接

1. 确认服务运行：`curl http://10.1.20.3:28080/health`
2. 检查防火墙设置
3. 确认 MCP 配置中使用正确端口 **28080**
4. 确认 MCP 配置格式正确：
   ```json
   {
     "mcpServers": {
       "go-stock": {
         "url": "http://10.1.20.3:28080/sse"
       }
     }
   }
   ```

---

## 安全建议

### 1. 修改默认端口（GZ 服务器已使用 28080）

GZ 服务器由于端口 8080/18080 被占用，已使用端口 28080。

对于自定义服务器，如需修改端口：

```bash
ssh root@your-server
systemctl edit go-stock-mcp
```

添加：
```ini
[Service]
Environment="PORT=28080"
```

重启服务：
```bash
systemctl daemon-reload
systemctl restart go-stock-mcp
```

更新防火墙：
```bash
ufw delete allow 8080/tcp
ufw allow 28080/tcp
```

### 2. 配置 Nginx + SSL

参考 [DEPLOY.md](./DEPLOY.md) 中的宝塔面板配置。

### 3. 限制访问 IP

在防火墙中只允许特定 IP 访问：

```bash
ufw allow from YOUR_IP to any port 8080
```

---

## 获取帮助

- GitHub Issues: https://github.com/kore-01/go-stock-mcp/issues
- 部署文档: [DEPLOY.md](./DEPLOY.md)

---

## 下一步

1. ✅ 推送到 GitHub
2. ✅ 部署到服务器
3. ✅ 配置 MCP 客户端
4. ✅ 验证部署

现在你可以在 OpenClaw 或 Claude 中使用股票数据服务了！
