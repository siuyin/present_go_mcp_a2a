package main

import (
	"context"
	"testing"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// These set of tests work with ./03_mcpserver, with server "myserver".
func TestMCP(t *testing.T) {
	sess := mcpConnect("myserver")
	defer sess.Close()

	ctx := context.Background()
	var (
		lpr *mcp.ListPromptsResult
		gpr *mcp.GetPromptResult
		err error
	)

	t.Run("ListPrompts", func(t *testing.T) {
		lpr, err = sess.ListPrompts(ctx, &mcp.ListPromptsParams{}) // sets test global lpr
		if err != nil {
			t.Error(err)
		}
		if n := len(lpr.Prompts); n == 0 {
			t.Errorf("there should be at least one prompt listed, got %d prompts", n)
		}
	})

	t.Run("GetPrompt", func(t *testing.T) {
		p := lpr.Prompts[0]
		gpr, err = sess.GetPrompt(ctx, &mcp.GetPromptParams{Name: p.Name})
		if err != nil {
			t.Error(err)
		}
		if n := len(gpr.Messages); n == 0 {
			t.Errorf("there should be at least one prompt message, got %d", n)
		}
	})

	t.Run("PromptMessages", func(t *testing.T) {
		if n := len(gpr.Messages); n != 2 {
			t.Errorf("expected 2 messages, got %d", n)
		}

		m0 := gpr.Messages[0]
		if r := m0.Role; r != "system" {
			t.Errorf("expected system role, got %s", r)
		}
		if tc := textContent(m0.Content); tc.Text != "System prompt" {
			t.Errorf("unexpected prompt, got %q", tc.Text)
		}

		m1 := gpr.Messages[1]
		if r := m1.Role; r != "user" {
			t.Errorf("expected user role, got %s", r)
		}
		if tc := textContent(m1.Content); tc.Text != "User prompt" {
			t.Errorf("unexpected prompt, got %q", tc.Text)
		}

	})
}
