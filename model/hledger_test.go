package model

import (
	"bytes"
	"testing"
	"time"
)

var (
	d1 = time.Date(2017, time.January, 12, 0, 0, 0, 0, time.UTC)
)

func TestSerializeHledger_topLine(t *testing.T) {
	tests := []struct {
		desc string
		txn  *Transaction
		want string
	}{
		{desc: "nil Transaction, empty topline"},
		{
			desc: "txn with date only",
			txn:  &Transaction{Date: d1},
			want: "2017/01/12",
		},
		{
			desc: "txn with date and payee only",
			txn:  &Transaction{Date: d1, Payee: "Dave"},
			want: "2017/01/12 Dave",
		},
		{
			desc: "txn with date and description only",
			txn:  &Transaction{Date: d1, Description: "Groceries"},
			want: "2017/01/12 Groceries",
		},
		{
			desc: "txn with date, payee and deascription",
			txn:  &Transaction{Date: d1, Payee: "Dave", Description: "Groceries"},
			want: "2017/01/12 Dave | Groceries",
		},
		{
			desc: "txn with date and payee only in pending state",
			txn:  &Transaction{Date: d1, Payee: "Dave", Status: Pending},
			want: "2017/01/12 ! Dave",
		},
		{
			desc: "txn with date and description only in cleared state",
			txn:  &Transaction{Date: d1, Description: "Groceries", Status: Cleared},
			want: "2017/01/12 * Groceries",
		},
		{
			desc: "txn with date, payee and description in cleared state",
			txn:  &Transaction{Date: d1, Payee: "Dave", Description: "Groceries", Status: Cleared},
			want: "2017/01/12 * Dave | Groceries",
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			if got, want := test.txn.topLine(), test.want; got != want {
				t.Errorf("topLine()=%q want %q", got, want)
			}
		})
	}
}

func TestSerializeHledger(t *testing.T) {
	tests := []struct {
		desc    string
		txn     *Transaction
		want    string
		wantErr bool
	}{
		{desc: "nil Transaction, does nothing"},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			var got bytes.Buffer
			err := test.txn.SerializeHledger(&got)
			if gotErr := err != nil; gotErr != test.wantErr {
				t.Errorf("SerializeHledger() got err? %t want %t (err=%v)", gotErr, test.wantErr, err)
			}
			if got.String() != test.want {
				t.Errorf("SerializeHledger()=%q want %q", got.String(), test.want)
			}
		})
	}
}
