package internal

import "github.com/IBM/sarama"

type MessageHandler interface {
	Handle(msg *sarama.ConsumerMessage) error
}

type Consumer interface {
	Setup(sarama.ConsumerGroupSession) error
	Cleanup(sarama.ConsumerGroupSession) error
	ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error
	MakeReady()
	WaitReady()
	CloseReady()
}
