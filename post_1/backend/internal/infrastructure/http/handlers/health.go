package handlers

import (
	"net/http"
	"personal_website/pkg/utils"
)

// HealthcheckHandler godoc
// @Summary Health check endpoint
// @Description Get the health status and system information of the API
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} utils.Envelope{status=string,system_info=object} "API health status"
// @Failure 500 {object} string "Internal server error"
// @Router /health [get]
func (h *Handler) HealthcheckHandler(w http.ResponseWriter, r *http.Request) {
	health_envelope := utils.Envelope{
		"status": "available",
		"system_info": map[string]string{
			"environment": h.config.Environment,
			"version":     h.config.Version,
		},
	}
	err := utils.WriteJSON(w, http.StatusOK, health_envelope)
	if err != nil {
		h.errorResponder.ServerErrorResponse(w, r, err)
	}
}
