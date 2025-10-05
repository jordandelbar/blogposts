package domain

type Session struct {
	UserID      int
	Email       string
	Permissions Permissions
	Activated   bool
}

var AnonymousSession = &Session{
	UserID:      0,
	Email:       "",
	Permissions: Permissions{},
	Activated:   false,
}

func (u *Session) GetUserID() int {
	return u.UserID
}

func (u *Session) IsAnonymous() bool {
	return u == AnonymousSession
}
