package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"

	"github.com/phad/msmtohl/converter"
	"github.com/phad/msmtohl/model"
	"github.com/phad/msmtohl/parser/qif"
)

var (
	inFiles = flag.String("in_files", "", "Comma-separated input files in QIF format.")
	outFile = flag.String("out_file", "", "Output file in hledger format.")
	max     = flag.Int("max", 0, "Maximum number of rows to output (0=output all)")
)

func main() {
	flag.Parse()

	fmt.Println("QIF Converter")

	hlf, err := os.Create(*outFile)
	if err != nil {
		panic(fmt.Errorf("Creating %q error: %v", *outFile, err))
	}
	defer hlf.Close()

	var allTxns []*model.Transaction
	inFileNames, err := filepath.Glob(*inFiles)
	if err != nil {
		panic(fmt.Errorf("filepath.Glob(%q) error: %v", *inFiles, err))
	}

	for _, inf := range inFileNames {
		fmt.Printf(" .. opening %s\n", inf)

		qf, err := os.Open(inf)
		if err != nil {
			panic(fmt.Errorf("Opening %q error: %v", inf, err))
		}
		defer qf.Close()

		fmt.Printf(" .. parsing QIF from %s\n", inf)

		rs, err := qif.NewRecordSet(qf)
		if err != nil {
			log.Fatalf("Reading file %q got error: %v", inf, err)
		}
		log.Printf(" .. parsed %d QIF records.", len(rs.Records))

		fmt.Printf(" .. converting to ledger from %s\n", inf)

		txns, err := converter.FromQIF(rs)
		if err != nil {
			log.Fatalf("Converting QIF RecordSet got error: %v", err)
		}
		log.Printf(" .. converted %d records for account %q.\n\n", len(txns), rs.AccountName())

		allTxns = append(allTxns, txns...)
	}

	sort.Slice(allTxns, func(l, r int) bool {
		return allTxns[l].Date.Before(allTxns[r].Date)
	})

	for i, txn := range allTxns {
		if err = txn.SerializeHledger(hlf); err != nil {
			panic(err)
		}
		if *max > 0 && *max == i {
			break
		}
	}
}
