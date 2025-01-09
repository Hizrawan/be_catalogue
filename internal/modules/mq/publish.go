package mq

import (
	"encoding/json"
	"fmt"

	"github.com/nsqio/go-nsq"
)

func PublishMessage(producer *nsq.Producer, topic string, msg any) error {
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("[PublishMessage] marshal: %w", err)
	}

	err = producer.Publish(topic, jsonMsg)
	if err != nil {
		return fmt.Errorf("[PublishMessage] publish: %w", err)
	}

	return nil
}
