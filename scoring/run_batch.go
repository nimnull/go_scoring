package scoring

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"unicode/utf8"

	"github.com/buger/jsonparser"
	"go_scoring/csv"
	"time"
)

type Batch struct {
	idx  int
	data *bytes.Buffer
}

const (
	MAX_BATCH_SIZE = 5*1024 ^ 2
	CLIENT_HEADERS = "datarobot_batch_scoring/%s|Golang/%s|system/%s"
)

var scoringEndpoint string

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

func getCSVInput(file *os.File) *csvtools.CSVInput {
	dt := csvtools.NewDetector()
	delimiters := dt.DetectDelimiter(file, '"')

	r, _ := utf8.DecodeRuneInString(delimiters[0])
	// rewind file after detector finished it's work
	file.Seek(0, 0)
	opts := csvtools.CSVInputOptions{true, r, file}
	csvInput, err := csvtools.NewCSVInput(&opts)
	check(err, "Failed to initialize csv reader")

	return csvInput
}

func sendScoringBatch(compression bool, data io.Reader) (int, []byte) {
	req, err := http.NewRequest("POST", scoringEndpoint, data)
	check(err, "Failed to prepare request")
	for k, v := range getHttpHeaders(compression) {
		req.Header.Set(k, v)
	}
	netClient := GetHttpClient()

	resp, err := netClient.Do(req)
	check(err, "Failed to send request")
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
	scoringEndpoint = baseUrl.String()
	encoding = investigateEncoding(dataset, encoding, delimiter, delimiter)

	file, err := os.Open(dataset)
	check(err, "Failed to open predict source")
	defer file.Close()

	csvInput := getCSVInput(file)

	csvHeader := csvInput.Header()
	buff := bytes.NewBufferString(strings.Join(csvHeader, ",") + "\n")
	firstRow := csvInput.ReadRecord()
	buff.WriteString(strings.Join(firstRow, ",") + "\n")

	//statusCode, body := sendScoringBatch(false, buff)
	//fmt.Println(statusCode)
	//processStatusCodes(statusCode, body)

	//fmt.Println(string(body))

	//if execTime, err := jsonparser.GetInt(body, "execution_time"); err != nil {
	//	log.Printf("Failed to read execution time: %s\n", err)
	//} else {
	//	fmt.Printf("Execution time: %d\n", execTime)
	//}

	queue := make(chan Batch, concurrent)
	done := make(chan bool, 1)

	for i := 0; i < concurrent; i++ {
		go sendLine(i, queue)
	}

	for i := 0; ; i++ {
		buff := bytes.NewBufferString(strings.Join(csvHeader, ",") + "\n")
		isLast := false

		for buff.Len() < MAX_BATCH_SIZE {
			line := csvInput.ReadRecord()
			//fmt.Println(line)
			if line != nil {
				buff.WriteString(strings.Join(line, ",") + "\n")
			} else {
				isLast = true
				break
			}
		}

		queue <- Batch{idx: i, data: buff}

		if isLast {
			log.Println("Last batch sent", i)
			close(queue)
			break
		}

	}
	<-done

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

func sendLine(workerIdx int, batches <-chan Batch) {
	log.Println("Spawned worker", workerIdx)

	for batch := range batches {
		fmt.Println(batch.idx)
		time.Sleep(time.Second * 3)
	}
	//_, body := sendScoringBatch(false, buff)
	//fmt.Println(string(body))

}
