package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/ollama/ollama/api"
	"github.com/siuyin/dflt"
	"github.com/siuyin/mcptry/olamtl"
	"github.com/siuyin/present_go_mcp_a2a/ollam"
)

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

	sess := mcpConnect("myserver")
	defer sess.Close()

	tools := listTools(sess)

	chat := newChat(host, model, messages, tools, sess)
	for n := 0; n < 2; n++ {
		chat.complete()
	}
	fmt.Println()

	//listOllam(tools)
	//listPrompts(sess)
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

func mcpConnect(mcpServer string) *mcp.ClientSession {
	ctx := context.Background()
	client := mcp.NewClient(&mcp.Implementation{Name: "mcp-client", Version: "v1.0.0"}, nil)
	transport := &mcp.CommandTransport{Command: exec.Command(mcpServer)}
	sess, err := client.Connect(ctx, transport, nil)
	if err != nil {
		log.Fatal(err)
	}

	return sess
}

func listTools(sess *mcp.ClientSession) []api.Tool {
	listToolsResult, err := sess.ListTools(context.Background(), &mcp.ListToolsParams{})
	if err != nil {
		log.Fatal("list tools: ", err)
	}

	tools, _ := olamtl.FromMCP(listToolsResult.Tools)
	return tools
}

func listOllam(tools []api.Tool) {
	for _, t := range tools {
		f := t.Function
		fmt.Printf("Name: %s, Description: %s\n", f.Name, f.Description)
	}
}

func newChat(host string, model string, messages []api.Message,
	tools []api.Tool, sess *mcp.ClientSession) *myChat {

	c := myChat{}
	c.cl = ollam.Client(host)
	c.msgs = messages
	c.model = model
	c.tools = tools
	c.sess = sess // HL
	return &c
}

type myChat struct {
	cl    *api.Client
	msgs  []api.Message
	model string
	tools []api.Tool
	sess  *mcp.ClientSession // HL
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
			output := mcpCallTool(m.sess, // HL
				&mcp.CallToolParams{Name: fn.Name, Arguments: fn.Arguments}) // HL
			m.msgs = append(m.msgs, api.Message{
				Role:    "tool",
				Content: output,
			})
		}
		return nil
	}

	ollam.ChatTools(m.cl, m.model, m.tools, m.msgs, responseFunction)

}

func mcpCallTool(session *mcp.ClientSession, params *mcp.CallToolParams) string {
	ctx := context.Background()
	res, err := session.CallTool(ctx, params) // HL
	if err != nil {
		return fmt.Sprintf("mcpCallTool failed: %v", err)
	}
	if res.IsError {
		log.Fatal("tool failed")
	}
	s := ""
	for _, c := range res.Content {
		s += c.(*mcp.TextContent).Text
	}
	log.Printf("\tTool: %s called with args: %v. resp: %s", params.Name, params.Arguments, s)
	return s
}

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
