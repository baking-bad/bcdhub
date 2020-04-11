package migrations

// Migration - intreface need to realize for migrate
type Migration interface {
	Do(ctx *Context) error
	Description() string
}
