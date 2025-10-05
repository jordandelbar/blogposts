package handlers

import (
	"net/http"
	"personal_website/internal/infrastructure/http/dto"
	"personal_website/internal/infrastructure/http/mappers"
	"personal_website/pkg/utils"
)

// ContactHandler godoc
// @Summary Submit contact form
// @Description Submit a contact form message that will be sent via email
// @Tags contact
// @Accept json
// @Produce json
// @Param contact body dto.ContactForm true "Contact form data"
// @Success 200 {object} utils.Envelope{message=string} "Message sent successfully"
// @Failure 400 {object} string "Invalid request data"
// @Failure 500 {object} string "Email sending failed"
// @Router /v1/contact [post]
func (h *Handler) ContactHandler(w http.ResponseWriter, r *http.Request) {
	var contactForm dto.ContactForm

	h.logger.Info("Contact form submission received", "method", r.Method, "user_agent", r.UserAgent())

	err := utils.ReadJSON(w, r, &contactForm)
	if err != nil {
		h.errorResponder.BadRequestResponse(w, r, err)
		return
	}

	h.logger.Info("Contact form data received", "name", contactForm.Name, "email", contactForm.Email, "message_length", len(contactForm.Message))

	if !h.validateDTO(w, r, contactForm, "contact form") {
		return
	}

	h.logger.Info("Attempting to send email", "recipient_email", contactForm.Email)
	contactMessage := mappers.ContactFormToDomain(contactForm)

	ctx := r.Context()
	err = h.emailService.SendContactEmail(ctx, contactMessage)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	h.logger.Info("Email sent successfully", "from", contactForm.Email, "name", contactForm.Name)
	response := utils.Envelope{
		"message": "Message sent successfully!",
	}

	err = utils.WriteJSON(w, http.StatusOK, response)
	if err != nil {
		h.errorResponder.ServerErrorResponse(w, r, err)
	}
}
