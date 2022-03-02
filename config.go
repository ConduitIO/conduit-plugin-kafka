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

package kafka

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

const (
	Servers            = "servers"
	Topic              = "topic"
	SecurityProtocol   = "securityProtocol"
	Acks               = "acks"
	DeliveryTimeout    = "deliveryTimeout"
	ReadFromBeginning  = "readFromBeginning"
	ClientCert         = "clientCert"
	ClientKey          = "clientKey"
	CACert             = "caCert"
	InsecureSkipVerify = "insecureSkipVerify"
)

var Required = []string{Servers, Topic}

// Config contains all the possible configuration parameters for Kafka sources and destinations.
// When changing this struct, please also change the plugin specification (in main.go) as well as the ReadMe.
type Config struct {
	// A list of bootstrap servers, which will be used to discover all the servers in a cluster.
	Servers []string
	Topic   string
	// Required acknowledgments when writing messages to a topic:
	// Can be: 0, 1, -1 (all)
	Acks            kafka.RequiredAcks
	DeliveryTimeout time.Duration
	// Read all messages present in a source topic.
	// Default value: false (only new messages are read)
	ReadFromBeginning bool
	// TLS
	ClientCert         string
	ClientKey          string
	CACert             string
	InsecureSkipVerify bool
}

func Parse(cfg map[string]string) (Config, error) {
	err := checkRequired(cfg)
	// todo check if values are valid, e.g. hosts are valid etc.
	if err != nil {
		return Config{}, err
	}
	// parse servers
	servers, err := split(cfg[Servers])
	if err != nil {
		return Config{}, fmt.Errorf("invalid servers: %w", err)
	}
	var parsed = Config{
		Servers: servers,
		Topic:   cfg[Topic],
	}
	// parse acknowledgment setting
	ack, err := parseAcks(cfg[Acks])
	if err != nil {
		return Config{}, fmt.Errorf("couldn't parse ack: %w", err)
	}
	parsed.Acks = ack

	// parse and validate ReadFromBeginning
	readFromBeginning, err := parseBool(cfg, ReadFromBeginning, false)
	if err != nil {
		return Config{}, fmt.Errorf("invalid value for ReadFromBeginning: %w", err)
	}
	parsed.ReadFromBeginning = readFromBeginning

	// parse and validate delivery DeliveryTimeout
	timeout, err := parseDuration(cfg, DeliveryTimeout, 10*time.Second)
	if err != nil {
		return Config{}, fmt.Errorf("invalid delivery timeout: %w", err)
	}
	// it makes no sense to expect a message to be delivered immediately
	if timeout == 0 {
		return Config{}, errors.New("invalid delivery timeout: has to be > 0ms")
	}
	parsed.DeliveryTimeout = timeout

	err = setTLSConfigs(&parsed, cfg)
	if err != nil {
		return Config{}, fmt.Errorf("invalid TLS config: %w", err)
	}
	return parsed, nil
}

func setTLSConfigs(parsed *Config, cfg map[string]string) error {
	// All three values should be set so that TLS works
	// If none of the three values are set, then TLS should not be used.
	tlsCfgOk := (cfg[ClientCert] == "") == (cfg[ClientKey] == "")
	tlsCfgOk = tlsCfgOk && (cfg[ClientKey] == "") == (cfg[CACert] == "")
	if !tlsCfgOk {
		return errors.New("TLS config not OK (all values need to be set or all values need to be unset)")
	}
	parsed.ClientCert = cfg[ClientCert]
	parsed.ClientKey = cfg[ClientKey]
	parsed.CACert = cfg[CACert]
	// Parse InsecureSkipVerify, default is 'false'
	insecureString, ok := cfg[InsecureSkipVerify]
	if ok {
		insecure, err := strconv.ParseBool(insecureString)
		if err != nil {
			return fmt.Errorf("value %q for InsecureSkipVerify is not valid", insecureString)
		}
		parsed.InsecureSkipVerify = insecure
	}
	return nil
}

func parseAcks(ack string) (kafka.RequiredAcks, error) {
	// when ack is empty, return default (which is 'all')
	if ack == "" {
		return kafka.RequireAll, nil
	}
	acks := kafka.RequiredAcks(0)
	err := acks.UnmarshalText([]byte(ack))
	if err != nil {
		return 0, fmt.Errorf("unknown ack mode: %w", err)
	}
	return acks, nil
}

func parseBool(cfg map[string]string, key string, defaultVal bool) (bool, error) {
	boolString, exists := cfg[key]
	if !exists {
		return defaultVal, nil
	}
	parsed, err := strconv.ParseBool(boolString)
	if err != nil {
		return false, fmt.Errorf("value for key %s cannot be parsed: %w", key, err)
	}
	return parsed, nil
}

func parseDuration(cfg map[string]string, key string, defaultVal time.Duration) (time.Duration, error) {
	timeoutStr, exists := cfg[key]
	if !exists {
		return defaultVal, nil
	}
	timeout, err := time.ParseDuration(timeoutStr)
	if err != nil {
		return 0, fmt.Errorf("duration cannot be parsed: %w", err)
	}
	return timeout, nil
}

func checkRequired(cfg map[string]string) error {
	for _, reqKey := range Required {
		_, exists := cfg[reqKey]
		if !exists {
			return requiredConfigErr(reqKey)
		}
	}
	return nil
}

func requiredConfigErr(name string) error {
	return fmt.Errorf("%q config value must be set", name)
}

func split(serversString string) ([]string, error) {
	split := strings.Split(serversString, ",")
	servers := make([]string, 0)
	for i, s := range split {
		if strings.Trim(s, " ") == "" {
			return nil, fmt.Errorf("empty %d. server", i)
		}
		servers = append(servers, s)
	}
	return servers, nil
}
