// Copyright © 2023 Meroxa, Inc.
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

package destination

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/conduitio/conduit-commons/csync"
	sdk "github.com/conduitio/conduit-connector-sdk"
	"github.com/conduitio/conduit-connector-sdk/kafkaconnect"
	"github.com/goccy/go-json"
	"github.com/twmb/franz-go/pkg/kgo"
)

type FranzProducer struct {
	client     *kgo.Client
	keyEncoder dataEncoder

	// getTopic is a function that returns the topic for a record. If nil, the
	// producer will use the default topic. This function is not safe for
	// concurrent use.
	getTopic func(sdk.Record) (string, error)
}

var _ Producer = (*FranzProducer)(nil)

func NewFranzProducer(ctx context.Context, cfg Config) (*FranzProducer, error) {
	opts := cfg.FranzClientOpts(sdk.Logger(ctx))
	opts = append(opts, []kgo.Opt{
		kgo.AllowAutoTopicCreation(),
		kgo.RecordDeliveryTimeout(cfg.DeliveryTimeout),
		kgo.RequiredAcks(cfg.RequiredAcks()),
		kgo.ProducerBatchCompression(cfg.CompressionCodecs()...),
		kgo.ProducerBatchMaxBytes(cfg.BatchBytes),
	}...)

	var topicFn func(sdk.Record) (string, error)
	if strings.Contains(cfg.Topic, "{{") && strings.Contains(cfg.Topic, "}}") {
		// If the topic contains a template, the topic will be determined for
		// each record individually, so we can't set the default topic here.
		t, err := template.New("topic").Funcs(sprig.FuncMap()).Parse(cfg.Topic)
		if err != nil {
			return nil, fmt.Errorf("failed to parse topic template: %w", err)
		}
		var buf bytes.Buffer
		topicFn = func(r sdk.Record) (string, error) {
			buf.Reset()
			if err := t.Execute(&buf, r); err != nil {
				return "", fmt.Errorf("failed to execute topic template: %w", err)
			}
			return buf.String(), nil
		}
	} else {
		opts = append(opts, kgo.DefaultProduceTopic(cfg.Topic))
	}

	if cfg.RequiredAcks() != kgo.AllISRAcks() {
		sdk.Logger(ctx).Warn().Msgf("disabling idempotent writes because \"acks\" is set to %v", cfg.Acks)
		opts = append(opts, kgo.DisableIdempotentWrite())
	}

	cl, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create kafka client: %w", err)
	}

	var keyEncoder dataEncoder = bytesEncoder{}
	if cfg.useKafkaConnectKeyFormat {
		keyEncoder = kafkaConnectEncoder{}
	}

	return &FranzProducer{
		client:     cl,
		keyEncoder: keyEncoder,
		getTopic:   topicFn,
	}, nil
}

func (p *FranzProducer) Produce(ctx context.Context, records []sdk.Record) (int, error) {
	if len(records) == 1 {
		// Fast path for a single record.
		rec, err := p.prepareRecord(records[0])
		if err != nil {
			return 0, fmt.Errorf("failed to prepare record: %w", err)
		}
		_, err = p.client.ProduceSync(ctx, rec).First()
		if err != nil {
			return 0, fmt.Errorf("failed to produce record: %w", err)
		}
		return 1, nil
	}

	var (
		wg       csync.WaitGroup
		results  = make([]error, 0, len(records))
		errIndex = -1
		err      error
		rec      *kgo.Record
	)

	for i, r := range records {
		rec, err = p.prepareRecord(r)
		if err != nil {
			errIndex = i
			err = fmt.Errorf("failed to prepare record: %w", err)
			break
		}

		wg.Add(1)
		p.client.Produce(
			ctx,
			rec,
			func(_ *kgo.Record, err error) {
				results = append(results, err)
				wg.Done()
			},
		)
	}

	err = wg.Wait(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to wait for all records to be produced: %w", err)
	}

	for i, err := range results {
		if err != nil {
			return i, fmt.Errorf("failed to produce record %v: %w", i, err)
		}
	}

	if err != nil {
		// We failed to prepare a record, return the error and the index of the
		// record that failed.
		return errIndex, err
	}

	return len(results), nil
}

func (p *FranzProducer) prepareRecord(r sdk.Record) (*kgo.Record, error) {
	encodedKey, err := p.keyEncoder.Encode(r.Key)
	if err != nil {
		return nil, fmt.Errorf("could not encode key: %w", err)
	}

	var topic string
	if p.getTopic != nil {
		topic, err = p.getTopic(r)
		if err != nil {
			return nil, fmt.Errorf("could not get topic: %w", err)
		}
	}
	return &kgo.Record{
		Key:   encodedKey,
		Value: r.Bytes(),
		Topic: topic,
	}, nil
}

func (p *FranzProducer) Close(_ context.Context) error {
	if p.client != nil {
		p.client.Close()
	}
	return nil
}

// dataEncoder is similar to a sdk.Encoder, which takes data and encodes it in
// a certain format. The producer uses this to encode the key of the kafka
// message.
type dataEncoder interface {
	Encode(sdk.Data) ([]byte, error)
}

// bytesEncoder is a dataEncoder that simply calls data.Bytes().
type bytesEncoder struct{}

func (bytesEncoder) Encode(data sdk.Data) ([]byte, error) {
	return data.Bytes(), nil
}

// kafkaConnectEncoder encodes the data into a kafka connect JSON with schema
// (NB: this is not the same as JSONSchema).
type kafkaConnectEncoder struct{}

func (e kafkaConnectEncoder) Encode(data sdk.Data) ([]byte, error) {
	sd := e.toStructuredData(data)
	schema := kafkaconnect.Reflect(sd)
	if schema == nil {
		// s is nil, let's write an empty struct in the schema
		schema = &kafkaconnect.Schema{
			Type:     kafkaconnect.TypeStruct,
			Optional: true,
		}
	}

	env := kafkaconnect.Envelope{
		Schema:  *schema,
		Payload: sd,
	}
	// TODO add support for other encodings than JSON
	return json.Marshal(env)
}

// toStructuredData tries its best to return StructuredData.
func (kafkaConnectEncoder) toStructuredData(d sdk.Data) sdk.Data {
	switch d := d.(type) {
	case nil:
		return nil
	case sdk.StructuredData:
		return d
	case sdk.RawData:
		// try parsing the raw data as json
		var sd sdk.StructuredData
		err := json.Unmarshal(d, &sd)
		if err != nil {
			// it's not JSON, nothing more we can do
			return d
		}
		return sd
	default:
		// should not be possible
		panic(fmt.Errorf("unknown data type: %T", d))
	}
}
