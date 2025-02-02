package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/wonksing/go-tutorials/cache/valkey/distlock"
	"github.com/wonksing/go-tutorials/cache/valkey/errorz"
	"github.com/wonksing/go-tutorials/cache/valkey/reserve/adapter"
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
	r, err := u.a.Zrange(ctx, userId)
	if err != nil {
		if !errors.Is(err, errorz.ErrResourceNotFound) {
			return "", err
		}
		// load
	} else {
		if r != "" {
			return r, nil
		}

		// check
		err = u.a.Exists(ctx, userId)
		if err != nil {
			if !errors.Is(err, errorz.ErrResourceNotFound) {
				return "", err
			}
			// load
		} else {
			return "", nil
		}
	}

	// load from storage
	return u.load(ctx, userId)
}

func (u *ApppushReserve) SetReserve(ctx context.Context, userId uint64, liveId uint64) (string, error) {
	var err error
	err = u.a.Exists(ctx, userId)
	if errors.Is(err, errorz.ErrResourceNotFound) {
		_, err = u.load(ctx, userId)
		if err != nil {
			return "", err
		}
	} else if err != nil {
		return "", fmt.Errorf("set exists reserve: %v", err)
	}

	return u.a.CasZadd(ctx, userId, liveId)
}

func (u *ApppushReserve) load(ctx context.Context, userId uint64) (string, error) {
	// load from storage
	fmt.Println("load from storage")

	lockKey := fmt.Sprintf("reserve:%d", userId)
	err := u.l.Lock(ctx, lockKey, "get-reserve-lock")
	if err != nil {
		v, err := u.a.Zrange(ctx, userId)
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
	// return "", nil

	// read from storage
	var someLiveId uint64 = 90203
	// cache the result from storage
	res, err := u.a.CasZadd(ctx, userId, someLiveId)
	if err != nil {
		return "", err
	}
	return res, nil
}
