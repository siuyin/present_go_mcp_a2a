package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/siuyin/a2atry/msg"
	"github.com/siuyin/a2atry/ptr"
	"github.com/siuyin/dflt"
	"github.com/siuyin/present_go_mcp_a2a/db"
	"trpc.group/trpc-go/trpc-a2a-go/log"
	spec "trpc.group/trpc-go/trpc-a2a-go/protocol"
	"trpc.group/trpc-go/trpc-a2a-go/server"
	tm "trpc.group/trpc-go/trpc-a2a-go/taskmanager"
)

// inventoryLookupAgent satisfies the tm.MessageProcessor interface.
type inventoryLookupAgent struct{}

func (a *inventoryLookupAgent) ProcessMessage(ctx context.Context, m spec.Message,
	opts tm.ProcessOptions, handler tm.TaskHandler) (*tm.MessageProcessingResult, error) {

	txt := msg.Text(m)
	log.Info("received input: ", txt)

	s := inventoryLookup(txt)
	resp := spec.NewMessage(
		spec.MessageRoleAgent,
		[]spec.Part{spec.NewTextPart(s)},
	)

	log.Info("sending output: ", s)
	return &tm.MessageProcessingResult{Result: &resp}, nil
}

func main() {
	port := dflt.EnvString("PORT", "8080")
	log.Infof("PORT=%s", port)
	log.Infof("curl http://localhost:%s/.well-known/agent.json for agent card", port)

	svr, err := server.NewA2AServer(myAgentCard(port), myTaskManager(&inventoryLookupAgent{})) // HL
	if err != nil {
		log.Fatal("new server:", err)
	}

	log.Fatal(svr.Start(":" + port))
}

func myAgentCard(port string) server.AgentCard {
	return server.AgentCard{
		Name:        "inventoryLookupAgent",
		Description: "Looks up inventory given a product name, returns product ID, price in USD and stock level.",
		URL:         fmt.Sprintf("http://localhost:%s/", port),
		Version:     "1.0.0",
		Capabilities: server.AgentCapabilities{
			Streaming:              ptr.Bool(true),
			PushNotifications:      ptr.Bool(false),
			StateTransitionHistory: ptr.Bool(true),
		},
		Skills: []server.AgentSkill{{
			ID:   "lookup_inventory",
			Name: "looks up inventory",
			Description: ptr.String(`Looks up inventory given a product name, 
returns product ID, price in USD and stock level.`),
		}},
	}
}

func myTaskManager(mp tm.MessageProcessor) tm.TaskManager {
	mgr, err := tm.NewMemoryTaskManager(mp)
	if err != nil {
		log.Fatal("new task manager: ", err)
	}

	return mgr
}

func inventoryLookup(txt string) string {
	ltxt := strings.ToLower(txt)
	prodList := db.List()
	for _, prd := range prodList {
		if strings.Contains(ltxt, prd) {
			return db.Get(prd)
		}
	}
	return fmt.Sprintf("With regard to %q. We do not stock the product.", txt)
}
