package ast

// Miguel kinds
const (
	MiguelKindCreate = "create"
	MiguelKindUpdate = "update"
	MiguelKindDelete = "delete"
)

const (
	valueKindString = iota + 1
	valueKindBytes
	valueKindInt
)
