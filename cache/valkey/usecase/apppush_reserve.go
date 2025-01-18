package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/wonksing/go-tutorials/cache/valkey/adapter"
	"github.com/wonksing/go-tutorials/cache/valkey/distlock"
	"github.com/wonksing/go-tutorials/cache/valkey/errorz"
)

type ApppushReserve struct {
	l *distlock.DistLockValkeyV2
	a *adapter.ReserveValkey
}

func NewApppushReserve(l *distlock.DistLockValkeyV2, a *adapter.ReserveValkey) *ApppushReserve {
	return &ApppushReserve{
		l: l,
		a: a,
	}
}

func (u *ApppushReserve) GetReserve(ctx context.Context, userId uint64) (string, error) {
	r, err := u.a.ZGetReserve(ctx, userId)
	if err == nil {
		// return immediately if found in cache
		return r, nil
	}

	if !errors.Is(err, errorz.ErrResourceNotFound) {
		return "", err
	}

	// load from storage
	return u.load(ctx, userId)
}

func (u *ApppushReserve) SetReserve(ctx context.Context, userId uint64, liveId uint64) (string, error) {
	var err error
	_, err = u.GetReserve(ctx, userId)
	if err != nil {
		return "", fmt.Errorf("get reserve: %v", err)
	}

	return u.a.ZSetReserve(ctx, userId, liveId)
}

func (u *ApppushReserve) load(ctx context.Context, userId uint64) (string, error) {
	// load from storage
	fmt.Println("load from storage")

	lockKey := fmt.Sprintf("reserve:%d", userId)
	err := u.l.Lock(ctx, lockKey, "get-reserve-lock")
	if err != nil {
		v, err := u.a.ZGetReserve(ctx, userId)
		if err != nil {
			if !errors.Is(err, errorz.ErrResourceNotFound) {
				return "", err
			}
		}
		return v, nil
	}
	defer u.l.Unlock(ctx, lockKey, "get-reserve-unlock")

	// read from storage
	// return "", nil if no data found from storage
	return "", nil

	/*
		// read from storage
		var someLiveId uint64 = 90203
		// cache the result from storage
		res, err := u.a.ZSetReserve(ctx, userId, someLiveId)
		if err != nil {
			return "", err
		}
		return res, nil
	*/
}
