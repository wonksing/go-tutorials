package internal

import (
	"log"

	"github.com/IBM/sarama"
)

// GroupConsumer represents a Sarama consumer group consumer
type GroupConsumer struct {
	ready      chan bool
	msgHandler MessageHandler
}

func NewGroupConsumer(msgHandler MessageHandler) *GroupConsumer {
	c := &GroupConsumer{}
	c.MakeReady()
	c.msgHandler = msgHandler

	return c
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer *GroupConsumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	consumer.CloseReady()
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer *GroupConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer *GroupConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				log.Println("message channel was closed")
				return nil
			}
			// log.Printf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
			if err := consumer.msgHandler.Handle(message); err == nil {
				session.MarkMessage(message, "")
			}

		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/Shopify/sarama/issues/1192
		case <-session.Context().Done():
			return nil
		}
	}
}

func (c *GroupConsumer) MakeReady() {
	c.ready = make(chan bool)
}

func (c *GroupConsumer) WaitReady() {
	<-c.ready
}

func (c *GroupConsumer) CloseReady() {
	close(c.ready)
}
