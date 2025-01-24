package pulsarmq

import (
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/cocktail828/go-tools/z/variadic"
)

type inVariadic struct{ variadic.Assigned }

type subscriptionTypeKey struct{}

// default Exclusive
func SubscriptionType(val pulsar.SubscriptionType) variadic.Option {
	return variadic.SetValue(subscriptionTypeKey{}, val)
}
func (iv inVariadic) SubscriptionType() pulsar.SubscriptionType {
	return variadic.GetValue[pulsar.SubscriptionType](iv, subscriptionTypeKey{})
}

type dlqPolicyKey struct{}

func DLQPolicy(val *pulsar.DLQPolicy) variadic.Option { return variadic.SetValue(dlqPolicyKey{}, val) }
func (iv inVariadic) DLQPolicy() *pulsar.DLQPolicy {
	return variadic.GetValue[*pulsar.DLQPolicy](iv, dlqPolicyKey{})
}

type disableBatchKey struct{}

func DisableBatch() variadic.Option      { return variadic.SetValue(disableBatchKey{}, true) }
func (iv inVariadic) DisableBatch() bool { return variadic.GetValue[bool](iv, disableBatchKey{}) }

type compressTypeKey struct{}

func CompressType(val pulsar.CompressionType) variadic.Option {
	return variadic.SetValue(compressTypeKey{}, val)
}
func (iv inVariadic) CompressType() pulsar.CompressionType {
	return variadic.GetValue[pulsar.CompressionType](iv, compressTypeKey{})
}

type compressLevelKey struct{}

func CompressLevel(val pulsar.CompressionLevel) variadic.Option {
	return variadic.SetValue(compressLevelKey{}, val)
}
func (iv inVariadic) CompressLevel() pulsar.CompressionLevel {
	return variadic.GetValue[pulsar.CompressionLevel](iv, compressLevelKey{})
}
