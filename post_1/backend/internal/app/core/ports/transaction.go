package ports

type Transaction interface {
	UserRepo() UserRepository
	Commit() error
	Rollback() error
}
