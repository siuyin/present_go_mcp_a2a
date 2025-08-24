package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func listPrompts(sess *mcp.ClientSession) {
	ctx := context.Background()
	lp, err := sess.ListPrompts(ctx, &mcp.ListPromptsParams{})
	if err != nil {
		log.Fatal("list prompts: ", err)
	}
	for _, p := range lp.Prompts {
		fmt.Printf("prompt: %s\n", p.Name)
		gpr, err := sess.GetPrompt(ctx, &mcp.GetPromptParams{Name: p.Name})
		if err != nil {
			log.Fatal("get prompt: ", err)
		}

		for _, m := range gpr.Messages {
			tc := textContent(m.Content)
			fmt.Printf("%v %v\n", m.Role, tc.Text)
		}
	}
}

func textContent(c mcp.Content) *mcp.TextContent {
	dat, err := c.MarshalJSON()
	if err != nil {
		log.Println("textContent marshal: ", err)
		return &mcp.TextContent{}
	}

	var tc mcp.TextContent
	if err := json.Unmarshal(dat, &tc); err != nil {
		log.Println("textContent unmarshal: ", err)
	}

	return &tc
}
