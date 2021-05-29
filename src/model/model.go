package model

import (
	"sort"

	"github.com/shopspring/decimal"
)

const (
	DirectionIn  = "IN"
	DirectionOut = "OUT"
)

type TransactionType int

const (
	ExternalDeposit TransactionType = iota
	ExternalWithdrawal
	ForeignTransfer
	InternalAssetExchange
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

// SortAssetsByCode sorts the Assets array in place by code
func SortAssetsByCode(a []*Asset) {
	sort.Slice(a, func(i, j int) bool {
		return a[i].Code < a[j].Code
	})
}

type AssetExchange struct {
	From *Asset
	To   *Asset
}

type Transaction struct {
	ID        int32           `json:"id"`
	Asset     string          `json:"asset"`
	Amount    decimal.Decimal `json:"amount"`
	Direction string          `json:"direction"`

	RelatedCustodianID            int32 `json:"related_custodian_id,omitempty"`
	RelatedCustodianTransactionID int32 `json:"related_custodian_transaction_id,omitempty"`
}

// TransactionsByID returns transactions indexed by ID for faster lookup
func IndexTransactionsByID(txs []*Transaction) map[int32]*Transaction {
	m := make(map[int32]*Transaction, len(txs))
	for _, tx := range txs {
		m[tx.ID] = tx
	}
	return m
}

// GetTransactionType returns the Transaction Type
func (c *Custodian) GetTransactionType(t *Transaction) TransactionType {

	// It's implemented on Custodian instead of Transaction because it needs the CustodianId
	// to determine if the transaction is internal or not
	switch {
	case t.RelatedCustodianID == 0 &&
		t.RelatedCustodianTransactionID == 0 &&
		t.Direction == "IN":
		return ExternalDeposit
	case t.RelatedCustodianID == 0 &&
		t.RelatedCustodianTransactionID == 0 &&
		t.Direction == "OUT":
		return ExternalWithdrawal
	case t.RelatedCustodianID != c.ID:
		return ForeignTransfer
	case t.RelatedCustodianID == c.ID:
		return InternalAssetExchange
	default:
		panic("Invalid transaction type")
	}
}

// FilterTransactionsByType filters transactions by Type
func (c *Custodian) FilterTransactionsByType(txtype TransactionType) []*Transaction {
	txl := make([]*Transaction, 0)
	for _, tx := range c.Transactions {
		if c.GetTransactionType(tx) == txtype {
			txl = append(txl, tx)
		}
	}
	return txl
}

// GetAssetExchanges returns the internal asset exchanges
func (c *Custodian) GetAssetExchanges() []AssetExchange {
	txs := c.FilterTransactionsByType(InternalAssetExchange)
	txsmap := IndexTransactionsByID(txs)

	var ex []AssetExchange

	for _, tx := range txs {
		if tx.Direction == "OUT" { // start with the OUT tx
			relatedTx := txsmap[tx.RelatedCustodianTransactionID]
			ex = append(ex, AssetExchange{
				From: &Asset{Code: tx.Asset, Balance: tx.Amount},
				To:   &Asset{Code: relatedTx.Asset, Balance: relatedTx.Amount},
			})
		}
	}
	return ex
}

type User struct {
	ID         int32   `json:"id"`
	Custodians []int32 `json:"custodians"`
}

func NewUser(id int32) *User {
	u := &User{id, []int32{}}
	return u
}

// AssetList makes adding assets together much easier
type AssetList struct {
	assetsMap map[string]*Asset
}

// NewAssetList creates a new AssetList
func NewAssetList(assets ...*Asset) *AssetList {
	res := &AssetList{}
	res.assetsMap = make(map[string]*Asset)

	for _, a := range assets {
		res.AddAssetValue(a)
	}
	return res
}

// AddAssetValue adds the value of an Asset to the AssetList
func (al *AssetList) add(code string, value decimal.Decimal) {
	if asset, found := al.assetsMap[code]; found {
		asset.Balance = asset.Balance.Add(value)
		return
	}
	// if not found, create the Asset in our map
	al.assetsMap[code] = &Asset{Code: code, Balance: value}
}

// AddAssetValue adds the asset and its value to the AssetList
func (al *AssetList) AddAssetValue(a *Asset) {
	al.add(a.Code, a.Balance)
}

// AddTransaction allows aggregation of transactions into an AssetList
func (al *AssetList) AddTransaction(tx *Transaction) {
	al.add(tx.Asset, tx.Amount)
}

// GetAssets returns the list of assets with total holding, Sorted by Asset Code
func (al *AssetList) GetAssets() []*Asset {
	result := make([]*Asset, 0, len(al.assetsMap))
	for _, a := range al.assetsMap {
		result = append(result, a)
	}
	SortAssetsByCode(result) // Sort by Code, makes testing easier, also prettier
	return result
}
