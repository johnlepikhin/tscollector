package transaction

import (
	"strconv"
)

type TransactionValueIntT struct {
	TransactionValueT
	Value uint64
}

func (v TransactionValueIntT) GetValue() uint64 {
	return v.Value
}

func (v *TransactionValueIntT) SetValue(newval string) error {
	if parsedval, err := strconv.ParseUint(newval, 10, 64); err == nil {
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
	return strconv.FormatUint(v.Value, 10)
}

func (v TransactionValueIntT) GetUInt64() uint64 {
	return v.GetValue()
}

func (v *TransactionValueIntT) SetUInt64(value uint64) {
	v.Value = value
}




type TransactionValueIntAvgT struct {
	TransactionValueIntT
	InsertionsCount uint64
}

func (v *TransactionValueIntAvgT) SetValue(newval string) error {
	if parsedval, err := strconv.ParseUint(newval, 10, 64); err == nil {
		v.Value += parsedval
		v.InsertionsCount++
	} else {
		return err
	}

	return nil
}

func (v TransactionValueIntAvgT) GetValue() uint64 {
	if v.InsertionsCount == 0 {
		return 0
	}

	return v.Value / v.InsertionsCount
}

func (v *TransactionValueIntAvgT) Cleanup() {
	v.Value = 0
	v.InsertionsCount = 0
}

func (v TransactionValueIntAvgT) GetUInt64() uint64 {
	return v.GetValue()
}

func (v *TransactionValueIntAvgT) SetUInt64(value uint64) {
	v.Value = value
	v.InsertionsCount = 1
}