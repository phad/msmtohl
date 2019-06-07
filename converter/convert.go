// Package converter contains functions to convert transactions from one format to another.
package converter

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/golang/glog"

	"github.com/phad/msmtohl/model"
	"github.com/phad/msmtohl/parser/qif"
)

// FromQIF converts the QIF RecordSet provided into a set of Transactions.
func FromQIF(rs *qif.RecordSet) ([]*model.Transaction, error) {
	var txns []*model.Transaction
	fromPosting, err := fromOpening(rs.Opening)
	if err != nil {
		return nil, err
	}
	for _, r := range rs.Records {
		t, err := fromQIFRecord(r, fromPosting)
		if err != nil {
			glog.Errorf("Converting from QIF %v error: %v", r, err)
			return nil, err
		}
		txns = append(txns, t)
	}
	return txns, nil
}

func fromQIFRecord(r *qif.Record, fromPosting *model.Posting) (*model.Transaction, error) {
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
	if len(r.Splits) > 0 {
		for _, s := range r.Splits {
			p, err := fromSplit(s)
			if err != nil {
				return nil, err
			}
			txn.Postings = append(txn.Postings, *p)
		}
		txn.Postings = append(txn.Postings, *fromPosting)
		return txn, nil
	}
	// Regular, unsplit transaction.  This can include inter-account transfers,
	// if we find one we fix up to mention transfer_account.
	var p *model.Posting
	category := reformatCategory(r.Label)
	if r.Transfer {
		category = "transfer_account"
		if strings.HasPrefix(r.Amount, "-") {
			txn.Comment = fmt.Sprintf("transfer-to:%q", r.Label)
		} else {
			txn.Comment = fmt.Sprintf("transfer-from:%q", r.Label)
		}
	}
	p, err = fromSplit(&qif.Split{
		Amount:   r.Amount,
		Category: category,
	})
	if err != nil {
		return nil, err
	}
	txn.Postings = append(txn.Postings, []model.Posting{*p, *fromPosting}...)
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

func fromOpening(op *qif.Record) (*model.Posting, error) {
	return fromSplit(&qif.Split{Category: reformatCategory(op.Label), Amount: "0"})
}

func fromSplit(s *qif.Split) (*model.Posting, error) {
	amount, err := strconv.ParseFloat(sanitizeAmount(s.Amount), 64)
	if err != nil {
		return nil, err
	}
	var ac model.Account
	if s.Category == "" {
		ac = append(ac, "((unknown account))")
	} else {
		for _, s := range strings.Split(s.Category, ":") {
			ac = append(ac, s)
		}
	}
	glog.Infof("fromSplit: account=%q", ac)
	return &model.Posting{Amount: -amount, Account: ac}, nil
}

func sanitizeAmount(a string) string {
	// Remove , 1000s separator
	return strings.Replace(a, ",", "", -1)
}

func reformatCategory(c string) string {
	switch c {
	case "Abbey ex-TESSA":
		return "assets:bank:abbey national:paul:ex-tessa"
	case "Joint - smile Current":
		return "assets:bank:smile:joint:current"
	case "Joint - smile Savings":
		return "assets:bank:smile:joint:savings"
	case "Joint - smile Savings 2 (house)":
		return "assets:bank:smile:joint:savings 2 (house)"
	case "Miranda - Halifax  Savings":
		return "assets:bank:halifax:miranda:savings"
	case "Oscar - goHenry":
		return "assets:bank:gohenry:oscar"
	case "Oscar - Halifax Ch Regular Saver":
		return "assets:bank:halifax:oscar:ch regular saver"
	case "Oscar - Halifax Save4It":
			return "assets:bank:halifax:oscar:save4it"
	case "Paul - Barclays Current":
		return "assets:bank:barclays:paul:current"
	case "Paul - Cahoot Credit Card":
		return "liabilities:bank:cahoot:paul:credit card"
	case "Paul - Cahoot Current":
		return "assets:bank:cahoot:paul:current"
	case "Paul - Cahoot Savings":
		return "assets:bank:cahoot:paul:savings"
	case "Paul - Halifax Savings":
			return "assets:bank:halifax:paul:savings"
	case "Paul - Monese Current":
		return "assets:bank:monese:paul:current"
	case "Paul - Monese Prepay - Old":
		return "assets:bank:monese:paul:prepay - old"
	case "Paul - Monzo Current":
		return "assets:bank:monzo:paul:current"
	case "Paul - Monzo Mastercard":
		return "assets:bank:monzo:paul:mastercard"
	case "Paul - smile cash mini-ISA":
		return "assets:bank:smile:paul:cash mini-isa"
	case "Paul - smile Current":
		return "assets:bank:smile:paul:current"
	case "Rachel - Barclays Savings":
		return "assets:bank:barclays:rachel:savings"
	case "Rachel - HSBC Current":
		return "assets:bank:hsbc:rachel:current"
	case "Rachel - HSBC Savings":
		return "assets:bank:hsbc:rachel:savings"
	case "Rachel - Monzo Cureent":
		return "assets:bank:monzo:rachel:current"
	case "Rachel - Smile Current":
		return "assets:bank:smile:rachel:current"
	case "Rachel - Smile Mini ISA":
		return "assets:bank:smile:rachel:mini_-isa"
	}
	// glog.Infof("Didn't reformat %s", c)
	return c
}
