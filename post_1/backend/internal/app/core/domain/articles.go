package domain

import "time"

type Article struct {
	ID          int32
	Title       string
	Slug        string
	Content     string
	CreatedAt   time.Time
	PublishedAt time.Time
	DeletedAt   time.Time
	IsPublished bool
}
