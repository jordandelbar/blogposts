package handlers

import (
	"context"
	"net/http"
	"strconv"
	"time"
)

// ResumeHandler godoc
// @Summary Download resume
// @Description Download the latest resume as a PDF file
// @Tags resume
// @Accept json
// @Produce application/pdf
// @Success 200 {file} binary "PDF resume file"
// @Failure 404 {object} string "Resume file not found"
// @Failure 500 {object} string "Resume storage service unavailable or internal server error"
// @Router /v1/resume [get]
func (h *Handler) ResumeHandler(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Resume download request received", "method", r.Method, "user_agent", r.UserAgent())

	ctx, cancel := context.WithTimeout(r.Context(), 30*time.Second)
	defer cancel()

	// Get the resume PDF from Minio
	resumeData, err := h.resumeService.GetResume(ctx, "resume_without_personal_data.pdf")
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	h.logger.Info("Resume fetched successfully", "size_bytes", len(resumeData))

	// Set appropriate headers for PDF download
	w.Header().Set("Content-Type", "application/pdf")
	w.Header().Set("Content-Disposition", "attachment; filename=\"jordan_delbar_resume.pdf\"")
	w.Header().Set("Content-Length", strconv.Itoa(len(resumeData)))

	// Write the PDF data
	_, err = w.Write(resumeData)
	if err != nil {
		h.logger.Error("Failed to write PDF response", "error", err)
		return
	}

	h.logger.Info("Resume download completed successfully")
}
