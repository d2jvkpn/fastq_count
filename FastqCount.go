package main

import (
	"bufio"
	"flag"
	"fmt"
	gzip "github.com/klauspost/pgzip" //"compress/gzip"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const USAGE = `
Usage: FastqCount  [-phred value]  [-o tsv]  <input1.fastq input2.fastq.gz>
  output (tsv) header: Total reads  Total bases  N bases  Q20  Q30  GC
  note:
    1. When input is -, read standard input;
    2. "pigz -dc *.fastq.gz | FastqCount -" is recommended for gzipped file(s).
`

const LISENSE = `
author: d2jvkpn
version: 0.9.3
release: 2019-04-02
project: https://github.com/d2jvkpn/FastqCount
lisense: GPLv3 (https://www.gnu.org/licenses/gpl-3.0.en.html)
`

func main() {
	output := flag.String("o", "", "output summary to a tsv file, default: stdout")
	phred := flag.Int("phred", 33, "set phred value")

	flag.Usage = func() {
		fmt.Println(USAGE)
		flag.PrintDefaults()
		fmt.Println(LISENSE)
	}

	flag.Parse()
	inputs := flag.Args()

	if len(os.Args) == 1 {
		flag.Usage()
		os.Exit(2)
	}

	var err error
	var wt io.Writer
	ch := make(chan [2]string, 10000)

	ct := new(Counter)

	go func() {
		for _, s := range inputs {
			log.Printf("FastqCount read sequences from %s\n", s)

			ci, err := NewCmdInput(s)
			if err != nil {
				log.Fatal(err)
			}
			defer ci.Close()

			for {
				var blk [2]string
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
		}
		close(ch)
	}()

	for k := range ch {
		ct.RN++
		ct.BN += len(k[0])
		ct.NN += strings.Count(k[0], "N")
		ct.GC += strings.Count(k[0], "G") + strings.Count(k[0], "C")

		for _, q := range k[1] {
			if int(q)-*phred >= 20 {
				ct.Q20++
			} else {
				continue
			}
			if int(q)-*phred >= 30 {
				ct.Q30++
			}
		}
	}

	wt = os.Stdout

	if *output != "" {
		err = os.MkdirAll(filepath.Dir(*output), 0755)
		if err != nil {
			log.Println(err)
		} else {
			wt, err = os.Create(*output)
			if err != nil {
				log.Println(err)
				wt = os.Stdout
			}
		}
	}

	// fmt.Println(ct)

	fmt.Fprintln(wt, "Reads\tBases\tN-bases\tQ20\tQ30\tGC")

	fmt.Fprintf(wt, "%.2fM\t%.2fG\t%.2f%%\t%.2f%%\t%.2f%%\t%.2f%%\n",
		float64(ct.RN)/float64(1E+6),
		float64(ct.BN)/float64(1E+9),
		float64(ct.NN*100)/float64(ct.BN),
		float64(ct.Q20*100)/float64(ct.BN),
		float64(ct.Q30*100)/float64(ct.BN),
		float64(ct.GC*100)/float64(ct.BN))

	fmt.Fprintf(wt, "%d\t%d\t%d\t%d\t%d\t%d\n", ct.RN, ct.BN, ct.NN,
		ct.Q20, ct.Q30, ct.GC)

	if *output != "" {
		log.Printf("Saved FastqCount result to %s\n", *output)
	}
}

func (ct *Counter) String() (s string) {
	s = "Reads\tBases\tN-bases\tQ20\tQ30\tGC\n"
	s += fmt.Sprintf("%d\t%d\t%d\t%d\t%d\t%d", 
		ct.RN, ct.BN, ct.NN, ct.Q20, ct.Q30, ct.GC)

	return
}

type Counter struct {
	RN, BN, Q20, Q30, GC, NN int
}

type CmdInput struct {
	Name    string
	File    *os.File
	Reader  *gzip.Reader
	Scanner *bufio.Scanner
}

func (ci *CmdInput) Close() {
	ci.Reader.Close()
	ci.File.Close()
}

func NewCmdInput(name string) (ci *CmdInput, err error) {
	ci = new(CmdInput)
	ci.Name = name

	if ci.Name == "-" {
		ci.Scanner = bufio.NewScanner(os.Stdin)
		return
	}

	ci.File, err = os.Open(ci.Name)

	if err != nil {
		return
	}

	if strings.HasSuffix(ci.Name, ".gz") {
		if ci.Reader, err = gzip.NewReader(ci.File); err != nil {
			return
		}
		ci.Scanner = bufio.NewScanner(ci.Reader)
	} else {
		ci.Scanner = bufio.NewScanner(ci.File)
	}

	return
}
