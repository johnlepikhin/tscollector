package transaction

import "github.com/johnlepikhin/tscollector/config"

type TransactionValue interface {
	SetValue(value string) error
	GetConfigValue() config.MeasureT
	SetConfigValue(config.MeasureT)
	Printable() string
	Cleanup()

	GetInt64() int64
	SetInt64(value int64)
	GetFloat64() float64
	SetFloat64(value float64)
}

type TransactionValueT struct {
	ConfigValue config.MeasureT
	Value interface{}
}

func (v TransactionValueT) GetConfigValue() config.MeasureT {
	return v.ConfigValue
}

func (v *TransactionValueT) SetConfigValue(configValue config.MeasureT) {
	v.ConfigValue = configValue
}

func (transaction TransactionValueT) GetInt64() int64 {
	panic("Call to unimplemented GetInt64")
}

func (transaction TransactionValueT) SetInt64(value int64) {
	panic("Call to unimplemented SetInt64")
}

func (transaction TransactionValueT) GetFloat64() float64 {
	panic("Call to unimplemented GetFloat64")
}

func (transaction TransactionValueT) SetFloat64(value float64) {
	panic("Call to unimplemented SetFloat64")
}
