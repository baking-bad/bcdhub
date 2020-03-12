package tzkt

import "time"

// Head -
type Head struct {
	Level     int64     `json:"level"`
	Hash      string    `json:"hash"`
	Timestamp time.Time `json:"timestamp"`
}

// Address -
type Address struct {
	Alias   string `json:"alias,omitempty"`
	Address string `json:"address"`
	Active  bool   `json:"active"`
}

// Origination -
type Origination struct {
	ID                 int64     `json:"id"`
	Type               string    `json:"type"`
	Level              int64     `json:"level"`
	Timestamp          time.Time `json:"timestamp"`
	Hash               string    `json:"hash"`
	Counter            int64     `json:"counter"`
	Sender             Address   `json:"sender"`
	GasLimit           int64     `json:"gasLimit"`
	GasUsed            int64     `json:"gasUsed"`
	StorageLimit       int64     `json:"storageLimit"`
	StorageUsed        int64     `json:"storageUsed"`
	BakerFee           int64     `json:"bakerFee"`
	StorageFee         int64     `json:"storageFee"`
	AllocationFee      int64     `json:"allocationFee"`
	ContractBalance    int64     `json:"contractBalance"`
	ContractManager    Address   `json:"contractManager"`
	ContractDelegate   Address   `json:"contractDelegate,omitempty"`
	Status             string    `json:"status"`
	OriginatedContract struct {
		Kind    string `json:"kind"`
		Address string `json:"address"`
	} `json:"originatedContract"`
}

// SystemOperation -
type SystemOperation struct {
	Type          string    `json:"type"`
	ID            int64     `json:"id"`
	Level         int64     `json:"level"`
	Timestamp     time.Time `json:"timestamp"`
	Kind          string    `json:"kind"`
	Account       Address   `json:"account"`
	BalanceChange int64     `json:"balanceChange"`
}

// Account -
type Account struct {
	Type              string    `json:"type"`
	Kind              string    `json:"kind"`
	Alias             string    `json:"alias"`
	Address           string    `json:"address"`
	Balance           int64     `json:"balance"`
	Creator           *Address  `json:"creator,omitempty"`
	Manager           *Address  `json:"manager,omitempty"`
	Delegate          *Address  `json:"delegate,omitempty"`
	DelegationLevel   int64     `json:"delegationLevel"`
	DelegationTime    time.Time `json:"delegationTime"`
	NumContracts      int64     `json:"numContracts"`
	NumDelegations    int64     `json:"numDelegations"`
	NumOriginations   int64     `json:"numOriginations"`
	NumTransactions   int64     `json:"numTransactions"`
	NumReveals        int64     `json:"numReveals"`
	NumMigrations     int64     `json:"numMigrations"`
	FirstActivity     int64     `json:"firstActivity"`
	FirstActivityTime time.Time `json:"firstActivityTime"`
	LastActivity      int64     `json:"lastActivity"`
	LastActivityTime  time.Time `json:"lastActivityTime"`
}
