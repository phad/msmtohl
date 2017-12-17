package qif

import (
	"reflect"
	"strings"
	"testing"
)

func TestNext(t *testing.T) {
	tests := []struct {
		desc     string
		qif      string
		wantNum  int
		wantRecs []*Record
		wantErrs []bool
		wantEOF  bool
	}{
		{
			desc:    "empty",
			wantEOF: true,
		},
		{
			desc:     "empty record",
			qif:      `^`,
			wantRecs: []*Record{{}},
			wantErrs: []bool{false},
			wantEOF:  true,
		},
		{
			desc:     "text following record separator is ignored",
			qif:      `^ignored`,
			wantRecs: []*Record{{}},
			wantErrs: []bool{false},
			wantEOF:  true,
		},
		{
			desc: "several empty records",
			qif: `^
^
^`,
			wantRecs: []*Record{{}, {}, {}},
			wantErrs: []bool{false, false, false},
			wantEOF:  true,
		},
		{
			desc: "unclosed record",
			qif: `!Type:Foo
D15/03'2003
`,
			wantEOF: true,
		},
		{
			desc: "! Type line",
			qif: `!Type:Foo
^
`,
			wantRecs: []*Record{{Type: "Type:Foo"}},
			wantErrs: []bool{false},
			wantEOF:  true,
		},
		{
			desc: "D Date line",
			qif: `D15/03'2003
^
`,
			wantRecs: []*Record{{Date: "15/03'2003"}},
			wantErrs: []bool{false},
			wantEOF:  true,
		},
		{
			desc: "T Transaction amount line",
			qif: `T10.00
^
`,
			wantRecs: []*Record{{Amount: "10.00"}},
			wantErrs: []bool{false},
			wantEOF:  true,
		},
		{
			desc: "U Transaction amount line (alternative to T)",
			qif: `U10.00
^
`,
			wantRecs: []*Record{{Amount: "10.00"}},
			wantErrs: []bool{false},
			wantEOF:  true,
		},
		{
			desc: "N check number line",
			qif: `N123456
^
`,
			wantRecs: []*Record{{Number: "123456"}},
			wantErrs: []bool{false},
			wantEOF:  true,
		},
		{
			desc: "C Cleared status line",
			qif: `CX
^
`,
			wantRecs: []*Record{{Cleared: "X"}},
			wantErrs: []bool{false},
			wantEOF:  true,
		},
		{
			desc: "P Payee line",
			qif: `PJohn Lewis
^
`,
			wantRecs: []*Record{{Payee: "John Lewis"}},
			wantErrs: []bool{false},
			wantEOF:  true,
		},
		{
			desc: "L Label line",
			qif: `LFood:Groceries
^
`,
			wantRecs: []*Record{{Label: "Food:Groceries"}},
			wantErrs: []bool{false},
			wantEOF:  true,
		},
		{
			desc: "M Memo line",
			qif: `MShopping
^
`,
			wantRecs: []*Record{{Memo: "Shopping"}},
			wantErrs: []bool{false},
			wantEOF:  true,
		},
		{
			desc: "complete record",
			qif: `D15/03'2003
CX
MPaint
T-26.07
NVISA
PHomebase
LHousing:Improvements
^
`,
			wantRecs: []*Record{
				{Date: "15/03'2003", Amount: "-26.07", Number: "VISA", Cleared: "X", Payee: "Homebase", Label: "Housing:Improvements", Memo: "Paint"},
			},
			wantErrs: []bool{false},
			wantEOF:  true,
		},
		{
			desc: "record with splits",
			qif: `D24/11'2004
CX
MLunch/early dinner at Heathrow for me and R
T-14.40
NVISA
PThe Bridge Bar
LFood:Dining Out
SFood:Dining Out
ELunch/early dinner
$-10.00
SDrink
EBeer & juice
$-4.40
^
`,
			wantRecs: []*Record{
				{
					Date:    "24/11'2004",
					Amount:  "-14.40",
					Number:  "VISA",
					Cleared: "X",
					Payee:   "The Bridge Bar",
					Label:   "Food:Dining Out",
					Memo:    "Lunch/early dinner at Heathrow for me and R",
					Splits: []*Split{
						{
							Category: "Food:Dining Out",
							Memo:     "Lunch/early dinner",
							Amount:   "-10.00",
						},
						{
							Category: "Drink",
							Memo:     "Beer & juice",
							Amount:   "-4.40",
						},
					},
				},
			},
			wantErrs: []bool{false},
			wantEOF:  true,
		},
		{
			desc: "unsupported Split percentage field",
			qif: `D24/11'2004
SFood:Dining Out                               
ELunch/early dinner                   
%25.00           
^   
`,
			wantEOF:  true,
			wantRecs: []*Record{nil},
			wantErrs: []bool{true},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			rd := strings.NewReader(test.qif)
			qif := New(rd)
			count := 0
			for {
				r, e := qif.Next()
				if e == ErrEOF {
					if !test.wantEOF || count < len(test.wantRecs) {
						t.Errorf("%s: Next()=_,EOF prematurely at count=%d want count=%d", test.desc, count, len(test.wantRecs))
					}
					break
				}
				wantErr := test.wantErrs[count]
				if gotErr := e != nil; gotErr != wantErr {
					t.Errorf("%s: Next()=_,err? %t wantErr? %t (err=%v)", test.desc, gotErr, wantErr, e)
				}
				if got, want := r, test.wantRecs[count]; !reflect.DeepEqual(got, want) {
					t.Errorf("%s: Next()=%v,_ want %v", test.desc, got, want)
				}
				count++
			}
		})
	}
}
