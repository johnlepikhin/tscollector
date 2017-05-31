package config

import (
	"time"
	"encoding/json"
	"io/ioutil"
	"errors"
	"fmt"
)

type Auth struct {
	Username string
	Password string
	AllowAddValues bool
	AllowAddNewMeasures bool
	AllowReadValues bool
}

type Listen struct {
	Address string
	Auth []Auth
}

type ValueType string

const (
	IntLast ValueType = "IntLast"
	IntAvg = "IntAvg"
	FloatLast = "FloatLast"
)

type MeasureT struct {
	ID int
	Type ValueType
	Key MeasureKey
}

type Configuration struct {
	Listen           []Listen
	Storage          string
	TickSizeMs       time.Duration
	SavePeriodMs     time.Duration
	MaxGetIntervalMs time.Duration
}

var ConfigFile string

var Config Configuration = Configuration{
	TickSizeMs: time.Second,
	SavePeriodMs: time.Minute/time.Millisecond,
	MaxGetIntervalMs: time.Hour*24/time.Millisecond,
}

var ValuesFile string

var Values []MeasureT

func ValuesParser() error {
	raw, err := ioutil.ReadFile(ValuesFile)
	if err != nil {
		return err
	}

	err = json.Unmarshal(raw, &Values)
	if err != nil {
		return err
	}

	for id, v := range Values {
		v.ID = id
	}

	return nil
}

func ConfigParser() error {
	raw, err := ioutil.ReadFile(ConfigFile)
	if err != nil {
		return err
	}

	json.Unmarshal(raw, &Config)
	if err != nil {
		return err
	}

	if len(Config.Listen) == 0 {
		return errors.New("No listens defined in configuration file")
	}

	if (Config.Storage == "") {
		return errors.New("No storage defined in configuration file")
	}

	if (Config.TickSizeMs <= 0) {
		return errors.New("TickSize cannot be zero ot negative in configuration file")
	}

	if (Config.SavePeriodMs <= 0) {
		return errors.New("SavePeriod cannot be zero or negative in configuration file")
	}

	if (Config.SavePeriodMs < Config.TickSizeMs) {
		return errors.New("SavePeriod cannot smaller than TickSize in configuration file")
	}

	if (Config.SavePeriodMs % Config.TickSizeMs > 0) {
		return errors.New("SavePeriod must be multiple of TickSize in configuration file")
	}

	Config.TickSizeMs *= time.Millisecond
	Config.SavePeriodMs *= time.Millisecond
	Config.MaxGetIntervalMs *= time.Millisecond

	return nil
}

func ValueTypeParser(input string) (ValueType, error) {
	switch input {
	case "IntLast": return IntLast, nil
	case "IntAvg": return IntAvg, nil
	case "FloatLast": return FloatLast, nil
	}

	return IntLast, errors.New(fmt.Sprintf("Invalid type '%s'", input))
}

func AddNewMeasure(key MeasureKey, valueType string) error {
	parsedValueType, err := ValueTypeParser(valueType)
	if err != nil {
		return errors.New(fmt.Sprintf("Cannot create new measure '%s': %s", key, err.Error()))
	}

	id := len(Values)

	measure := MeasureT{
		ID: id,
		Type: parsedValueType,
		Key: key,
	}

	values := append(Values, measure)

	jsonValues, err := json.MarshalIndent(values, "", "  ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(ValuesFile, jsonValues, 0644)
	if err != nil {
		return err
	}

	ValuesParser()

	return nil
}