package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/valkey-io/valkey-go"
	"github.com/wonksing/go-tutorials/cache/valkey/distlock"
	"github.com/wonksing/go-tutorials/cache/valkey/reserve/adapter"
	"github.com/wonksing/go-tutorials/cache/valkey/reserve/usecase"
)

var (
	client valkey.Client
)

func init() {
	var err error
	client, err = valkey.NewClient(valkey.ClientOption{InitAddress: []string{"127.0.0.1:6379"}})
	if err != nil {
		panic(err)
	}
	_ = rand.Int()
}

func main() {
	defer client.Close()

	zaddTest()
}

func zaddTest() {

	// var err error
	ctx := context.Background()
	timeout := 15 * time.Second
	// loadLock := distlock.NewDistLockValkeyV2(client, "key-prefix:load:", "chan-prefix:load:", timeout, 0)
	// setLock := distlock.NewDistLockValkeyV2(client, "key-prefix:zadd:", "chan-prefix:zadd:", timeout, 3)
	loadLock := distlock.NewDistLockValkeyV3(ctx, client, "key-prefix:load:", "chan-prefix:load:", timeout, 0)
	defer loadLock.Close()
	setLock := distlock.NewDistLockValkeyV3(ctx, client, "key-prefix:zadd:", "chan-prefix:zadd:", timeout, 3)
	defer setLock.Close()
	a := adapter.NewReserveValkey(client, "reserve:", 10)
	u := usecase.NewApppushReserveV2(loadLock, setLock, a)

	// simple
	// res, err := u.SetReserve(ctx, 1, 10108)
	// if err != nil {
	// 	fmt.Printf("err: set reserve: %v\n", err)
	// 	return
	// }
	// fmt.Println("set reserve:", res)

	// multi
	// userIds := []uint64{1, 1, 1, 1, 1}
	// liveIds := []uint64{10108, 10109, 10110, 10111, 10112, 10112, 10112}
	// var suc atomic.Int64
	// for _, userId := range userIds {
	// 	for _, liveId := range liveIds {
	// 		_, err := u.SetReserve(ctx, userId, liveId)
	// 		if err != nil {
	// 			fmt.Printf("err: set reserve(%v, %v): %v\n", userId, liveId, err)
	// 			return
	// 		}
	// 		// fmt.Printf("set reserve(%d, %v, %v): %v\n", i, userId, liveId, res)
	// 		suc.Add(1)
	// 	}
	// }
	// fmt.Printf("done... %d\n", suc.Load())

	// multi sequential
	// userIds := []uint64{1, 2, 3, 4, 5}
	// liveIds := []uint64{10108, 10109, 10110, 10111, 10112}
	// numTests := 30
	// var wg sync.WaitGroup
	// wg.Add(numTests * len(userIds))
	// var suc, fail atomic.Int64
	// for _, userId := range userIds {
	// 	for i := 0; i < numTests; i++ {
	// 		go func(i int, userId uint64) {
	// 			defer wg.Done()
	// 			liveId := liveIds[rand.Int63n(int64(len(liveIds)))]
	// 			_, err := u.SetReserve(ctx, userId, liveId)
	// 			if err != nil {
	// 				fmt.Printf("err: set reserve(%d, %v, %v): %v\n", i, userId, liveId, err)
	// 				fail.Add(1)
	// 				return
	// 			}
	// 			// fmt.Printf("set reserve(%d, %v, %v): %v\n", i, userId, liveId, res)
	// 			suc.Add(1)
	// 		}(i, userId)
	// 	}
	// }
	// wg.Wait()
	// fmt.Printf("done... success=%d, failure=%d\n", suc.Load(), fail.Load())

	// multi random
	// userIds := []uint64{1, 2, 3, 4, 5}
	userIds := []uint64{1}
	liveIds := []uint64{10108, 10109, 10110, 10111, 10112}
	numTests := 500
	numOperations := 1000
	var success, fail atomic.Int64
	for j := 0; j < numTests; j++ {
		var wg sync.WaitGroup
		wg.Add(numOperations)
		for i := 0; i < numOperations; i++ {
			go func(i int) {
				defer wg.Done()
				userId := userIds[rand.Int63n(int64(len(userIds)))]
				liveId := liveIds[rand.Int63n(int64(len(liveIds)))]
				_, err := u.SetReserve(ctx, userId, liveId)
				if err != nil {
					fmt.Printf("err: set reserve(%d, %v, %v): %v\n", i, userId, liveId, err)
					fail.Add(1)
					return
				}
				// fmt.Printf("set reserve(%d, %v, %v): %v\n", i, userId, liveId, res)
				success.Add(1)
			}(i)
		}
		wg.Wait()
		fmt.Printf("done: total=%d, success=%d, failure=%d, setCnt=%d, setFailCnt=%d\n",
			success.Load()+fail.Load(), success.Load(), fail.Load(), u.SetCnt(), u.SetFailCnt())
	}
	fmt.Printf("done: total=%d, success=%d, failure=%d, setCnt=%d, setFailCnt=%d\n",
		success.Load()+fail.Load(), success.Load(), fail.Load(), u.SetCnt(), u.SetFailCnt())
}

func casZaddTest() {

	// var err error
	ctx := context.Background()
	timeout := 15 * time.Second
	l := distlock.NewDistLockValkeyV2(client, "key-prefix:", "chan-prefix:", timeout, 0)
	a := adapter.NewReserveValkey(client, "reserve:", 10)
	u := usecase.NewApppushReserve(l, a)

	// simple
	// res, err := u.SetReserve(ctx, 1, 10108)
	// if err != nil {
	// 	fmt.Printf("err: set reserve: %v\n", err)
	// 	return
	// }
	// fmt.Println("set reserve:", res)

	// multi
	// userIds := []uint64{1, 1, 1, 1, 1}
	// liveIds := []uint64{10108, 10109, 10110, 10111, 10112, 10112, 10112}
	// var suc atomic.Int64
	// for _, userId := range userIds {
	// 	for _, liveId := range liveIds {
	// 		_, err := u.SetReserve(ctx, userId, liveId)
	// 		if err != nil {
	// 			fmt.Printf("err: set reserve(%v, %v): %v\n", userId, liveId, err)
	// 			return
	// 		}
	// 		// fmt.Printf("set reserve(%d, %v, %v): %v\n", i, userId, liveId, res)
	// 		suc.Add(1)
	// 	}
	// }
	// fmt.Printf("done... %d\n", suc.Load())

	// multi sequential
	// userIds := []uint64{1, 2, 3, 4, 5}
	// liveIds := []uint64{10108, 10109, 10110, 10111, 10112}
	// numTests := 30
	// var wg sync.WaitGroup
	// wg.Add(numTests * len(userIds))
	// var suc atomic.Int64
	// for _, userId := range userIds {
	// 	for i := 0; i < numTests; i++ {
	// 		go func(i int, userId uint64) {
	// 			defer wg.Done()
	// 			liveId := liveIds[rand.Int63n(int64(len(liveIds)))]
	// 			_, err := u.SetReserve(ctx, userId, liveId)
	// 			if err != nil {
	// 				fmt.Printf("err: set reserve(%d, %v, %v): %v\n", i, userId, liveId, err)
	// 				return
	// 			}
	// 			// fmt.Printf("set reserve(%d, %v, %v): %v\n", i, userId, liveId, res)
	// 			suc.Add(1)
	// 		}(i, userId)
	// 	}
	// }
	// wg.Wait()
	// fmt.Printf("done... %d\n", suc.Load())

	// multi random
	userIds := []uint64{1, 2, 3, 4, 5}
	// userIds := []uint64{1}
	liveIds := []uint64{10108, 10109, 10110, 10111, 10112}
	numTests := 300
	numClients := 3
	var wg sync.WaitGroup
	wg.Add(numTests * numClients)
	var suc atomic.Int64
	for j := 0; j < numClients; j++ {
		for i := 0; i < numTests; i++ {
			go func(i int) {
				defer wg.Done()
				userId := userIds[rand.Int63n(int64(len(userIds)))]
				liveId := liveIds[rand.Int63n(int64(len(liveIds)))]
				_, err := u.SetReserve(ctx, userId, liveId)
				if err != nil {
					fmt.Printf("err: set reserve(%d, %v, %v): %v\n", i, userId, liveId, err)
					return
				}
				// fmt.Printf("set reserve(%d, %v, %v): %v\n", i, userId, liveId, res)
				suc.Add(1)
			}(i)
		}
	}
	wg.Wait()
	fmt.Printf("done... %d\n", suc.Load())
}

func locakTest1() {
	timeout := 15 * time.Second
	l := distlock.NewDistLockValkeyV2(client, "key-prefix:", "chan-prefix:", timeout, 0)

	numTests := 1
	numClientsPerTest := 5
	var wg sync.WaitGroup
	wg.Add(numTests)
	for i := 0; i < numTests; i++ {
		// lockTest(l, i, 1)

		go func(i int) {
			defer wg.Done()
			_lockTest(l, i, numClientsPerTest)
		}(i)
	}
	wg.Wait()
	fmt.Println("done...")
}

func _lockTest(l *distlock.DistLockValkeyV2, i int, numClients int) {
	ctx := context.Background()
	key := fmt.Sprintf("user-1113-%d", i)
	msg := "some-message"

	//////////////////////////////////////////////////////////////////////////////////////////
	// this locks and holds for n duration
	err := l.Lock(ctx, key, msg)
	if err != nil {
		fmt.Printf("err: lock: %v\n", err)
		return
	}
	fmt.Println("acquired lock")
	go func() {
		time.Sleep(1500 * time.Millisecond)
		if err := l.Unlock(ctx, key, msg); err != nil {
			fmt.Printf("err: unlock: %v\n", err)
		}
	}()
	//////////////////////////////////////////////////////////////////////////////////////////

	var cnt atomic.Int64
	var wg sync.WaitGroup
	wg.Add(numClients)
	for i := 0; i < numClients; i++ {
		go func(i int) {
			defer wg.Done()
			err = l.Lock(ctx, key, fmt.Sprintf("%s-%d", msg, i))
			if err != nil {
				log.Printf("err: lock(%d): %v, %v\n", i, err, time.Now())
				return
			}
			cnt.Add(1)
			defer l.Unlock(ctx, key, fmt.Sprintf("%s-%d", msg, i))
			log.Printf("acquired lock(%d), %v\n", i, time.Now())
		}(i)
	}
	wg.Wait()
	if cnt.Load() == 0 {
		panic("cnt == 0")
	}
}

func exampleLockV2() {
	client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{"127.0.0.1:6379"}})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	timeout := 5 * time.Second
	l := distlock.NewDistLockValkeyV2(client, "key-prefix:", "chan-prefix:", timeout, 0)

	ctx := context.Background()
	key := "user-1112"
	msg := "some-message"

	err = l.Lock(ctx, key, msg)
	if err != nil {
		fmt.Printf("err: lock: %v\n", err)
		return
	}
	go func() {
		time.Sleep(timeout - 1*time.Second)
		if err := l.Unlock(ctx, key, msg); err != nil {
			fmt.Printf("err: unlock: %v\n", err)
		}
	}()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			err = l.Lock(ctx, key, msg)
			if err != nil {
				fmt.Printf("err: lock(%d): %v\n", i, err)
				return
			}
			defer l.Unlock(ctx, key, msg)
			fmt.Printf("acquired lock(%d)\n", i)
		}(i)
	}
	wg.Wait()

	err = l.Lock(ctx, key, msg)
	if err != nil {
		fmt.Printf("err: lock(2): %v\n", err)
		return
	}
	defer l.Unlock(ctx, key, msg)
}
func example() {
	fmt.Println("Hello, World!")

	client, err := valkey.NewClient(valkey.ClientOption{InitAddress: []string{"127.0.0.1:6379"}})
	if err != nil {
		panic(err)
	}
	defer client.Close()

	ctx := context.Background()

	// SET key val NX
	err = client.Do(ctx, client.B().Set().Key("key").Value("val").Nx().Build()).Error()
	if err != nil {
		fmt.Printf("err: setnx: %v\n", err)
	}

	res := client.Do(ctx, client.B().Set().Key("key2").Value("some-value").Build())
	if res.Error() != nil {
		fmt.Printf("err: set: %v\n", err)
	} else {
		fmt.Printf("set key: %v\n", res)
	}

	// HGETALL hm
	hm, err := client.Do(ctx, client.B().Hgetall().Key("hm").Build()).AsStrMap()
	if err != nil {
		fmt.Printf("err: hgetall: %v\n", err)
	} else {
		fmt.Println(hm)
	}

	//////////////////////////////////////////////////////////////////////////////////////////
	// PUB/SUB
	ctx2, cancel2 := context.WithCancel(ctx)
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup, ctx context.Context) {
		defer wg.Done()
		err := client.Receive(ctx, client.B().Subscribe().Channel("ch1", "ch2").Build(), func(msg valkey.PubSubMessage) {
			// Handle the message. Note that if you want to call another `client.Do()` here, you need to do it in another goroutine or the `client` will be blocked.
			fmt.Printf("Received message: %v\n", msg)
		})
		if err != nil {
			fmt.Printf("err: subscribe: %v\n", err)
		}
	}(&wg, ctx2)

	time.Sleep(200 * time.Millisecond)
	res2 := client.Do(ctx, client.B().Publish().Channel("ch1").Message("hello").Build())
	if res2.Error() != nil {
		fmt.Printf("err: publish: %v\n", res2.Error())
	} else {
		fmt.Printf("publish: %v\n", res2)
	}

	cancel2()
	wg.Wait()
	//////////////////////////////////////////////////////////////////////////////////////////

	fmt.Println("finished, bye")
}
