package model

import (
	"fmt"
	"io"
	"strings"
)

func (t *Transaction) SerializeHledger(w io.Writer) error {
	if t == nil {
		return nil
	}
	topLine := t.topLine() + "\n"
	if _, err := w.Write([]byte(topLine)); err != nil {
		return err
	}
	for _, p := range t.Postings {
		entLine := fmt.Sprintf("  %s %f\n", p.Account, p.Amount)
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
