package core

import (
	"testing"

	"github.com/baking-bad/bcdhub/internal/models"
)

func TestEq_String(t *testing.T) {
	type fields struct {
	}
	tests := []struct {
		name  string
		field string
		value interface{}
		want  string
	}{
		{
			name:  "int",
			field: "test",
			value: 100,
			want:  "test = 100",
		}, {
			name:  "string",
			field: "test",
			value: "value",
			want:  "test = value",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			eq := NewEq(tt.field, tt.value)
			if got := eq.String(); got != tt.want {
				t.Errorf("Eq.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLte_String(t *testing.T) {
	tests := []struct {
		name  string
		field string
		value int64
		want  string
	}{
		{
			name:  "first",
			field: "test",
			value: 100,
			want:  "test <= 100",
		}, {
			name:  "second",
			field: "test",
			value: -10,
			want:  "test <= -10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lte := NewLte(tt.field, tt.value)
			if got := lte.String(); got != tt.want {
				t.Errorf("Lte.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGt_String(t *testing.T) {
	tests := []struct {
		name  string
		field string
		value int64
		want  string
	}{
		{
			name:  "first",
			field: "test",
			value: 100,
			want:  "test > 100",
		}, {
			name:  "second",
			field: "test",
			value: -10,
			want:  "test > -10",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gt := NewGt(tt.field, tt.value)
			if got := gt.String(); got != tt.want {
				t.Errorf("Gt.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAnd_String(t *testing.T) {
	tests := []struct {
		name string
		and  *And
		want string
	}{
		{
			name: "default",
			and:  NewAnd(NewEq("test", "100"), NewGt("field", 100)),
			want: "(test = 100 AND field > 100)",
		}, {
			name: "from map",
			and: NewAndFromMap(map[string]interface{}{
				"field": "a",
				"test":  100,
			}),
			want: "(field = a AND test = 100)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.and.String(); got != tt.want {
				t.Errorf("And.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOr_String(t *testing.T) {
	tests := []struct {
		name string
		or   *Or
		want string
	}{
		{
			name: "default",
			or:   NewOr(NewEq("test", "100"), NewGt("field", 100)),
			want: "(test = 100 OR field > 100)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.or.String(); got != tt.want {
				t.Errorf("Or.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIn_String(t *testing.T) {
	tests := []struct {
		name string
		in   *In
		want string
	}{
		{
			name: "default",
			in:   NewIn("test", "value1", "value2"),
			want: "test IN (value1, value2)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.in.String(); got != tt.want {
				t.Errorf("In.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestBuilder_String(t *testing.T) {
	tests := []struct {
		name    string
		builder *Builder
		want    string
	}{
		{
			name: "balanceupdate.GetBalance",
			builder: NewBuilder().SelectAll(models.DocBalanceUpdates).And(
				NewEq("network", "mainnet"),
				NewEq("contract", "address"),
			).End(),
			want: `SELECT * FROM balance_update WHERE (network = mainnet AND contract = address);`,
		}, {
			name: "bigmapaction.Get",
			builder: NewBuilder().SelectAll(models.DocBigMapActions).And(
				NewEq("network", "mainnet"),
				NewOr(
					NewEq("source_ptr", 1),
					NewEq("destination_ptr", 1),
				),
			).SortDesc("indexed_time").End(),
			want: `SELECT * FROM bigmapaction WHERE (network = mainnet AND (source_ptr = 1 OR destination_ptr = 1)) ORDER BY indexed_time DESC;`,
		}, {
			name: "blocks.GetBlock",
			builder: NewBuilder().SelectAll(models.DocBlocks).And(
				NewEq("network", "mainnet"),
				NewEq("level", 100),
			).One().End(),
			want: `SELECT * FROM block WHERE (network = mainnet AND level = 100) TOP 1;`,
		}, {
			name: "blocks.GetLastBlock",
			builder: NewBuilder().SelectAll(models.DocBlocks).And(
				NewEq("network", "mainnet"),
			).SortDesc("level").One().End(),
			want: `SELECT * FROM block WHERE (network = mainnet) ORDER BY level DESC TOP 1;`,
		}, {
			name: "blocks.GetNetworkAlias",
			builder: NewBuilder().SelectAll(models.DocBlocks).And(
				NewEq("chain_id", "chainID"),
			).One().End(),
			want: `SELECT * FROM block WHERE (chain_id = chainID) TOP 1;`,
		}, {
			name: "DeleteByLevelAndNetwork",
			builder: NewBuilder().Delete("transfer").And(
				NewGt("level", 1000),
				NewEq("network", "mainnet"),
			).End().Delete("operation").And(
				NewGt("level", 1000),
				NewEq("network", "mainnet"),
			).End(),
			want: `DELETE FROM transfer WHERE (level > 1000 AND network = mainnet);DELETE FROM operation WHERE (level > 1000 AND network = mainnet);`,
		}, {
			name:    "DeleteByLevelAndNetwork",
			builder: NewBuilder().Drop("transfer").End().Drop("operation").End(),
			want:    `DROP TABLE transfer;DROP TABLE operation;`,
		}, {
			name:    "Count",
			builder: NewBuilder().Count(models.DocTezosDomains).And(NewEq("network", "mainnet")).End(),
			want:    `SELECT COUNT(id) FROM tezos_domain WHERE (network = mainnet);`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.builder.String(); got != tt.want {
				t.Errorf("Builder.String() = %v, want %v", got, tt.want)
			}
		})
	}
}
