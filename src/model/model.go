package model

import (
	"github.com/shopspring/decimal"
)

const (
	DirectionIn  = "IN"
	DirectionOut = "OUT"
)

type Custodian struct {
	ID int32 `json:"id"`

	Assets       []*Asset       `json:"assets,omitempty"`
	Transactions []*Transaction `json:"transactions,omitempty"`
}

// AddTransaction adds one or more new transactions to the custodian
func (c *Custodian) AddTransaction(t ...*Transaction) {
	// determine the last ID we added
	startID := int32(0)
	if len(c.Transactions) > 0 {
		startID = c.Transactions[len(c.Transactions)-1].ID
	}

	// Set the IDs of the transactions we're adding
	for idx, transaction := range t {
		transaction.ID = startID + int32(idx) + 1
	}

	c.Transactions = append(c.Transactions, t...)
}

type Asset struct {
	Code    string          `json:"code"`
	Balance decimal.Decimal `json:"balance"`
}

type Transaction struct {
	ID        int32           `json:"id"`
	Asset     string          `json:"asset"`
	Amount    decimal.Decimal `json:"amount"`
	Direction string          `json:"direction"`

	RelatedCustodianID            int32 `json:"related_custodian_id,omitempty"`
	RelatedCustodianTransactionID int32 `json:"related_custodian_transaction_id,omitempty"`
}

type User struct {
	ID         int32   `json:"id"`
	Custodians []int32 `json:"custodians"`
}

func NewUser(id int32) *User {
	u := &User{id, []int32{}}
	return u
}
