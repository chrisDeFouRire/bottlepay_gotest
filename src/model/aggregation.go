package model

// AggregateHoldings returns the total holdings accross all assets of all custodians passed in params.
// Returns an Array of Assets, sorted by Asset.Code
func AggregateHoldings(custodians []*Custodian) []*Asset {
	al := NewAssetList()

	for _, c := range custodians {
		for _, asset := range c.Assets {
			al.AddAssetValue(asset)
		}
	}
	return al.GetAssets()
}
