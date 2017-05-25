package config

type DataType int

type MeasureKey string

const (
	Int64 DataType = iota
	Float64 = iota
)

func ValueTypeToDataType(valueType ValueType) DataType {
	switch valueType {
	case IntAvg:
		return Int64
	case IntLast:
		return Int64
	case FloatLast:
		return Float64
	}

	return -1
}
