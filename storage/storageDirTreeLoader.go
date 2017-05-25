package storage

import (
	"bufio"
	"tscollector/transaction"
	"errors"
	"tscollector/config"
	"time"
	"math"
	"strconv"
	"os"
)

func (storage StorageDirTreeT) ReadUInt64(stream *bufio.Reader) (uint64, error) {
	first_byte, err := stream.ReadByte()
	if err != nil {
		return 0, err
	}

	if first_byte < 249 {
		return uint64(first_byte), nil
	}

	length := int(first_byte - 247)

	ret := uint64(0)

	for pos:=0; pos<length; pos++ {
		b, err := stream.ReadByte()
		if err != nil {
			return 0, err
		}

		ret |= uint64(b) << uint64(8*pos)
	}

	return ret, nil
}

func (storage StorageDirTreeT) ReadInt64(stream *bufio.Reader) (int64, error) {
	v, err := storage.ReadUInt64(stream)
	if err != nil {
		return 0, err
	}

	if v & 1 == 1 {
		return -int64(v >> 1), nil
	}

	return int64(v) >> 1, nil
}


func (storage StorageDirTreeT) ReadFloat64(stream *bufio.Reader) (float64, error) {
	bits, err := storage.ReadUInt64(stream)
	if err != nil {
		return 0, err
	}

	r := math.Float64frombits(bits)
	return r, nil
}

type bit int

type configValuesPosition int

type bitmask map[bit]configValuesPosition

type group struct {
	ID   uint64
	Mask bitmask
}

type groups []group

func (storage StorageDirTreeT) LoadGroup(stream *bufio.Reader, fileTimeOffset time.Time, bitmap bitmask) (transaction.Transaction, error) {
	record := transaction.NewTransaction()

	first_byte, err := stream.ReadByte()
	if err != nil {
		return record, err
	}

	if first_byte & 0xf0 != 0xa0 {
		return record, errors.New("Invalid record header")
	}

	if (first_byte & 0xf) >> 2 != Version {
		return record, errors.New("Invalid record version")
	}

	recordTickOffset, err := storage.ReadUInt64(stream)
	if err != nil {
		return record, err
	}

	_, err = storage.ReadUInt64(stream)
	if err != nil {
		return record, err
	}

	mask, err := storage.ReadUInt64(stream)
	if err != nil {
		return record, err
	}

	for bit, configValuesPosition := range bitmap {
		configValue := config.Values[configValuesPosition]

		if (1 << uint(bit)) & mask == 0 {
			continue
		}

		var transactionValue transaction.TransactionValue

		switch config.ValueTypeToDataType(configValue.Type) {
		case config.Int64:
			value, err := storage.ReadInt64(stream)
			if err != nil {
				return record, err
			}

			switch configValue.Type {
			case config.IntAvg:
				transactionValue = new(transaction.TransactionValueIntAvgT)
			case config.IntLast:
				transactionValue = new(transaction.TransactionValueIntT)
			}
			transactionValue.SetInt64(value)

		case config.Float64:
			value, err := storage.ReadFloat64(stream)
			if err != nil {
				return record, err
			}

			switch configValue.Type {
			case config.FloatLast:
				transactionValue = new(transaction.TransactionValueFloatT)
			}
			transactionValue.SetFloat64(value)
		}
		transactionValue.SetConfigValue(configValue)
		record.PutValue(configValue.Key, transactionValue)
	}

	timestamp := fileTimeOffset.Add(time.Duration(int64(recordTickOffset))*config.Config.TickSize)
	record.SetTimeStamp(timestamp)

	return record, nil
}

func makeBitMask(groupID uint64) bitmask {
	ret := bitmask{}

	minID := int(groupID*GroupSize)
	maxID := int(minID+GroupSize-1)

	for id := range config.Values {
		if id < minID || id > maxID {
			continue
		}

		var idInGroup = bit(id - minID)

		ret[idInGroup] = configValuesPosition(id)
	}

	return ret
}

func makeGroup(groupID uint64) group {
	return group{
		ID: groupID,
		Mask: makeBitMask(groupID),
	}
}

func makeGroups() groups {
	maxID := 0
	for _, configValue := range config.Values {
		if configValue.ID > maxID {
			maxID = configValue.ID
		}
	}

	maxGroupID := uint64(math.Ceil(float64(maxID) / float64(GroupSize)))

	groups := []group{}
	for groupID:=uint64(0); groupID<=maxGroupID; groupID++ {
		group := makeGroup(groupID)
		groups = append(groups, group)
	}

	return groups
}

func (storage StorageDirTreeT) generageFullFileNamePrefix(levels []time.Time) string {
	directories := storage.GenerateDirectories(levels)
	fname := storage.GenerateFileNamePrefix(levels)

	return directories + string(os.PathSeparator) + fname + "_"
}

func (storage StorageDirTreeT) LoadTimeSeries(start time.Time, end time.Time) (transaction.TimeSeries, error) {
	groups := makeGroups()

	if start.UnixNano() >= end.UnixNano() {
		return transaction.TimeSeries{}, errors.New("Cannot load timeseries: start time >= end time")
	}

	fileStep := storage.Levels[len(storage.Levels)-1]

	timeSeries := transaction.TimeSeries{}

	for timeStamp:=start; timeStamp.UnixNano()<=end.UnixNano(); timeStamp=time.Unix(0, int64(fileStep)+timeStamp.UnixNano()) {
		levels := storage.GenerateLevels(timeStamp)
		filePrefix := storage.generageFullFileNamePrefix(levels)

		LoopGroup:
		for _, group := range groups {
			fileName := filePrefix + strconv.Itoa(int(group.ID))

			file, err := os.Open(fileName)
			if err != nil {
				continue LoopGroup
			}

			stream := bufio.NewReader(file)

			LoopRecord:
			for {
				readTransaction, err := storage.LoadGroup(stream, levels[len(levels)-1], group.Mask)
				if err != nil {
					break LoopRecord
				}

				if readTransaction.GetTimeStamp().UnixNano() > end.UnixNano() {
					break LoopRecord
				}

				if readTransaction.GetTimeStamp().UnixNano() < start.UnixNano() {
					continue LoopRecord
				}

				roundedTimeStamp := readTransaction.GetTimeStamp()

				_, ok := timeSeries[roundedTimeStamp]
				if ok {
					timeSeries[roundedTimeStamp].MergeValues(readTransaction)
				} else {
					timeSeries[roundedTimeStamp] = readTransaction
				}
			}
			file.Close()
		}
	}

	return timeSeries, nil
}
