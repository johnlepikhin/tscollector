package transaction

import (
	"tscollector/config"
	"time"
	"errors"
)

type InputValue struct {
	Value string
	Type string
}

type InputValues map[config.MeasureKey]InputValue

type Transaction interface {
	GetValues() map[config.MeasureKey]TransactionValue
	GetTimeStamp() time.Time
	AddPlainValue(auth config.Auth, key config.MeasureKey, value InputValue) error
	AddPlainValues(auth config.Auth, values InputValues) error
	SetTimeStamp(timestamp time.Time)
	PutValue(key config.MeasureKey, value TransactionValue)
	Cleanup()
	MergeValues(transaction Transaction)
}

type TransactionT struct {
	TimeStamp time.Time
	Values map[config.MeasureKey]TransactionValue
}

type TimeSeries map[time.Time]Transaction

func (transaction TransactionT) addValue(configValue config.MeasureT, key config.MeasureKey, value InputValue) error {
	if transactionValue, ok := transaction.Values[key]; ok {
		return transactionValue.SetValue(value.Value)
	} else {
		var transactionValue TransactionValue
		switch configValue.Type {
		case config.IntLast:
			transactionValue = new(TransactionValueIntT)
		case config.IntAvg:
			transactionValue = new(TransactionValueIntAvgT)
		case config.FloatLast:
			transactionValue = new(TransactionValueFloatT)
		}
		err := transactionValue.SetValue(value.Value)
		if err != nil {
			return err
		}
		transactionValue.SetConfigValue(configValue)
		transaction.Values[key] = transactionValue
	}

	return nil
}

func (transaction TransactionT) GetValues() map[config.MeasureKey]TransactionValue {
	return transaction.Values
}

func (transaction TransactionT) AddPlainValue(auth config.Auth, key config.MeasureKey, value InputValue) error {
	for _, configValue := range config.Values {
		if configValue.Key == key {
			return transaction.addValue(configValue, key, value)
		}
	}


	if !auth.AllowAddNewMeasures {
		return errors.New(string("Measure '" + key + "' is unknown and no access to add new measures"))
	}
	err := config.AddNewMeasure(key, value.Type)
	if err != nil {
		return err
	}

	transaction.AddPlainValue(auth, key, value)

	return nil
}

func (transaction TransactionT) AddPlainValues(auth config.Auth, values InputValues) error {
	for k, v := range values {
		err := transaction.AddPlainValue(auth, k, v)
		if err != nil {
			return err
		}
	}

	return nil
}

func (transaction TransactionT) Cleanup() {
	for _, v := range transaction.Values {
		v.Cleanup()
	}
}

func (transaction *TransactionT) SetTimeStamp(timestamp time.Time) {
	transaction.TimeStamp = timestamp
}

func (transaction TransactionT) GetTimeStamp() time.Time {
	return transaction.TimeStamp
}

func (transaction TransactionT) PutValue(key config.MeasureKey, value TransactionValue) {
	transaction.Values[key] = value
}

func (transaction1 *TransactionT) MergeValues(transaction2 Transaction) {
	for _, v := range transaction2.GetValues() {
		transaction1.Values[v.GetConfigValue().Key] = v
	}
}

func NewTransaction() Transaction {
	return &TransactionT{
		TimeStamp: time.Unix(0, 0),
		Values: make(map[config.MeasureKey]TransactionValue),
	}
}