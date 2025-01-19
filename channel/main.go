package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"
)

func fail() {
	ch := make(chan string)

	numReceivers := 100
	numSenders := 100

	var l sync.RWMutex
	var wg sync.WaitGroup
	wg.Add(numSenders)
	for i := 0; i < numSenders; i++ {
		go func(i int) {
			defer wg.Done()

			n := rand.Int63n(500)
			time.Sleep(time.Duration(n) * time.Millisecond)
			select {
			case ch <- fmt.Sprintf("Hello from goroutine %d", i):
			default:
				if len(ch) == 0 {
					l.Lock()
					fmt.Println("Channel is empty")
					close(ch)
					ch = make(chan string)
					l.Unlock()
				} else {
					fmt.Println("Channel is full")
				}

			}
		}(i)
	}

	wg.Add(numReceivers)
	for i := 0; i < numReceivers; i++ {
		go func(i int) {
			defer wg.Done()

			n := rand.Int63n(500)
			time.Sleep(time.Duration(n) * time.Millisecond)

			l.Lock()
			msg := <-ch
			fmt.Println(msg)
			l.Unlock()
		}(i)
	}

	wg.Wait()

	fmt.Printf("numReceivers: %d, numSenders: %d\n", numReceivers, numSenders)
}

func main() {
	pubsub := make(chan string)
	select {
	case pubsub <- "key":
	default:
		fmt.Println("channel is not ready")
	}
	numReceivers := 10000

	var wg sync.WaitGroup
	wg.Add(numReceivers)
	for i := 0; i < numReceivers; i++ {
		go func() {
			defer wg.Done()
			randSleep()
			pubsub <- "key"
			randSleep()
			ch := addLockChan("key")

			select {
			case msg := <-ch:
				fmt.Println(msg)
			case <-time.After(5 * time.Second):
				fmt.Println("timeout")
			}
		}()
	}

	var wg2 sync.WaitGroup
	wg2.Add(1)
	go func() {
		wg2.Done()
		for key := range pubsub {
			// randSleep()
			releaseLockChan(key)
		}
	}()
	wg.Wait()
	close(pubsub)
	wg2.Wait()
}

func randSleep() {
	time.Sleep(time.Duration(20+rand.Intn(100)) * time.Millisecond)
}

var (
	lockChansMutex sync.RWMutex
	lockChans      = make(map[string]chan string)
)

func addLockChan(key string) chan string {
	lockChansMutex.Lock()
	defer lockChansMutex.Unlock()

	if c, ok := lockChans[key]; ok {
		return c
	}
	ch := make(chan string)
	lockChans[key] = ch
	// lockMeasure.IncChanCreated()
	return ch
}

func releaseLockChan(key string) {
	lockChansMutex.Lock()
	defer lockChansMutex.Unlock()
	// randSleep()
	if c, ok := lockChans[key]; ok {
		select {
		case c <- key:
		default:
			if len(c) == 0 {
				close(c)
				delete(lockChans, key)
				// lockMeasure.IncChanDeleted()
				log.Println("channel is empty!!!!!!!!!!")
				return
			}
			log.Println("channel is not empty")
			return
		}
	}
}
