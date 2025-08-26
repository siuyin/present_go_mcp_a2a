package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ollama/ollama/api"
	"github.com/siuyin/a2atry/ptr"
	"github.com/siuyin/dflt"
	"github.com/siuyin/present_go_mcp_a2a/ollam"
	"trpc.group/trpc-go/trpc-a2a-go/client"
	"trpc.group/trpc-go/trpc-a2a-go/protocol"
	spec "trpc.group/trpc-go/trpc-a2a-go/protocol"
)

func main() {
	prompt := dflt.EnvString("PROMPT", "Do you have the iPhone 14?")
	res := inventoryLookup(prompt)
	formatResponse(fmt.Sprintf(`user query: %s
	response from inventory lookup: %s
	Please answer the user query in a concise and professional manner`, prompt, res))
}

func inventoryLookup(qry string) string {
	timeout := 20 * time.Second
	cl := newA2AInvLookupClient(timeout)

	return sendMsg(cl, qry, timeout)
}

func newA2AInvLookupClient(timeout time.Duration) *client.A2AClient {
	url := dflt.EnvString("INV_URL", "http://localhost:8080/")
	cl, err := client.NewA2AClient(url, client.WithTimeout(timeout))
	if err != nil {
		log.Fatalf("Failed to create A2A client: %v to connect to agent at: %s", err, url)
	}
	return cl
}

func sendMsg(cl *client.A2AClient, qry string, timeout time.Duration) string {
	msg := spec.NewMessage(spec.MessageRoleUser, []spec.Part{spec.NewTextPart(qry)})
	params := protocol.SendMessageParams{
		Message: msg,
		Configuration: &protocol.SendMessageConfiguration{
			Blocking:            ptr.Bool(true), // Non-blocking for streaming, blocking for standard
			AcceptedOutputModes: []string{"text"},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	res, err := cl.SendMessage(ctx, params)
	if err != nil {
		log.Fatalf("Failed to send message: %v", err)
	}
	resMsg, ok := res.Result.(*spec.Message)
	if !ok {
		log.Fatal("did not received expected message from agent")
	}

	s := ""
	for _, p := range resMsg.Parts {
		s += p.(*spec.TextPart).Text
	}
	return s
}

func formatResponse(prompt string) {
	host := dflt.EnvString("OLLAMA_HOST", "http://localhost:11434")
	model := dflt.EnvString("MODEL", "gemma3:1b")
	sys := dflt.EnvString("SYS", "Strictly use only the data provided by inventory lookup. Response in text format.")
	log.Printf("OLLAMA_HOST=%s MODEL=%s SYS=%q PROMPT=%q ", host, model, sys, prompt)

	messages := []api.Message{
		{Role: "system", Content: sys},
		{Role: "user", Content: prompt},
	}

	chat := newChat(host, model, messages)
	chat.complete()
}

func newChat(host string, model string, messages []api.Message) *myChat {
	c := myChat{}
	c.cl = ollam.Client(host)
	c.msgs = messages
	c.model = model
	return &c
}

type myChat struct {
	cl    *api.Client
	msgs  []api.Message
	model string
}

func (m *myChat) complete() {
	f := createThinkFile("/tmp/j")
	defer f.Close()

	responseFunction := func(r api.ChatResponse) error {
		if r.Message.Thinking != "" {
			fmt.Fprint(f, r.Message.Thinking)
			return nil
		}

		fmt.Print(r.Message.Content)
		return nil
	}

	ollam.Chat(m.cl, m.model, m.msgs, responseFunction)
	fmt.Println()

}

func createThinkFile(name string) *os.File {
	f, err := os.Create(name)
	if err != nil {
		log.Fatal("createThinkFile: ", err)
	}

	return f
}
