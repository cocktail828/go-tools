package pulsarmq

import (
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/cocktail828/go-tools/z/variadic"
)

type subscriptionTypeKey struct{}

// default Exclusive
func SubscriptionType(val pulsar.SubscriptionType) variadic.Option {
	return variadic.Set(subscriptionTypeKey{}, val)
}
func getSubscriptionType(c variadic.Container) pulsar.SubscriptionType {
	return variadic.Value[pulsar.SubscriptionType](c, subscriptionTypeKey{})
}

type dlqPolicyKey struct{}

func DLQPolicy(val *pulsar.DLQPolicy) variadic.Option { return variadic.Set(dlqPolicyKey{}, val) }
func getDLQPolicy(c variadic.Container) *pulsar.DLQPolicy {
	return variadic.Value[*pulsar.DLQPolicy](c, dlqPolicyKey{})
}

type disableBatchKey struct{}

func DisableBatch() variadic.Option             { return variadic.Set(disableBatchKey{}, true) }
func getDisableBatch(c variadic.Container) bool { return variadic.Value[bool](c, disableBatchKey{}) }

type compressTypeKey struct{}

func CompressType(val pulsar.CompressionType) variadic.Option {
	return variadic.Set(compressTypeKey{}, val)
}
func getCompressType(c variadic.Container) pulsar.CompressionType {
	return variadic.Value[pulsar.CompressionType](c, compressTypeKey{})
}

type compressLevelKey struct{}

func CompressLevel(val pulsar.CompressionLevel) variadic.Option {
	return variadic.Set(compressLevelKey{}, val)
}
func getCompressLevel(c variadic.Container) pulsar.CompressionLevel {
	return variadic.Value[pulsar.CompressionLevel](c, compressLevelKey{})
}
