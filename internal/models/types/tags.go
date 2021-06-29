package types

import "database/sql/driver"

// Tags -
type Tags int64

// NewTags -
func NewTags(value []string) Tags {
	t := Tags(0)
	for i := range value {
		switch value[i] {
		case ContractFactoryStringTag:
			t.Set(ContractFactoryTag)
		case DelegatableStringTag:
			t.Set(DelegatableTag)
		case DelegatorStringTag:
			t.Set(DelegatorTag)
		case ChainAwareStringTag:
			t.Set(ChainAwareTag)
		case CheckSigStringTag:
			t.Set(CheckSigTag)
		case SaplingStringTag:
			t.Set(SaplingTag)
		case FA1StringTag:
			t.Set(FA1Tag)
		case FA12StringTag:
			t.Set(FA12Tag)
		case FA2StringTag:
			t.Set(FA2Tag)
		case UpgradableStringTag:
			t.Set(UpgradableTag)
		case MultisigStringTag:
			t.Set(MultisigTag)
		case ViewAddressStringTag:
			t.Set(ViewAddressTag)
		case ViewBalanceOfStringTag:
			t.Set(ViewBalanceOfTag)
		case ViewNatStringTag:
			t.Set(ViewNatTag)
		case LedgerStringTag:
			t.Set(LedgerTag)
		}
	}
	return t
}

// String -
func (t Tags) ToArray() []string {
	value := make([]string, 0)
	for _, tag := range []Tags{
		ContractFactoryTag,
		DelegatableTag,
		DelegatorTag,
		ChainAwareTag,
		CheckSigTag,
		SaplingTag,
		FA1Tag,
		FA12Tag,
		FA2Tag,
		UpgradableTag,
		MultisigTag,
		ViewAddressTag,
		ViewBalanceOfTag,
		ViewNatTag,
		LedgerTag,
	} {
		if t.Has(tag) {
			switch tag {
			case ContractFactoryTag:
				value = append(value, ContractFactoryStringTag)
			case DelegatableTag:
				value = append(value, DelegatableStringTag)
			case DelegatorTag:
				value = append(value, DelegatorStringTag)
			case ChainAwareTag:
				value = append(value, ChainAwareStringTag)
			case CheckSigTag:
				value = append(value, CheckSigStringTag)
			case SaplingTag:
				value = append(value, SaplingStringTag)
			case FA1Tag:
				value = append(value, FA1StringTag)
			case FA12Tag:
				value = append(value, FA12StringTag)
			case FA2Tag:
				value = append(value, FA2StringTag)
			case UpgradableTag:
				value = append(value, UpgradableStringTag)
			case MultisigTag:
				value = append(value, MultisigStringTag)
			case ViewAddressTag:
				value = append(value, ViewAddressStringTag)
			case ViewBalanceOfTag:
				value = append(value, ViewBalanceOfStringTag)
			case ViewNatTag:
				value = append(value, ViewNatStringTag)
			case LedgerTag:
				value = append(value, LedgerStringTag)
			}
		}
	}
	return value
}

// Set -
func (t *Tags) Set(flag Tags) { *t |= flag }

// Clear -
func (t *Tags) Clear(flag Tags) { *t &^= flag }

// Toggle -
func (t *Tags) Toggle(flag Tags) { *t ^= flag }

// Has -
func (t *Tags) Has(flag Tags) bool { return *t&flag != 0 }

// Scan -
func (t *Tags) Scan(value interface{}) error {
	*t = Tags(value.(int64))
	return nil
}

// Value -
func (t Tags) Value() (driver.Value, error) { return int(t), nil }

// Tags name
const (
	ContractFactoryStringTag  = "CREATE_CONTRACT"
	DelegatableStringTag      = "SET_DELEGATE"
	DelegatorStringTag        = "delegator"
	ChainAwareStringTag       = "CHAIN_ID"
	CheckSigStringTag         = "CHECK_SIGNATURE"
	SaplingStringTag          = "sapling"
	FA1StringTag              = "fa1"
	FA12StringTag             = "fa1-2"
	FA2StringTag              = "fa2"
	UpgradableStringTag       = "upgradable"
	MultisigStringTag         = "multisig"
	ViewAddressStringTag      = "view_address"
	ViewBalanceOfStringTag    = "view_balance_of"
	ViewNatStringTag          = "view_nat"
	LedgerStringTag           = "ledger"
	ContractMetadataStringTag = "contract_metadata"
	TokenMetadataStringTag    = "token_metadata"
)

// Tags
const (
	ContractFactoryTag Tags = 1 << iota
	DelegatableTag
	DelegatorTag
	ChainAwareTag
	CheckSigTag
	SaplingTag
	FA1Tag
	FA12Tag
	FA2Tag
	UpgradableTag
	MultisigTag
	ViewAddressTag
	ViewBalanceOfTag
	ViewNatTag
	LedgerTag
	ContractMetadataTag
	TokenMetadataTag
)

const (
	EmptyTag Tags = 0
)
