package main

import (
	"bufio"
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
version: 1.0.0
release: 2021-03-04
project: https://github.com/d2jvkpn/fastq_count
lisense: GPLv3 (https://www.gnu.org/licenses/gpl-3.0.en.html)
`

func main() {
	var (
		output string
		inputs []string
		err    error
		ct     *Counter
	)

	ct = new(Counter)
	flag.StringVar(&output, "out", "", "output summary to a tsv file, default: stdout")
	flag.IntVar(&ct.Phred, "phred", 33, "set phred value")

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

	if err = ct.Output(output); err != nil {
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
	Phred int

	RN  int64 // read number
	BN  int64 // base number
	Q20 int64 // Q20 number
	Q30 int64 // Q30 number
	GC  int64 // base number of G and C
	NN  int64 // base number of N
}

func (ct *Counter) String() string {
	return fmt.Sprintf(
		"Reads\tBases\tN-bases\tQ20\tQ30\tGC\n%d\t%d\t%d\t%d\t%d\t%d",
		ct.RN, ct.BN, ct.NN, ct.Q20, ct.Q30, ct.GC,
	)
}

func (ct *Counter) Write(wt io.Writer) {
	fmt.Fprintln(wt, "Reads\tBases\tN-bases\tQ20\tQ30\tGC")

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

func (ct *Counter) Output(output string) (err error) {
	defer func() {
		ct.Write(os.Stdout)
	}()

	if output == "" {
		return nil
	}

	if err = os.MkdirAll(filepath.Dir(output), 0755); err != nil {
		return err
	}

	var file *os.File
	if file, err = os.Create(output); err != nil {
		return err
	}

	ct.Write(file)
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
