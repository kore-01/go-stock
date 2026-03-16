package main

import (
	"encoding/json"
	"fmt"
	"github.com/mark3labs/mcp-go/mcp"
)

func main() {
	result1 := mcp.NewToolResultText("test data")
	data1, _ := json.MarshalIndent(result1, "", "  ")
	fmt.Println("mcp.NewToolResultText:")
	fmt.Println(string(data1))
	fmt.Println()

	result2 := &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: "test data",
			},
		},
	}
	data2, _ := json.MarshalIndent(result2, "", "  ")
	fmt.Println("手动构建:")
	fmt.Println(string(data2))
}
