package dto

type ArticleRequest struct {
	Title   string `json:"title" validate:"required,min=5,max=200"`
	Slug    string `json:"slug" validate:"required,min=3,max=100,alphanum_hyphen"`
	Content string `json:"content" validate:"required,min=50,max=80000"`
}

type ArticlePreview struct {
	ID           int32   `json:"id"`
	Title        string  `json:"title"`
	Slug         string  `json:"slug"`
	Created_at   *string `json:"created_at"`
	Published_at *string `json:"published_at"`
	Deleted_at   *string `json:"deleted_at"`
	IsPublished  bool    `json:"is_published"`
}

type ArticleResponse struct {
	ID           int32   `json:"id"`
	Title        string  `json:"title"`
	Slug         string  `json:"slug"`
	Content      string  `json:"content"`
	Created_at   *string `json:"created_at"`
	Published_at *string `json:"published_at"`
	Deleted_at   *string `json:"deleted_at"`
	IsPublished  bool    `json:"is_published"`
}
