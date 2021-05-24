package store

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"sync"

	"github.com/bottlepay/portfolio-data/model"
	"github.com/shopspring/decimal"
)

type Store struct {
	custodians    []*model.Custodian
	custodiansMap map[int32]*model.Custodian

	lock sync.RWMutex

	stateFile string
}

// IsEmpty returns true if the store is currently empty
func (s *Store) IsEmpty() bool {
	return len(s.custodians) == 0
}

// AddCustodian adds one or more new custodians to the data store
func (s *Store) AddCustodian(c ...*model.Custodian) {
	s.lock.Lock()
	defer s.lock.Unlock()

	// determine the last ID we added
	startID := int32(0)
	if len(s.custodians) > 0 {
		startID = s.custodians[len(s.custodians)-1].ID
	}

	// Set the IDs of the custodians we're adding
	for idx, custodian := range c {
		custodian.ID = startID + int32(idx) + 1
		s.custodiansMap[custodian.ID] = custodian
	}

	s.custodians = append(s.custodians, c...)
}

// GetCustodian gets a custodian by its ID
func (s *Store) GetCustodian(id int32) *model.Custodian {
	s.lock.RLock()
	defer s.lock.RUnlock()

	custodian, _ := s.custodiansMap[id]
	return custodian
}

// GetCustodiansWithoutTransactions gets a clone of the custodians list without their transactions
func (s *Store) GetCustodiansWithoutTransactions() []*model.Custodian {
	s.lock.RLock()
	defer s.lock.RUnlock()

	ret := make([]*model.Custodian, 0, len(s.custodians))

	for _, c := range s.custodians {
		ret = append(ret, &model.Custodian{
			ID:     c.ID,
			Assets: c.Assets,
		})
	}

	return ret
}

// Snapshot saves a snapshot to the configured state file
func (s *Store) Snapshot() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	data, err := json.MarshalIndent(s.custodians, "", "	")
	if err != nil {
		return err
	}

	return ioutil.WriteFile(s.stateFile, data, 0777)
}

// AddRandomEvent adds a new random event to one or more custodians
func (s *Store) AddRandomEvent() error {
	s.lock.Lock()
	defer s.lock.Unlock()

	// Pick the first custodian
	custodian := s.custodians[rand.Intn(len(s.custodians))]

	// Pick the asset
	asset := custodian.Assets[rand.Intn(len(custodian.Assets))]

	var otherCustodian *model.Custodian
	var otherAsset *model.Asset

	// Determine if we want another custodian involved
	if rand.Intn(100) < 60 {
		// Pick another custodian. They must be:
		// - the current custodian, if it supports multiple assets
		// - another custodian, if it has the same asset
	selectorLoop:
		for {
			otherCustodian = s.custodians[rand.Intn(len(s.custodians))]
			if otherCustodian.ID == custodian.ID && len(custodian.Assets) < 2 {
				continue
			}

			// If it's the same custodian, pick another asset
			if otherCustodian.ID == custodian.ID {
				for _, otherAsset = range otherCustodian.Assets {
					if otherAsset.Code != asset.Code {
						break selectorLoop
					}
				}
			}

			// Otherwise pick a separate asset
			for _, otherAsset = range otherCustodian.Assets {
				if otherAsset.Code == asset.Code {
					break selectorLoop
				}
			}
		}
	}

	// Determine the amount to use for this transaction.
	// We'll use up to 20% of the current balance.
	percentage := rand.Intn(20)
	amount := asset.Balance.Mul(decimal.NewFromFloat(float64(percentage) / 100))

	// If there's another custodian involved then it's a simple transfer
	if otherCustodian != nil {
		transactionOut := &model.Transaction{
			Asset:              asset.Code,
			Amount:             amount,
			Direction:          model.DirectionOut,
			RelatedCustodianID: otherCustodian.ID,
		}

		transactionIn := &model.Transaction{
			Asset:              otherAsset.Code,
			Amount:             amount,
			Direction:          model.DirectionIn,
			RelatedCustodianID: custodian.ID,
		}

		// If the assets are different then do forex
		if asset.Code != otherAsset.Code {
			rate, err := ForexRate(asset.Code, otherAsset.Code)
			if err != nil {
				return fmt.Errorf("error converting %s to %s", asset.Code, otherAsset.Code)
			}

			transactionIn.Amount = transactionIn.Amount.Mul(rate)
		}

		// Add the transactions to the custodians
		custodian.AddTransaction(transactionOut)
		otherCustodian.AddTransaction(transactionIn)

		// Modify the asset balances
		asset.Balance = asset.Balance.Sub(transactionOut.Amount).RoundBank(8)
		otherAsset.Balance = otherAsset.Balance.Add(transactionIn.Amount).RoundBank(8)

		// Make the transactions reference each other
		transactionOut.RelatedCustodianTransactionID = transactionIn.ID
		transactionIn.RelatedCustodianTransactionID = transactionOut.ID

		return nil
	}

	// If there's no other custodian involved then it's a deposit or withdrawal
	transaction := &model.Transaction{
		Asset:     asset.Code,
		Amount:    amount,
		Direction: model.DirectionIn,
	}
	if rand.Intn(10) < 5 {
		transaction.Direction = model.DirectionOut
		asset.Balance = asset.Balance.Sub(amount).RoundBank(8)
	} else {
		transaction.Amount = amount.Mul(decimal.NewFromInt(4))
		asset.Balance = asset.Balance.Add(transaction.Amount).RoundBank(8)
	}
	custodian.AddTransaction(transaction)

	return nil
}

// NewStore creates a new data Store, persisted to stateFile
func NewStore(stateFile string) (*Store, error) {
	// Read the existing state file (if it exists)
	stateData, err := ioutil.ReadFile(stateFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}

	store := &Store{
		stateFile:     stateFile,
		custodiansMap: make(map[int32]*model.Custodian),
	}

	// Unmarshal the data if we have any
	if len(stateData) != 0 {
		if err = json.Unmarshal(stateData, &store.custodians); err != nil {
			return nil, err
		}
	}

	// Setup the map of IDs -> Custodians
	for _, c := range store.custodians {
		store.custodiansMap[c.ID] = c
	}

	return store, nil
}
