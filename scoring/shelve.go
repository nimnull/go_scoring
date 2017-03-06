package scoring

import (
	"encoding/binary"
	"fmt"
	"github.com/golang/leveldb"
)


var (
	db *leveldb.DB
)


func incrementCounter(key string)  error {
	storedCounter, err := db.Get([]byte(key), nil)
	if err != nil {
		return err
	}
	counter := binary.BigEndian.Uint64(storedCounter) + 1
	binary.BigEndian.PutUint64(storedCounter, counter)
	err = db.Set([]byte(key), storedCounter, nil)
	if err != nil {
		return err
	}
	return nil
}

func getCounter(key string) (uint64, error) {
	storedCounter, err := db.Get([]byte(key), nil)
	if err != nil {
		return 0, err
	} else {
		return binary.BigEndian.Uint64(storedCounter), nil
	}
}


func storeBatchPred( idx int, source, prediction []byte) {
	err := db.Set([]byte(fmt.Sprintf("source_%d", idx)), source,nil)
	check(err, fmt.Sprintf("Failed to store batch: %d", idx))


	err = db.Set([]byte(fmt.Sprintf("result_%d", idx)), prediction,nil)
	check(err, fmt.Sprintf("Failed to store batch %d results", idx))

	err = incrementCounter("batches")
	check(err, "Failed to increment counter")
}
