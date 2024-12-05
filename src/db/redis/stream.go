package redis

import (
	"context"
	"crypto/rand"
	"fmt"
	"sync"
	"time"

	"github.com/laper32/goose/logging"

	"github.com/redis/go-redis/v9"
)

const (
	maxLen            = 100
	idLen             = 6
	comsumerSpawnTime = 1 * time.Second
	retryTime         = 5 * time.Second
	groupExistErr     = "BUSYGROUP Consumer Group name already exists"
	canceledErr       = "context canceled"
)

type Stream struct {
	client *redis.Client
	group  string
	ctx    context.Context
	closer func()
	wg     sync.WaitGroup
}

type StreamCallback func(context.Context, string)

var stream *Stream

func NewStream(client *redis.Client, service string) *Stream {
	stream = &Stream{
		client: client,
		group:  service,
		wg:     sync.WaitGroup{},
	}
	stream.ctx, stream.closer = context.WithCancel(context.Background())
	return stream
}

func (s *Stream) Publish(ctx context.Context, topic, msg string) (err error) {
	return s.client.XAdd(ctx, &redis.XAddArgs{
		Stream: topic,
		MaxLen: maxLen,
		Values: []string{"msg", msg},
	}).Err()
}

func (s *Stream) Subscribe(topic string, callback StreamCallback) (err error) {
	// Create group
	err = s.client.XGroupCreateMkStream(s.ctx, topic, s.group, "$").Err()
	if err != nil {
		if err.Error() != groupExistErr {
			return
		}
		err = nil
	}

	// Create comsumer
	comsumer := ""
	for i := 0; i < idLen; i++ {
		comsumer = getRandomID()
		res, err := s.client.XGroupCreateConsumer(s.ctx, topic, s.group, comsumer).Result()

		// Create successfully
		if res == 1 {
			break
		}

		if err != nil {
			logging.Error(err)
		} else {
			// Not things wrong, just keep go on!
			i--
		}
		time.Sleep(comsumerSpawnTime)
	}

	// Loop message from topic
	// Add wait group for waiting block cancel and delete consumer
	s.wg.Add(1)
	go func() {
		for {
			// Read message
			res, err := s.client.XReadGroup(s.ctx, &redis.XReadGroupArgs{
				Group:    s.group,
				Consumer: comsumer,
				Streams:  []string{topic, ">"},
				Count:    1,
				Block:    0,
				NoAck:    true,
			}).Result()

			if err != nil {
				// Go to cancel
				if err.Error() == canceledErr {
					_ = s.client.XGroupDelConsumer(context.Background(), topic, s.group, comsumer)
					s.wg.Done()
					return
				}
				logging.Error("redis XReadGroup:", err)
				time.Sleep(retryTime)
			} else {
				for _, xs := range res {
					for _, ms := range xs.Messages {
						for _, m := range ms.Values {
							v, _ := m.(string)
							s.wg.Add(1)
							go func() {
								callback(s.ctx, v)
								s.wg.Done()
							}()
						}
					}
				}
			}
		}
	}()

	return
}

func (s *Stream) Close() {
	s.closer()
	s.wg.Wait()
}

func getRandomID() string {
	randBytes := make([]byte, idLen/2)
	_, _ = rand.Read(randBytes)
	return fmt.Sprintf("%x", randBytes)
}
