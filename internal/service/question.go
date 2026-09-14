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

type QuestionService struct {
	questions repository.QuestionRepository
}

func NewQuestionService(questions repository.QuestionRepository) *QuestionService {
	return &QuestionService{questions: questions}
}

func (s *QuestionService) Create(ctx context.Context, userID int64, req reqDto.CreateQuestion) (*resDto.QuestionItem, error) {
	options, answer, err := validateQuestionPayload(req.Type, req.Options, req.Answer)
	if err != nil {
		return nil, err
	}

	question := model.Question{
		Type:      req.Type,
		Content:   req.Content,
		Options:   options,
		Answer:    answer,
		Analysis:  req.Analysis,
		Tags:      req.Tags,
		Status:    model.QuestionStatusEnabled,
		CreatedBy: userID,
	}
	if req.Difficulty != nil {
		question.Difficulty = *req.Difficulty
	}
	if req.DefaultScore != nil {
		question.DefaultScore = *req.DefaultScore
	}

	created, err := s.questions.Create(ctx, question)
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	return toQuestionItem(created), nil
}

func (s *QuestionService) Update(ctx context.Context, id int64, req reqDto.UpdateQuestion) (*resDto.QuestionItem, error) {
	question, err := s.questions.FindByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, response.NewBizError(response.CodeQuestionNotFound)
	}
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	applyQuestionUpdate(question, req)

	options, answer, err := validateQuestionPayload(question.Type, mustOptions(question.Options), json.RawMessage(question.Answer))
	if err != nil {
		return nil, err
	}
	question.Options = options
	question.Answer = answer

	if err := s.questions.Update(ctx, *question); err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	return toQuestionItem(question), nil
}

func applyQuestionUpdate(question *model.Question, req reqDto.UpdateQuestion) {
	if req.Type != nil {
		question.Type = *req.Type
	}
	if req.Content != nil {
		question.Content = *req.Content
	}
	if req.Options != nil {
		options, _ := json.Marshal(*req.Options)
		question.Options = strPtr(string(options))
	}
	if req.Answer != nil {
		question.Answer = string(req.Answer)
	}
	if req.Analysis != nil {
		question.Analysis = req.Analysis
	}
	if req.Tags != nil {
		question.Tags = req.Tags
	}
	if req.Difficulty != nil {
		question.Difficulty = *req.Difficulty
	}
	if req.DefaultScore != nil {
		question.DefaultScore = *req.DefaultScore
	}
	if req.Status != nil {
		question.Status = *req.Status
	}
}

func (s *QuestionService) GetByID(ctx context.Context, id int64) (*resDto.QuestionItem, error) {
	question, err := s.questions.FindByID(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, response.NewBizError(response.CodeQuestionNotFound)
	}
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	return toQuestionItem(question), nil
}

func (s *QuestionService) Delete(ctx context.Context, id int64) error {
	refs, err := s.questions.CountPaperRefs(ctx, id)
	if err != nil {
		return response.NewBizError(response.CodeDBError)
	}
	if refs > 0 {
		return response.NewBizError(response.CodeQuestionReferenced)
	}

	err = s.questions.Delete(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return response.NewBizError(response.CodeQuestionNotFound)
	}
	if err != nil {
		return response.NewBizError(response.CodeDBError)
	}

	return nil
}

func (s *QuestionService) List(ctx context.Context, filter reqDto.QuestionFilter) (*resDto.QuestionPage, error) {
	questions, total, err := s.questions.List(ctx, filter)
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	items := make([]*resDto.QuestionItem, 0, len(questions))
	for i := range questions {
		items = append(items, toQuestionItem(&questions[i]))
	}

	return &resDto.QuestionPage{List: items, Total: total}, nil
}

// validateQuestionPayload 校验题型、选项、答案的组合合法性，并返回规范化后的 JSON 存储形式
func validateQuestionPayload(qType model.QuestionType, options []model.QuestionOption, answer json.RawMessage) (*string, string, error) {
	invalid := func(format string, args ...interface{}) error {
		return response.NewBizErrorWithDetails(response.CodeQuestionInvalid, fmt.Sprintf(format, args...))
	}

	switch qType {
	case model.QuestionTypeJudge:
		if len(options) > 0 {
			return nil, "", invalid("判断题不能设置选项")
		}
		var b bool
		if err := json.Unmarshal(answer, &b); err != nil {
			return nil, "", invalid("判断题答案必须是 true 或 false")
		}
		answerJSON, _ := json.Marshal(b)
		return nil, string(answerJSON), nil

	case model.QuestionTypeSingle, model.QuestionTypeMultiple:
		if len(options) < 2 {
			return nil, "", invalid("选择题至少需要 2 个选项")
		}
		keys := make(map[string]struct{}, len(options))
		for _, o := range options {
			if o.Key == "" || o.Text == "" {
				return nil, "", invalid("选项的 key 和 text 不能为空")
			}
			if _, dup := keys[o.Key]; dup {
				return nil, "", invalid("选项 key 重复: %s", o.Key)
			}
			keys[o.Key] = struct{}{}
		}

		optionsJSON, err := json.Marshal(options)
		if err != nil {
			return nil, "", invalid("选项序列化失败")
		}

		if qType == model.QuestionTypeSingle {
			var key string
			if err := json.Unmarshal(answer, &key); err != nil {
				return nil, "", invalid("单选题答案必须是选项 key 字符串")
			}
			if _, ok := keys[key]; !ok {
				return nil, "", invalid("单选题答案 %q 不在选项中", key)
			}
			answerJSON, _ := json.Marshal(key)
			return strPtr(string(optionsJSON)), string(answerJSON), nil
		}

		var inKeys []string
		if err := json.Unmarshal(answer, &inKeys); err != nil || len(inKeys) == 0 {
			return nil, "", invalid("多选题答案必须是非空的选项 key 数组")
		}
		uniq := make([]string, 0, len(inKeys))
		seen := make(map[string]struct{}, len(inKeys))
		for _, k := range inKeys {
			if _, ok := keys[k]; !ok {
				return nil, "", invalid("多选题答案 %q 不在选项中", k)
			}
			if _, dup := seen[k]; !dup {
				seen[k] = struct{}{}
				uniq = append(uniq, k)
			}
		}
		answerJSON, _ := json.Marshal(uniq)
		return strPtr(string(optionsJSON)), string(answerJSON), nil

	default:
		return nil, "", invalid("不支持的题型: %s", qType)
	}
}

func mustOptions(raw *string) []model.QuestionOption {
	if raw == nil {
		return nil
	}
	var options []model.QuestionOption
	if err := json.Unmarshal([]byte(*raw), &options); err != nil {
		return nil
	}
	return options
}

func toQuestionItem(q *model.Question) *resDto.QuestionItem {
	item := &resDto.QuestionItem{
		ID:           q.ID,
		Type:         q.Type,
		Content:      q.Content,
		Answer:       json.RawMessage(q.Answer),
		Analysis:     q.Analysis,
		Tags:         q.Tags,
		Difficulty:   q.Difficulty,
		DefaultScore: q.DefaultScore,
		Status:       q.Status,
		CreatedAt:    q.CreatedAt,
	}
	if q.Options != nil {
		var options []model.QuestionOption
		if err := json.Unmarshal([]byte(*q.Options), &options); err == nil {
			item.Options = options
		}
	}
	return item
}

func strPtr(s string) *string {
	return &s
}
