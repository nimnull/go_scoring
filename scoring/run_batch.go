package scoring

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"unicode/utf8"

	"github.com/buger/jsonparser"
	"github.com/golang/leveldb"

	"go_scoring/csv"
)

type Batch struct {
	Idx    int    `json:"id"`
	Data   []byte `json:"data"`
	Result []byte `json:"result,omitempty"`
	isLast bool   `json:"-"`
}

const (
	MAX_BATCH_SIZE  = 2 * (1024 ^ 2)
	CLIENT_HEADERS  = "datarobot_batch_scoring/%s|Golang/%s|system/%s"
	BATCH_COUNTER   = "batches_cnt"
	FAILED_LIST_KEY = "failed"
	SHELVE_PATH     = "./shelve"
)

var (
	predictionEndpoint string
	totalBatches       int
)

func getHttpHeaders(compression bool) map[string]string {
	headers := make(map[string]string)
	headers["Content-Type"] = "text/csv"
	headers["User-Agent"] = fmt.Sprintf(CLIENT_HEADERS, "0.1", runtime.Version(), runtime.GOOS)
	if compression {
		headers["Content-Encoding"] = "gzip"
	}

	return headers
}

func check(e error, msg string) {
	if e != nil {
		log.Fatalf("%s: %s\n", msg, e)
	}
}

func getCSVInput(file *os.File) (*csvtools.CSVInput, string) {
	dt := csvtools.NewDetector()
	delimiters := dt.DetectDelimiter(file, '"')

	r, _ := utf8.DecodeRuneInString(delimiters[0])
	// rewind file after detector finished it's work
	file.Seek(0, 0)
	opts := csvtools.CSVInputOptions{true, r, file}
	csvInput, err := csvtools.NewCSVInput(&opts)
	check(err, "Failed to initialize csv reader")

	return csvInput, delimiters[0]
}

func requestScoring(compression bool, data []byte) (int, []byte) {
	req, err := http.NewRequest("POST", predictionEndpoint, bytes.NewReader(data))
	check(err, "Failed to prepare request")
	for k, v := range getHttpHeaders(compression) {
		req.Header.Set(k, v)
	}
	netClient := GetHttpClient()
	resp, err := netClient.Do(req)
	for err != nil {
		req, err := http.NewRequest("POST", predictionEndpoint, bytes.NewReader(data))
		log.Printf("Failed to send request %s. %s\n", data, err)
		netClient := GetHttpClient()
		resp, err = netClient.Do(req)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	check(err, "Failed to read response")

	return resp.StatusCode, body
}

func processStatusCodes(statusCode int, respBody []byte) {
	switch statusCode {
	case 400, 404:
		var errorMsg string
		if msg, err := jsonparser.GetString(respBody, "message"); err != nil {
			errorMsg = string(respBody)

		} else {
			errorMsg = msg
		}
		log.Fatalf("Failed with client error: %s\n", errorMsg)
	case 403:
		log.Fatalf("Failed with message:\n\t%s\n", respBody)
	case 401:
		log.Fatalf("failed to authenticate -- "+
			"please check your: datarobot_key (if required), "+
			"username/password and/or api token. Contact "+
			"customer support if the problem persists "+
			"message:\n%s\n", respBody)
	case 405:
		log.Fatalln("failed to request endpoint -- please check your " +
			"'--host' argument")
	case 502:
		log.Fatalln("problem with the gateway -- please check your " +
			"'--host' argument and contact customer support" +
			"if the problem persists.")
	}
}

func RunBatch(
	baseUrl *url.URL, importId, dataset, encoding, delimiter string,
	maxBatchSize, concurrent int, compression, fastMode bool) {

	if maxBatchSize == 0 || maxBatchSize > MAX_BATCH_SIZE {
		maxBatchSize = MAX_BATCH_SIZE
	}

	baseUrl.Path = fmt.Sprintf("%s/%s/predict", baseUrl.Path, importId)
	predictionEndpoint = baseUrl.String()
	encoding = investigateEncoding(dataset, encoding, delimiter, delimiter)

	file, err := os.Open(dataset)
	check(err, "Failed to open predict source")
	defer file.Close()
	csvInput, delimiter := getCSVInput(file)

	csvHeader := csvInput.Header()
	buff := bytes.NewBufferString(strings.Join(csvHeader, ",") + "\n")
	firstRow := csvInput.ReadRecord()
	buff.WriteString(strings.Join(firstRow, ",") + "\n")

	statusCode, body := requestScoring(false, buff.Bytes())
	processStatusCodes(statusCode, body)

	if execTime, err := jsonparser.GetInt(body, "execution_time"); err != nil {
		log.Printf("Failed to read execution time: %s\n", err)
	} else {
		fmt.Printf("Execution time: %d\n", execTime)
	}

	queue := make(chan Batch, 100)
	finisher := make(chan bool, 1)
	// init storage handler
	storage, err = leveldb.Open(SHELVE_PATH, nil)
	check(err, "DB Error:")
	defer storage.Close()
	defer os.Remove(SHELVE_PATH)

	for j := 0; j < concurrent; j++ {
		go batchSender(j, queue, finisher)
	}

	for i := 0; ; i++ {
		isLast := false
		totalBatches = i

		buff := bytes.NewBufferString(strings.Join(csvHeader, ",") + "\n")

		for buff.Len() < MAX_BATCH_SIZE {
			line := csvInput.ReadRecord()

			if line != nil {
				var escapedLine []string
				for _, record := range line {
					if strings.Contains(record, delimiter) {
						escapedLine = append(escapedLine, fmt.Sprintf("\"%s\"", record))
					} else {
						escapedLine = append(escapedLine, record)
					}
				}
				buff.WriteString(strings.Join(escapedLine, ",") + "\n")
			} else {
				isLast = true
				break
			}
		}

		queue <- Batch{Idx: i, Data: buff.Bytes(), isLast: isLast}

		if isLast {
			log.Println("Last batch sent", i)
			close(queue)
			break
		}

	}

	<-finisher

	var failedListParsed []int
	failedList, err := storage.Get([]byte("failed"), nil)
	err = json.Unmarshal(failedList, failedListParsed)
	fmt.Printf("Failed batches: %#v\n", len(failedList))

	if len(failedList) {

	}

	fmt.Println("results assembling goes here")

	//fmt.Printf("Resp: %#v\n", string(body))

	//
	//fmt.Println(csvInput)

	//firstRow := peekRow(dataset, encoding, delimiter, fastMode)
	//fmt.Println(firstRow)
	//queueSize := concurrent * 2
	//r.Header.Set("Authorization", "Basic "+basicAuth(username, password))
	//auth := username + ":" + password
	//return base64.StdEncoding.EncodeToString([]byte(auth))

}

func batchSender(idx int, queue <-chan Batch, finisher chan bool) {
	fmt.Println("Spawned worker", idx)
	for batch := range queue {
		statusCode, body := requestScoring(false, batch.Data)
		if statusCode == 200 {
			batch.Result = body
			storeFinishedBatch(&batch)
		} else {
			storeFailedBatch(&batch)
		}

		counter, err := incrementCounter(BATCH_COUNTER)
		check(err, "Failed to increment counter")
		if counter == uint64(totalBatches) {
			finisher <- true
		}

		//processStatusCodes(statusCode, body)
	}
}
