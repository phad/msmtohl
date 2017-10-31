// Package converter contains functions to convert transactions from one format to another.
package converter

import (
	"github.com/phad/msmtohl/model"
	"github.com/phad/msmtohl/parser/qif"
)

// FromQIF converts the QIF RecordSet provided into a set of Transactions.
func FromQIF(rs *qif.RecordSet) ([]*model.Transaction, error) {
	var txns []*model.Transaction
	for _, r := range rs.Records {
		t, err := fromQIFRecord(r)
		if err != nil {
			return nil, err
		}
		txns = append(txns, t)
	}
	return txns, nil
}

func fromQIFRecord(r *qif.Record) (*model.Transaction, error) {
	d, err := qif.ParseDate(r.Date)
	if err != nil {
		return nil, err
	}
	return &model.Transaction{
		Date:        d,
		Status:      fromQIFStatus(r.Cleared),
		Payee:       r.Payee,
		Description: r.Memo,
	}, nil	
}

func fromQIFStatus(qs string) model.Status {
	switch qs {
	case " ":
		return model.Unmarked
	case "*", "C":
		return model.Pending
	case "X", "R":
		return model.Cleared
	}
	return model.Unknown
}
