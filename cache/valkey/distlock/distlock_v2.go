package distlock

import (
	"context"
	"time"

	"github.com/valkey-io/valkey-go"
)

type DistLockValkeyV2 struct {
	client        valkey.Client
	keyPrefix     string
	channelPrefix string
	timeout       time.Duration

	defaultExpiry time.Duration
	retry         int8
}

func NewDistLockValkeyV2(client valkey.Client, keyPrefix, channelPrefix string, timeout time.Duration, retry int8) *DistLockValkeyV2 {
	return &DistLockValkeyV2{
		client:        client,
		keyPrefix:     keyPrefix,
		channelPrefix: channelPrefix,
		timeout:       timeout,
		// defaultExpiry: timeout + (80 * time.Millisecond),
		defaultExpiry: 3 * time.Minute,
		retry:         retry,
	}
}

func (d *DistLockValkeyV2) Lock(ctx context.Context, key string, value string) error {
	return d.LockWithExpiry(ctx, key, value, d.defaultExpiry)
}

func (d *DistLockValkeyV2) LockWithExpiry(ctx context.Context, key string, value string, expiry time.Duration) error {
	if d.retry == 0 {
		return d.lockWithExpiry(ctx, key, value, expiry)
	}

	for i := int8(0); i < d.retry; i++ {
		err := d.lockWithExpiry(ctx, key, value, expiry)
		if err == nil {
			return nil
		}
	}
	return NewAcquireLockError("retry limit reached")
}

func (d *DistLockValkeyV2) Unlock(ctx context.Context, key string, value string) error {
	c, cancel := d.client.Dedicate()
	defer cancel()

	err := c.Do(ctx,
		c.B().Del().Key(d.keyPrefix+key).Build()).Error()
	if err != nil {
		return err
	}

	err = c.Do(ctx,
		c.B().Publish().Channel(d.channelPrefix+key).Message(value).Build()).Error()
	if err != nil {
		return err
	}
	return nil
}

func (d *DistLockValkeyV2) lockWithExpiry(ctx context.Context, key string, value string, expiry time.Duration) error {
	client, cancelClient := d.client.Dedicate()
	defer cancelClient()

	err := client.Do(ctx,
		client.B().Set().Key(d.keyPrefix+key).Value(value).Nx().Ex(expiry).Build()).Error()
	if err == nil {
		// fmt.Println("lock in first attempt")
		return nil
	}

	pubsubClient, cancelPubsubClient := d.client.Dedicate()
	defer cancelPubsubClient()

	ctxSub, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	wait := make(chan error, 1)
	w := pubsubClient.SetPubSubHooks(valkey.PubSubHooks{
		OnMessage: func(msg valkey.PubSubMessage) {
			// Handle the message. Note that if you want to call another `client.Do()` here, you need to do it in another goroutine or the `client` will be blocked.
			// fmt.Printf("message: %v\n", msg)
			wait <- nil
		},
	})
	err = pubsubClient.Do(ctxSub, pubsubClient.B().Subscribe().Channel(d.channelPrefix+key).Build()).Error()
	if err != nil {
		return NewAcquireLockError("subscribe: " + err.Error())
	}

	select {
	case <-ctxSub.Done():
		return NewAcquireLockError("subscribe done: " + ctxSub.Err().Error())
	case err := <-w:
		if err != nil {
			return NewAcquireLockError("disconnected subscriber: " + err.Error())
		}
		return NewAcquireLockError("disconnected subscriber: unknown error")
	case <-wait:
		err = client.Do(ctx,
			client.B().Set().Key(d.keyPrefix+key).Value(value).Nx().Ex(expiry).Build()).Error()
		if err != nil {
			return NewAcquireLockError(err.Error())
		}
		// fmt.Println("lock after second attempt")
		return nil
	}
}
