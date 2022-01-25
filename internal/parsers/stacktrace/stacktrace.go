package stacktrace

import (
	"fmt"
	"io"
	"strings"

	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/operation"
	"github.com/baking-bad/bcdhub/internal/models/types"
)

// Item -
type Item struct {
	ParentID   int64
	Entrypoint types.NullString

	children     []int64
	source       string
	destination  string
	contentIndex int64
	nonce        *int64
}

// NewItem -
func NewItem(operation operation.Operation, parentID int64) *Item {
	item := &Item{
		ParentID:     parentID,
		children:     make([]int64, 0),
		source:       operation.Source.Address,
		destination:  operation.Destination.Address,
		contentIndex: operation.ContentIndex,
		nonce:        operation.Nonce,
		Entrypoint:   operation.Entrypoint,
	}
	return item
}

// GetID -
func (sti *Item) GetID() int64 {
	return computeID(sti.contentIndex, sti.nonce)
}

// String -
func (sti *Item) String() string {
	s := sti.Entrypoint.String()
	if len(s) < 20 {
		s += strings.Repeat(" ", 20-len(s))
	}
	return fmt.Sprintf("| %s [%s] => [%s]\n", s, sti.source, sti.destination)
}

// AddChild -
func (sti *Item) AddChild(child *Item) {
	sti.children = append(sti.children, child.GetID())
}

// IsNext -
func (sti *Item) IsNext(operation operation.Operation) bool {
	if !sti.gtNonce(operation.Nonce) {
		return false
	}
	return sti.destination == operation.Source.Address
}

func (sti *Item) gtNonce(nonce *int64) bool {
	if nonce == nil {
		return false
	}
	if sti.nonce == nil {
		return true
	}
	return *sti.nonce < *nonce
}

func computeID(contentIndex int64, nonce *int64) int64 {
	id := contentIndex * 1000
	if nonce != nil {
		id += (*nonce + 1)
	}
	return id
}

// StackTrace -
type StackTrace struct {
	tree  map[int64]*Item
	order []*Item
}

// New -
func New() *StackTrace {
	return &StackTrace{
		tree:  make(map[int64]*Item),
		order: make([]*Item, 0),
	}
}

// Get -
func (st *StackTrace) Get(operation operation.Operation) *Item {
	id := computeID(operation.ContentIndex, operation.Nonce)
	result, ok := st.tree[id]
	if !ok {
		return nil
	}
	return result
}

// GetByID -
func (st *StackTrace) GetByID(id int64) *Item {
	result, ok := st.tree[id]
	if !ok {
		return nil
	}
	return result
}

// Add -
func (st *StackTrace) Add(operation operation.Operation) {
	var parent *Item
	for i := len(st.order) - 1; i >= 0; i-- {
		if st.order[i].IsNext(operation) {
			parent = st.order[i]
			break
		}
	}
	parentID := int64(-1)
	if parent != nil {
		parentID = parent.GetID()
	}
	sti := NewItem(operation, parentID)
	st.tree[sti.GetID()] = sti
	st.order = append(st.order, sti)
	if parent != nil {
		parent.AddChild(sti)
	}
}

// Empty -
func (st *StackTrace) Empty() bool {
	return len(st.tree) == 0
}

// String -
func (st *StackTrace) String() string {
	builder := strings.Builder{}
	builder.WriteString("\nStackTrace:\n")

	topLevel := make([]int64, 0)
	for _, sti := range st.tree {
		if sti.ParentID == -1 {
			topLevel = append(topLevel, sti.GetID())
		}
	}

	if err := st.print(topLevel, 1, &builder); err != nil {
		logger.Err(err)
	}
	return builder.String()
}

func (st *StackTrace) print(arr []int64, depth int, builder io.StringWriter) error {
	for i := range arr {
		if item, ok := st.tree[arr[i]]; ok {
			if _, err := builder.WriteString(strings.Repeat("  ", depth)); err != nil {
				return err
			}
			if _, err := builder.WriteString(item.String()); err != nil {
				return err
			}
			st.print(item.children, depth+1, builder)
		}
	}
	return nil
}

// Fill -
func (st *StackTrace) Fill(repo operation.Repository, op operation.Operation) error {
	ops, err := repo.Get(map[string]interface{}{
		"operation.network": op.Network,
		"hash":              op.Hash,
	}, 0, true)
	if err != nil {
		return err
	}
	for i := range ops {
		st.Add(ops[i])
	}
	return nil
}
