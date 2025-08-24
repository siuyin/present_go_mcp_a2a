package ollam

import (
	"context"
	"fmt"
	"log"

	"github.com/ollama/ollama/api"
	"github.com/siuyin/dflt"
)

// Client creates an ollama API client from environment variable OLLAMA_HOST.
func Client(host string) *api.Client {
	client, err := api.ClientFromEnvironment()
	if err != nil {
		log.Fatal(err)
	}
	return client
}

// Chat requests a ollama chat completion with chat history, messages, and streaming response via respFunc.
func Chat(client *api.Client, model string, messages []api.Message, respFunc func(r api.ChatResponse) error) {
	req := chatReq(model, messages)

	ctx := context.Background()
	if err := client.Chat(ctx, req, respFunc); err != nil {
		log.Fatal("chat: ", err)
	}

}

// ChatTools requests a ollama chat completion with tools, chat history, messages, and streaming response via respFunc.
func ChatTools(client *api.Client, model string, tools []api.Tool, messages []api.Message, respFunc func(r api.ChatResponse) error) {
	req := chatReqTools(model, messages, tools)

	ctx := context.Background()
	if err := client.Chat(ctx, req, respFunc); err != nil {
		log.Fatal("chat: ", err)
	}

}

func chatReq(model string, messages []api.Message) *api.ChatRequest {
	think := dflt.EnvString("THINK", "aloud")

	return &api.ChatRequest{
		Model:    model,
		Messages: messages,
		Options:  map[string]any{"temperature": 0.1},
		Think:    thinkValueFor(think),
	}
}

func chatReqTools(model string, messages []api.Message, tools []api.Tool) *api.ChatRequest {
	think := dflt.EnvString("THINK", "false")

	return &api.ChatRequest{
		Model:    model,
		Messages: messages,
		Options:  map[string]any{"temperature": 0.1},
		Think:    thinkValueFor(think),
		Tools:    tools,
	}
}

func thinkValueFor(think string) *api.ThinkValue {
	var tv any = false
	switch think {
	case "true":
		tv = true
	case "aloud":
		tv = nil
	default:
		tv = false
	}

	return &api.ThinkValue{tv}
}

// DumpMessages prints out []api.Messsges in a user friendly manner.
func DumpMessages(msgs []api.Message) {
	for _, m := range msgs {
		fmt.Printf("%s: %s: toolCalls: %d\n", m.Role, m.Content, len(m.ToolCalls))
	}
}
