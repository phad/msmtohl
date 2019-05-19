// Package converter contains functions to convert transactions from one format to another.
package converter

import (
	"strings"
	"strconv"

	"github.com/golang/glog"

	"github.com/phad/msmtohl/model"
	"github.com/phad/msmtohl/parser/qif"
)

// FromQIF converts the QIF RecordSet provided into a set of Transactions.
func FromQIF(rs *qif.RecordSet) ([]*model.Transaction, error) {
	var txns []*model.Transaction
	for _, r := range rs.Records {
		t, err := fromQIFRecord(r, rs.Opening)
		if err != nil {
			glog.Errorf("Converting from QIF %v error: %v", r, err)
			return nil, err
		}
		txns = append(txns, t)
	}
	return txns, nil
}

func fromQIFRecord(r, op *qif.Record) (*model.Transaction, error) {
	d, err := qif.ParseDate(r.Date)
	if err != nil {
		return nil, err
	}
	txn := &model.Transaction{
		Date:        d,
		Status:      fromQIFStatus(r.Cleared),
		Payee:       r.Payee,
		Description: r.Memo,
	}
	from, err := fromSplit(&qif.Split{Category: op.Label, Amount: "0"})
	if err != nil {
		return nil, err
	}
	if len(r.Splits) > 0 {
		for _ ,s := range r.Splits {
			p, err := fromSplit(s)
			if err != nil {
				return nil, err
			}
			txn.Postings = append(txn.Postings, p)
		}
		txn.Postings = append(txn.Postings, from)
		return txn, nil
	}
	// Regular, unsplit transaction.
	p, err := fromSplit(&qif.Split{
		Amount:   r.Amount,
		Category: r.Label,
	})
	if err != nil {
		return nil, err
	}
	txn.Postings = append(txn.Postings, []model.Posting{p, from}...)
	return txn, err	
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

func fromSplit(s *qif.Split) (model.Posting, error) {
	amount, err := strconv.ParseFloat(sanitizeAmount(s.Amount), 64)
        if err != nil {
                return model.Posting{}, err
        }
	var ac model.Account
	for _, s := range strings.Split(s.Category, ":") {
		ac = append(ac, s)
	}
	return model.Posting{Amount: -amount, Account: ac}, nil
}

func sanitizeAmount(a string) string {
	// Remove , 1000s separator
	return strings.Replace(a, ",", "", -1)
}
