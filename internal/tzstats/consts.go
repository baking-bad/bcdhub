package tzstats

// Avaliable URLs
const (
	MainNetURL    = "https://api.tzstats.com"
	ZeroNetURL    = "https://api.zeronet.tzstats.com"
	BabylonNetURL = "https://api.babylonnet.tzstats.com"
)

// Operation types
const (
	OperationTypeAll         = "n_ops"
	OperationTypeTx          = "n_tx"
	OperationTypeDelegation  = "n_delegation"
	OperationTypeOrigination = "n_origination"
	OperationTypeProposal    = "n_proposal"
	OperationTypeBallot      = "n_ballot"
)

// Operation kinds
const (
	OperationKindTransaction         = "transaction"
	OperationKindDelegation          = "delegation"
	OperationKindOrigination         = "origination"
	OperationKindReveal              = "reveal"
	OperationKindProposals           = "proposals"
	OperationKindBallot              = "ballot"
	OperationKindActivateAccount     = "activate_account"
	OperationKindDBEvidence          = "double_baking_evidence"
	OperationKindDEEvidence          = "double_endorsement_evidence"
	OperationKindSeedNonceRevelation = "seed_nonce_revelation"
)

// Tables
const (
	TableChain     = "chain"
	TableSupply    = "supply"
	TableBlock     = "block"
	TableOperation = "op"
	TableAccount   = "account"
	TableContract  = "contract"
	TableFlow      = "flow"
	TableRights    = "rights"
	TableSnapshot  = "snapshot"
	TableIncome    = "income"
	TableElection  = "election"
	TableProposal  = "proposal"
	TableVote      = "vote"
	TableBallot    = "ballot"
)
