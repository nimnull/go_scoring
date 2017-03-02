package scoring

import (
	"compress/gzip"
	"github.com/saintfish/chardet"
	"log"
	"os"
	"strings"
)

const (
	DETECT_SAMPLE_SIZE_FAST int = 209715 // int(0.2 * 1024 ^ 2)
	DETECT_SAMPLE_SIZE_SLOW int = 1024 * 1024
	AUTO_SAMPLE_SIZE        int = 524288 // int(0.5 * 1024 ^ 2)
	AUTO_SMALL_SAMPLES      int = 500
	AUTO_GOAL_SIZE          int = 2621440 // int(2.5 * 1024 ^ 2) size we want per batch
)

func investigateEncoding(dataset, encoding, inDelimiter, outDelimiter string) string {
	var (
		sample []byte
		result *chardet.Result
	)
	sample = make([]byte, DETECT_SAMPLE_SIZE_FAST, DETECT_SAMPLE_SIZE_SLOW)

	if dsF, err := os.Open(dataset); err != nil {
		log.Fatal(err)
	} else {
		defer dsF.Close()
		if strings.HasSuffix(dataset, ".gz") {
			if reader, err := gzip.NewReader(dsF); err != nil {
				log.Fatal(err)
			} else {
				defer reader.Close()
				if _, err := reader.Read(sample); err != nil {
					log.Fatal(err)
				}
			}
		} else {
			dsF.Read(sample)
		}
		detector := chardet.NewTextDetector()
		if result, err = detector.DetectBest(sample); err != nil {
			log.Fatal(err)
		}
	}
	return result.Charset
}
