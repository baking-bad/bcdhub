package database

import (
	"github.com/jinzhu/gorm"
	// postgres driver
	_ "github.com/jinzhu/gorm/dialects/postgres"
)

// DB -
type DB interface {
	IAccount
	IAlias
	IAssessment
	ICompilationTask
	IDApp
	IDeployment
	ISubscription
	IToken
	IUser
	IVerification

	Close()
}

// IAccount -
type IAccount interface {
	GetOrCreateAccount(*Account) error
}

// IAlias -
type IAlias interface {
	GetAliases(network string) ([]Alias, error)
	GetAlias(address, network string) (Alias, error)
	GetOperationAliases(src, dst, network string) (OperationAliases, error)
	GetAliasesMap(network string) (map[string]string, error)
	CreateAlias(string, string, string) error
	CreateOrUpdateAlias(a *Alias) error
	GetBySlug(string) (Alias, error)
}

// IAssessment -
type IAssessment interface {
	CreateAssessment(a *Assessments) error
	CreateOrUpdateAssessment(a *Assessments) error
	GetAssessmentsWithValue(uint, uint, uint) ([]Assessments, error)
	GetUserCompletedAssesments(userID uint) (count int, err error)
}

// ICompilationTask -
type ICompilationTask interface {
	ListCompilationTasks(userID, limit, offset uint, kind string) ([]CompilationTask, error)
	GetCompilationTask(taskID uint) (*CompilationTask, error)
	GetCompilationTaskBy(address, network, status string) (*CompilationTask, error)
	CreateCompilationTask(ct *CompilationTask) error
	UpdateTaskStatus(taskID uint, status string) error
	UpdateTaskResults(task *CompilationTask, status string, results []CompilationTaskResult) error
	CountCompilationTasks(userID uint) (int64, error)
}

// IDApp -
type IDApp interface {
	GetDApps() ([]DApp, error)
	GetDApp(id uint) (DApp, error)
	GetDAppBySlug(slug string) (dapp DApp, err error)
}

// IDeployment -
type IDeployment interface {
	ListDeployments(userID, limit, offset uint) ([]Deployment, error)
	CreateDeployment(dt *Deployment) error
	GetDeploymentBy(opHash string) (*Deployment, error)
	UpdateDeployment(dt *Deployment) error
	CountDeployments(userID uint) (int64, error)
}

// ISubscription -
type ISubscription interface {
	GetSubscription(userID uint, address, network string) (Subscription, error)
	GetSubscriptions(address, network string) ([]Subscription, error)
	ListSubscriptions(userID uint) ([]Subscription, error)
	UpsertSubscription(*Subscription) error
	DeleteSubscription(*Subscription) error
	GetSubscriptionsCount(address, network string) (int, error)
}

// IToken -
type IToken interface {
	GetTokens() ([]Token, error)
}

// IUser -
type IUser interface {
	GetOrCreateUser(*User, string) error
	GetUser(uint) (*User, error)
	UpdateUserMarkReadAt(uint, int64) error
}

// IVerification -
type IVerification interface {
	ListVerifications(userID, limit, offset uint) ([]Verification, error)
	CreateVerification(v *Verification) error
	GetVerificationBy(address, network string) (*Verification, error)
	CountVerifications(userID uint) (int64, error)
}

type db struct {
	*gorm.DB
}

// New - creates db connection
func New(connectionString string) (DB, error) {
	gormDB, err := gorm.Open("postgres", connectionString)
	if err != nil {
		return nil, err
	}

	gormDB.LogMode(false)

	gormDB.AutoMigrate(
		&User{},
		&Subscription{},
		&Alias{},
		&Assessments{},
		&Account{},
		&Picture{},
		&DApp{},
		&Token{},
		&CompilationTask{},
		&CompilationTaskResult{},
		&Verification{},
		&Deployment{},
	)

	gormDB = gormDB.Set("gorm:auto_preload", false)

	return &db{gormDB}, nil
}

func (d *db) Close() {
	d.DB.Close()
}
