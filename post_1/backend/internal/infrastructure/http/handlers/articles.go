package handlers

import (
	"net/http"
	"personal_website/internal/infrastructure/http/mappers"
	"personal_website/pkg/utils"
)

// ListArticles godoc
// @Summary List published articles
// @Description Get a list of all published articles with previews
// @Tags articles
// @Accept json
// @Produce json
// @Success 200 {object} utils.Envelope{data=[]dto.ArticlePreview} "List of published articles"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/articles [get]
func (h *Handler) ListArticles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	articles, err := h.datastore.ArticleRepo().ListArticles(ctx)
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

// ListAllArticles godoc
// @Summary List all articles (including unpublished)
// @Description Get a list of all articles including unpublished ones (admin endpoint)
// @Tags articles
// @Accept json
// @Produce json
// @Security Bearer
// @Success 200 {object} utils.Envelope{data=[]dto.ArticlePreview} "List of all articles"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/articles/all [get]
func (h *Handler) ListAllArticles(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	articles, err := h.datastore.ArticleRepo().ListAllArticles(ctx)
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
