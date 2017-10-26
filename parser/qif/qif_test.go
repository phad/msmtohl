package qif

import (
	"reflect"
	"strings"
	"testing"
)

func TestNext(t *testing.T) {
	tests := []struct{
		desc     string
		qif      string
		wantNum  int
		wantRecs []*Record
		wantErrs []bool
		wantEof  bool
	}{
		{
			desc: "empty",
			wantEof: true,
		},
		{
			desc: "empty record",
			qif: `^`,
			wantRecs: []*Record{&Record{}},
			wantErrs: []bool{false},
			wantEof: true,
		},
		{
			desc: "text following record separator is ignored",
                        qif: `^ignored`,
                        wantRecs: []*Record{&Record{}},
                        wantErrs: []bool{false},
                        wantEof: true,
		},
		{
			desc: "several empty records",
                        qif: `^
^
^`,
                        wantRecs: []*Record{&Record{}, &Record{}, &Record{}},
                        wantErrs: []bool{false, false, false},
                        wantEof: true,
		},
		{
			desc: "unclosed record",
			qif: `!Type:Foo
D15/03'2003
`,
                        wantEof: true,
		},
                {
                        desc: "! Type line",
                        qif: `!Type:Foo
^
`,
                        wantRecs: []*Record{&Record{Type: "Type:Foo"}},
                        wantErrs: []bool{false},			
                        wantEof: true,
                },
                {
                        desc: "D Date line",
                        qif: `D15/03'2003
^
`,
                        wantRecs: []*Record{&Record{Date: "15/03'2003"}},
                        wantErrs: []bool{false},
                        wantEof: true,
                },
                {
                        desc: "T Transaction amount line",
                        qif: `T10.00
^
`,
                        wantRecs: []*Record{&Record{Amount: "10.00"}},
                        wantErrs: []bool{false},
                        wantEof: true,
                },
                {
                        desc: "U Transaction amount line (alternative to T)",
                        qif: `U10.00
^
`,
                        wantRecs: []*Record{&Record{Amount: "10.00"}},
                        wantErrs: []bool{false},
                        wantEof: true,
                },
                {
                        desc: "N check number line",
                        qif: `N123456
^
`,
                        wantRecs: []*Record{&Record{Number: "123456"}},
                        wantErrs: []bool{false},
                        wantEof: true,
                },
                {
                        desc: "C Transaction status line", 
                        qif: `CX
^
`,
                        wantRecs: []*Record{&Record{Status: "X"}},
                        wantErrs: []bool{false},
                        wantEof: true,
                },
                {
                        desc: "P Payee line",
                        qif: `PJohn Lewis
^
`,
                        wantRecs: []*Record{&Record{Payee: "John Lewis"}},
                        wantErrs: []bool{false},
                        wantEof: true,
                },
                {
                        desc: "L Label line",
                        qif: `LFood:Groceries
^
`,
                        wantRecs: []*Record{&Record{Label: "Food:Groceries"}},
                        wantErrs: []bool{false},
                        wantEof: true,
                },
                {
                        desc: "M Memo line",
                        qif: `MShopping
^
`,
                        wantRecs: []*Record{&Record{Memo: "Shopping"}},
                        wantErrs: []bool{false},
                        wantEof: true,
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
				&Record{Date:"15/03'2003", Amount: "-26.07", Number: "VISA", Status: "X", Payee: "Homebase", Label: "Housing:Improvements", Memo: "Paint"}, 
			},
                        wantErrs: []bool{false},
                        wantEof: true,
                },

	}

	for _, test := range tests {
		rd := strings.NewReader(test.qif)
		qif := New(rd)
		count := 0
		for {
			r, e := qif.Next()
			t.Logf("Next()=%d %v %v", count, r, e)
			if e == ErrEOF {
				if !test.wantEof || count < len(test.wantRecs) {
					t.Errorf("%s: Next()=_,EOF prematurely at count=%d want count=%d", test.desc, count, len(test.wantRecs))
				}
				break
			}
			if gotErr, wantErr := e != nil, test.wantErrs[count]; gotErr != wantErr {
				t.Errorf("%s: Next()=_,err? %t wantErr? %t (err=%v)", test.desc, gotErr, wantErr, e)
			}
			if got, want := r, test.wantRecs[count]; !reflect.DeepEqual(got, want) {
				t.Errorf("%s: Next()=%v,_ want %v", test.desc, got, want)
			}
			count++
		}
	}
}

