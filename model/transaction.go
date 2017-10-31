package model

import (
	"time"
)

// Status represents the 'cleared' status of a Transaction
type Status int

// Defined statuses from http://hledger.org/journal.html
const (
	Unknown  Status = 0 // Unknown represents an unknown status.
	Unmarked Status = 1 // Unmarked is used for Transactions with no cleared status marking.
	Pending  Status = 2 // Pending is used for Transactions awaiting reconciliation.
	Cleared  Status = 3 // Cleared is used for Transactions that are cleared, i.e. reconciled.
)

// String conforms with Stringer for Status values.
func (s Status) String() string {
	switch s {
	case Unknown, Unmarked:
		return ""
	case Pending:
		return "!"
	case Cleared:
		return "*"
	}
	return ""
}

// Account is an account name, modelled as a label hierarchy.
type Account []string

// Posting models a credit to, or debit from, a particular Account.
type Posting struct {
	Account Account
	Amount  float64
}

// Transaction represents the movement of funds between two or more Accounts.
type Transaction struct {
	Date        time.Time // The date on which the Transaction occurred (assumed UTC).
	Status      Status    // The status of the transaction.
	Payee       string    // The transaction payee
	Description string    // A descriptive label for the transaction
	Postings    []Posting // Two or more Accounts that were involved in the Transaction.
}
