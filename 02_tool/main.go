package main

import (
	"fmt"
	"gomcp/db"
	"gomcp/ollam"
	"log"

	"github.com/ollama/ollama/api"
	"github.com/siuyin/dflt"
)

type ToolParams struct {
	Type       string                      `json:"type"`
	Defs       any                         `json:"$defs,omitempty"`
	Items      any                         `json:"items,omitempty"`
	Required   []string                    `json:"required"`
	Properties map[string]api.ToolProperty `json:"properties"`
}

func main() {
	host := dflt.EnvString("OLLAMA_HOST", "http://localhost:11434")
	model := dflt.EnvString("MODEL", "qwen3:0.6b")
	prompt := dflt.EnvString("PROMPT", "lookup stock level of iPhone 14")
	sys := dflt.EnvString("SYS", `You are a helpful assistant with access to lookupInventory tool which you MUST call.
	Strictly use only data from the database and provide concise answers.
	Example: "how much is the geeWhiz?", geeWhiz is the product_name.
	Example2: "is iphone 14 in stock", "iphone 14" is the product_name.
	Example3: "product id for simpleX phone", "simpleX" is the product_name.
	`)
	log.Printf("OLLAMA_HOST=%s MODEL=%s SYS=%q PROMPT=%q ", host, model, sys, prompt)

	messages := []api.Message{
		{Role: "system", Content: sys},
		{Role: "user", Content: prompt},
	}

	lookupInventory := api.Tool{
		Type: "function",
		Function: api.ToolFunction{
			Name:        "lookupInventory",
			Description: "look up inventory to get product ID, Name and Price in USD given a product_name",
			Parameters: ToolParams{
				Type:     "object",
				Required: []string{"product_name"},
				Properties: map[string]api.ToolProperty{
					"product_name": api.ToolProperty{Type: []string{"string"}},
				},
			},
		},
	}
	chat := newChat(host, model, messages, []api.Tool{lookupInventory})
	for n := 0; n < 2; n++ {
		chat.complete()
	}
	fmt.Println()

}

func newChat(host string, model string, messages []api.Message, tools []api.Tool) *myChat {
	c := myChat{}
	c.cl = ollam.Client(host)
	c.msgs = messages
	c.model = model
	c.tools = tools
	return &c
}

type myChat struct {
	cl    *api.Client
	msgs  []api.Message
	model string
	tools []api.Tool
}

func (m *myChat) complete() {
	responseFunction := func(r api.ChatResponse) error {
		if len(r.Message.ToolCalls) == 0 {
			fmt.Print(r.Message.Content)
			return nil
		}

		for _, tc := range r.Message.ToolCalls {
			fn := tc.Function
			log.Printf("Model wants to call tool: %s with args: %v", fn.Name, fn.Arguments)
			switch fn.Name {
			case "lookupInventory":
				pn := fn.Arguments["product_name"].(string)
				output := db.Get(pn)
				m.msgs = append(m.msgs, api.Message{
					Role:    "tool",
					Content: output,
				})
				log.Printf("\tTool: %s called with args: %s. resp: %s", fn.Name, fn.Arguments, output)
			}
		}
		return nil
	}

	ollam.ChatTools(m.cl, m.model, m.tools, m.msgs, responseFunction)

}
