package handlers

import (
	"net/http"
	"personal_website/internal/infrastructure/http/mappers"
	"personal_website/pkg/utils"
)

// RestoreArticle godoc
// @Summary Restore a deleted article
// @Description Restore an article from trash
// @Tags articles
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Success 200 "Article restored successfully"
// @Failure 400 {object} string "Invalid ID parameter"
// @Failure 404 {object} string "Article not found"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/articles/id/{id}/restore [post]
func (h *Handler) RestoreArticle(w http.ResponseWriter, r *http.Request) {
	id, ok := h.extractIDParam(w, r)
	if !ok {
		return
	}

	ctx := r.Context()
	err := h.datastore.ArticleRepo().RestoreArticle(ctx, id)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ListDeletedArticles godoc
// @Summary List deleted articles
// @Description Get a list of all soft-deleted articles in trash
// @Tags articles
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} utils.Envelope{data=[]dto.ArticlePreview} "List of deleted articles"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/articles/trash [get]
func (h *Handler) ListDeletedArticles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	articles, err := h.datastore.ArticleRepo().ListDeletedArticles(ctx)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	articleList := mappers.ArticlesToPreviews(articles)

	data := utils.Envelope{"data": articleList}
	err = utils.WriteJSON(w, http.StatusOK, data)
	if err != nil {
		h.errorResponder.ServerErrorResponse(w, r, err)
	}
}
