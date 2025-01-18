package distlock

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valkey-io/valkey-go"
)

type DistLockValkeyV3 struct {
	client        valkey.Client
	keyPrefix     string
	channelPrefix string
	timeout       time.Duration

	defaultExpiry time.Duration
	retry         int8

	lockChans     map[string]chan string
	lockChansLock sync.RWMutex

	subChan      chan string
	cancelSubCtx context.CancelFunc
	subWg        *sync.WaitGroup

	lockMeasure *lockMeasure
}

type lockMeasure struct {
	Locked   atomic.Int64 `json:"locked"`
	Unlocked atomic.Int64 `json:"unlocked"`

	ChanCreated atomic.Int64 `json:"chan_created"`
	ChanDeleted atomic.Int64 `json:"chan_deleted"`
}

func (l *lockMeasure) String() string {
	return fmt.Sprintf("locked: %d, unlocked: %d, chan_created: %d, chan_deleted: %d",
		l.Locked.Load(), l.Unlocked.Load(), l.ChanCreated.Load(), l.ChanDeleted.Load())
}

func (l *lockMeasure) IncLock() {
	l.Locked.Add(1)
}
func (l *lockMeasure) IncUnlock() {
	l.Unlocked.Add(1)
}
func (l *lockMeasure) IncChanCreated() {
	l.ChanCreated.Add(1)
}
func (l *lockMeasure) IncChanDeleted() {
	l.ChanDeleted.Add(1)
}

func NewDistLockValkeyV3(ctx context.Context, client valkey.Client, keyPrefix, channelPrefix string, timeout time.Duration, retry int8) *DistLockValkeyV3 {
	lockChans := make(map[string]chan string)
	subChan := make(chan string, 1)
	subCtx, cancelSubCtx := context.WithCancel(ctx)

	l := &DistLockValkeyV3{
		client:        client,
		keyPrefix:     keyPrefix,
		channelPrefix: channelPrefix,
		timeout:       timeout,
		// defaultExpiry: timeout + (80 * time.Millisecond),
		defaultExpiry: 3 * time.Minute,
		retry:         retry,

		lockChans:    lockChans,
		subChan:      subChan,
		cancelSubCtx: cancelSubCtx,
		subWg:        &sync.WaitGroup{},

		lockMeasure: &lockMeasure{},
	}

	// l.subWg.Add(1)
	// go l.readMsg()
	l.subWg.Add(1)
	go l.startSubscribe(subCtx)

	return l
}

func (d *DistLockValkeyV3) addLockChans(key string) chan string {
	d.lockChansLock.Lock()
	defer d.lockChansLock.Unlock()

	if c, ok := d.lockChans[key]; ok {
		return c
	}

	ch := make(chan string)
	d.lockChans[key] = ch
	d.lockMeasure.IncChanCreated()
	return ch
}

func (d *DistLockValkeyV3) releaseLockChans(key string) {
	d.lockChansLock.Lock()
	defer d.lockChansLock.Unlock()

	if c, ok := d.lockChans[key]; ok {
		// close(c)
		// delete(d.lockChans, key)
		select {
		case c <- key:
		default:
			if len(c) == 0 {
				close(c)
				delete(d.lockChans, key)
				d.lockMeasure.IncChanDeleted()
				return
			}
			log.Println("channel is not empty")
			return
		}
	}

	// fmt.Println("channel not found")
}

func (d *DistLockValkeyV3) Lock(ctx context.Context, key string, value string) error {
	return d.LockWithExpiry(ctx, key, value, d.defaultExpiry)
}

func (d *DistLockValkeyV3) LockWithExpiry(ctx context.Context, key string, value string, expiry time.Duration) error {
	if d.retry == 0 {
		return d.lockWithExpiry(ctx, key, value, expiry)
	}

	var err error
	for i := int8(0); i < d.retry; i++ {
		err = d.lockWithExpiry(ctx, key, value, expiry)
		if err == nil {
			return nil
		}
	}
	return AsAcquireLockError("retry limit reached", err)
}

func (d *DistLockValkeyV3) Unlock(ctx context.Context, key string, value string) error {
	c, cancel := d.client.Dedicate()
	defer cancel()

	err := c.Do(ctx,
		c.B().Del().Key(d.keyPrefix+key).Build()).Error()
	if err != nil {
		return err
	}

	// d.addLockChans(d.keyPrefix + key)
	err = c.Do(ctx,
		c.B().Publish().Channel(d.channelPrefix).Message(d.keyPrefix+key).Build()).Error()
	if err != nil {
		return err
	}
	d.lockMeasure.IncUnlock()
	return nil
}

func (d *DistLockValkeyV3) lockWithExpiry(ctx context.Context, key string, value string, expiry time.Duration) error {
	client, cancelClient := d.client.Dedicate()
	defer cancelClient()

	err := client.Do(ctx,
		client.B().Set().Key(d.keyPrefix+key).Value(value).Nx().Ex(expiry).Build()).Error()
	if err == nil {
		d.lockMeasure.IncLock()
		return nil
	}

	ch := d.addLockChans(d.keyPrefix + key)
	// wait
	// fmt.Printf("waiting for lock: %v\n", key)
	select {
	case <-ch:
	case <-time.After(d.timeout):
		return NewAcquireLockError("timeout")
	case <-ctx.Done():
		return AsAcquireLockError("context done: ", ctx.Err())
	}

	err = client.Do(ctx,
		client.B().Set().Key(d.keyPrefix+key).Value(value).Nx().Ex(expiry).Build()).Error()
	if err != nil {
		return AsAcquireLockError("trying to acquire lock", err)
	}
	d.lockMeasure.IncLock()
	return nil
}

func (d *DistLockValkeyV3) startSubscribe(ctx context.Context) error {
	defer d.subWg.Done()

	pubsubClient, cancelPubsubClient := d.client.Dedicate()
	defer func() {
		cancelPubsubClient()
		close(d.subChan)
	}()

	err := pubsubClient.Receive(ctx, pubsubClient.B().Subscribe().Channel(d.channelPrefix).Build(), func(msg valkey.PubSubMessage) {
		// Handle the message. Note that if you want to call another `client.Do()` here, you need to do it in another goroutine or the `client` will be blocked.
		// fmt.Printf("Received message: %v\n", msg)
		d.releaseLockChans(msg.Message)
		// select {
		// case d.subChan <- msg.Message:
		// default:
		// }
	})
	if err != nil && !errors.Is(err, context.Canceled) {
		return NewAcquireLockError("subscribe: " + err.Error())
	}
	return nil
}

// func (d *DistLockValkeyV3) readMsg() {
// 	defer d.subWg.Done()
// 	for msg := range d.subChan {
// 		fmt.Printf("message: %v\n", msg)
// 	}
// }

func (d *DistLockValkeyV3) Close() error {
	d.cancelSubCtx()
	d.subWg.Wait()
	fmt.Println(d.lockMeasure.String())
	return nil
}
