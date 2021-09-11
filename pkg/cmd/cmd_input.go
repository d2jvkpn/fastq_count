package cmd

import (
	"bufio"
	"os"
	"strings"

	gzip "github.com/klauspost/pgzip" //"compress/gzip"
)

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
