package scoring

import (
	"encoding/binary"
	"fmt"

	"github.com/golang/leveldb"
)

var (
	storage *leveldb.DB
)

func incrementCounter(key string) (uint64, error) {
	var counter uint64
	storedCounter, err := storage.Get([]byte(key), nil)
	if err != nil {
		counter = 0
		storedCounter = make([]byte, 64)
	} else {
		counter = binary.BigEndian.Uint64(storedCounter) + 1
	}

	binary.BigEndian.PutUint64(storedCounter, counter)
	err = storage.Set([]byte(key), storedCounter, nil)
	if err != nil {
		return 0, err
	}
	return counter, nil
}

func getCounter(key string) (uint64, error) {
	storedCounter, err := storage.Get([]byte(key), nil)
	if err != nil {
		return 0, err
	} else {
		return binary.BigEndian.Uint64(storedCounter), nil
	}
}

func storeFinishedBatch(batch *Batch) {
	sourceKey := fmt.Sprintf("source_%d", batch.Idx)

	err := storage.Set([]byte(sourceKey), batch.Data, nil)
	check(err, fmt.Sprintf("Failed to store batch: %d", batch.Idx))

	resultKey := fmt.Sprintf("result_%d", batch.Idx)
	err = storage.Set([]byte(resultKey), batch.Result, nil)
	check(err, fmt.Sprintf("Failed to store batch %d results", batch.Idx))
}

func storeFailedBatch(batch *Batch) {
	failedKey := fmt.Sprintf("failed_%d", batch.Idx)

	err := storage.Set([]byte(failedKey), batch.Data, nil)
	check(err, fmt.Sprintf("Failed to store batch %d result", batch.Idx))
}
