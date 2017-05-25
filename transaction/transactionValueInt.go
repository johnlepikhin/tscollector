package transaction

import (
	"strconv"
)

type TransactionValueIntT struct {
	TransactionValueT
	Value int64
}

func (v TransactionValueIntT) GetValue() int64 {
	return v.Value
}

func (v *TransactionValueIntT) SetValue(newval string) error {
	if parsedval, err := strconv.ParseInt(newval, 10, 64); err == nil {
		v.Value = parsedval
	} else {
		return err
	}

	return nil
}

func (v *TransactionValueIntT) Cleanup() {
	v.Value = 0
}

func (v TransactionValueIntT) Printable() string {
	return strconv.FormatInt(v.Value, 10)
}

func (v TransactionValueIntT) GetInt64() int64 {
	return v.GetValue()
}

func (v *TransactionValueIntT) SetInt64(value int64) {
	v.Value = value
}




type TransactionValueIntAvgT struct {
	TransactionValueIntT
	InsertionsCount int64
}

func (v *TransactionValueIntAvgT) SetValue(newval string) error {
	if parsedval, err := strconv.ParseInt(newval, 10, 64); err == nil {
		v.Value += parsedval
		v.InsertionsCount++
	} else {
		return err
	}

	return nil
}

func (v TransactionValueIntAvgT) GetValue() int64 {
	if v.InsertionsCount == 0 {
		return 0
	}

	return v.Value / v.InsertionsCount
}

func (v *TransactionValueIntAvgT) Cleanup() {
	v.Value = 0
	v.InsertionsCount = 0
}

func (v TransactionValueIntAvgT) GetInt64() int64 {
	return v.GetValue()
}

func (v *TransactionValueIntAvgT) SetInt64(value int64) {
	v.Value = value
	v.InsertionsCount = 1
}