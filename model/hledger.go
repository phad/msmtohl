// Package model contains the types used to model transactions and serialise them in the hledger journal format.
package model

import (
	"fmt"
	"io"
	"strings"
)

// SerializeHledger writes a text representation in the hledger format of the Transaction to the given Writer.
func (t *Transaction) SerializeHledger(w io.Writer) error {
	if t == nil {
		return nil
	}
	topLine := t.topLine() + "\n"
	if _, err := w.Write([]byte(topLine)); err != nil {
		return err
	}
	for i, p := range t.Postings {
		last := i == len(t.Postings)-1
		entLine := fmt.Sprintf("  %s\n", p.postingLine(last))
		if _, err := w.Write([]byte(entLine)); err != nil {
			return err
		}
	}
	return nil
}

func (t *Transaction) topLine() string {
	if t == nil {
		return ""
	}
	items := []string{t.Date.Format("2006/01/02")}
	if t.Status != Unknown && t.Status != Unmarked {
		items = append(items, t.Status.String())
	}
	if len(t.Payee) > 0 {
		items = append(items, t.Payee)
	}
	if len(t.Payee) > 0 && len(t.Description) > 0 {
		items = append(items, "|")
	}
	if len(t.Description) > 0 {
		items = append(items, t.Description)
	}
	return strings.Join(items, " ")
}

func (p *Posting) postingLine(last bool) string {
	ac := ""
	for i, a := range p.Account {
		ac += strings.ToLower(strings.Replace(a, " ", "_", -1))
		if i < len(p.Account) - 1 {
			ac += ":"
		}
	}
	if last {
		return ac
	}
	// TODO(phad): optional status at start.
	return fmt.Sprintf("%s  %f", ac, p.Amount)
}

