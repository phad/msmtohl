package main

import (
	"flag"
	"fmt"
	"log"
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

	rs, err := qif.NewRecordSet(qf)
	if err != nil {
		log.Fatalf("Reading file %q got error: %v", *inFile, err)
	}
	log.Printf("For account %q read %d records.", rs.AccountName(), len(rs.Records))
}
