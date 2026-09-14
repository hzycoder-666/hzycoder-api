package handler

import (
	"github.com/gin-gonic/gin"
	"hzycoder.com/lion/internal/middleware"
	reqDto "hzycoder.com/lion/internal/model/dto/request"
	"hzycoder.com/lion/internal/service"
	"hzycoder.com/lion/pkg/response"
)

type PaperHandler struct {
	papers *service.PaperService
}

func NewPaperHandler(papers *service.PaperService) *PaperHandler {
	return &PaperHandler{papers: papers}
}

type createPaperRequest struct {
	Title       string                      `json:"title" binding:"required,min=2,max=128"`
	Description *string                     `json:"description" binding:"omitempty,max=512"`
	Questions   []reqDto.PaperQuestionInput `json:"questions" binding:"required,min=1,max=200"`
}

func (r createPaperRequest) toDTO() reqDto.CreatePaper {
	return reqDto.CreatePaper{
		Title:       r.Title,
		Description: r.Description,
		Questions:   r.Questions,
	}
}

type updatePaperRequest struct {
	Title       *string                     `json:"title" binding:"omitempty,min=2,max=128"`
	Description *string                     `json:"description" binding:"omitempty,max=512"`
	Questions   []reqDto.PaperQuestionInput `json:"questions" binding:"omitempty,min=1,max=200"`
}

func (r updatePaperRequest) toDTO() reqDto.UpdatePaper {
	return reqDto.UpdatePaper{
		Title:       r.Title,
		Description: r.Description,
		Questions:   r.Questions,
	}
}

type listPaperRequest struct {
	Status   string `form:"status" binding:"omitempty,oneof=draft published offline"`
	Keyword  string `form:"keyword" binding:"omitempty,max=64"`
	Page     int    `form:"page" binding:"omitempty,min=1"`
	PageSize int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

func (r listPaperRequest) toDTO() reqDto.PaperFilter {
	filter := reqDto.PaperFilter{
		Keyword:  r.Keyword,
		Page:     r.Page,
		PageSize: r.PageSize,
	}
	if r.Status != "" {
		status := r.Status
		filter.Status = &status
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	return filter
}

// Create 管理员建卷
// @Summary 新建试卷（草稿）
// @Description 从题库选题组卷，初始状态为 draft；questions 传题目 ID、顺序与分值
// @Tags 试卷管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body createPaperRequest true "试卷信息"
// @Success 200 {object} response.Resp{data=hzycoder_com_lion_internal_model_dto_response.PaperDetail}
// @Router /v1/admin/papers [post]
func (h *PaperHandler) Create(c *gin.Context) {
	var req createPaperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	detail, err := h.papers.Create(c.Request.Context(), middleware.GetUserID(c), req.toDTO())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, detail)
}

// Update 管理员编辑草稿试卷
// @Summary 编辑试卷（仅草稿）
// @Description 仅 draft 状态可编辑；questions 整体替换
// @Tags 试卷管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "试卷 ID"
// @Param request body updatePaperRequest true "待更新字段"
// @Success 200 {object} response.Resp{data=hzycoder_com_lion_internal_model_dto_response.PaperDetail}
// @Router /v1/admin/papers/{id} [put]
func (h *PaperHandler) Update(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	var req updatePaperRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	detail, err := h.papers.Update(c.Request.Context(), id, req.toDTO())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, detail)
}

// Publish
// @Summary 发布试卷
// @Description 仅 draft 状态可发布；发布后用户端可见可作答
// @Tags 试卷管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "试卷 ID"
// @Success 200 {object} response.Resp
// @Router /v1/admin/papers/{id}/publish [post]
func (h *PaperHandler) Publish(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	if err := h.papers.Publish(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, gin.H{"id": id, "status": "published"})
}

// Offline
// @Summary 下架试卷
// @Description 仅 published 状态可下架；下架后用户端不可见
// @Tags 试卷管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "试卷 ID"
// @Success 200 {object} response.Resp
// @Router /v1/admin/papers/{id}/offline [post]
func (h *PaperHandler) Offline(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	if err := h.papers.Offline(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, gin.H{"id": id, "status": "offline"})
}

// GetByID 管理员查看试卷详情（含答案与解析）
// @Summary 试卷详情（管理端，含答案解析）
// @Tags 试卷管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "试卷 ID"
// @Success 200 {object} response.Resp{data=hzycoder_com_lion_internal_model_dto_response.PaperDetail}
// @Router /v1/admin/papers/{id} [get]
func (h *PaperHandler) GetByID(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	detail, err := h.papers.GetByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, detail)
}

// List
// @Summary 试卷分页列表（管理端）
// @Tags 试卷管理
// @Produce json
// @Security BearerAuth
// @Param status query string false "状态：draft published offline" Enums(draft,published,offline)
// @Param keyword query string false "标题关键词"
// @Param page query int false "页码，默认 1"
// @Param page_size query int false "每页条数，默认 20"
// @Success 200 {object} response.Resp{data=hzycoder_com_lion_internal_model_dto_response.PaperPage}
// @Router /v1/admin/papers [get]
func (h *PaperHandler) List(c *gin.Context) {
	var req listPaperRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}

	page, err := h.papers.List(c.Request.Context(), req.toDTO())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, page)
}

// ListPublished 用户端已发布试卷列表
// @Summary 已发布试卷列表（用户端）
// @Description 仅返回 published 状态的试卷摘要
// @Tags 答题
// @Produce json
// @Security BearerAuth
// @Param keyword query string false "标题关键词"
// @Param page query int false "页码，默认 1"
// @Param page_size query int false "每页条数，默认 20"
// @Success 200 {object} response.Resp{data=hzycoder_com_lion_internal_model_dto_response.PaperPage}
// @Router /v1/papers [get]
func (h *PaperHandler) ListPublished(c *gin.Context) {
	var req listPaperRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}

	page, err := h.papers.ListPublished(c.Request.Context(), req.toDTO())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, page)
}

// GetPublished 用户端答题视图：不含答案与解析
// @Summary 试卷答题视图（用户端）
// @Description 仅包含题干与选项，不含答案与解析
// @Tags 答题
// @Produce json
// @Security BearerAuth
// @Param id path int true "试卷 ID"
// @Success 200 {object} response.Resp{data=hzycoder_com_lion_internal_model_dto_response.PaperDetail}
// @Router /v1/papers/{id} [get]
func (h *PaperHandler) GetPublished(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	detail, err := h.papers.GetPublished(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, detail)
}
