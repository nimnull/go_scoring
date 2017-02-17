package scoring

import (
	"net/url"
	"fmt"
	"runtime"
	"os"
	"log"
	"unicode/utf8"
	"go_scoring/csv"
	"bytes"
	"strings"
	"github.com/buger/jsonparser"
	"net/http"
	"io/ioutil"
)

const (
	MAX_BATCH_SIZE = 5*1024 ^ 2
	CLIENT_HEADERS = "datarobot_batch_scoring/%s|Golang/%s|system/%s"
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

func RunBatch(
	baseUrl *url.URL, importId, dataset, encoding, delimiter string,
	maxBatchSize, concurrent int, compression, fastMode bool) {

	if maxBatchSize == 0 || maxBatchSize > MAX_BATCH_SIZE {
		maxBatchSize = MAX_BATCH_SIZE
	}

	baseUrl.Path = fmt.Sprintf("%s/%s/predict", baseUrl.Path, importId)
	encoding = investigateEncoding(dataset, encoding, delimiter, delimiter)

	file, err := os.Open(dataset)
	check(err, "Failed to open predict source")
	defer file.Close()

	dt := csvtools.NewDetector()
	delimiters := dt.DetectDelimiter(file,'"')

	r, _ := utf8.DecodeRuneInString(delimiters[0])
	file.Seek(0,0)
	opts := csvtools.CSVInputOptions{true, r, file}
	csvInput, err := csvtools.NewCSVInput(&opts)
	check(err, "Failed to initialize csv reader")

	buff := bytes.NewBufferString(strings.Join(csvInput.Header(), ",") + "\n")
	buff.WriteString(strings.Join(csvInput.ReadRecord(), ",") + "\n")


	req, err := http.NewRequest("POST", baseUrl.String(), buff)
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

	switch resp.StatusCode {
	case 400:
		var errorMsg string
		if msg, err := jsonparser.GetString(body, "message"); err != nil {
			errorMsg = string(body)

		} else {
			errorMsg = msg
		}
		log.Fatalf("Failed with client error: %s\n", errorMsg)
	case 403:
		log.Fatalf("Failed with message:\n\t%s\n", body)
	case 401:
		log.Fatalf("failed to authenticate -- "+
			"please check your: datarobot_key (if required), "+
			"username/password and/or api token. Contact "+
			"customer support if the problem persists "+
            "message:\n%s\n", body)
	case 405:
		log.Fatalln("failed to request endpoint -- please check your "+
			"'--host' argument")
	case 502:
		log.Fatalln("problem with the gateway -- please check your "+
			"'--host' argument and contact customer support"+
			"if the problem persists.")
	}
	fmt.Println(string(body))

	execTime, err := jsonparser.GetInt(body, "execution_time")
	check(err, "Failed to read exec time")
	fmt.Printf("Execution time: %d\n", execTime)

	//done := make(chan bool, 1)

	//for i := 0;;i++ {
	//	line := csvInput.ReadRecord()
	//	buff = bytes.NewBufferString(strings.Join(csvInput.Header(), ",") + "\n")
	//	buff.WriteString(strings.Join(line, ",") + "\n")
	//	println(buff.Len())


		//go sendLine(i, csvInput.Header(), line, done)
		//<-done
	//}

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

func sendLine(step int, header, line []string, done chan bool) {
	fmt.Println(strings.Join(line, ","))
	if line == nil {
		done <- true
	}
}

func createBatch() {

}
