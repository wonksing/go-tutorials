package adapter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/valkey-io/valkey-go"
	"github.com/wonksing/go-tutorials/cache/valkey/errorz"
)

type ReserveValkey struct {
	client    valkey.Client
	retry     int8
	keyPrefix string
}

func NewReserveValkey(client valkey.Client, keyPrefix string, retry int8) *ReserveValkey {
	return &ReserveValkey{
		client:    client,
		keyPrefix: strings.TrimSuffix(keyPrefix, ":") + ":",
		retry:     retry,
	}
}

func (a *ReserveValkey) Key(userId uint64) string {
	return fmt.Sprintf("%s%d", a.keyPrefix, userId)
}

func (r *ReserveValkey) Get(ctx context.Context, userId uint64) (string, error) {
	c, cancel := r.client.Dedicate()
	defer cancel()

	res := c.Do(ctx, c.B().Get().Key(fmt.Sprintf("%d", userId)).Build())
	if res.Error() != nil {
		if valkey.IsValkeyNil(res.Error()) {
			return "", errorz.ErrResourceNotFound
		}
		return "", res.Error()
	}
	return res.ToString()
}

func (r *ReserveValkey) Set(ctx context.Context, userId uint64, liveId uint64) (string, error) {
	c, cancel := r.client.Dedicate()
	defer cancel()

	var err error

	key := fmt.Sprintf("%d", userId)
	if err = c.Do(ctx, c.B().Watch().Key(key).Build()).Error(); err != nil {
		return "", err
	}

	res := c.Do(ctx, c.B().Get().Key(key).Build())
	if res.Error() != nil {
		if !valkey.IsValkeyNil(res.Error()) {
			return "", res.Error()
		}
	}

	lives := ""
	lives, err = res.ToString()
	if err != nil {
		return "", err
	}
	if lives != "" {
		lives = fmt.Sprintf("%s,%d", lives, liveId)
	} else {
		lives = fmt.Sprintf("%d", liveId)
	}
	res2 := c.DoMulti(
		ctx,
		c.B().Multi().Build(),
		c.B().Set().Key(key).Value(lives).Build(),
		c.B().Exec().Build(),
	)
	for _, r := range res2 {
		if r.Error() != nil {
			return "", r.Error()
		}
	}

	return lives, nil
}

func (a *ReserveValkey) Zrange(ctx context.Context, userId uint64) (string, error) {
	c, cancel := a.client.Dedicate()
	defer cancel()

	key := a.Key(userId)

	// returns empty slice if the key does not exist
	res, err := c.Do(ctx, c.B().Zrange().Key(key).Min("0").Max("-1").Build()).AsStrSlice()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return "", errorz.ErrResourceNotFound
		}
		return "", err
	}

	// DO CHECK WITH 'EXISTS' COMMAND
	// if len(res) == 0 {
	// 	return "", errorz.ErrResourceNotFound
	// }

	return strings.Join(res, ","), nil
}

func (a *ReserveValkey) Zadd(ctx context.Context, userId uint64, liveId uint64) (int64, error) {
	c, cancel := a.client.Dedicate()
	defer cancel()

	key := a.Key(userId)
	res, err := c.Do(ctx, c.B().Zadd().Key(key).ScoreMember().ScoreMember(float64(liveId), fmt.Sprintf("%d", liveId)).Build()).AsInt64()
	if err != nil {
		return 0, err
	}

	return res, nil
}

func (a *ReserveValkey) CasZadd(ctx context.Context, userId uint64, liveId uint64) (string, error) {
	if a.retry == 0 {
		return a.casZadd(ctx, userId, liveId)
	}

	var res string
	var err error
	for i := int8(0); i < a.retry; i++ {
		res, err = a.casZadd(ctx, userId, liveId)
		if err == nil {
			return res, nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	return "", fmt.Errorf("set reserve: retry limit reached: %v", err)
}

func (a *ReserveValkey) Exists(ctx context.Context, userId uint64) error {

	c, cancel := a.client.Dedicate()
	defer cancel()

	key := a.Key(userId)

	// returns empty slice if the key does not exist
	res, err := c.Do(ctx, c.B().Exists().Key(key).Build()).AsInt64()
	if err != nil {
		if valkey.IsValkeyNil(err) {
			return errorz.ErrResourceNotFound
		}
		return err
	}

	if res == 0 {
		return errorz.ErrResourceNotFound
	}

	return nil
}

func (a *ReserveValkey) casZadd(ctx context.Context, userId uint64, liveId uint64) (string, error) {
	c, cancel := a.client.Dedicate()
	defer cancel()

	var err error

	key := a.Key(userId)
	if err = c.Do(ctx, c.B().Watch().Key(key).Build()).Error(); err != nil {
		return "", err
	}

	err = c.Do(ctx, c.B().Zrange().Key(key).Min("0").Max("-1").Build()).Error()
	if err != nil {
		if !valkey.IsValkeyNil(err) {
			return "", err
		}
	}

	res2 := c.DoMulti(
		ctx,
		c.B().Multi().Build(),
		c.B().Zadd().Key(key).ScoreMember().ScoreMember(float64(liveId), fmt.Sprintf("%d", liveId)).Build(),
		c.B().Expire().Key(key).Seconds(600).Nx().Build(),
		c.B().Exec().Build(),
	)
	for i, r := range res2 {
		if valkey.IsValkeyNil(r.Error()) {
			// "valkey nil message" error is returned when the value is beging modified by another client.
			// So, we need to retry.
			return "", errorz.ErrNeedRetry
		}

		if r.Error() != nil {
			return "", fmt.Errorf("zadd(resInd=%d): %v", i, r.Error())
		}
	}

	existingReserves, err := c.Do(ctx, c.B().Zrange().Key(key).Min("0").Max("-1").Build()).AsStrSlice()
	if err != nil {
		return "", err
	}

	return strings.Join(existingReserves, ","), nil
}
