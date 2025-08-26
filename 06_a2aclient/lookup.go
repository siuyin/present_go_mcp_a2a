package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/siuyin/a2atry/ptr"
	"github.com/siuyin/dflt"
	"trpc.group/trpc-go/trpc-a2a-go/client"
	spec "trpc.group/trpc-go/trpc-a2a-go/protocol"
)

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
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	params := formatAsParam(qry)
	res, err := cl.SendMessage(ctx, params)
	if err != nil {
		return fmt.Sprintf("Sorry, the inventory lookup can't be contacted at the moment: %v\n", err)
	}

	resMsg, ok := res.Result.(*spec.Message)
	if !ok {
		return "did not received expected message from inventory lookup agent"
	}

	s := ""
	for _, p := range resMsg.Parts {
		s += p.(*spec.TextPart).Text
	}
	return s
}

func formatAsParam(qry string) spec.SendMessageParams {
	msg := spec.NewMessage(spec.MessageRoleUser, []spec.Part{spec.NewTextPart(qry)})
	params := spec.SendMessageParams{
		Message: msg,
		Configuration: &spec.SendMessageConfiguration{
			Blocking:            ptr.Bool(true), // Non-blocking for streaming, blocking for standard
			AcceptedOutputModes: []string{"text"},
		},
	}
	return params
}
