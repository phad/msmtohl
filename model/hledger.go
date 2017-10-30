package model

import (
	"fmt"
	"io"
)

func (t *Transaction) SerializeHledger(w io.Writer) error {
	topLine := fmt.Sprintf("%s %s %s | %s\n", t.Date.Format("02/01/2006"), t.Status, t.Payee, t.Description)
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