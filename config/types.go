package config

type DataType int

type MeasureKey string

const (
	UInt64 DataType = iota
)

func ValueTypeToDataType(valueType ValueType) DataType {
	switch valueType {
	case IntAvg:
		return UInt64
	case IntLast:
		return UInt64
	}

	return -1
}
