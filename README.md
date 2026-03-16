# Go-Stock MCP Server

股票数据服务 MCP Server，提供实时行情、K线数据、新闻、研报、大盘指数等功能。

## 🚀 在线服务

**已部署服务地址**（GZ 内网服务器）：
- SSE 端点：`http://10.1.20.3:28080/sse`
- 健康检查：`http://10.1.20.3:28080/health`
- 服务状态：✅ 运行中

> 注意：默认端口 8080/18080 被占用，实际使用端口 **28080**

## 功能特性

- 股票实时行情查询
- 股票搜索
- K线数据（支持多种周期）
- 股票相关新闻
- 市场热点新闻
- 个股研报
- 大盘指数
- 热门股票排行

## 快速开始

### 本地运行

```bash
# 克隆仓库
git clone https://github.com/kore-01/go-stock-mcp.git
cd go-stock-mcp

# 编译
go build -ldflags="-s -w" -o go-stock-mcp main.go sse_server.go

# 运行（STDIO 模式）
./go-stock-mcp

# 或 SSE 模式
./go-stock-mcp -mode=sse
```

### Docker 部署

```bash
# 构建镜像
docker build -t go-stock-mcp .

# 运行容器
docker run -d -p 8080:8080 --name go-stock-mcp go-stock-mcp
```

### 服务器部署

使用一键部署脚本：

```bash
# 下载并运行部署脚本
curl -fsSL https://raw.githubusercontent.com/kore-01/go-stock-mcp/main/deploy/install.sh | bash
```

详细部署文档：[DEPLOY.md](./DEPLOY.md)

## 配置

### 环境变量

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `MCP_MODE` | 运行模式：stdio 或 sse | `stdio` |
| `PORT` | SSE 模式监听端口 | `8080` |
| `BASE_URL` | SSE 服务基础 URL | `http://localhost:8080` |
| `LOG_LEVEL` | 日志级别 | `info` |

### MCP 客户端配置

#### OpenClaw（已部署服务）

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

#### Claude Desktop（已部署服务）

```json
{
  "mcpServers": {
    "go-stock": {
      "url": "http://10.1.20.3:28080/sse"
    }
  }
}
```

#### 本地 STDIO 模式

**OpenClaw**:
```json
{
  "mcpServers": {
    "go-stock": {
      "command": "/usr/local/bin/go-stock-mcp",
      "args": [],
      "env": {}
    }
  }
}
```

## 可用工具

| 工具名 | 描述 |
|--------|------|
| `get_stock_realtime` | 获取股票实时行情 |
| `search_stocks` | 搜索股票 |
| `get_stock_detail` | 获取股票详细信息 |
| `get_kline_data` | 获取K线数据 |
| `get_stock_news` | 获取股票相关新闻 |
| `get_market_news` | 获取市场热点新闻 |
| `get_research_reports` | 获取个股研报 |
| `get_market_indices` | 获取大盘指数 |
| `get_hot_stocks` | 获取热门股票排行 |

## 文档

- [部署指南](./DEPLOY.md) - 详细的服务器部署教程
- [快速开始](./QUICKSTART.md) - 5分钟快速上手指南

## 技术栈

- Go 1.24+
- MCP Go SDK
- 新浪财经 API
- 东方财富 API

## License

MIT License
