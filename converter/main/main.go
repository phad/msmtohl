package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/phad/msmtohl/converter"
	"github.com/phad/msmtohl/parser/qif"
)

var (
	inFile  = flag.String("in_file", "", "Input file in QIF format.")
	outFile = flag.String("out_file", "", "Output file in hledger format.")
	max     = flag.Int("max", 0, "Maximum number of rows to output (0=output all)")
)

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

	for i, txn := range txns {
		if err = txn.SerializeHledger(hlf); err != nil {
			panic(err)
		}
		if *max > 0 && *max == i {
			break
		}
	}
}

