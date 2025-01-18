package usecase

import (
	"context"
	"errors"
	"fmt"

	"github.com/wonksing/go-tutorials/cache/valkey/distlock"
	"github.com/wonksing/go-tutorials/cache/valkey/errorz"
	"github.com/wonksing/go-tutorials/cache/valkey/reserve/adapter"
)

type ApppushReserveV2 struct {
	loadLock   *distlock.DistLockValkeyV3
	setLock    *distlock.DistLockValkeyV3
	a          *adapter.ReserveValkey
	setCnt     int64
	setFailCnt int64
}

func NewApppushReserveV2(loadLock, setLock *distlock.DistLockValkeyV3, a *adapter.ReserveValkey) *ApppushReserveV2 {
	return &ApppushReserveV2{
		loadLock: loadLock,
		setLock:  setLock,
		a:        a,
	}
}

func (u *ApppushReserveV2) GetReserve(ctx context.Context, userId uint64) (string, error) {
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

func (u *ApppushReserveV2) SetReserve(ctx context.Context, userId uint64, liveId uint64) (string, error) {
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

	lockKey := fmt.Sprintf("reserve:%d", userId)
	err = u.setLock.Lock(ctx, lockKey, "set-reserve-lock")
	if err != nil {
		return "", err
	}
	defer u.setLock.Unlock(ctx, lockKey, "set-reserve-unlock")

	_, err = u.a.Zadd(ctx, userId, liveId)
	if err != nil {
		u.setFailCnt++
		return "", err
	}
	u.setCnt++
	// fmt.Printf("applied: %v\n", applied)
	return u.GetReserve(ctx, userId)
}

func (u *ApppushReserveV2) load(ctx context.Context, userId uint64) (string, error) {

	lockKey := fmt.Sprintf("reserve:%d", userId)
	err := u.loadLock.Lock(ctx, lockKey, "get-reserve-lock")
	if err != nil {
		v, err := u.a.Zrange(ctx, userId)
		if err != nil {
			if !errors.Is(err, errorz.ErrResourceNotFound) {
				return "", err
			}
		}
		return v, nil
	}
	defer u.loadLock.Unlock(ctx, lockKey, "get-reserve-unlock")

	// load from storage
	fmt.Println("load from storage")

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
func (u *ApppushReserveV2) SetCnt() int64 {
	return u.setCnt
}
func (u *ApppushReserveV2) SetFailCnt() int64 {
	return u.setFailCnt
}
