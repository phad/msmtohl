package converter

import (
	"reflect"
	"testing"
	"time"

	"github.com/phad/msmtohl/model"
	"github.com/phad/msmtohl/parser/qif"
)

func TestFromQIFRecord(t *testing.T) {
	tests := []struct{
		desc string
		qifRec *qif.Record
		opening *model.Posting
		want *model.Transaction
		wantErr bool
	}{
		{
			desc: "mangled date",
			qifRec: &qif.Record{
				Date: "12'02/2016",
				Cleared: "C",
				Payee: "Dave",
				Memo: "New shoes",
			},
			opening: &model.Posting{},
			wantErr: true,
		},
		{
			desc: "record with no splits",
			qifRec: &qif.Record{
				Date: "12/02'2016",
				Cleared: "C",
				Payee: "Dave",
				Amount: "123",
				Label: "Clothes:Shoes",
				Memo: "New shoes",
				Splits: []*qif.Split{},
			},
			opening: &model.Posting{
				Account: []string{"smile", "current"},
			},
			want: &model.Transaction{
				Date:        time.Date(2016, time.February, 12, 0, 0, 0, 0, time.UTC),
				Status:      model.Pending,
				Payee:       "Dave",
				Description: "New shoes",
				Postings:    []model.Posting{
					{Amount: -123.0, Account: []string{"Clothes", "Shoes"}},
					{Account: []string{"smile", "current"}},
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.desc, func(t *testing.T) {
			txn, err := fromQIFRecord(test.qifRec, test.opening)
			if gotErr := err != nil; gotErr != test.wantErr {
				t.Errorf("fromQIFRecord()=_, err? %t want? %t (err=%v)", gotErr, test.wantErr, err)
			}
			if !reflect.DeepEqual(txn, test.want) {
				t.Errorf("fromQIFRecord()=%v want %v", txn, test.want)
			}
		})
	}
}

func TestFromQIFStatus(t *testing.T) {
	tests := []struct{
		inputs []string
		want model.Status
	}{
		{
			inputs: []string{"", "A", "blah"},
			want: model.Unknown,
		},
		{
			inputs: []string{" "},
			want: model.Unmarked,
		},
		{
			inputs: []string{"*", "C"},
			want: model.Pending,
		},
		{
			inputs: []string{"R", "X"},
			want: model.Cleared,
		},
	}
	for _, test := range tests {
		for _, in := range test.inputs {
			if got, want := fromQIFStatus(in), test.want; got != want {
				t.Errorf("fromQIFStatus(%q)=%v want %v", in, got, want)
			}
		}
	}
}
