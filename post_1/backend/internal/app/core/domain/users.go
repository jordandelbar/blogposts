package domain

import "time"

type User struct {
	ID        int
	CreatedAt time.Time
	Name      string
	Email     string
	Password  password
	Activated bool
}

func (u *User) IsAnonymous() bool {
	return u.ID == 0
}
