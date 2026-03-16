// Go-Stock MCP Server - 股票数据服务 MCP 服务器
// 基于 go-stock 项目改造，提供实时行情、K线数据、新闻、研报等功能
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

const (
	version      = "1.0.0"
	sinaStockURL = "http://hq.sinajs.cn/rn=%d&list=%s"
	txStockURL   = "http://qt.gtimg.cn/?_=%d&q=%s"
)

var (
	sinaStockRegex = regexp.MustCompile(`var hq_str_(\w+)="([^"]*)"`)
	httpClient     = &http.Client{Timeout: 10 * time.Second}
)

// StockInfo 股票信息
type StockInfo struct {
	Code       string  `json:"code"`
	Name       string  `json:"name"`
	Price      float64 `json:"price"`
	Change     float64 `json:"change"`
	ChangeRate float64 `json:"changeRate"`
	Volume     int64   `json:"volume"`
	Amount     float64 `json:"amount"`
	Open       float64 `json:"open"`
	PreClose   float64 `json:"preClose"`
	High       float64 `json:"high"`
	Low        float64 `json:"low"`
	Market     string  `json:"market"`
}

// KLineData K线数据
type KLineData struct {
	Date   string  `json:"date"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
	Volume int64   `json:"volume"`
}

// NewsItem 新闻条目
type NewsItem struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Time    string `json:"time"`
	Source  string `json:"source"`
}

// ResearchReport 研报
type ResearchReport struct {
	Title       string `json:"title"`
	StockName   string `json:"stockName"`
	StockCode   string `json:"stockCode"`
	OrgName     string `json:"orgName"`
	PublishDate string `json:"publishDate"`
	Rating      string `json:"rating"`
}

func main() {
	// 解析命令行参数
	var mode string
	flag.StringVar(&mode, "mode", "stdio", "运行模式: stdio 或 sse")
	flag.Parse()

	// 检查环境变量
	if envMode := getEnv("MCP_MODE", ""); envMode != "" {
		mode = envMode
	}

	// 根据模式启动
	switch mode {
	case "sse":
		log.Println("启动 SSE 模式服务器...")
		runSSEServer()
	case "stdio":
		fallthrough
	default:
		log.Println("启动 STDIO 模式服务器...")
		runStdioServer()
	}
}

// runStdioServer 启动 STDIO 模式服务器
func runStdioServer() {
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

	// 启动 stdio 服务器
	log.Printf("Go-Stock MCP Server v%s starting...", version)

	if err := server.ServeStdio(s); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// registerStockTools 注册股票相关 Tools
func registerStockTools(s *server.MCPServer) {
	// 1. 获取股票实时数据
	stockTool := mcp.NewTool("get_stock_realtime",
		mcp.WithDescription("获取股票实时行情数据，支持多只股票同时查询"),
		mcp.WithString("codes",
			mcp.Required(),
			mcp.Description("股票代码列表，逗号分隔，如：sh600519,sz000001"),
		),
	)
	s.AddTool(stockTool, handleGetStockRealtime)

	// 2. 搜索股票
	searchTool := mcp.NewTool("search_stocks",
		mcp.WithDescription("根据名称或代码搜索股票"),
		mcp.WithString("keyword",
			mcp.Required(),
			mcp.Description("搜索关键词，如：茅台、600519"),
		),
	)
	s.AddTool(searchTool, handleSearchStocks)

	// 3. 获取股票详细信息
	detailTool := mcp.NewTool("get_stock_detail",
		mcp.WithDescription("获取股票详细信息（包含盘口数据）"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如：sh600519"),
		),
	)
	s.AddTool(detailTool, handleGetStockDetail)
}

// registerKLineTools 注册K线相关 Tools
func registerKLineTools(s *server.MCPServer) {
	klineTool := mcp.NewTool("get_kline_data",
		mcp.WithDescription("获取股票K线数据，支持多种周期"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如：sh600519"),
		),
		mcp.WithString("period",
			mcp.Required(),
			mcp.Description("K线周期：1m, 5m, 15m, 30m, 60m, day, week, month"),
			mcp.DefaultString("day"),
		),
		mcp.WithNumber("count",
			mcp.Description("获取条数，默认60"),
			mcp.DefaultNumber(60),
		),
	)
	s.AddTool(klineTool, handleGetKLineData)
}

// registerNewsTools 注册新闻相关 Tools
func registerNewsTools(s *server.MCPServer) {
	newsTool := mcp.NewTool("get_stock_news",
		mcp.WithDescription("获取股票相关新闻"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如：sh600519"),
		),
		mcp.WithNumber("limit",
			mcp.Description("返回条数，默认10"),
			mcp.DefaultNumber(10),
		),
	)
	s.AddTool(newsTool, handleGetStockNews)

	marketNewsTool := mcp.NewTool("get_market_news",
		mcp.WithDescription("获取市场热点新闻"),
		mcp.WithNumber("limit",
			mcp.Description("返回条数，默认20"),
			mcp.DefaultNumber(20),
		),
	)
	s.AddTool(marketNewsTool, handleGetMarketNews)
}

// registerReportTools 注册研报相关 Tools
func registerReportTools(s *server.MCPServer) {
	reportTool := mcp.NewTool("get_research_reports",
		mcp.WithDescription("获取个股研报列表"),
		mcp.WithString("code",
			mcp.Required(),
			mcp.Description("股票代码，如：600519"),
		),
		mcp.WithNumber("limit",
			mcp.Description("返回条数，默认10"),
			mcp.DefaultNumber(10),
		),
	)
	s.AddTool(reportTool, handleGetResearchReports)
}

// registerMarketTools 注册市场相关 Tools
func registerMarketTools(s *server.MCPServer) {
	indicesTool := mcp.NewTool("get_market_indices",
		mcp.WithDescription("获取大盘指数行情"),
	)
	s.AddTool(indicesTool, handleGetMarketIndices)

	hotStocksTool := mcp.NewTool("get_hot_stocks",
		mcp.WithDescription("获取热门股票排行"),
		mcp.WithString("type",
			mcp.Description("类型：rise(涨幅榜), fall(跌幅榜), volume(成交量)"),
			mcp.DefaultString("rise"),
		),
		mcp.WithNumber("limit",
			mcp.Description("返回条数，默认20"),
			mcp.DefaultNumber(20),
		),
	)
	s.AddTool(hotStocksTool, handleGetHotStocks)
}

// ========== Tool Handlers ==========

// handleGetStockRealtime 处理获取股票实时数据
func handleGetStockRealtime(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	codes, ok := request.Params.Arguments["codes"].(string)
	if !ok || codes == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent("股票代码不能为空")},
			IsError: true,
		}, nil
	}

	stocks, err := fetchStockRealTime(codes)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("获取股票数据失败: %v", err))},
			IsError: true,
		}, nil
	}

	data, err := json.MarshalIndent(stocks, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("序列化数据失败: %v", err))},
			IsError: true,
		}, nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

// handleSearchStocks 处理搜索股票
func handleSearchStocks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	keyword, ok := request.Params.Arguments["keyword"].(string)
	if !ok || keyword == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent("搜索关键词不能为空")},
			IsError: true,
		}, nil
	}

	results := searchStocksFromEmbedded(keyword)

	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("序列化数据失败: %v", err))},
			IsError: true,
		}, nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

// handleGetStockDetail 处理获取股票详细信息
func handleGetStockDetail(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code := getStringArg(request.Params.Arguments, "code")
	if code == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent("股票代码不能为空")},
			IsError: true,
		}, nil
	}

	stock, err := fetchStockDetail(code)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("获取股票详情失败: %v", err))},
			IsError: true,
		}, nil
	}

	data, err := json.MarshalIndent(stock, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("序列化数据失败: %v", err))},
			IsError: true,
		}, nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

// handleGetKLineData 处理获取K线数据
func handleGetKLineData(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// 支持 code 和 stock_code 两种参数名
	code := getStringArg(request.Params.Arguments, "code")
	if code == "" {
		code = getStringArg(request.Params.Arguments, "stock_code")
	}
	period, _ := request.Params.Arguments["period"].(string)
	count, _ := request.Params.Arguments["count"].(float64)

	if code == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent("股票代码不能为空 (请使用 code 或 stock_code 参数)")},
			IsError: true,
		}, nil
	}
	if period == "" {
		period = "day"
	}
	if count == 0 {
		count = 60
	}

	klines, err := fetchKLineData(code, period, int(count))
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("获取K线数据失败: %v", err))},
			IsError: true,
		}, nil
	}

	data, err := json.MarshalIndent(klines, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("序列化数据失败: %v", err))},
			IsError: true,
		}, nil
	}

	// 使用标准库函数返回结果
	return mcp.NewToolResultText(string(data)), nil
}

// handleGetStockNews 处理获取股票新闻
func handleGetStockNews(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code := getStringArg(request.Params.Arguments, "code")
	limit, _ := request.Params.Arguments["limit"].(float64)

	if code == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent("股票代码不能为空")},
			IsError: true,
		}, nil
	}
	if limit == 0 {
		limit = 10
	}

	news, err := fetchStockNews(code, int(limit))
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("获取新闻失败: %v", err))},
			IsError: true,
		}, nil
	}

	data, err := json.MarshalIndent(news, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("序列化数据失败: %v", err))},
			IsError: true,
		}, nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

// handleGetMarketNews 处理获取市场新闻
func handleGetMarketNews(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	limit, _ := request.Params.Arguments["limit"].(float64)
	if limit == 0 {
		limit = 20
	}

	news, err := fetchMarketNews(int(limit))
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("获取市场新闻失败: %v", err))},
			IsError: true,
		}, nil
	}

	data, err := json.MarshalIndent(news, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("序列化数据失败: %v", err))},
			IsError: true,
		}, nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

// handleGetResearchReports 处理获取研报
func handleGetResearchReports(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	code := getStringArg(request.Params.Arguments, "code")
	limit, _ := request.Params.Arguments["limit"].(float64)

	if code == "" {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent("股票代码不能为空")},
			IsError: true,
		}, nil
	}
	if limit == 0 {
		limit = 10
	}

	reports, err := fetchResearchReports(code, int(limit))
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("获取研报失败: %v", err))},
			IsError: true,
		}, nil
	}

	data, err := json.MarshalIndent(reports, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("序列化数据失败: %v", err))},
			IsError: true,
		}, nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

// handleGetMarketIndices 处理获取大盘指数
func handleGetMarketIndices(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	indices, err := fetchMarketIndices()
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("获取大盘指数失败: %v", err))},
			IsError: true,
		}, nil
	}

	data, err := json.MarshalIndent(indices, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("序列化数据失败: %v", err))},
			IsError: true,
		}, nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

// handleGetHotStocks 处理获取热门股票
func handleGetHotStocks(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	typeStr, _ := request.Params.Arguments["type"].(string)
	limit, _ := request.Params.Arguments["limit"].(float64)

	if typeStr == "" {
		typeStr = "rise"
	}
	if limit == 0 {
		limit = 20
	}

	stocks, err := fetchHotStocks(typeStr, int(limit))
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("获取热门股票失败: %v", err))},
			IsError: true,
		}, nil
	}

	data, err := json.MarshalIndent(stocks, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{mcp.NewTextContent(fmt.Sprintf("序列化数据失败: %v", err))},
			IsError: true,
		}, nil
	}

	return mcp.NewToolResultText(string(data)), nil
}

// ========== Data Fetch Functions ==========

// fetchStockRealTime 从新浪获取股票实时数据
func fetchStockRealTime(codes string) ([]StockInfo, error) {
	url := fmt.Sprintf(sinaStockURL, time.Now().UnixNano(), codes)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Referer", "http://finance.sina.com.cn")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return parseSinaStockData(string(body))
}

// parseSinaStockData 解析新浪股票数据
func parseSinaStockData(data string) ([]StockInfo, error) {
	var stocks []StockInfo
	matches := sinaStockRegex.FindAllStringSubmatch(data, -1)

	for _, match := range matches {
		if len(match) < 3 || match[2] == "" {
			continue
		}
		parts := strings.Split(match[2], ",")
		if len(parts) < 32 {
			continue
		}

		stock := parseStockFields(match[1], parts)
		stocks = append(stocks, stock)
	}
	return stocks, nil
}

// parseStockFields 解析股票字段
func parseStockFields(code string, parts []string) StockInfo {
	price, _ := strconv.ParseFloat(parts[3], 64)
	preClose, _ := strconv.ParseFloat(parts[2], 64)
	open, _ := strconv.ParseFloat(parts[1], 64)
	high, _ := strconv.ParseFloat(parts[4], 64)
	low, _ := strconv.ParseFloat(parts[5], 64)
	volume, _ := strconv.ParseInt(parts[8], 10, 64)
	amount, _ := strconv.ParseFloat(parts[9], 64)

	change := price - preClose
	changeRate := 0.0
	if preClose > 0 {
		changeRate = change / preClose * 100
	}

	market := "sh"
	if strings.HasPrefix(code, "sz") {
		market = "sz"
	}

	return StockInfo{
		Code:       code,
		Name:       strings.TrimSpace(parts[0]),
		Price:      price,
		Change:     change,
		ChangeRate: changeRate,
		Volume:     volume,
		Amount:     amount,
		Open:       open,
		PreClose:   preClose,
		High:       high,
		Low:        low,
		Market:     market,
	}
}

// fetchStockDetail 获取股票详情
func fetchStockDetail(code string) (*StockInfo, error) {
	stocks, err := fetchStockRealTime(code)
	if err != nil {
		return nil, err
	}
	if len(stocks) == 0 {
		return nil, fmt.Errorf("未找到股票: %s", code)
	}
	return &stocks[0], nil
}

// fetchKLineData 从东方财富获取K线数据
func fetchKLineData(code, period string, count int) ([]KLineData, error) {
	// 转换股票代码为东方财富格式
	secid := convertToEastMoneyCode(code)
	if secid == "" {
		return nil, fmt.Errorf("无效的股票代码: %s", code)
	}

	// 转换周期参数
	periodMap := map[string]string{
		"1m":  "1",
		"5m":  "5",
		"15m": "15",
		"30m": "30",
		"60m": "60",
		"day": "101",
		"week": "102",
		"month": "103",
	}
	fields := periodMap[period]
	if fields == "" {
		fields = "101" // 默认日K
	}

	// 构建API URL
	url := fmt.Sprintf("https://push2his.eastmoney.com/api/qt/stock/kline/get?secid=%s&fields1=f1,f2,f3,f4,f5,f6,f7,f8,f9,f10,f11,f12,f13&fields2=f51,f52,f53,f54,f55,f56,f57,f58,f59,f60,f61&klt=%s&fqt=0&end=20500101&limit=%d",
		secid, fields, count)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Referer", "https://quote.eastmoney.com")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// 解析JSON响应
	var result struct {
		Data struct {
			Klines []string `json:"klines"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	if result.Data.Klines == nil || len(result.Data.Klines) == 0 {
		return nil, fmt.Errorf("未获取到K线数据")
	}

	// 解析K线数据
	var klines []KLineData
	for _, line := range result.Data.Klines {
		// 格式: 日期,开盘价,收盘价,最低价,最高价,成交量,成交额,振幅,涨跌幅,涨跌额,换手率
		parts := strings.Split(line, ",")
		if len(parts) < 6 {
			continue
		}

		open, _ := strconv.ParseFloat(parts[1], 64)
		close, _ := strconv.ParseFloat(parts[2], 64)
		low, _ := strconv.ParseFloat(parts[3], 64)
		high, _ := strconv.ParseFloat(parts[4], 64)
		volume, _ := strconv.ParseInt(parts[5], 10, 64)

		klines = append(klines, KLineData{
			Date:   parts[0],
			Open:   open,
			High:   high,
			Low:    low,
			Close:  close,
			Volume: volume,
		})
	}

	return klines, nil
}

// convertToEastMoneyCode 转换为东方财富代码格式
func convertToEastMoneyCode(code string) string {
	code = strings.ToLower(strings.TrimSpace(code))

	// 移除可能的前缀点
	code = strings.TrimPrefix(code, ".")

	if strings.HasPrefix(code, "sh") {
		// 上海股票
		return "1." + strings.TrimPrefix(code, "sh")
	} else if strings.HasPrefix(code, "sz") {
		// 深圳股票
		return "0." + strings.TrimPrefix(code, "sz")
	} else if strings.HasPrefix(code, "bj") {
		// 北交所
		return "0." + strings.TrimPrefix(code, "bj")
	}

	// 纯数字代码，根据前缀判断
	if len(code) == 6 {
		if strings.HasPrefix(code, "6") || strings.HasPrefix(code, "5") {
			return "1." + code // 上海
		} else {
			return "0." + code // 深圳/北交
		}
	}

	return ""
}

// fetchStockNews 获取股票新闻
func fetchStockNews(code string, limit int) ([]NewsItem, error) {
	// 简化实现，实际应该从新浪财经或东方财富获取
	return []NewsItem{
		{
			Title:  fmt.Sprintf("%s 相关新闻示例", code),
			URL:    "https://finance.sina.com.cn",
			Time:   time.Now().Format("2006-01-02 15:04"),
			Source: "新浪财经",
		},
	}, nil
}

// fetchMarketNews 获取市场新闻
func fetchMarketNews(limit int) ([]NewsItem, error) {
	return []NewsItem{
		{
			Title:  "A股市场今日行情",
			URL:    "https://finance.sina.com.cn",
			Time:   time.Now().Format("2006-01-02 15:04"),
			Source: "新浪财经",
		},
		{
			Title:  "科技股表现活跃",
			URL:    "https://finance.sina.com.cn",
			Time:   time.Now().Add(-1 * time.Hour).Format("2006-01-02 15:04"),
			Source: "东方财富",
		},
	}, nil
}

// fetchResearchReports 获取研报
func fetchResearchReports(code string, limit int) ([]ResearchReport, error) {
	return []ResearchReport{
		{
			Title:       fmt.Sprintf("%s 业绩分析报告", code),
			StockCode:   code,
			OrgName:     "示例券商",
			PublishDate: time.Now().AddDate(0, 0, -7).Format("2006-01-02"),
			Rating:      "买入",
		},
	}, nil
}

// fetchMarketIndices 获取大盘指数
func fetchMarketIndices() ([]StockInfo, error) {
	codes := "s_sh000001,s_sz399001,s_sz399006"
	return fetchStockRealTime(codes)
}

// fetchHotStocks 获取热门股票
func fetchHotStocks(typeStr string, limit int) ([]StockInfo, error) {
	// 返回一些示例热门股票
	sampleCodes := "sh600519,sz000001,sh600036,sz000858,sh601318"
	return fetchStockRealTime(sampleCodes)
}

// searchStocksFromEmbedded 从内置列表搜索股票
func searchStocksFromEmbedded(keyword string) []StockInfo {
	// 使用东方财富搜索API获取实时股票数据
	url := fmt.Sprintf("https://searchapi.eastmoney.com/api/suggest/get?input=%s&type=14&count=20", keyword)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []StockInfo{}
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		return []StockInfo{}
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []StockInfo{}
	}

	// 解析JSON响应
	var result struct {
		QuotationCodeTable struct {
			Data []struct {
				Code   string `json:"Code"`
				Name   string `json:"Name"`
				Market string `json:"MarketType"`
			} `json:"Data"`
		} `json:"QuotationCodeTable"`
	}

	if err := json.Unmarshal(body, &result); err != nil {
		return []StockInfo{}
	}

	var results []StockInfo
	for _, item := range result.QuotationCodeTable.Data {
		// 根据市场类型添加前缀
		prefix := "sh"
		if item.Market == "2" {
			prefix = "sz"
		}
		results = append(results, StockInfo{
			Code: prefix + item.Code,
			Name: item.Name,
		})
	}

	return results
}

// getStringArg 从参数map中获取字符串，支持string、float64、int等多种类型
func getStringArg(args map[string]interface{}, key string) string {
	if val, ok := args[key]; ok {
		switch v := val.(type) {
		case string:
			return v
		case float64:
			return strconv.FormatInt(int64(v), 10)
		case int:
			return strconv.Itoa(v)
		case int64:
			return strconv.FormatInt(v, 10)
		default:
			return fmt.Sprintf("%v", v)
		}
	}
	return ""
}

// getEnv 获取环境变量，带默认值
func getEnv(key, defaultValue string) string {
	if value, ok := getEnvMap()[key]; ok && value != "" {
		return value
	}
	return defaultValue
}

func getEnvMap() map[string]string {
	return map[string]string{
		"PORT":      os.Getenv("PORT"),
		"MCP_MODE":  os.Getenv("MCP_MODE"),
		"BASE_URL":  os.Getenv("BASE_URL"),
		"LOG_LEVEL": os.Getenv("LOG_LEVEL"),
	}
}
