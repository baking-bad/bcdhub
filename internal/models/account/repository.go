package account

//go:generate mockgen -source=$GOFILE -destination=../mock/account/mock.go -package=account -typed
type Repository interface {
	Get(address string) (Account, error)
}
