package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	// "sync/atomic"

	gzip "github.com/klauspost/pgzip" //"compress/gzip"
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
version: 1.1.0
release: 2021-06-18
project: https://github.com/d2jvkpn/fastq_count
lisense: GPLv3 (https://www.gnu.org/licenses/gpl-3.0.en.html)
`

func main() {
	var (
		jsonFormat bool
		output     string
		inputs     []string
		err        error
		ct         *Counter
	)

	ct = new(Counter)
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
	ch := make(chan [2]string, 1000)
	wg := new(sync.WaitGroup)

	go func() {
		for i := range inputs {
			wg.Add(1)
			input := inputs[i]
			go ReadBlocks(input, ch, wg)
		}

		wg.Wait()
		close(ch)
	}()

	ct.Counting(ch, nil)

	if err = ct.Output(output, jsonFormat); err != nil {
		log.Fatalln(err)
	}
}

func ReadBlocks(input string, ch chan<- [2]string, wg *sync.WaitGroup) {
	defer wg.Done()

	var (
		err error
		blk [2]string
		ci  *CmdInput
	)

	log.Printf("fastq_count read sequences from %s\n", input)
	if ci, err = NewCmdInput(input); err != nil {
		log.Println(err)
		return
	}

	for {
		ci.Scanner.Scan()
		ci.Scanner.Scan()
		blk[0] = ci.Scanner.Text()
		ci.Scanner.Scan()
		if !ci.Scanner.Scan() {
			break
		}
		blk[1] = ci.Scanner.Text()
		ch <- blk
	}

	ci.Close()
}

/// Counter
type Counter struct {
	Phred int `json:"phred"`

	RN  int64 `json:"reads"` // read number
	BN  int64 `json:"bases"` // base number
	NN  int64 `json:"n"`     // base number of N
	GC  int64 `json:"gc"`    // base number of G and C
	Q20 int64 `json:"q20"`   // Q20 number
	Q30 int64 `json:"q30"`   // Q30 number
}

func (ct *Counter) String() string {
	return fmt.Sprintf(
		"Reads\tBases\tN-bases\tGC\tQ20\tQ30\n%d\t%d\t%d\t%d\t%d\t%d",
		ct.RN, ct.BN, ct.NN, ct.GC, ct.Q20, ct.Q30,
	)
}

func (ct *Counter) Write(wt io.Writer, jsonFormat bool) {
	if jsonFormat {
		bts, _ := json.Marshal(ct)
		fmt.Fprintf(wt, string(bts)+"\n")
		return
	}

	fmt.Fprintln(wt, "Reads\tBases\tN-bases\tGC\tQ20\tQ30")
	fmt.Fprintf(wt, "%.2fM\t%.2fG\t%.2f%%\t%.2f%%\t%.2f%%\t%.2f%%\n",
		float64(ct.RN)/float64(1e+6),
		float64(ct.BN)/float64(1e+9),
		float64(ct.NN*100)/float64(ct.BN),
		float64(ct.Q20*100)/float64(ct.BN),
		float64(ct.Q30*100)/float64(ct.BN),
		float64(ct.GC*100)/float64(ct.BN),
	)

	fmt.Fprintf(wt, "%d\t%d\t%d\t%d\t%d\t%d\n", ct.RN, ct.BN, ct.NN, ct.Q20, ct.Q30, ct.GC)
}

func (ct *Counter) Counting(ch <-chan [2]string, wg *sync.WaitGroup) {
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()

	var v int64

	for k := range ch {
		ct.RN++
		v = int64(len(k[0]))
		ct.BN += v
		// atomic.AddInt64(&ct.BN, v)

		v = int64(strings.Count(k[0], "N"))
		ct.NN += v

		v = int64(strings.Count(k[0], "G") + strings.Count(k[0], "C"))
		ct.GC += v

		for _, q := range k[1] {
			if int(q)-ct.Phred >= 20 {
				ct.Q20++
			} else {
				continue
			}
			if int(q)-ct.Phred >= 30 {
				ct.Q30++
			}
		}
	}
}

func (ct *Counter) Output(output string, jsonFormat bool) (err error) {
	if output == "" {
		ct.Write(os.Stdout, jsonFormat)
		return nil
	}

	if err = os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return err
	}

	var file *os.File
	if file, err = os.Create(output); err != nil {
		return err
	}

	ct.Write(file, jsonFormat)
	file.Close()

	return nil
}

/// CmdInput
type CmdInput struct {
	Name    string
	File    *os.File
	Reader  *gzip.Reader
	Scanner *bufio.Scanner
}

func (ci *CmdInput) Close() {
	if ci.Reader != nil {
		ci.Reader.Close()
	}

	if ci.File != nil {
		ci.File.Close()
	}
}

func NewCmdInput(name string) (ci *CmdInput, err error) {
	ci = new(CmdInput)
	ci.Name = name

	if ci.Name == "-" { // from stdin
		ci.Scanner = bufio.NewScanner(os.Stdin)
		return
	}

	if ci.File, err = os.Open(ci.Name); err != nil {
		return
	}

	if strings.HasSuffix(ci.Name, ".gz") { // read gzipped file
		if ci.Reader, err = gzip.NewReader(ci.File); err != nil {
			return
		}
		ci.Scanner = bufio.NewScanner(ci.Reader)
	} else {
		ci.Scanner = bufio.NewScanner(ci.File) // read text file
	}

	return
}
