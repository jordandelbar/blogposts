package handlers

import (
	"net/http"
)

// SoftDeleteArticle godoc
// @Summary Soft delete an article
// @Description Move an article to trash (soft delete)
// @Tags articles
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Success 200 "Article moved to trash successfully"
// @Failure 400 {object} string "Invalid ID parameter"
// @Failure 404 {object} string "Article not found"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/articles/id/{id} [delete]
func (h *Handler) SoftDeleteArticle(w http.ResponseWriter, r *http.Request) {
	id, ok := h.extractIDParam(w, r)
	if !ok {
		return
	}

	ctx := r.Context()
	err := h.datastore.ArticleRepo().SoftDeleteArticle(ctx, id)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteArticle godoc
// @Summary Permanently delete an article
// @Description Permanently delete an article from the database. The article must be soft-deleted first.
// @Tags articles
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Success 200 "Article permanently deleted"
// @Failure 400 {object} string "Invalid ID parameter"
// @Failure 404 {object} string "Article not found"
// @Failure 409 {object} string "Article must be soft deleted before permanent deletion"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/articles/id/{id}/permanent [delete]
func (h *Handler) DeleteArticle(w http.ResponseWriter, r *http.Request) {
	id, ok := h.extractIDParam(w, r)
	if !ok {
		return
	}

	ctx := r.Context()
	err := h.datastore.ArticleRepo().DeleteArticle(ctx, id)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
