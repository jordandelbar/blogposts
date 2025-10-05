package dto

type ContactForm struct {
	Name    string `json:"name" validate:"required,min=2,max=100,no_html"`
	Email   string `json:"email" validate:"required,email,max=254"`
	Message string `json:"message" validate:"required,min=10,max=5000,no_script_tags"`
}
