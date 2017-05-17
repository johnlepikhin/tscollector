package storage

import (
	"tscollector/transaction"
	"regexp"
	"fmt"
	"time"
	"errors"
)

type StorageI interface {
	ParseConfiguration(input string)
	SaveTransaction(t transaction.Transaction)
	LoadTimeSeries(start time.Time, end time.Time) (transaction.TimeSeries, error)
}

var Storage StorageI

func StorageParser(input string) error {
	var splitInput = regexp.MustCompile(`^([^:]+):(.*)`)

	splittedInput := splitInput.FindStringSubmatch(input)
	if splittedInput == nil {
		panic("Invalid storage format in configuration file")
	}

	if (splittedInput[1] == "directoryTree") {
		Storage = new(StorageDirTreeT)
		Storage.ParseConfiguration(splittedInput[2])
	} else {
		return errors.New(fmt.Sprintf("Invalid storage type '%s' in configuration file", splittedInput[1]))
	}

	return nil
}
