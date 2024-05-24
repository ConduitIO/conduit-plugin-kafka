// Code generated by paramgen. DO NOT EDIT.
// Source: github.com/ConduitIO/conduit-connector-sdk/tree/main/cmd/paramgen

package destination

import (
	sdk "github.com/conduitio/conduit-connector-sdk"
)

func (Config) Parameters() map[string]sdk.Parameter {
	return map[string]sdk.Parameter{
		"acks": {
			Default:     "all",
			Description: "acks defines the number of acknowledges from partition replicas required before receiving a response to a produce request. None = fire and forget, one = wait for the leader to acknowledge the writes, all = wait for the full ISR to acknowledge the writes.",
			Type:        sdk.ParameterTypeString,
			Validations: []sdk.Validation{
				sdk.ValidationInclusion{List: []string{"none", "one", "all"}},
			},
		},
		"batchBytes": {
			Default:     "1000012",
			Description: "batchBytes limits the maximum size of a request in bytes before being sent to a partition. This mirrors Kafka's max.message.bytes.",
			Type:        sdk.ParameterTypeInt,
			Validations: []sdk.Validation{},
		},
		"caCert": {
			Default:     "",
			Description: "caCert is the Kafka broker's certificate.",
			Type:        sdk.ParameterTypeString,
			Validations: []sdk.Validation{},
		},
		"clientCert": {
			Default:     "",
			Description: "clientCert is the Kafka client's certificate.",
			Type:        sdk.ParameterTypeString,
			Validations: []sdk.Validation{},
		},
		"clientID": {
			Default:     "conduit-connector-redpanda",
			Description: "clientID is a unique identifier for client connections established by this connector.",
			Type:        sdk.ParameterTypeString,
			Validations: []sdk.Validation{},
		},
		"clientKey": {
			Default:     "",
			Description: "clientKey is the Kafka client's private key.",
			Type:        sdk.ParameterTypeString,
			Validations: []sdk.Validation{},
		},
		"compression": {
			Default:     "snappy",
			Description: "compression set the compression codec to be used to compress messages.",
			Type:        sdk.ParameterTypeString,
			Validations: []sdk.Validation{
				sdk.ValidationInclusion{List: []string{"none", "gzip", "snappy", "lz4", "zstd"}},
			},
		},
		"deliveryTimeout": {
			Default:     "",
			Description: "deliveryTimeout for write operation performed by the Writer.",
			Type:        sdk.ParameterTypeDuration,
			Validations: []sdk.Validation{},
		},
		"insecureSkipVerify": {
			Default:     "",
			Description: "insecureSkipVerify defines whether to validate the broker's certificate chain and host name. If 'true', accepts any certificate presented by the server and any host name in that certificate.",
			Type:        sdk.ParameterTypeBool,
			Validations: []sdk.Validation{},
		},
		"saslMechanism": {
			Default:     "",
			Description: "saslMechanism configures the connector to use SASL authentication. If empty, no authentication will be performed.",
			Type:        sdk.ParameterTypeString,
			Validations: []sdk.Validation{
				sdk.ValidationInclusion{List: []string{"PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512"}},
			},
		},
		"saslPassword": {
			Default:     "",
			Description: "saslPassword sets up the password used with SASL authentication.",
			Type:        sdk.ParameterTypeString,
			Validations: []sdk.Validation{},
		},
		"saslUsername": {
			Default:     "",
			Description: "saslUsername sets up the username used with SASL authentication.",
			Type:        sdk.ParameterTypeString,
			Validations: []sdk.Validation{},
		},
		"servers": {
			Default:     "",
			Description: "servers is a list of Kafka bootstrap servers, which will be used to discover all the servers in a cluster.",
			Type:        sdk.ParameterTypeString,
			Validations: []sdk.Validation{
				sdk.ValidationRequired{},
			},
		},
		"tls.enabled": {
			Default:     "",
			Description: "tls.enabled defines whether TLS is needed to communicate with the Kafka cluster.",
			Type:        sdk.ParameterTypeBool,
			Validations: []sdk.Validation{},
		},
		"topic": {
			Default:     "{{ index .Metadata \"opencdc.collection\" }}",
			Description: "topic is the Kafka topic. It can contain a [Go template](https://pkg.go.dev/text/template) that will be executed for each record to determine the topic. By default, the topic is the value of the `opencdc.collection` metadata field.",
			Type:        sdk.ParameterTypeString,
			Validations: []sdk.Validation{},
		},
	}
}
