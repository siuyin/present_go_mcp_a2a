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

	client := ollam.Client(host)

	messages := []api.Message{
		{Role: "system", Content: sys},
		{Role: "user", Content: prompt},
	}

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

	ollam.Chat(client, model, messages, responseFunction)
	fmt.Println()
}

func createThinkFile(name string) *os.File {
	f, err := os.Create(name)
	if err != nil {
		log.Fatal("createThinkFile: ", err)
	}

	return f
}
