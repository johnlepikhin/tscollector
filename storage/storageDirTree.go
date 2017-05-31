package storage

import (
	"fmt"
	"regexp"
	"strings"
	"strconv"
	"github.com/johnlepikhin/tscollector/transaction"
	"github.com/johnlepikhin/tscollector/config"
	"sort"
	"time"
	"os"
	"math"
)

const Version = 1

const GroupSize = 63

func Floor(tm time.Time, duration time.Duration) time.Time {
	return time.Unix(0, (tm.UnixNano() / duration.Nanoseconds()) * duration.Nanoseconds())
}

type StorageDirTreeT struct {
	RootDirectory string
	Levels        []time.Duration
}

type EncodedValue []byte

type GroupValue struct {
	EncodedValue
	transaction.TransactionValue
	IDInGroup int
}

type Group map[int]GroupValue

type Groups map[int]Group

func (storage *StorageDirTreeT) ParseConfiguration(input string) {
	var splitInput = regexp.MustCompile(`([^=]+)=([^;]*);?`)

	splittedInput := splitInput.FindAllStringSubmatch(input, -1)
	if splittedInput == nil {
		panic(fmt.Sprintf("Invalid params for storage 'directoryTree': %s", input))
	}

	for _, rec := range splittedInput {
		k := rec[1]
		v := rec[2]
		switch k {
		case "rootDirectory":
			storage.RootDirectory = v
		case "levels":
			splittedLevels := strings.Split(v, ",")
			for _, v := range splittedLevels {
				i, err := strconv.ParseUint(v, 10, 64)
				if err != nil {
					panic(fmt.Sprintf("Cannot parse level '%s' in configuration of storage 'directoryTree'. Unsigned integer expected, error was: %s", v, err.Error()))
				}
				level := time.Duration(i) * time.Second
				storage.Levels = append(storage.Levels, level)
			}
		}
	}

	if len(storage.RootDirectory) == 0 {
		panic("'rootDirectory=' cannot be empty in configuration of storage 'directoryTree'")
	}

	if len(storage.Levels) == 0 {
		panic("'levels=' cannot be empty in configuration of storage 'directoryTree'")
	}
}

func (Storage StorageDirTreeT) EncodeUInt64(v uint64) EncodedValue {
	if 249 > v {
		var b [1]byte
		b[0] = byte(v)
		return b[:]
	}

	var b [9]byte
	var length = 2

	b[1] = byte(v)
	b[2] = byte(v >> 8)
	if 1<<16 <= v { length++; b[3] = byte(v >> 16) }
	if 1<<24 <= v { length++; b[4] = byte(v >> 24) }
	if 1<<32 <= v { length++; b[5] = byte(v >> 32) }
	if 1<<40 <= v { length++; b[6] = byte(v >> 40) }
	if 1<<48 <= v { length++; b[7] = byte(v >> 48) }
	if 1<<56 <= v { length++; b[8] = byte(v >> 56) }

	b[0] = byte(247 + length)

	return b[:length+1]
}

func (Storage StorageDirTreeT) EncodeInt64(v int64) EncodedValue {

	if (v < 0) {
		return Storage.EncodeUInt64(uint64(((-v) << 1) | 1))
	}

	return Storage.EncodeUInt64(uint64(v) << 1)
}

func (Storage StorageDirTreeT) EncodeFloat64(v float64) EncodedValue {
	return Storage.EncodeUInt64(math.Float64bits(v))
}


func (storage StorageDirTreeT) Encode(v transaction.TransactionValue) EncodedValue {
	dataType := config.ValueTypeToDataType(v.GetConfigValue().Type)

	switch dataType {
	case config.Int64:
		var intval = v.GetInt64()
		if intval == 0 {
			return nil
		}
		return storage.EncodeInt64(intval)
	case config.Float64:
		var floatVal = v.GetFloat64()
		if floatVal == 0 {
			return nil
		}
		return storage.EncodeFloat64(floatVal)
	}

	return nil
}

func (storage StorageDirTreeT) prepareGroups(transaction transaction.Transaction) Groups {
	var groups = make(Groups)

	for _, v := range transaction.GetValues() {
		var encoded = storage.Encode(v)
		if encoded == nil {
			continue
		}

		var id = v.GetConfigValue().ID
		var groupId = id / GroupSize
		var valueInGroupID = id % GroupSize
		var groupValue = GroupValue{ encoded, v, valueInGroupID }

		if group, ok := groups[groupId]; ok {
			group[valueInGroupID] = groupValue
		} else {
			var group = make(Group)
			group[valueInGroupID] = groupValue
			groups[groupId] = group
		}
	}

	return groups
}

func timeListReverse (lst []time.Time) []time.Time {
	lstlen := len(lst)
	var ret = make([]time.Time, lstlen)

	for i, elt := range lst {
		ret[lstlen-i-1] = elt
	}

	return ret
}

func durationListReverse (lst []time.Duration) []time.Duration {
	lstlen := len(lst)
	var ret = make([]time.Duration, lstlen)

	for i, elt := range lst {
		ret[lstlen-i-1] = elt
	}

	return ret
}

func (storage StorageDirTreeT) GenerateLevels(tm time.Time) []time.Time {
	revLst := durationListReverse(storage.Levels)

	var ret = make([]time.Time, len(revLst))
	for pos, level := range revLst {
		ret[pos] = Floor(tm, level)
	}

	return timeListReverse(ret)
}

func (storage StorageDirTreeT) GenerateDirectories(levels []time.Time) string {
	levelsStrings := []string{}
	for _, i := range levels {
		levelsStrings = append(levelsStrings, strconv.FormatInt(i.Unix(), 10))
	}

	directories := storage.RootDirectory + string(os.PathSeparator) +
		strings.Join(levelsStrings[:len(levelsStrings) - 1], string(os.PathSeparator))

	return directories
}

func (storage StorageDirTreeT) GenerateFileNamePrefix(levels []time.Time) string {
	return strconv.FormatInt(levels[len(levels)-1].Unix(), 10)
}

func (storage StorageDirTreeT) PrepareDirectories(levels []time.Time) string {
	directories := storage.GenerateDirectories(levels)

	os.MkdirAll(directories, 0755)

	return directories
}

func (storage StorageDirTreeT) SaveGroup(groupId int, group Group) {
	var mask uint64
	var ids = make([]int, 0)
	for valueId := range group {
		mask |= 1 << uint(valueId)
		ids = append(ids, valueId)
	}

	sort.Ints(ids)

	var encodedValues []byte
	for _, valueIDInGroup := range ids {
		var value = group[valueIDInGroup]
		encodedValues = append(encodedValues, value.EncodedValue...)
	}

	mask_bytes := storage.EncodeUInt64(mask)

	encodedValues = append(mask_bytes, encodedValues...)

	// Generate header

	first_byte := []byte{0xa0 | (Version << 2)}

	now_time := time.Now()
	file_start_time := Floor(now_time, storage.Levels[len(storage.Levels)-1])
	now_time_tick_offset := (now_time.UnixNano() - file_start_time.UnixNano()) / config.Config.TickSize.Nanoseconds()

	now_time_tick_offset_bytes := storage.EncodeUInt64(uint64(now_time_tick_offset))

	length_bytes := storage.EncodeUInt64(uint64(len(encodedValues)))

	record := append(first_byte, now_time_tick_offset_bytes...)
	record = append(record, length_bytes...)
	record = append(record, encodedValues...)

	levels := storage.GenerateLevels(now_time)

	directories := storage.PrepareDirectories(levels)

	fname := storage.GenerateFileNamePrefix(levels) + "_" + strconv.Itoa(groupId)
	fname = directories + string(os.PathSeparator) + fname

	f, err := os.OpenFile(fname, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0640)

	if err != nil {
		panic(err)
	}

	defer f.Close()

	if _, err = f.Write(record); err != nil {
		panic(err)
	}
}

func (storage StorageDirTreeT) SaveGroups(groups Groups) {
	for groupId, group := range groups {
		storage.SaveGroup(groupId, group)
	}
}


func (storage StorageDirTreeT) SaveTransaction(transaction transaction.Transaction) {
	var groups = storage.prepareGroups(transaction)

	storage.SaveGroups(groups)
}