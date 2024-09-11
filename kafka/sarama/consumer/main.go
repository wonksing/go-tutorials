package main

import (
	"context"
	"flag"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/IBM/sarama"
	"github.com/wonksing/go-tutorials/kafka/sarama/consumer/internal"
)

// Sarama configuration options
var (
	brokers = ""
	version = ""
	topic   = ""
	group   = ""
	verbose = false
)

func init() {
	flag.StringVar(&brokers, "brokers", "testkafkahost:9092", "Kafka bootstrap brokers to connect to, as a comma separated list")
	flag.StringVar(&version, "version", sarama.DefaultVersion.String(), "Kafka cluster version")
	flag.StringVar(&topic, "topic", "tp-txn", "Kafka topics where records will be copied from topics.")
	flag.StringVar(&group, "group", "tp-txn-grp01", "")
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

	config, err := buildConfig()
	if err != nil {
		log.Printf("failed to build config: %v\n", err)
		os.Exit(1)
	}

	client, err := sarama.NewConsumerGroup([]string{brokers}, group, config)
	if err != nil {
		return
	}
	consumer := internal.NewGroupConsumer(&messageHandler{})
	cg := internal.NewConsumerGroup(client, consumer, []string{topic})
	consumerLoaded := make(chan bool, 1)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		err := cg.Start(consumerLoaded)
		if err != nil {
			log.Printf("failed starting consumer: %v", err)
		}
		cancel()
	}(&wg)
	<-consumerLoaded

	ticker := time.NewTicker(3 * time.Second)
	wg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				log.Printf("number of consumed messages: %d\n", consumed)
			}
		}
	}(&wg)

	wg.Wait()
	ticker.Stop()
}

func buildConfig() (*sarama.Config, error) {

	config := sarama.NewConfig()
	v, err := sarama.ParseKafkaVersion(version)
	if err != nil {
		return nil, err
	}
	config.Version = v
	config.Consumer.Group.Rebalance.Strategy = sarama.NewBalanceStrategyRange()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	config.Consumer.Group.Session.Timeout = time.Duration(60) * time.Second
	config.Consumer.Group.Heartbeat.Interval = time.Duration(60/3) * time.Second

	config.Metadata.RefreshFrequency = 1 * time.Minute // consumer starts rebalancing right after started, and it freezes
	return config, nil
}

var consumed int64

type messageHandler struct{}

func (h *messageHandler) Handle(message *sarama.ConsumerMessage) error {
	atomic.AddInt64(&consumed, 1)
	return nil
}
