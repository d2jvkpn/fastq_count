package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

const USAGE = `
Usage: fastq_count  [-phred value]  [-out tsv]  <input1.fastq input2.fastq.gz>
  output (tsv) header: Total reads  Total bases  N bases  Q20  Q30  GC
  note:
    1. When input is -, read standard input;
    2. "pigz -dc *.fastq.gz | fastq_count -" is recommended for gzipped file(s).
`

const LISENSE = `
author: d2jvkpn
version: 1.2.0
release: 2021-07-04
project: https://github.com/d2jvkpn/fastq_count
lisense: GPLv3 (https://www.gnu.org/licenses/gpl-3.0.en.html)
`

const (
	RFC3339ms = "2006-01-02T15:04:05.000Z07:00"
)

func init() {
	SetLogRFC3339()
}

func main() {
	var (
		jsonFormat bool
		output     string
		inputs     []string
		err        error
		start      time.Time
		ct         *Counter
	)

	ct = NewCounter()
	flag.StringVar(&output, "output", "", "save result to file, default: stdout")
	flag.IntVar(&ct.Phred, "phred", 33, "set phred value")
	flag.BoolVar(&jsonFormat, "json_format", false, "output json format")

	flag.Usage = func() {
		fmt.Println(USAGE)
		flag.PrintDefaults()
		fmt.Println(LISENSE)
	}
	flag.Parse()

	inputs = flag.Args()
	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(2)
	}

	/// run
	start = time.Now()
	wg := new(sync.WaitGroup)
	go func() {
		for i := range inputs {
			wg.Add(1)
			input := inputs[i]
			go ReadBlocks(input, ct, wg)
		}

		wg.Wait()
	}()

	ct.Counting()

	if err = ct.Output(output, jsonFormat); err != nil {
		log.Fatalln(err)
	}
	log.Printf("fastq count elapsed: %v\n", time.Since(start))
}

func SetLogRFC3339() {
	log.SetFlags(0)
	log.SetOutput(new(logWriter))
}

type logWriter struct{}

func (writer *logWriter) Write(bts []byte) (int, error) {
	// time.RFC3339
	return fmt.Print(time.Now().Format(RFC3339ms) + "  " + string(bts))
}
