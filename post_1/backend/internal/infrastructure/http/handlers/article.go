package handlers

import (
	"fmt"
	"net/http"
	_ "personal_website/internal/infrastructure/http/docs"
	"personal_website/internal/infrastructure/http/dto"
	"personal_website/internal/infrastructure/http/mappers"
	"personal_website/pkg/utils"
)

// CreateArticle godoc
// @Summary Create a new article
// @Description Create a new article with title, slug, and content
// @Tags articles
// @Accept json
// @Produce json
// @Security Bearer
// @Param article body dto.ArticleRequest true "Article data"
// @Success 201 "Article created successfully"
// @Failure 400 {object} string "Invalid JSON body or validation error"
// @Failure 409 {object} string "article already exists"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/articles [post]
func (h *Handler) CreateArticle(w http.ResponseWriter, r *http.Request) {
	var dtoArticle dto.ArticleRequest

	err := utils.ReadJSON(w, r, &dtoArticle)
	if err != nil {
		h.errorResponder.BadRequestResponse(w, r, err)
		return
	}

	if !h.validateDTO(w, r, dtoArticle, "create article") {
		return
	}

	ctx := r.Context()
	domainArticle := mappers.ArticleRequestToDomain(dtoArticle)

	err = h.datastore.ArticleRepo().CreateArticle(ctx, domainArticle)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

// UpdateArticle godoc
// @Summary Update an existing article
// @Description Update an existing article by ID
// @Tags articles
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Param article body dto.ArticleRequest true "Updated article data"
// @Success 200 "Article updated successfully"
// @Failure 400 {object} string "Invalid JSON body or validation error"
// @Failure 404 {object} string "Article not found"
// @Failure 409 {object} string "Article with slug already exists"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/articles/id/{id} [put]
func (h *Handler) UpdateArticle(w http.ResponseWriter, r *http.Request) {
	id, ok := h.extractIDParam(w, r)
	if !ok {
		return
	}

	var dtoArticle dto.ArticleRequest

	err := utils.ReadJSON(w, r, &dtoArticle)
	if err != nil {
		h.errorResponder.BadRequestResponse(w, r, err)
		return
	}

	if !h.validateDTO(w, r, dtoArticle, "update article") {
		return
	}

	ctx := r.Context()
	domainArticle := mappers.ArticleRequestToDomainWithID(dtoArticle, id)

	err = h.datastore.ArticleRepo().UpdateArticle(ctx, domainArticle)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetArticleBySlug godoc
// @Summary Get article by slug
// @Description Retrieve a published article by its URL slug
// @Tags articles
// @Accept json
// @Produce json
// @Param slug path string true "Article slug"
// @Success 200 {object} utils.Envelope{data=dto.ArticleResponse} "Article details"
// @Failure 404 {object} string "Article not found"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/articles/slug/{slug} [get]
func (h *Handler) GetArticleBySlug(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	if slug == "" {
		h.errorResponder.BadRequestResponse(w, r, fmt.Errorf("slug parameter is required"))
		return
	}

	ctx := r.Context()
	article, err := h.datastore.ArticleRepo().GetArticleBySlug(ctx, slug)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	data := mappers.ArticleToResponse(article)

	err = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": data})
	if err != nil {
		h.errorResponder.ServerErrorResponse(w, r, err)
	}
}

// GetArticleById godoc
// @Summary Get article by ID
// @Description Retrieve an article by its ID
// @Tags articles
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Success 200 {object} utils.Envelope{data=dto.ArticlePreview} "Article preview"
// @Failure 400 {object} string "Invalid ID parameter"
// @Failure 404 {object} string "Article not found"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/articles/id/{id} [get]
func (h *Handler) GetArticleById(w http.ResponseWriter, r *http.Request) {
	id, ok := h.extractIDParam(w, r)
	if !ok {
		return
	}

	ctx := r.Context()
	article, err := h.datastore.ArticleRepo().GetArticleByID(ctx, id)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	data := mappers.ArticleToPreview(article)

	err = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": data})
	if err != nil {
		h.errorResponder.ServerErrorResponse(w, r, err)
	}
}

// GetArticleForEdit godoc
// @Summary Get article for editing
// @Description Retrieve full article content by ID for editing purposes
// @Tags articles
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Success 200 {object} utils.Envelope{data=dto.ArticleResponse} "Full article content"
// @Failure 400 {object} string "Invalid ID parameter"
// @Failure 404 {object} string "Article not found"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/articles/id/edit/{id} [get]
func (h *Handler) GetArticleForEdit(w http.ResponseWriter, r *http.Request) {
	id, ok := h.extractIDParam(w, r)
	if !ok {
		return
	}

	ctx := r.Context()
	article, err := h.datastore.ArticleRepo().GetArticleByID(ctx, id)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	data := mappers.ArticleToResponse(article)

	err = utils.WriteJSON(w, http.StatusOK, utils.Envelope{"data": data})
	if err != nil {
		h.errorResponder.ServerErrorResponse(w, r, err)
	}
}

// PublishArticle godoc
// @Summary Publish an article
// @Description Mark an article as published
// @Tags articles
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Success 200 "Article published successfully"
// @Failure 400 {object} string "Invalid ID parameter"
// @Failure 404 {object} string "Article not found"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/articles/id/{id}/publish [patch]
func (h *Handler) PublishArticle(w http.ResponseWriter, r *http.Request) {
	id, ok := h.extractIDParam(w, r)
	if !ok {
		return
	}

	ctx := r.Context()
	err := h.datastore.ArticleRepo().PublishArticle(ctx, id)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// UnpublishArticle godoc
// @Summary Unpublish an article
// @Description Mark an article as unpublished
// @Tags articles
// @Accept json
// @Produce json
// @Security Bearer
// @Param id path int true "Article ID"
// @Success 200 "Article unpublished successfully"
// @Failure 400 {object} string "Invalid ID parameter"
// @Failure 404 {object} string "Article not found"
// @Failure 500 {object} string "Internal server error"
// @Router /v1/articles/id/{id}/unpublish [patch]
func (h *Handler) UnpublishArticle(w http.ResponseWriter, r *http.Request) {
	id, ok := h.extractIDParam(w, r)
	if !ok {
		return
	}

	ctx := r.Context()
	err := h.datastore.ArticleRepo().UnpublishArticle(ctx, id)
	if err != nil {
		h.HandleDomainError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusOK)
}
