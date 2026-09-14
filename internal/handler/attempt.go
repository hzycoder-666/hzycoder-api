package handler

import (
	"github.com/gin-gonic/gin"
	"hzycoder.com/lion/internal/middleware"
	reqDto "hzycoder.com/lion/internal/model/dto/request"
	"hzycoder.com/lion/internal/service"
	"hzycoder.com/lion/pkg/response"
)

type AttemptHandler struct {
	attempts *service.AttemptService
}

func NewAttemptHandler(attempts *service.AttemptService) *AttemptHandler {
	return &AttemptHandler{attempts: attempts}
}

// Submit 提交答卷（练习模式，一次提交即时评分并保存）
// @Summary 提交答卷
// @Description 逐题作答并即时评分保存；answer 传 null 表示未作答（计 0 分）
// @Tags 答题
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "试卷 ID"
// @Param request body submitRequest true "逐题作答"
// @Success 200 {object} response.Resp{data=hzycoder_com_lion_internal_model_dto_response.SubmitResult}
// @Router /v1/papers/{id}/submit [post]
func (h *AttemptHandler) Submit(c *gin.Context) {
	paperID, ok := parseIDParam(c)
	if !ok {
		return
	}

	var req submitRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	result, err := h.attempts.Submit(c.Request.Context(), middleware.GetUserID(c), paperID, req.toDTO())
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, result)
}

// List 当前用户的答题历史
// @Summary 我的答题历史
// @Tags 答题
// @Produce json
// @Security BearerAuth
// @Param page query int false "页码，默认 1"
// @Param page_size query int false "每页条数，默认 10"
// @Success 200 {object} response.Resp{data=hzycoder_com_lion_internal_model_dto_response.AttemptPage}
// @Router /v1/attempts [get]
func (h *AttemptHandler) List(c *gin.Context) {
	var req attemptListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		_ = c.Error(err)
		return
	}

	f := req.toDTO()
	page, err := h.attempts.List(c.Request.Context(), middleware.GetUserID(c), f.Page, f.PageSize)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, page)
}

// GetByID 当前用户某次作答的详情（答案+解析）
// @Summary 答卷详情
// @Description 含逐题正误、我的答案、标准答案与解析
// @Tags 答题
// @Produce json
// @Security BearerAuth
// @Param id path int true "答卷 ID"
// @Success 200 {object} response.Resp{data=hzycoder_com_lion_internal_model_dto_response.AttemptDetail}
// @Router /v1/attempts/{id} [get]
func (h *AttemptHandler) GetByID(c *gin.Context) {
	id, ok := parseIDParam(c)
	if !ok {
		return
	}

	detail, err := h.attempts.GetByID(c.Request.Context(), middleware.GetUserID(c), id)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, detail)
}

// AdminListByPaper 管理端按试卷查看成绩列表
// @Summary 按试卷查看成绩列表
// @Tags 成绩统计
// @Produce json
// @Security BearerAuth
// @Param id path int true "试卷 ID"
// @Param page query int false "页码，默认 1"
// @Param page_size query int false "每页条数，默认 10"
// @Success 200 {object} response.Resp{data=hzycoder_com_lion_internal_model_dto_response.AdminAttemptPage}
// @Router /v1/admin/papers/{id}/attempts [get]
func (h *AttemptHandler) AdminListByPaper(c *gin.Context) {
	paperID, ok := parseIDParam(c)
	if !ok {
		return
	}

	var req attemptListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		_ = c.Error(err)
		return
	}

	f := req.toDTO()
	page, err := h.attempts.ListByPaper(c.Request.Context(), paperID, f.Page, f.PageSize)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, page)
}

// AdminStats 管理端单卷最简统计
// @Summary 单卷最简统计
// @Description 返回作答人数、平均分、最高分
// @Tags 成绩统计
// @Produce json
// @Security BearerAuth
// @Param id path int true "试卷 ID"
// @Success 200 {object} response.Resp{data=hzycoder_com_lion_internal_model_dto_response.PaperStats}
// @Router /v1/admin/papers/{id}/stats [get]
func (h *AttemptHandler) AdminStats(c *gin.Context) {
	paperID, ok := parseIDParam(c)
	if !ok {
		return
	}

	stats, err := h.attempts.StatsByPaper(c.Request.Context(), paperID)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response.Success(c, stats)
}

type submitRequest struct {
	Answers []reqDto.SubmitAnswer `json:"answers" binding:"required,dive"`
}

func (r submitRequest) toDTO() reqDto.SubmitPaper {
	return reqDto.SubmitPaper{
		Answers: r.Answers,
	}
}

type attemptListRequest struct {
	Page     int `form:"page" binding:"omitempty,min=1"`
	PageSize int `form:"page_size" binding:"omitempty,min=1,max=100"`
}

func (r attemptListRequest) toDTO() reqDto.AttemptFilter {
	f := reqDto.AttemptFilter{
		Page:     r.Page,
		PageSize: r.PageSize,
	}
	if f.Page <= 0 {
		f.Page = 1
	}
	if f.PageSize <= 0 {
		f.PageSize = 10
	}
	return f
}
