package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/eapache/go-resiliency/breaker"
)

// Sarama configuration options
var (
	brokers   = ""
	version   = ""
	topic     = ""
	producers = 1
	verbose   = false

	recordsNumber int64 = 1
)

func init() {
	flag.StringVar(&brokers, "brokers", "testkafkahost:9092", "Kafka bootstrap brokers to connect to, as a comma separated list")
	flag.StringVar(&version, "version", sarama.DefaultVersion.String(), "Kafka cluster version")
	flag.StringVar(&topic, "topic", "tp-txn", "Kafka topics where records will be copied from topics.")
	flag.IntVar(&producers, "producers", 10, "Number of concurrent producers")
	flag.Int64Var(&recordsNumber, "records-number", 10000, "Number of records sent per loop")
	flag.BoolVar(&verbose, "verbose", false, "Sarama logging")
	flag.Parse()

	if len(brokers) == 0 {
		panic("no Kafka bootstrap brokers defined, please set the -brokers flag")
	}

	if len(topic) == 0 {
		panic("no topic given to be consumed, please set the -topic flag")
	}
}
func main() {

	if verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	version, err := sarama.ParseKafkaVersion(version)
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}

	config := sarama.NewConfig()
	config.ClientID = "test_retry"
	config.Version = version
	config.Producer.Idempotent = true
	config.Producer.Retry.Max = 10
	config.Producer.Retry.BackoffFunc = func(retries int, maxRetries int) time.Duration {
		v := (1 << retries) * 100 * time.Millisecond
		if v > 10000*time.Millisecond {
			v = 10000 * time.Millisecond
		}
		return v
	}
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true
	config.Net.ReadTimeout = 10 * time.Second
	config.Net.WriteTimeout = 10 * time.Second
	config.Net.DialTimeout = 10 * time.Second
	config.Net.MaxOpenRequests = 1
	config.Metadata.Retry.Max = 10
	config.Metadata.Retry.BackoffFunc = func(retries int, maxRetries int) time.Duration {
		v := (1 << retries) * 100 * time.Millisecond
		if v > 10000*time.Millisecond {
			v = 10000 * time.Millisecond
		}
		return v
	}
	config.Metadata.RefreshFrequency = 1 * time.Minute

	brokersArr := strings.Split(brokers, ",")
	producer, err := sarama.NewSyncProducer(brokersArr, config)
	if err != nil {
		log.Printf("failed to create sync producer: %v\n", err)
		os.Exit(1)
	}
	defer producer.Close()

	var wg sync.WaitGroup

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var success, failure int64
	numProducers := 4
	for i := 0; i < numProducers; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, i int) {
			defer wg.Done()
			ticker := time.NewTicker(80 * time.Millisecond)
			defer ticker.Stop()

			j := 0
			for {
				select {
				case <-ticker.C:
					err := produce(producer, &sarama.ProducerMessage{
						Topic: topic,
						Key:   sarama.ByteEncoder(fmt.Sprintf("%d-%d", i, j)),
						Value: sarama.StringEncoder(fmt.Sprintf("message %d %d", i, j)),
					})
					if err != nil {
						log.Printf("error: failed to produce: %v\n", err)
						atomic.AddInt64(&failure, 1)
						continue
					}
					atomic.AddInt64(&success, 1)
				case <-ctx.Done():
					return
				}
			}
		}(&wg, i)
	}

	timer := time.NewTimer(180 * time.Second)
	// <-timer.C
	select {
	case <-timer.C:
	case <-sigterm:
	}
	cancel()
	wg.Wait()
	fmt.Printf("success=%d, failure=%d\n", success, failure)
}

func produce(producer sarama.SyncProducer, message *sarama.ProducerMessage) error {
	var err error
	attempts := 0
	wait := 100 * time.Millisecond
	for attempts < 5 {
		attempts++
		_, _, err = producer.SendMessage(message)
		if err != nil {
			if isRetryableError(err) {
				wait *= 2
				log.Printf("retrying in %d (%d/5)\n", wait, attempts)
				time.Sleep(wait)
				continue
			}
			return err
		}
		return nil
	}
	return err
}

func isRetryableError(err error) bool {
	switch err {
	case sarama.ErrBrokerNotAvailable,
		sarama.ErrLeaderNotAvailable,
		sarama.ErrReplicaNotAvailable,
		sarama.ErrRequestTimedOut,
		sarama.ErrNotEnoughReplicas,
		// sarama.ErrNotEnoughReplicasAfterAppend, // "kafka server: Messages are written to the log, but to fewer in-sync replicas than required"
		// sarama.ErrNetworkException, // "kafka server: The server disconnected before a response was received"
		sarama.ErrOutOfBrokers,
		sarama.ErrOutOfOrderSequenceNumber,
		sarama.ErrNotController,
		sarama.ErrNotLeaderForPartition,
		breaker.ErrBreakerOpen:
		return true
	default:
		return false
	}
}
