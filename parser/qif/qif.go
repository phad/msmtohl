package qif

import (
	"bufio"
	"errors"
	"fmt"
	"io"
)

type QIF struct {
	scanner *bufio.Scanner
	linesRead int
}

type Record struct {
	Type string
	Date string
	Amount string
	Status string
	Payee string
	Label string
	Message string
}

func (r *Record) String() string {
	return fmt.Sprintf("type %q date %q amount %q status %q payee %q label %q message %q",
		r.Type, r.Date, r.Amount, r.Status, r.Payee, r.Label, r.Message)
}

func New(qifData io.Reader) *QIF {
    return &QIF{scanner: bufio.NewScanner(qifData)}
}

var ErrEOF = errors.New("QIF end of file")

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
			case "T":
				// Transaction amount line
				r.Amount = rest
			case "C":
				// Transaction status line
				r.Status = rest
			case "P":
				// Payee line
				r.Payee = rest
			case "L":
				// Label (category) line
				r.Label = rest
			case "M":
				// Message (description) line
				r.Message = rest
			case "^":
				// Record separator line
				return r, nil
		}
	}
	return nil, ErrEOF
}
