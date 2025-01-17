package distlock

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/valkey-io/valkey-go"
)

type DistLockValkey struct {
	Client        valkey.Client
	KeyPrefix     string
	ChannelPrefix string
	Timeout       time.Duration
}

func (d *DistLockValkey) Lock(ctx context.Context, key string, value string) error {
	return d.LockWithExpiry(ctx, key, value, d.Timeout+(100*time.Millisecond))
}

func (d *DistLockValkey) LockWithExpiry(ctx context.Context, key string, value string, expiry time.Duration) error {
	err := d.Client.Do(ctx, d.Client.B().Set().Key(d.KeyPrefix+key).Value(value).Nx().Ex(expiry).Build()).Error()
	if err == nil {
		return nil
	}

	ctxSub, cancel := context.WithTimeout(context.Background(), d.Timeout)
	defer cancel()

	wait := make(chan error, 1)
	go func() {
		err = d.Client.Receive(ctxSub, d.Client.B().Subscribe().Channel(d.ChannelPrefix+key).Build(), func(msg valkey.PubSubMessage) {
			// Handle the message. Note that if you want to call another `client.Do()` here, you need to do it in another goroutine or the `client` will be blocked.
			// fmt.Printf("Received message: %v\n", msg)
			wait <- nil
		})
		if err != nil {
			if errors.Is(err, context.Canceled) {
				return
			}
			fmt.Printf("err: subscribe: %v\n", err)
			wait <- err
		}
	}()

	select {
	case <-time.After(d.Timeout + (50 * time.Millisecond)):
		cancel()
		return errors.New("acquire lock: timeout")
	case <-ctxSub.Done():
		return errors.New("acquire lock: " + ctxSub.Err().Error())
	case err := <-wait:
		if err != nil {
			return errors.New("acquire lock: " + err.Error())
		}

		err = d.Client.Do(ctx,
			d.Client.B().Set().Key(d.KeyPrefix+key).Value(value).Nx().Ex(expiry).Build()).Error()
		if err != nil {
			return errors.New("acquire lock: " + err.Error())
		}
		return nil
	}
}

func (d *DistLockValkey) Unlock(ctx context.Context, key string, value string) error {
	err := d.Client.Do(ctx,
		d.Client.B().Del().Key(d.KeyPrefix+key).Build()).Error()
	if err != nil {
		return err
	}

	err = d.Client.Do(ctx,
		d.Client.B().Publish().Channel(d.ChannelPrefix+key).Message(value).Build()).Error()
	if err != nil {
		return err
	}
	return nil
}
