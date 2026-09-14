package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"hzycoder.com/lion/internal/model"
	reqDto "hzycoder.com/lion/internal/model/dto/request"
	resDto "hzycoder.com/lion/internal/model/dto/response"
	"hzycoder.com/lion/internal/repository"
	"hzycoder.com/lion/pkg/response"
)

type PaperService struct {
	papers    repository.PaperRepository
	questions repository.QuestionRepository
}

func NewPaperService(papers repository.PaperRepository, questions repository.QuestionRepository) *PaperService {
	return &PaperService{papers: papers, questions: questions}
}

func (s *PaperService) Create(ctx context.Context, userID int64, req reqDto.CreatePaper) (*resDto.PaperDetail, error) {
	items, totalScore, err := s.buildPaperItems(ctx, req.Questions)
	if err != nil {
		return nil, err
	}

	paper := model.Paper{
		Title:         req.Title,
		Description:   req.Description,
		Status:        model.PaperStatusDraft,
		TotalScore:    totalScore,
		QuestionCount: len(items),
		CreatedBy:     userID,
	}

	created, err := s.papers.Create(ctx, paper, items)
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	return s.GetByID(ctx, created.ID)
}

func (s *PaperService) Update(ctx context.Context, id int64, req reqDto.UpdatePaper) (*resDto.PaperDetail, error) {
	paper, err := s.papers.FindByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, response.NewBizError(response.CodePaperNotFound)
	}
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}
	if paper.Status != model.PaperStatusDraft {
		return nil, response.NewBizError(response.CodePaperStatusInvalid, "仅草稿状态的试卷可编辑，请先下线")
	}

	if req.Title != nil {
		paper.Title = *req.Title
	}
	if req.Description != nil {
		paper.Description = req.Description
	}

	var items []model.PaperQuestion
	if req.Questions != nil {
		items, paper.TotalScore, err = s.buildPaperItems(ctx, req.Questions)
		if err != nil {
			return nil, err
		}
		paper.QuestionCount = len(items)
	} else {
		rows, err := s.papers.FindQuestions(ctx, id)
		if err != nil {
			return nil, response.NewBizError(response.CodeDBError)
		}
		for _, row := range rows {
			items = append(items, row.PaperQuestion)
		}
	}

	if err := s.papers.Update(ctx, *paper, items); err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	return s.GetByID(ctx, id)
}

// buildPaperItems 校验选题合法（存在、启用、不重复），按数组顺序生成题号，空分值用题目默认分值
func (s *PaperService) buildPaperItems(ctx context.Context, inputs []reqDto.PaperQuestionInput) ([]model.PaperQuestion, int, error) {
	ids := make([]int64, len(inputs))
	seen := make(map[int64]struct{}, len(inputs))
	for i, in := range inputs {
		if _, dup := seen[in.QuestionID]; dup {
			return nil, 0, response.NewBizErrorWithDetails(response.CodePaperInvalid, fmt.Sprintf("题目 %d 重复选择", in.QuestionID))
		}
		seen[in.QuestionID] = struct{}{}
		ids[i] = in.QuestionID
	}

	questions, err := s.questions.FindByIDs(ctx, ids)
	if err != nil {
		return nil, 0, response.NewBizError(response.CodeDBError)
	}
	if len(questions) != len(ids) {
		return nil, 0, response.NewBizErrorWithDetails(response.CodePaperInvalid, "存在无效或已删除的题目")
	}

	qmap := make(map[int64]model.Question, len(questions))
	for _, q := range questions {
		if q.Status != model.QuestionStatusEnabled {
			return nil, 0, response.NewBizErrorWithDetails(response.CodePaperInvalid, fmt.Sprintf("题目 %d 已停用", q.ID))
		}
		qmap[q.ID] = q
	}

	items := make([]model.PaperQuestion, 0, len(inputs))
	totalScore := 0
	for i, in := range inputs {
		score := qmap[in.QuestionID].DefaultScore
		if in.Score != nil {
			score = *in.Score
		}
		items = append(items, model.PaperQuestion{
			QuestionID: in.QuestionID,
			Seq:        i + 1,
			Score:      score,
		})
		totalScore += score
	}

	return items, totalScore, nil
}

func (s *PaperService) GetByID(ctx context.Context, id int64) (*resDto.PaperDetail, error) {
	return s.getDetail(ctx, id, false)
}

func (s *PaperService) GetPublished(ctx context.Context, id int64) (*resDto.PaperDetail, error) {
	return s.getDetail(ctx, id, true)
}

func (s *PaperService) getDetail(ctx context.Context, id int64, publishedOnly bool) (*resDto.PaperDetail, error) {
	paper, err := s.papers.FindByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, response.NewBizError(response.CodePaperNotFound)
	}
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}
	if publishedOnly && paper.Status != model.PaperStatusPublished {
		return nil, response.NewBizError(response.CodePaperNotFound)
	}

	rows, err := s.papers.FindQuestions(ctx, id)
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	detail := &resDto.PaperDetail{
		ID:            paper.ID,
		Title:         paper.Title,
		Description:   paper.Description,
		Status:        paper.Status,
		TotalScore:    paper.TotalScore,
		QuestionCount: paper.QuestionCount,
		CreatedAt:     paper.CreatedAt,
		Questions:     make([]*resDto.PaperQuestionView, 0, len(rows)),
	}
	for _, row := range rows {
		view := &resDto.PaperQuestionView{
			QuestionID: row.QuestionID,
			Seq:        row.Seq,
			Score:      row.Score,
			Type:       row.Type,
			Content:    row.Content,
		}
		if row.Options != nil {
			var options []model.QuestionOption
			if err := json.Unmarshal([]byte(*row.Options), &options); err == nil {
				view.Options = options
			}
		}
		if !publishedOnly {
			view.Answer = json.RawMessage(row.Answer)
			view.Analysis = row.Analysis
		}
		detail.Questions = append(detail.Questions, view)
	}

	return detail, nil
}

func (s *PaperService) List(ctx context.Context, filter reqDto.PaperFilter) (*resDto.PaperPage, error) {
	papers, total, err := s.papers.List(ctx, filter)
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	items := make([]*resDto.PaperItem, 0, len(papers))
	for i := range papers {
		items = append(items, toPaperItem(&papers[i]))
	}

	return &resDto.PaperPage{List: items, Total: total}, nil
}

func (s *PaperService) ListPublished(ctx context.Context, filter reqDto.PaperFilter) (*resDto.PaperPage, error) {
	status := string(model.PaperStatusPublished)
	filter.Status = &status
	return s.List(ctx, filter)
}

func (s *PaperService) Publish(ctx context.Context, id int64) error {
	paper, err := s.papers.FindByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return response.NewBizError(response.CodePaperNotFound)
	}
	if err != nil {
		return response.NewBizError(response.CodeDBError)
	}
	if paper.Status != model.PaperStatusDraft {
		return response.NewBizError(response.CodePaperStatusInvalid, "仅草稿状态的试卷可发布")
	}

	if err := s.papers.UpdateStatus(ctx, id, model.PaperStatusPublished); err != nil {
		return response.NewBizError(response.CodeDBError)
	}
	return nil
}

func (s *PaperService) Offline(ctx context.Context, id int64) error {
	paper, err := s.papers.FindByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return response.NewBizError(response.CodePaperNotFound)
	}
	if err != nil {
		return response.NewBizError(response.CodeDBError)
	}
	if paper.Status != model.PaperStatusPublished {
		return response.NewBizError(response.CodePaperStatusInvalid, "仅已发布的试卷可下线")
	}

	if err := s.papers.UpdateStatus(ctx, id, model.PaperStatusOffline); err != nil {
		return response.NewBizError(response.CodeDBError)
	}
	return nil
}

func toPaperItem(p *model.Paper) *resDto.PaperItem {
	return &resDto.PaperItem{
		ID:            p.ID,
		Title:         p.Title,
		Description:   p.Description,
		Status:        p.Status,
		TotalScore:    p.TotalScore,
		QuestionCount: p.QuestionCount,
		CreatedAt:     p.CreatedAt,
	}
}
