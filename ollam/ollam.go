package ollam

import (
	"context"
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

func chatReq(model string, messages []api.Message) *api.ChatRequest {
	think := dflt.EnvString("THINK", "aloud")

	var tv any = false
	switch think {
	case "true":
		tv = true
	case "aloud":
		tv = nil
	default:
		tv = false
	}

	return &api.ChatRequest{
		Model:    model,
		Messages: messages,
		Options:  map[string]any{"temperature": 0.1},
		Think:    &api.ThinkValue{tv},
	}
}
