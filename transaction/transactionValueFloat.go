package transaction

import "strconv"

type TransactionValueFloatT struct {
	TransactionValueT
	Value float64
}

func (v TransactionValueFloatT) GetValue() float64 {
	return v.Value
}

func (v *TransactionValueFloatT) SetValue(newval string) error {
	if parsedval, err := strconv.ParseFloat(newval, 64); err == nil {
		v.Value = parsedval
	} else {
		return err
	}

	return nil
}

func (v *TransactionValueFloatT) Cleanup() {
	v.Value = 0
}

func (v TransactionValueFloatT) Printable() string {
	return strconv.FormatFloat(v.Value, 'f', 5, 64)
}

func (v TransactionValueFloatT) GetFloat64() float64 {
	return v.GetValue()
}

func (v *TransactionValueFloatT) SetFloat64(value float64) {
	v.Value = value
}
