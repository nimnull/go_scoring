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
	listMux sync.Mutex
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

func storeBatchResult(batch *Batch, listKey, recordTpl string) {
	recordKey := fmt.Sprintf(recordTpl, batch.Idx)
	recordVal, err := json.Marshal(batch)

	err = storage.Set([]byte(recordKey), recordVal, nil)
	check(err, fmt.Sprintf("Failed to store batch %d", batch.Idx))

	storeBatchList(listKey, batch.Idx)
}


func storeBatchList(listName string, batchId int) error {
	listMux.Lock()
	defer listMux.Unlock()

	var parsedList []int
	listKey := []byte(listName)
	storedList, err := storage.Get(listKey, nil)
	if err != nil {
		parsedList = make([]int, 0)
	} else {
		err = json.Unmarshal(storedList, parsedList)
		return err
	}

	storedLen := len(parsedList)
	parsedList = make([]int, len(parsedList) + 1)
	parsedList[storedLen] = batchId

	storedList, err = json.Marshal(parsedList)
	if err != nil {
		return err
	}
	storage.Set(listKey, storedList, nil)

	return nil
}
