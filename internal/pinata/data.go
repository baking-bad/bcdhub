package pinata

import "time"

// PinList -
type PinList struct {
	Count int `json:"count"`
	Rows  []struct {
		ID           string      `json:"id"`
		IpfsPinHash  string      `json:"ipfs_pin_hash"`
		Size         int         `json:"size"`
		UserID       string      `json:"user_id"`
		DatePinned   time.Time   `json:"date_pinned"`
		DateUnpinned interface{} `json:"date_unpinned"`
		Metadata     struct {
			Name      interface{} `json:"name"`
			Keyvalues interface{} `json:"keyvalues"`
		} `json:"metadata"`
		Regions []struct {
			RegionID                string `json:"regionId"`
			CurrentReplicationCount int    `json:"currentReplicationCount"`
			DesiredReplicationCount int    `json:"desiredReplicationCount"`
		} `json:"regions"`
	} `json:"rows"`
}

// PinJSONResponse -
type PinJSONResponse struct {
	IpfsHash  string    `json:"IpfsHash"`
	PinSize   int       `json:"PinSize"`
	Timestamp time.Time `json:"Timestamp"`
}
