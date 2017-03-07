package scoring

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/golang/leveldb"
)

var (
	storage *leveldb.DB
	cntMux  sync.Mutex
)

func incrementCounter(key string) (uint64, error) {
	var counter uint64
	cntMux.Lock()
	defer cntMux.Unlock()

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
	sourceKey := fmt.Sprintf("finished_%d", batch.Idx)
	sourceVal, err := json.Marshal(batch)

	err = storage.Set([]byte(sourceKey), sourceVal, nil)
	check(err, fmt.Sprintf("Failed to store batch: %d", string(sourceVal)))
}

func storeFailedBatch(batch *Batch) {
	var failedListParsed []int
	failedKey := fmt.Sprintf("failed_%d", batch.Idx)
	listKey := []byte(FAILED_LIST_KEY)
	failedVal, err := json.Marshal(batch)

	err = storage.Set([]byte(failedKey), failedVal, nil)
	check(err, fmt.Sprintf("Failed to store batch %d result", batch.Idx))

	failedList, err := storage.Get(listKey, nil)

	err = json.Unmarshal(failedList, failedListParsed)
	if err != nil {
		failedListParsed = []int{}
	}
	failedList, err = json.Marshal(failedListParsed)
	fmt.Printf("Failed batches: %#v\n", failedList)
	storage.Set(listKey, failedList, nil)

}
