package types

// MigrationKind -
type MigrationKind int

// NewMigrationKind -
func NewMigrationKind(value string) MigrationKind {
	switch value {
	case "bootstrap":
		return MigrationKindBootstrap
	case "lambda":
		return MigrationKindLambda
	case "update":
		return MigrationKindUpdate
	default:
		return 0
	}
}

// String -
func (kind MigrationKind) String() string {
	switch kind {
	case MigrationKindBootstrap:
		return "bootstrap"
	case MigrationKindLambda:
		return "lambda"
	case MigrationKindUpdate:
		return "update"
	default:
		return ""
	}
}

const (
	MigrationKindBootstrap MigrationKind = iota + 1
	MigrationKindLambda
	MigrationKindUpdate
)
