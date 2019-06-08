package qif

import (
	"reflect"
	"strings"
	"testing"
	"time"
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
			desc: "D Date line (older format)",
			qif: `D01/02/1996
^
`,
			wantRecs: []*Record{{Date: "01/02/1996"}},
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
			desc: "L Label line containing [account name]",
			qif: `L[Paul_-_smile_current]
^
`,
			wantRecs: []*Record{{Label: "Paul_-_smile_current", Transfer: true}},
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
		{
			desc: "funds transferred in",
			qif: `D28/11'2011
CX
MMonthly allowance in from joint ac
T800.00
PUs
L[Joint - smile Current]
^
`,
			wantRecs: []*Record{
				{Date: "28/11'2011", Amount: "800.00", Number: "", Cleared: "X", Payee: "Us", Label: "Joint - smile Current", Memo: "Monthly allowance in from joint ac", Transfer: true},
			},
			wantErrs: []bool{false},
			wantEOF:  true,
		},
		{
			desc: "funds transferred out",
			qif: `D28/11'2011
CX
MMonthly allowance out to paul ac
T-800.00
PPaul
L[Paul - smile Current]
^
`,
			wantRecs: []*Record{
				{Date: "28/11'2011", Amount: "-800.00", Number: "", Cleared: "X", Payee: "Paul", Label: "Paul - smile Current", Memo: "Monthly allowance out to paul ac", Transfer: true},
			},
			wantErrs: []bool{false},
			wantEOF:  true,
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

func TestSanitizeLabel(t *testing.T) {
	for _, tc := range []struct{
		in, wantOut  string
		wantTransfer bool
	}{
		{"", "", false},
		{"a", "a", false},
		{"[a", "[a", false},
		{"a]", "a]", false},
		{"[a]", "a", true},
		{"[]", "", true},
		{"[foo bar]", "foo bar", true},
	} {
		t.Run(tc.in, func(t *testing.T) {
			out, isTransfer := sanitizeLabel(tc.in)
			if out != tc.wantOut || isTransfer != tc.wantTransfer {
				t.Errorf("sanitizeLabel(%s)=%q,%t want %q,%t", tc.in, out, isTransfer, tc.wantOut, tc.wantTransfer)
			}
		})
	}
}

func TestParseDate(t *testing.T) {
	for _, tc := range []struct{
		in       string
		wantTime time.Time
		wantErr  bool
	}{
		{"", time.Time{}, true},
		{"not a date", time.Time{}, true},
		{"13-04-2006", time.Time{}, true},
		{"13/04'2006", time.Date(2006, time.April, 13, 0, 0, 0, 0, time.UTC), false},
		{"11/07/1970", time.Date(1970, time.July, 11, 0, 0, 0, 0, time.UTC), false},
	} {
		t.Run(tc.in, func(t *testing.T) {
			tm, err := ParseDate(tc.in)
			if gotErr := err != nil; gotErr != tc.wantErr {
				t.Errorf("ParseDate(%s)=err? %t want? %t (err=%v)", tc.in, gotErr, tc.wantErr, err)
			}
			if err != nil {
				return
			}
			if tm != tc.wantTime {
				t.Errorf("ParseDate(%s)=%v want %v", tc.in, tm, tc.wantTime)
			}
		})
	}
}
