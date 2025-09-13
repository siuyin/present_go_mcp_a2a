// Package bolttm implements the trpc-go's TaskManager interface
package bolttm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/boltdb/bolt"
	"github.com/nats-io/nuid"
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
	log.Printf("BOLTDB=%s", path)
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

	return &BoltDBTaskManager{}, nil
}

func (b *BoltDBTaskManager) OnSendMessage(ctx context.Context, r spec.SendMessageParams) (*spec.MessageResult, error) {
	msg := spec.Message{
		MessageID: nuid.Next(),
	}
	return &spec.MessageResult{Result: &msg}, nil
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
	bkt := tx.Bucket([]byte(MessageBucket))
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
