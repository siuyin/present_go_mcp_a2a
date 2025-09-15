// Package bolttm implements the trpc-go's TaskManager interface
package bolttm

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"testing"

	"github.com/boltdb/bolt"
	"github.com/siuyin/a2atry/msg"
	"github.com/siuyin/dflt"
	spec "trpc.group/trpc-go/trpc-a2a-go/protocol"
	tm "trpc.group/trpc-go/trpc-a2a-go/taskmanager"
)

var (
	db *bolt.DB
)

func init() {
	initDB()
	initBuckets()
}

func initDB() {
	var err error
	path := dflt.EnvString("BOLTDB", "/tmp/bolttm.db")
	if !testing.Testing() {
		log.Printf("BOLTDB=%s", path)
	}
	db, err = bolt.Open(path, 0600, nil)
	if err != nil {
		log.Fatal("bolt open: ", err)
	}
}

func initBuckets() {
	buckets := []string{MessageBucket, ConversationBucket, TaskBucket, SubscriberBucket, PushNotificationBucket}
	for _, b := range buckets {
		err := db.Update(func(tx *bolt.Tx) error {
			_, err := tx.CreateBucketIfNotExists([]byte(b))
			return err
		})
		if err != nil {
			log.Fatal("init bucket: ", b, ": ", err)
		}
	}
}

const (
	// Messages stores messages indexed by MessageID.
	MessageBucket = "msg"
	// Conversations stores the conversation message history, indexed by ContextID.
	ConversationBucket = "conv"
	// Tasks stores tasks, keyed by taskID
	TaskBucket = "task"
	// Subscriberes keyed by taskID
	SubscriberBucket = "sub"
	// PushNotifications stores push configs keyed by taskID
	PushNotificationBucket = "push"
)

type BoltDBTaskManager struct {
	// Processor is the user defined message processor.
	Processor tm.MessageProcessor
}

func NewBoltDBTaskManager(proc tm.MessageProcessor, opts ...BoltDBTaskManagerOption) (*BoltDBTaskManager, error) {
	if proc == nil {
		return nil, fmt.Errorf("processor cannot be nil")
	}

	return &BoltDBTaskManager{Processor: proc}, nil
}

func (b *BoltDBTaskManager) OnSendMessage(ctx context.Context, r spec.SendMessageParams) (*spec.MessageResult, error) {
	ret := &spec.MessageResult{}
	m := r.Message
	log.Printf("%#v", r.Message)
	log.Printf("%q", msg.Text(r.Message))
	b.setMessageIDIfEmpty(&m)
	b.setContextIDIfEmpty(&m)
	if err := b.appendConversation(&m); err != nil {
		return ret, err
	}
	if err := b.storeMessage(&m); err != nil {
		return ret, err
	}

	//FIXME: options should be configured from request, r
	options := tm.ProcessOptions{}
	//FIXME: handler should be specified. Currently not used.
	res, err := b.Processor.ProcessMessage(ctx, m, options, nil)
	if err != nil {
		return ret, err
	}

	rm, ok := res.Result.(*spec.Message)
	if !ok {
		return ret, err
	}
	rm.Role = spec.MessageRoleAgent

	b.setMessageIDIfEmpty(rm)
	b.setContextIDIfEmpty(rm)
	if err := b.appendConversation(rm); err != nil {
		return ret, err
	}
	if err := b.storeMessage(rm); err != nil {
		return ret, err
	}
	return &spec.MessageResult{Result: rm}, nil
}

func (b *BoltDBTaskManager) OnSendMessageStream(ctx context.Context, r spec.SendMessageParams) (<-chan spec.StreamingMessageEvent, error) {
	c := make(chan spec.StreamingMessageEvent)
	return c, nil
}

func (b *BoltDBTaskManager) OnGetTask(ctx context.Context, tq spec.TaskQueryParams) (spec.Task, error) {
	return spec.Task{}, nil
}

func (b *BoltDBTaskManager) OnCancelTask(ctx context.Context, tid spec.TaskIDParams) (spec.Task, error) {
	return spec.Task{}, nil
}

func (b *BoltDBTaskManager) OnPushNotificationSet(ctx context.Context, tid spec.TaskIDParams) (spec.TaskPushNotificationConfig, error) {
	return spec.TaskPushNotificationConfig{}, nil
}

func (b *BoltDBTaskManager) OnPushNotificationGet(ctx context.Context, tid spec.TaskIDParams) (spec.TaskPushNotificationConfig, error) {
	return spec.TaskPushNotificationConfig{}, nil
}

func (b *BoltDBTaskManager) OnResubscribe(ctx context.Context, tid spec.TaskIDParams) (<-chan spec.StreamingMessageEvent, error) {
	c := make(chan spec.StreamingMessageEvent)
	return c, nil
}

func (b *BoltDBTaskManager) setMessageIDIfEmpty(msg *spec.Message) {
	if msg.MessageID == "" {
		msg.MessageID = spec.GenerateMessageID()
	}
}
func (b *BoltDBTaskManager) setContextIDIfEmpty(msg *spec.Message) {
	if msg.ContextID == nil || *msg.ContextID == "" {
		cid := spec.GenerateContextID()
		msg.ContextID = &cid
	}
}
func (b *BoltDBTaskManager) storeMessage(msg *spec.Message) error {
	return db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(MessageBucket))
		dat, err := json.Marshal(msg)
		if err != nil {
			return err
		}
		if err := bkt.Put([]byte(msg.MessageID), dat); err != nil {
			return err
		}
		return nil
	})
}
func (b *BoltDBTaskManager) appendConversation(msg *spec.Message) error {
	if msg.ContextID == nil {
		return nil
	}

	cid := *msg.ContextID
	return db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket([]byte(ConversationBucket))
		cb, err := bkt.CreateBucketIfNotExists([]byte(cid))
		if err != nil {
			return fmt.Errorf("create bucket: %s: %v", cid, err)
		}

		seq, err := cb.NextSequence()
		if err != nil {
			return fmt.Errorf("nextSequence: %v", err)
		}

		key := itob(seq)
		if err := cb.Put(key, []byte(msg.MessageID)); err != nil {
			return fmt.Errorf("conversation put: %v", err)
		}

		return nil
	})
}

type BoltDBTaskManagerOption struct{}
type BoltDBCancellableTask struct{}
type BoltDBTaskSubscriber struct{}

type EchoProc struct{}

func (e *EchoProc) ProcessMessage(ctx context.Context, m spec.Message, opts tm.ProcessOptions, handler tm.TaskHandler) (*tm.MessageProcessingResult, error) {
	res := &spec.Message{
		Role:  spec.MessageRoleAgent,
		Parts: []spec.Part{spec.NewTextPart("Echo: " + msg.Text(m))},
	}
	return &tm.MessageProcessingResult{Result: res}, nil
}

func msgFrom(s string) spec.Message {
	return spec.Message{
		Role:  spec.MessageRoleUser,
		Parts: []spec.Part{spec.NewTextPart(s)},
	}
}

func itob(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
func getText(message spec.Message) string {
	for _, part := range message.Parts {
		if textPart, ok := part.(*spec.TextPart); ok {
			return textPart.Text
		}
	}
	return ""
}
