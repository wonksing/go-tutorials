package internal

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/IBM/sarama"
)

func ConsumerGroupWithConfig(config *sarama.Config, brokers []string, group string) (sarama.ConsumerGroup, error) {

	client, err := sarama.NewConsumerGroup(brokers, group, config)

	return client, err
}

type ConsumerGroup struct {
	sarama.ConsumerGroup
	consumer Consumer
	topics   []string

	isPaused bool
	cancel   context.CancelFunc
}

func NewConsumerGroup(cg sarama.ConsumerGroup, consumer Consumer, topics []string) *ConsumerGroup {
	return &ConsumerGroup{
		ConsumerGroup: cg,
		consumer:      consumer,
		topics:        topics,
		isPaused:      false,
	}
}

func (cg *ConsumerGroup) Toggle() {
	cg.toggleConsumptionFlow()
}

func (cg *ConsumerGroup) Start(loaded chan bool) error {
	var ctx context.Context
	ctx, cg.cancel = context.WithCancel(context.Background())

	var startErr error = nil
	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := cg.Consume(ctx, cg.topics, cg.consumer); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					startErr = err
					return
				}
				log.Panicf("Error from consumer: %v\n", err)
				// log.Println("consume error: " + err.Error())
				// for i := 0; i < 12; i++ {
				// 	time.Sleep(time.Second * 1)
				// 	if ctx.Err() != nil {
				// 		startErr = ctx.Err()
				// 		return
				// 	}
				// }
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				startErr = ctx.Err()
				return
			}
			// consumer.ready = make(chan bool)
			cg.consumer.MakeReady()
		}
	}(&wg)

	cg.consumer.WaitReady() // Await till the consumer has been set up
	loaded <- true
	log.Println("Sarama consumers are up and running")

	sigusr1 := make(chan os.Signal, 1)
	signal.Notify(sigusr1, syscall.SIGUSR1)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	keepRunning := true
	for keepRunning {
		select {
		case <-ctx.Done():
			log.Println("terminating: context cancelled")
			keepRunning = false
		case <-sigterm:
			log.Println("terminating: via signal")
			keepRunning = false
		case <-sigusr1:
			cg.toggleConsumptionFlow()
		}
	}
	cg.cancel()
	wg.Wait()
	if err := cg.Close(); err != nil {
		if startErr == nil {
			startErr = err
		}
		log.Panicf("Error closing client: %v", err)
	}

	return startErr
}

func (cg *ConsumerGroup) Finish() error {
	if cg.cancel == nil {
		return errors.New("ConsumerGroup has not been started")
	}
	cg.cancel()
	return nil
}

func (cg *ConsumerGroup) Stop() error {
	return cg.Finish()
}

func (cg *ConsumerGroup) toggleConsumptionFlow() {
	if cg.isPaused {
		cg.ResumeAll()
		log.Println("Resuming consumption")
	} else {
		cg.PauseAll()
		log.Println("Pausing consumption")
	}

	cg.isPaused = !cg.isPaused
}
