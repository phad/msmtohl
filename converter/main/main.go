package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/phad/msmtohl/converter"
	"github.com/phad/msmtohl/parser/qif"
)

var inFile = flag.String("in_file", "", "Input file in QIF format.")
var outFile = flag.String("out_file", "", "Output file in hledger format.")

func main() {
	flag.Parse()

	fmt.Println("QIF Converter")

	qf, err := os.Open(*inFile)
	if err != nil {
		panic(fmt.Errorf("os.File error: %v", err))
	}
	defer qf.Close()

	hlf, err := os.Create(*outFile)
	if err != nil {
		panic(fmt.Errorf("os.File error: %v", err))
	}
	defer hlf.Close()

	rs, err := qif.NewRecordSet(qf)
	if err != nil {
		log.Fatalf("Reading file %q got error: %v", *inFile, err)
	}
	log.Printf("For account %q read %d records.", rs.AccountName(), len(rs.Records))

	txns, err := converter.FromQIF(rs)
	if err != nil {
		log.Fatalf("Converting QIF RecordSet got error: %v", err)
	}
	log.Printf("Converted %d records.", len(txns))

	for i := 0; i < 10; i++ {
		if err = txns[i].SerializeHledger(hlf); err != nil {
			panic(err)
		}
	}
}
