package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/d2jvkpn/fastq_count/pkg/cmd"
)

func ReadBlocks(input string, ct *Counter, wg *sync.WaitGroup) {
	defer func() {
		if wg != nil {
			wg.Done()
		}
	}()

	var (
		err error
		ci  *cmd.CmdInput
	)

	log.Printf("fastq count input: %q\n", input)
	if ci, err = cmd.NewCmdInput(input); err != nil {
		log.Println(err)
		return
	}

	i := 0
	for ci.Scanner.Scan() {
		i++
		switch i % 4 {
		case 2:
			text := ci.Scanner.Text()
			ct.ch1 <- &text
		case 0:
			text := ci.Scanner.Text()
			ct.ch2 <- &text
		}
	}

	close(ct.ch1)
	close(ct.ch2)

	ci.Close()
}

/// Counter
type Counter struct {
	Phred int `json:"phred"`

	RN  int64 `json:"reads"` // read number
	BN  int64 `json:"bases"` // base number
	NN  int64 `json:"nn"`    // base number of N
	GC  int64 `json:"gc"`    // base number of G and C
	Q20 int64 `json:"q20"`   // Q20 number
	Q30 int64 `json:"q30"`   // Q30 number

	ch1 chan *string `json:"-"`
	ch2 chan *string `json:"-"`
}

func NewCounter(phreds ...int) (counter *Counter) {
	counter = new(Counter)
	counter.ch1 = make(chan *string, 100)
	counter.ch2 = make(chan *string, 100)
	if len(phreds) > 0 {
		counter.Phred = phreds[0]
	}

	return counter
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
		float64(ct.GC*100)/float64(ct.BN),
		float64(ct.Q20*100)/float64(ct.BN),
		float64(ct.Q30*100)/float64(ct.BN),
	)

	fmt.Fprintf(wt, "%d\t%d\t%d\t%d\t%d\t%d\n", ct.RN, ct.BN, ct.NN, ct.GC, ct.Q20, ct.Q30)
}

func (ct *Counter) Counting() {
	wg := new(sync.WaitGroup)
	wg.Add(2)

	go func() {
		for k := range ct.ch1 {
			ct.RN++
			ct.BN += int64(len(*k))
			ct.NN += int64(strings.Count(*k, "N"))
			ct.GC += int64(strings.Count(*k, "G") + strings.Count(*k, "C"))
		}
		wg.Done()
	}()

	go func() {
		for k := range ct.ch2 {
			for _, q := range *k {
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
		wg.Done()
	}()

	wg.Wait()
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
