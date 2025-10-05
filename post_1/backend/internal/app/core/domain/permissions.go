package domain

import "slices"

type Permissions []string

func (p Permissions) Include(code string) bool {
	return slices.Contains(p, code)
}
