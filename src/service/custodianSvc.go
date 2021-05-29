package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/bottlepay/portfolio-data/model"
)

// CustodianSvc runs HTTP requests to get Custodian data
type CustodianSvc struct {
	url string
}

// NewCustodianSvc creates a new CustodianSvc with the specified base URL
func NewCustodianSvc(u string) *CustodianSvc {
	return &CustodianSvc{u}
}

// FetchFromCustodian will return Custodian records for the specified IDs
func (c *CustodianSvc) FetchFromCustodian(ctx context.Context, custodianIDs ...int32) ([]*model.Custodian, error) {
	results := make([]*model.Custodian, len(custodianIDs))

	for i, custID := range custodianIDs {
		custURL := c.url + strconv.Itoa(int(custID))
		req, err := http.NewRequestWithContext(ctx, "GET", custURL, nil)
		if err != nil {
			return nil, fmt.Errorf("Custodian GET request Error: %v", err)
		}

		client := http.DefaultClient
		res, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("Custodian GET Error: %v", err)
		}

		if res.StatusCode != 200 {
			return nil, fmt.Errorf("Custodian GET Error: status code== %d", res.StatusCode)
		}
		defer res.Body.Close()

		cust := &model.Custodian{}
		if err = json.NewDecoder(res.Body).Decode(cust); err != nil {
			return nil, fmt.Errorf("Custodian GET JSON: %v", err)
		}
		results[i] = cust
	}

	return results, nil
}
