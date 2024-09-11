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
)

// Sarama configuration options
var (
	brokers   = ""
	version   = ""
	topic     = ""
	producers = 1
	duraion   = 0
	verbose   = false

	brokerList []string
)

func init() {
	flag.StringVar(&brokers, "brokers", "testkafkahost:9092", "Kafka bootstrap brokers to connect to, as a comma separated list")
	flag.StringVar(&version, "version", sarama.DefaultVersion.String(), "Kafka cluster version")
	flag.StringVar(&topic, "topic", "tp-txn", "Kafka topics where records will be copied from topics.")
	flag.IntVar(&producers, "producers", 10, "Number of concurrent producers")
	flag.IntVar(&duraion, "duration", 10, "Duration in seconds to produce messages")
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

	producer, err := sarama.NewAsyncProducer(brokerList, config)
	if err != nil {
		log.Printf("failed to create sync producer: %v\n", err)
		os.Exit(1)
	}
	// defer producer.Close()

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var sent, success, failure int64
	var wgErr sync.WaitGroup
	wgErr.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		for pe := range producer.Errors() {
			log.Println("fail: " + pe.Error())
			atomic.AddInt64(&failure, 1)
		}
		log.Printf("exiting from receiving errors\n")
	}(&wgErr)

	log.Println("start producing messages asynchronously")
	var wg sync.WaitGroup
	for i := 0; i < producers; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup, i int) {
			defer wg.Done()
			ticker := time.NewTicker(50 * time.Millisecond)
			defer ticker.Stop()

			j := 0
			for {
				select {
				case <-ticker.C:
					producer.Input() <- &sarama.ProducerMessage{
						Topic: topic,
						Key:   sarama.ByteEncoder(fmt.Sprintf("%d-%d", i, j)),
						Value: sarama.StringEncoder(fmt.Sprintf("message %d %d", i, j)),
					}
					atomic.AddInt64(&sent, 1)
				case <-ctx.Done():
					log.Printf("exiting producer: %d\n", i)
					return
				}
			}
		}(&wg, i)
	}

	timer := time.NewTimer(time.Duration(duraion) * time.Second)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-sigterm:
	}

	cancel()
	wg.Wait()
	fmt.Printf("closing producer\n")
	err = producer.Close()
	if err != nil {
		if pe, ok := err.(sarama.ProducerErrors); ok {
			atomic.AddInt64(&failure, int64(len(pe)))
		}
	}
	wgErr.Wait()
	fmt.Printf("finished producing messages: sent=%d, success=%d, failure=%d\n", sent, success, failure)
}

func buildConfig() (*sarama.Config, error) {
	version, err := sarama.ParseKafkaVersion(version)
	if err != nil {
		// log.Panicf("Error parsing Kafka version: %v", err)
		return nil, err
	}

	config := sarama.NewConfig()
	config.ClientID = "test_retry"
	config.Version = version
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
	config.Producer.Return.Successes = false
	config.Producer.Return.Errors = true
	config.Net.ReadTimeout = 10 * time.Second
	config.Net.WriteTimeout = 10 * time.Second
	config.Net.DialTimeout = 10 * time.Second
	config.Net.MaxOpenRequests = 1
	config.ChannelBufferSize = 1024
	config.Producer.Timeout = 3000 * time.Millisecond
	config.Producer.Flush.Frequency = 3 * time.Second
	config.Producer.Flush.Messages = 512
	config.Producer.Flush.MaxMessages = 1024

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
