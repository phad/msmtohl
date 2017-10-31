// Package qif contains functions to parse transaction data presented in the QIF format.
package qif

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"time"
)

// QIF contains the scan state for a set of records in QIF format.
type QIF struct {
	scanner   *bufio.Scanner
	linesRead int
}

// Record groups the QIF attributes for a single transaction read in QIF format.
type Record struct {
	Type    string
	Date    string
	Amount  string
	Number  string
	Cleared string
	Payee   string
	Label   string
	Memo    string
}

// RecordSet is a group of QIF Records, with the opening Record separated.
type RecordSet struct {
	Opening *Record
	Records []*Record
}

// String conforms with Stringer for Records.
func (r *Record) String() string {
	return fmt.Sprintf("type %q date %q amount %q number %q cleared %q payee %q label %q memo %q",
		r.Type, r.Date, r.Amount, r.Number, r.Cleared, r.Payee, r.Label, r.Memo)
}

// New returns a QIF scanner for QIF data to be read from the given io.Reader.
func New(qifData io.Reader) *QIF {
	return &QIF{scanner: bufio.NewScanner(qifData)}
}

// ErrEOF is a condition used to signal that the parser reached the end of a QIF file.
var ErrEOF = errors.New("QIF end of file")

// Next is an iterator function that returns the next Record scanned from the QIF file.
func (q *QIF) Next() (*Record, error) {
	r := &Record{}
	for q.scanner.Scan() {
		line, err := q.scanner.Text(), q.scanner.Err()
		q.linesRead++
		if err != nil {
			return nil, fmt.Errorf("QIF: scanner error at line %d: %v", q.linesRead, err)
		}
		if len(line) == 0 {
			return nil, fmt.Errorf("QIF: empty line at line %d", q.linesRead)
		}
		switch spec, rest := line[0:1], line[1:]; spec {
		case "!":
			// 'Type' line
			r.Type = rest
		case "D":
			// Date line
			r.Date = rest
		case "T", "U":
			// Transaction amount line
			r.Amount = rest
		case "N":
			// Check number line, or other identifier eg. ATM
			r.Number = rest
		case "C":
			// Cleared status line
			r.Cleared = rest
		case "P":
			// Payee line
			r.Payee = rest
		case "L":
			// Label (category) line
			r.Label = rest
		case "M":
			// Memo (description) line
			r.Memo = rest
		case "^":
			// Record separator line
			return r, nil
		}
	}
	return nil, ErrEOF
}

// NewRecordSet returns a RecordSet for QIF records read from the given io.Reader.
func NewRecordSet(r io.Reader) (*RecordSet, error) {
	q := New(r)
	first, err := q.Next()
	if err != nil {
		return nil, fmt.Errorf("reading first QIF record, error: %v", err)
	}
	if first.Type != "Type:Bank" {
		return nil, fmt.Errorf("unsupported first record type: got %q want \"Type:Bank\" (record: %v)", first.Type, first)
	}
	rs := &RecordSet{Opening: first}
	cnt := 0
	for {
		cnt++
		r, err := q.Next()
		if err == ErrEOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("reading QIF record %d, error: %v", cnt, err)
		}
		rs.Records = append(rs.Records, r)
	}
	return rs, nil
}

// AccountName returns the name of the account described by the opening record of the RecordSet.
func (rs *RecordSet) AccountName() string {
	n := rs.Opening.Label
	return n[1 : len(n)-1]
}

// ParseDate parses date strings in the QIF exported by Microsoft Money 2000, which is dd/mm'yyyy
func ParseDate(d string) (time.Time, error) {
	return time.Parse("02/01'2006", d)
}
