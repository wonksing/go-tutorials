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

	brokerList []string
)

func init() {
	flag.StringVar(&brokers, "brokers", "testkafkahost:9092", "Kafka bootstrap brokers to connect to, as a comma separated list")
	flag.StringVar(&version, "version", sarama.DefaultVersion.String(), "Kafka cluster version")
	flag.StringVar(&topic, "topic", "tp-txn", "Kafka topics where records will be copied from topics.")
	flag.IntVar(&producers, "producers", 10, "Number of concurrent producers")
	flag.BoolVar(&verbose, "verbose", false, "Sarama logging")
	flag.Parse()

	if len(brokers) == 0 {
		panic("no Kafka bootstrap brokers defined, please set the -brokers flag")
	}

	if len(topic) == 0 {
		panic("no topic given to be consumed, please set the -topic flag")
	}

	brokerList = strings.Split(brokers, ",")
}

func main() {

	if verbose {
		sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	}

	config, err := buildConfig()
	if err != nil {
		log.Printf("failed to build config: %v\n", err)
		os.Exit(1)
	}

	producer, err := sarama.NewSyncProducer(brokerList, config)
	if err != nil {
		log.Printf("failed to create sync producer: %v\n", err)
		os.Exit(1)
	}
	defer producer.Close()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var success, failure int64
	var wg sync.WaitGroup
	for i := 0; i < producers; i++ {
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
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-sigterm:
	}

	cancel()
	wg.Wait()
	fmt.Printf("success=%d, failure=%d\n", success, failure)
}

func buildConfig() (*sarama.Config, error) {
	version, err := sarama.ParseKafkaVersion(version)
	if err != nil {
		return nil, err
	}

	config := sarama.NewConfig()
	config.ClientID = "test-sync-producer"
	config.Version = version

	config.Net.ReadTimeout = 30 * time.Second
	config.Net.WriteTimeout = 30 * time.Second
	config.Net.DialTimeout = 30 * time.Second
	config.Net.MaxOpenRequests = 1

	config.Producer.Timeout = 30 * time.Second
	config.Producer.Idempotent = true
	config.Producer.Retry.Max = 10
	config.Producer.Retry.BackoffFunc = func(retries int, maxRetries int) time.Duration {
		v := (1 << retries) * 250 * time.Millisecond
		if v > 10000*time.Millisecond {
			v = 10000 * time.Millisecond
		}
		return v
	}
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Errors = true
	config.Producer.Return.Successes = true

	config.Metadata.Retry.Max = 10
	config.Metadata.Retry.BackoffFunc = func(retries int, maxRetries int) time.Duration {
		v := (1 << retries) * 250 * time.Millisecond
		if v > 10000*time.Millisecond {
			v = 10000 * time.Millisecond
		}
		return v
	}
	config.Metadata.RefreshFrequency = 5 * time.Minute
	return config, nil
}

func produce(producer sarama.SyncProducer, message *sarama.ProducerMessage) error {
	var err error
	attempts := 0
	wait := 1000 * time.Millisecond
	for attempts < 5 {
		attempts++
		_, _, err = producer.SendMessage(message)
		if err != nil {
			if isRetryableError(err) {
				log.Printf("manual retrying in %.3f second(s): (%d/5)\n", wait.Seconds(), attempts)
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
