package model

import (
	"reflect"
	"testing"

	"github.com/shopspring/decimal"
)

func TestAggregateHoldings(t *testing.T) {
	type args struct {
		custodians []*Custodian
	}
	tests := []struct {
		name string
		args args
		want []*Asset
	}{
		{"empty", args{}, []*Asset{}},
		{"one empty custodian", args{[]*Custodian{
			{1, nil, nil},
		}}, []*Asset{}},
		{"one custodian one asset", args{[]*Custodian{
			{1, []*Asset{
				{Code: "USD", Balance: decimal.NewFromInt(1000)},
			}, nil},
		}}, []*Asset{
			{Code: "USD", Balance: decimal.NewFromInt(1000)},
		}},
		{"two custodians one asset", args{[]*Custodian{
			{1, []*Asset{
				{Code: "USD", Balance: decimal.NewFromInt(1000)},
			}, nil},
			{2, []*Asset{
				{Code: "USD", Balance: decimal.NewFromInt(500)},
			}, nil},
		}}, []*Asset{
			{Code: "USD", Balance: decimal.NewFromInt(1500)},
		}},
		{"two custodians one asset negative balance", args{[]*Custodian{
			{1, []*Asset{
				{Code: "USD", Balance: decimal.NewFromInt(1000)},
			}, nil},
			{2, []*Asset{
				{Code: "USD", Balance: decimal.NewFromInt(-150)},
			}, nil},
		}}, []*Asset{
			{Code: "USD", Balance: decimal.NewFromInt(850)},
		}},
		{"two custodians two assets", args{[]*Custodian{
			{1, []*Asset{
				{Code: "USD", Balance: decimal.NewFromInt(1000)},
				{Code: "BTC", Balance: decimal.NewFromInt(3)},
			}, nil},
			{2, []*Asset{
				{Code: "USD", Balance: decimal.NewFromInt(500)},
			}, nil},
		}}, []*Asset{
			{Code: "BTC", Balance: decimal.NewFromInt(3)}, // alpha order by Code
			{Code: "USD", Balance: decimal.NewFromInt(1500)},
		}},
		{"two custodians two assets variable orders", args{[]*Custodian{
			{1, []*Asset{
				{Code: "BTC", Balance: decimal.NewFromInt(3)},
				{Code: "USD", Balance: decimal.NewFromInt(1000)},
			}, nil},
			{2, []*Asset{
				{Code: "USD", Balance: decimal.NewFromInt(500)},
				{Code: "BTC", Balance: decimal.NewFromInt(1)},
			}, nil},
		}}, []*Asset{
			{Code: "BTC", Balance: decimal.NewFromInt(4)},
			{Code: "USD", Balance: decimal.NewFromInt(1500)},
		}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AggregateHoldings(tt.args.custodians); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AggregateHoldings() = %v, want %v", got, tt.want)
			}
		})
	}
}
