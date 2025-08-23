package main

import (
	"fmt"
	"gomcp/ollam"
	"log"
	"os"

	"github.com/ollama/ollama/api"
	"github.com/siuyin/dflt"
)

func main() {
	host := dflt.EnvString("OLLAMA_HOST", "http://localhost:11434")
	model := dflt.EnvString("MODEL", "gemma3:1b")
	prompt := dflt.EnvString("PROMPT", "What is the meaning of life?")
	sys := dflt.EnvString("SYS", "Provide concise responses.")
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
