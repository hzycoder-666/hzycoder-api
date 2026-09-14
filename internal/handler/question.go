package handler

import (
	"encoding/json"
	"strconv"

	"github.com/gin-gonic/gin"
	"hzycoder.com/lion/internal/middleware"
	"hzycoder.com/lion/internal/model"
	reqDto "hzycoder.com/lion/internal/model/dto/request"
	"hzycoder.com/lion/internal/service"
	"hzycoder.com/lion/pkg/response"
)

type QuestionHandler struct {
	questions *service.QuestionService
}

func NewQuestionHandler(questions *service.QuestionService) *QuestionHandler {
	return &QuestionHandler{questions: questions}
}

type createQuestionRequest struct {
	Type         model.QuestionType     `json:"type" binding:"required,oneof=single multiple judge"`
	Content      string                 `json:"content" binding:"required,min=2,max=2000"`
	Options      []model.QuestionOption `json:"options" binding:"max=10"`
	Answer       json.RawMessage        `json:"answer" binding:"required"`
	Analysis     *string                `json:"analysis" binding:"omitempty,max=2000"`
	Tags         *string                `json:"tags" binding:"omitempty,max=255"`
	Difficulty   *int                   `json:"difficulty" binding:"omitempty,min=1,max=5"`
	DefaultScore *int                   `json:"default_score" binding:"omitempty,min=1,max=100"`
}

func (r createQuestionRequest) toDTO() reqDto.CreateQuestion {
	return reqDto.CreateQuestion{
		Type:         r.Type,
		Content:      r.Content,
		Options:      r.Options,
		Answer:       r.Answer,
		Analysis:     r.Analysis,
		Tags:         r.Tags,
		Difficulty:   r.Difficulty,
		DefaultScore: r.DefaultScore,
	}
}

type updateQuestionRequest struct {
	Type         *model.QuestionType     `json:"type" binding:"omitempty,oneof=single multiple judge"`
	Content      *string                 `json:"content" binding:"omitempty,min=2,max=2000"`
	Options      *[]model.QuestionOption `json:"options" binding:"omitempty,max=10"`
	Answer       json.RawMessage         `json:"answer"`
	Analysis     *string                 `json:"analysis" binding:"omitempty,max=2000"`
	Tags         *string                 `json:"tags" binding:"omitempty,max=255"`
	Difficulty   *int                    `json:"difficulty" binding:"omitempty,min=1,max=5"`
	DefaultScore *int                    `json:"default_score" binding:"omitempty,min=1,max=100"`
	Status       *model.QuestionStatus   `json:"status" binding:"omitempty,oneof=enabled disabled"`
}

func (r updateQuestionRequest) toDTO() reqDto.UpdateQuestion {
	return reqDto.UpdateQuestion{
		Type:         r.Type,
		Content:      r.Content,
		Options:      r.Options,
		Answer:       r.Answer,
		Analysis:     r.Analysis,
		Tags:         r.Tags,
		Difficulty:   r.Difficulty,
		DefaultScore: r.DefaultScore,
		Status:       r.Status,
	}
}

type listQuestionRequest struct {
	Type       string `form:"type" binding:"omitempty,oneof=single multiple judge"`
	Status     string `form:"status" binding:"omitempty,oneof=enabled disabled"`
	Difficulty int    `form:"difficulty" binding:"omitempty,min=1,max=5"`
	Tag        string `form:"tag" binding:"omitempty,max=32"`
	Keyword    string `form:"keyword" binding:"omitempty,max=64"`
	Page       int    `form:"page" binding:"omitempty,min=1"`
	PageSize   int    `form:"page_size" binding:"omitempty,min=1,max=100"`
}

func (r listQuestionRequest) toDTO() reqDto.QuestionFilter {
	filter := reqDto.QuestionFilter{
		Type:     emptyNilQuestionType(r.Type),
		Status:   emptyNilQuestionStatus(r.Status),
		Tag:      r.Tag,
		Keyword:  r.Keyword,
		Page:     r.Page,
		PageSize: r.PageSize,
	}
	if r.Difficulty > 0 {
		filter.Difficulty = &r.Difficulty
	}
	if filter.Page <= 0 {
		filter.Page = 1
	}
	if filter.PageSize <= 0 {
		filter.PageSize = 20
	}
	return filter
}

// Create
// @Summary 新建题目
// @Description 创建单选/多选/判断题；tags 为逗号分隔字符串；answer 为 JSON（单选/判断传字符串，多选传字符串数组）
// @Tags 题库管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body createQuestionRequest true "题目内容"
// @Success 200 {object} response.Resp{data=hzycoder_com_lion_internal_model_dto_response.QuestionItem}
// @Router /v1/admin/questions [post]
func (h *QuestionHandler) Create(c *gin.Context) {
	var req createQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	userID, ok := middleware.GetUserID(c)
	if !ok {
		response.Abort(c, "missing user id")
		return
	}

	item, err := h.questions.Create(c.Request.Context(), userID, req.toDTO())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, item)
}

// Update
// @Summary 编辑题目
// @Description 全字段可选更新；status 可切换 enabled/disabled
// @Tags 题库管理
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "题目 ID"
// @Param request body updateQuestionRequest true "待更新字段"
// @Success 200 {object} response.Resp{data=hzycoder_com_lion_internal_model_dto_response.QuestionItem}
// @Router /v1/admin/questions/{id} [put]
func (h *QuestionHandler) Update(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	var req updateQuestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(err)
		return
	}

	item, err := h.questions.Update(c.Request.Context(), id, req.toDTO())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, item)
}

// Delete
// @Summary 删除题目
// @Description 已被试卷引用的题目不可删除
// @Tags 题库管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "题目 ID"
// @Success 200 {object} response.Resp
// @Router /v1/admin/questions/{id} [delete]
func (h *QuestionHandler) Delete(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	if err := h.questions.Delete(c.Request.Context(), id); err != nil {
		c.Error(err)
		return
	}

	response.Success(c, gin.H{"id": id})
}

// GetByID
// @Summary 题目详情
// @Tags 题库管理
// @Produce json
// @Security BearerAuth
// @Param id path int true "题目 ID"
// @Success 200 {object} response.Resp{data=hzycoder_com_lion_internal_model_dto_response.QuestionItem}
// @Router /v1/admin/questions/{id} [get]
func (h *QuestionHandler) GetByID(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	item, err := h.questions.GetByID(c.Request.Context(), id)
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, item)
}

// List
// @Summary 题目分页列表
// @Description 支持题型/状态/难度/标签筛选与关键词搜索（匹配题干）
// @Tags 题库管理
// @Produce json
// @Security BearerAuth
// @Param type query string false "题型：single multiple judge" Enums(single,multiple,judge)
// @Param status query string false "状态：enabled disabled" Enums(enabled,disabled)
// @Param difficulty query int false "难度 1-5"
// @Param tag query string false "标签（逗号分隔标签之一即命中）"
// @Param keyword query string false "题干关键词"
// @Param page query int false "页码，默认 1"
// @Param page_size query int false "每页条数，默认 20"
// @Success 200 {object} response.Resp{data=hzycoder_com_lion_internal_model_dto_response.QuestionPage}
// @Router /v1/admin/questions [get]
func (h *QuestionHandler) List(c *gin.Context) {
	var req listQuestionRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.Error(err)
		return
	}

	page, err := h.questions.List(c.Request.Context(), req.toDTO())
	if err != nil {
		c.Error(err)
		return
	}

	response.Success(c, page)
}

func parseIDParam(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.Error(response.NewBizError(response.CodeParamInvalid, "无效的ID"))
		return 0, false
	}
	return id, true
}

func emptyNilQuestionType(s string) *model.QuestionType {
	if s == "" {
		return nil
	}
	t := model.QuestionType(s)
	return &t
}

func emptyNilQuestionStatus(s string) *model.QuestionStatus {
	if s == "" {
		return nil
	}
	st := model.QuestionStatus(s)
	return &st
}
