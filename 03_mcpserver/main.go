package main

import (
	"context"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/siuyin/present_go_mcp_a2a/db"
)

func main() {
	log.Println("myserver running")
	server := mcp.NewServer(&mcp.Implementation{Name: "mymcp", Version: "v1.0.0"}, nil)

	mcp.AddTool(server,
		&mcp.Tool{Name: "lookupInventory",
			Description: "look up inventory to get product ID, Name and Price in USD given a product_name",
			Title:       "Lookup Inventory",
		}, lookup)
	server.AddPrompt(&mcp.Prompt{Name: "lookupInventorySystemPrompt"}, promptHandler)

	if err := server.Run(context.Background(), mcp.NewStdioTransport()); err != nil {
		log.Println("run: ", err)
	}
}

type lookupInput struct {
	Name string `json:"product_name" jsonschema:"the name of the product to lookup"`
}

func lookup(ctx context.Context, req *mcp.CallToolRequest, args lookupInput) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: db.Get(args.Name)}}}, nil, nil
}

func promptHandler(ctx context.Context, r *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	return &mcp.GetPromptResult{Messages: []*mcp.PromptMessage{
		&mcp.PromptMessage{Role: mcp.Role("system"),
			Content: &mcp.TextContent{Text: "System prompt"}},
		&mcp.PromptMessage{Role: mcp.Role("user"),
			Content: &mcp.TextContent{Text: "User prompt"}},
	}}, nil
}
