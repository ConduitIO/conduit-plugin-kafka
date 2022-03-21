// Copyright © 2022 Meroxa, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

//go:generate mockgen -destination mock/producer.go -package mock -mock_names=Producer=Producer . Producer

package kafka

import (
	"context"
	"fmt"

	"github.com/segmentio/kafka-go"
)

type Producer interface {
	// Send synchronously delivers a message.
	// Returns an error, if the message could not be delivered.
	Send(key []byte, payload []byte) error

	// Close this producer and the associated resources (e.g. connections to the broker)
	Close() error
}

type segmentProducer struct {
	writer *kafka.Writer
}

// NewProducer creates a new Kafka producer.
// The current implementation uses Segment's kafka-go client.
func NewProducer(cfg Config) (Producer, error) {
	if len(cfg.Servers) == 0 {
		return nil, ErrServersMissing
	}
	if cfg.Topic == "" {
		return nil, ErrTopicMissing
	}

	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Servers...),
		Topic:        cfg.Topic,
		BatchSize:    1,
		WriteTimeout: cfg.DeliveryTimeout,
		RequiredAcks: cfg.Acks,
		MaxAttempts:  3,
	}
	// TLS config
	if cfg.ClientCert != "" {
		tlsCfg, err := newTLSConfig(cfg.ClientCert, cfg.ClientKey, cfg.CACert, cfg.InsecureSkipVerify)
		if err != nil {
			return nil, fmt.Errorf("invalid TLS config: %w", err)
		}
		transport := &kafka.Transport{
			TLS: tlsCfg,
		}
		// todo move out
		if cfg.saslEnabled() {
			transportWithSASL(transport, cfg)
		}
		writer.Transport = transport
	}
	return &segmentProducer{writer: writer}, nil
}

func (c *segmentProducer) Send(key []byte, payload []byte) error {
	err := c.writer.WriteMessages(
		context.Background(),
		kafka.Message{
			Key:   key,
			Value: payload,
		},
	)

	if err != nil {
		return fmt.Errorf("message not delivered: %w", err)
	}
	return nil
}

func (c *segmentProducer) Close() error {
	if c.writer == nil {
		return nil
	}
	// this will also make the loops in the reader goroutines stop
	err := c.writer.Close()
	if err != nil {
		return fmt.Errorf("couldn't close writer: %w", err)
	}

	return nil
}
