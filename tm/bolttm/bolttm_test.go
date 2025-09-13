package bolttm

import (
	"context"
	"testing"

	"github.com/boltdb/bolt"
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

		testMessageRetrieval(t, msg)

		testSingleMessageIDStoredForConversation(t, msg)
	})
}

func testMessageRetrieval(t *testing.T, msg *spec.Message) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(MessageBucket))
		dat := b.Get([]byte(msg.MessageID))
		if dat == nil {
			t.Error("expected to be able to retrieve message")
		}
		//t.Errorf("%s\n", dat)
		if len(dat) == 0 {
			t.Error("expected message, got empty response")
		}
		return nil
	})
}

func testSingleMessageIDStoredForConversation(t *testing.T, msg *spec.Message) {
	db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(ConversationBucket))
		cb := b.Bucket([]byte(*msg.ContextID))
		c := cb.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			if string(v) != msg.MessageID {
				t.Error("incorrect message ID retrieved: ", msg.MessageID)
			}
		}

		return nil
	})
}
