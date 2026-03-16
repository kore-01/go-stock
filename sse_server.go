// SSE Server for Go-Stock MCP
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// runSSEServer 启动 SSE 模式服务器
func runSSEServer() {
	// 加载环境变量
	godotenv.Load()

	// 创建 MCP Server
	s := server.NewMCPServer(
		"Go-Stock MCP Server",
		version,
		server.WithResourceCapabilities(true, true),
		server.WithLogging(),
	)

	// 注册 Tools
	registerStockTools(s)
	registerKLineTools(s)
	registerNewsTools(s)
	registerReportTools(s)
	registerMarketTools(s)

	// 获取端口
	port := getEnv("PORT", "8080")
	baseURL := getEnv("BASE_URL", fmt.Sprintf("http://localhost:%s", port))

	// 创建 SSE 服务器
	sseServer := server.NewSSEServer(s,
		server.WithBaseURL(baseURL),
		server.WithSSEEndpoint("/sse"),
		server.WithMessageEndpoint("/message"),
	)

	// 设置路由
	http.Handle("/", sseServer)
	http.HandleFunc("/health", healthHandler)

	// 启动服务器
	addr := fmt.Sprintf("0.0.0.0:%s", port)
	log.Printf("Go-Stock MCP Server v%s starting...", version)
	log.Printf("SSE endpoint: http://%s/sse", addr)
	log.Printf("Health check: http://%s/health", addr)

	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// healthHandler 健康检查处理器
func healthHandler(w http.ResponseWriter, r *http.Request) {
	health := map[string]interface{}{
		"status":  "ok",
		"version": version,
		"time":    time.Now().Format(time.RFC3339),
		"service": "Go-Stock MCP Server",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(health)
}

// API 响应结构
type APIResponse struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// handleAPIError 统一错误处理
func handleAPIError(w http.ResponseWriter, message string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(APIResponse{
		Code:    code,
		Message: message,
		Data:    nil,
	})
}

// handleAPISuccess 统一成功响应
func handleAPISuccess(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(APIResponse{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// ToolRequest 工具调用请求
type ToolRequest struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

// ToolResponse 工具调用响应
type ToolResponse struct {
	Success bool        `json:"success"`
	Result  interface{} `json:"result,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// createMCPCallRequest 创建 MCP 调用请求
func createMCPCallRequest(toolName string, args map[string]interface{}) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Request: mcp.Request{
			Method: "tools/call",
		},
		Params: struct {
			Name      string                 `json:"name"`
			Arguments map[string]interface{} `json:"arguments,omitempty"`
			Meta      *struct {
				ProgressToken mcp.ProgressToken `json:"progressToken,omitempty"`
			} `json:"_meta,omitempty"`
		}{
			Name:      toolName,
			Arguments: args,
		},
	}
}

// executeTool 执行工具调用
func executeTool(ctx context.Context, s *server.MCPServer, toolName string, args map[string]interface{}) (*mcp.CallToolResult, error) {
	// 创建请求
	request := createMCPCallRequest(toolName, args)

	// 查找对应的 handler
	switch toolName {
	case "get_stock_realtime":
		return handleGetStockRealtime(ctx, request)
	case "search_stocks":
		return handleSearchStocks(ctx, request)
	case "get_stock_detail":
		return handleGetStockDetail(ctx, request)
	case "get_kline_data":
		return handleGetKLineData(ctx, request)
	case "get_stock_news":
		return handleGetStockNews(ctx, request)
	case "get_market_news":
		return handleGetMarketNews(ctx, request)
	case "get_research_reports":
		return handleGetResearchReports(ctx, request)
	case "get_market_indices":
		return handleGetMarketIndices(ctx, request)
	case "get_hot_stocks":
		return handleGetHotStocks(ctx, request)
	default:
		return nil, fmt.Errorf("unknown tool: %s", toolName)
	}
}
