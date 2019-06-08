// Package qif contains functions to parse transaction data presented in the QIF format.
package qif

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"time"

	"golang.org/x/text/encoding"
)

// QIF contains the scan state for a set of records in QIF format.
type QIF struct {
	scanner   *bufio.Scanner
	decoder   *encoding.Decoder
	linesRead int
	parseErr  error
}

// Record groups the QIF attributes for a single transaction read in QIF format.
type Record struct {
	Type     string
	Date     string
	Amount   string
	Number   string
	Cleared  string
	Payee    string
	Label    string
	Memo     string
	Splits   []*Split
	Transfer bool
}

// Split represents a single sub-transaction in a QIF Record that has >1 split.
type Split struct {
	Category string
	Memo     string
	Amount   string
	Percent  string // exists in QIF spec but not supported yet.
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
func New(qifData io.Reader, dec *encoding.Decoder) *QIF {
	return &QIF{scanner: bufio.NewScanner(qifData), decoder: dec}
}

// ErrEOF is a condition used to signal that the parser reached the end of a QIF file.
var ErrEOF = errors.New("QIF end of file")

// ErrNotSupported is returned if a QIF field type is encountered that this parser
// doesn't support.
type ErrNotSupported struct {
	Desc string
}

func (e *ErrNotSupported) Error() string {
	return fmt.Sprintf("QIF: %q not supported.", e.Desc)
}

// Next is an iterator function that returns the next Record scanned from the QIF file.
func (q *QIF) Next() (*Record, error) {
	r := &Record{}
	var s *Split
	for q.scanner.Scan() {
		line, err := q.scanner.Text(), q.scanner.Err()
		q.linesRead++
		if err != nil {
			return nil, fmt.Errorf("QIF: scanner error at line %d: %v", q.linesRead, err)
		}
		if len(line) == 0 {
			return nil, fmt.Errorf("QIF: empty line at line %d", q.linesRead)
		}
		utf8Line, err := q.decoder.String(line)
		if err != nil {
			return nil, fmt.Errorf("QIF: encoding.Decoder.String(%v): err %v", line, err)
		}
		switch spec, rest := utf8Line[0:1], utf8Line[1:]; spec {
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
			r.Label, r.Transfer = sanitizeLabel(rest)
		case "M":
			// Memo (description) line
			r.Memo = rest
		case "S":
			// Split: Category line
			if s != nil {
				r.Splits = append(r.Splits, s)
			}
			s = &Split{Category: rest}
		case "E":
			// Split: Memo line - we assume the Split opened with 'S'.
			s.Memo = rest
		case "$":
			// Split: Amount line - we assume the Split opened with 'S'
			s.Amount = rest
		case "%":
			// Split: percentage - used in place of Amount. Not supported.
			q.parseErr = &ErrNotSupported{Desc: "Field %"}
		case "^":
			// Record separator line. Store Split if one is in progress.
			if s != nil {
				r.Splits = append(r.Splits, s)
			}
			if q.parseErr != nil {
				e := q.parseErr
				q.parseErr = nil
				return nil, e
			}
			return r, nil
		}
	}
	return nil, ErrEOF
}

// NewRecordSet returns a RecordSet for QIF records read from the given io.Reader.
// Character set conversion from input to UTF-8 is performed by dec.
func NewRecordSet(r io.Reader, dec *encoding.Decoder) (*RecordSet, error) {
	q := New(r, dec)
	first, err := q.Next()
	if err != nil {
		return nil, fmt.Errorf("reading first QIF record, error: %v", err)
	}
	switch first.Type {
		case "Type:Bank":
	  case "Type:Cash":
		case "Type:CCard":
			break
		default:
			return nil, fmt.Errorf("unsupported first record type: got %q want \"Type:Bank\", \"Type:CCard\" or \"Type:Cash\", (record: %v)", first.Type, first)
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

// ParseDate parses date strings in the QIF format used by Microsoft Money 2000,
// which is dd/mm'yyyy or dd/mm/yyyy for pre-2000 dates.
func ParseDate(d string) (time.Time, error) {
	t, err := time.Parse("02/01'2006", d)
	if err != nil {
		if t, err = time.Parse("02/01/2006", d); err != nil {
			return time.Time{}, err
		}
	}
	return t, nil
}

// sanitizeLabel strips wrapping [ ] on label.  If present returns the stripped
// label and true; otherwise the original label and false.
func sanitizeLabel(l string) (string, bool) {
	if len(l) >= 2 && l[0] == '[' && l[len(l) - 1] == ']' {
		return l[1:len(l) - 1], true
	}
	return l, false
}
