package bolttm

import (
	"context"
	"testing"

	spec "trpc.group/trpc-go/trpc-a2a-go/protocol"
)

func TestNewBoltDBTaskManager(t *testing.T) {
	mgr, err := NewBoltDBTaskManager(&EchoProc{})
	if err != nil {
		t.Error(err)
	}

	if mgr == nil {
		t.Error("nil task manager returned")
	}
}

func TestTaskManager(t *testing.T) {
	mgr, _ := NewBoltDBTaskManager(&EchoProc{})
	ctx := context.Background()
	t.Run("OnSendMessage", func(t *testing.T) {
		req := spec.SendMessageParams{Message: msgFrom("Hello")}
		res, err := mgr.OnSendMessage(ctx, req)
		if err != nil {
			t.Error(err)
		}
		if res == nil {
			t.Error(err)
		}

		msg, ok := res.Result.(*spec.Message)
		if !ok {
			t.Error("expected a Message, got something else")
		}
		if msg.MessageID == "" {
			t.Error("expected message ID, got empty ID")
		}

		// check message is stored

	})
}
