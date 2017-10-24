package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/phad/msmtohl/parser/qif"
)

var inFile = flag.String("in_file", "", "Input file in QIF format.")

func main() {
	flag.Parse()

	fmt.Println("QIF Converter")

	qf, err := os.Open(*inFile)
	if err != nil {
		panic(fmt.Errorf("QIF error: %v", err))
	}
	defer qf.Close()

	q := qif.New(qf)
	for {
		rec, err := q.Next()
		if err == qif.ErrEOF {
			break
		} else if err != nil {
			fmt.Printf("Error: %v\n", err)
			break
		}
		fmt.Printf("Read QIF record: %v\n", rec)
	}
}
