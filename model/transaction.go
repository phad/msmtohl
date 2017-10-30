package model

import (
	"time"
)

type Status int

const(
	Unknown Status = 0
	Unmarked Status = 1
	Pending Status = 2
	Cleared Status = 3
)

func (s Status) String() string {
	switch s {
	case Unknown, Unmarked: return ""
	case Pending: return "!"
	case Cleared: return "*"
	}
	return ""
}

type Account []string

type Posting struct {
	Account Account
	Amount float64
}

type Transaction struct {
	Date time.Time
	Status Status
	Payee, Description string
	Postings []Posting	
}
