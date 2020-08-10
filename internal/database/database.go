package database

import (
	"github.com/jinzhu/gorm"
	// postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// DB -
type DB interface {
	// User
	GetOrCreateUser(*User) error
	GetUser(uint) (*User, error)
	UpdateUserMarkReadAt(uint, int64) error

	// Subscription
	GetSubscription(userID uint, address, network string) (Subscription, error)
	ListSubscriptions(userID uint) ([]Subscription, error)
	UpsertSubscription(*Subscription) error
	DeleteSubscription(*Subscription) error
	GetSubscriptionsCount(address, network string) (int, error)

	// Alias
	GetAliases(network string) ([]Alias, error)
	GetAlias(address, network string) (Alias, error)
	GetOperationAliases(src, dst, network string) (OperationAlises, error)
	GetAliasesMap(network string) (map[string]string, error)
	CreateAlias(string, string, string) error
	CreateOrUpdateAlias(a *Alias) error
	GetBySlug(string) (Alias, error)

	// Assessment
	CreateAssessment(a *Assessments) error
	CreateOrUpdateAssessment(a *Assessments) error
	GetAssessmentsWithValue(uint, uint, uint) ([]Assessments, error)
	GetUserCompletedAssesments(userID uint) (count int, err error)

	// Account
	GetOrCreateAccount(*Account) error

	// DApp
	GetDApps() ([]DApp, error)
	GetDApp(id uint) (DApp, error)

	Close()
}

type db struct {
	ORM *gorm.DB
}

// New - creates db connection
func New(connectionString string) (DB, error) {
	gormDB, err := gorm.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	gormDB.LogMode(false)

	gormDB.AutoMigrate(&User{}, &Subscription{}, &Alias{}, &Assessments{}, &Account{}, &Picture{}, &DApp{}, &Token{})

	gormDB = gormDB.Set("gorm:auto_preload", false)

	return &db{
		ORM: gormDB,
	}, nil
}

func (d *db) Close() {
	d.ORM.Close()
}
