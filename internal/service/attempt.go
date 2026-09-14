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

type AttemptService struct {
	attempts repository.AttemptRepository
	papers   repository.PaperRepository
}

func NewAttemptService(attempts repository.AttemptRepository, papers repository.PaperRepository) *AttemptService {
	return &AttemptService{attempts: attempts, papers: papers}
}

// Submit 练习模式提交：校验题目归属后整卷评分，单事务落库并返回逐题评价
func (s *AttemptService) Submit(ctx context.Context, userID, paperID int64, req reqDto.SubmitPaper) (*resDto.SubmitResult, error) {
	paper, err := s.papers.FindByID(ctx, paperID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, response.NewBizError(response.CodePaperNotFound)
	}
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}
	if paper.Status != model.PaperStatusPublished {
		return nil, response.NewBizError(response.CodePaperNotFound)
	}

	rows, err := s.papers.FindQuestions(ctx, paperID)
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	userAnswers, err := indexUserAnswers(req.Answers)
	if err != nil {
		return nil, err
	}
	validIDs := make(map[int64]struct{}, len(rows))
	for _, row := range rows {
		validIDs[row.QuestionID] = struct{}{}
	}
	for questionID := range userAnswers {
		if _, ok := validIDs[questionID]; !ok {
			return nil, response.NewBizErrorWithDetails(response.CodeParamInvalid, fmt.Sprintf("题目 %d 不属于该试卷", questionID))
		}
	}

	answers := make([]model.AttemptAnswer, 0, len(rows))
	results := make([]*resDto.AnswerResult, 0, len(rows))
	userScore := 0

	for _, row := range rows {
		raw, answered := userAnswers[row.QuestionID]

		var userRaw *string
		var userJSON json.RawMessage
		if answered && isProvidedAnswer(raw) {
			s := string(raw)
			userRaw = &s
			userJSON = raw
		}

		correct := answered && isProvidedAnswer(raw) && judgeCorrect(row.Type, row.Answer, raw)
		got := 0
		if correct {
			got = row.Score
			userScore += got
		}

		answers = append(answers, model.AttemptAnswer{
			QuestionID: row.QuestionID,
			UserAnswer: userRaw,
			IsCorrect:  correct,
			Score:      got,
		})
		results = append(results, &resDto.AnswerResult{
			QuestionID:    row.QuestionID,
			Seq:           row.Seq,
			Score:         row.Score,
			UserScore:     got,
			IsCorrect:     correct,
			UserAnswer:    userJSON,
			CorrectAnswer: json.RawMessage(row.Answer),
			Analysis:      row.Analysis,
		})
	}

	attempt, err := s.attempts.Create(ctx, model.ExamAttempt{
		UserID:     userID,
		PaperID:    paperID,
		TotalScore: userScore,
	}, answers)
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	return &resDto.SubmitResult{
		AttemptID:  attempt.ID,
		PaperID:    paper.ID,
		Title:      paper.Title,
		UserScore:  userScore,
		TotalScore: paper.TotalScore,
		Answers:    results,
	}, nil
}

func indexUserAnswers(inputs []reqDto.SubmitAnswer) (map[int64]json.RawMessage, error) {
	indexed := make(map[int64]json.RawMessage, len(inputs))
	for _, in := range inputs {
		if _, dup := indexed[in.QuestionID]; dup {
			return nil, response.NewBizErrorWithDetails(response.CodeParamInvalid, fmt.Sprintf("题目 %d 重复作答", in.QuestionID))
		}
		indexed[in.QuestionID] = in.Answer
	}
	return indexed, nil
}

// isProvidedAnswer 区分"未作答"（null / 缺省 / 空串）
func isProvidedAnswer(raw json.RawMessage) bool {
	return len(raw) > 0 && string(raw) != "null"
}

// judgeCorrect 按题型判分：判断比布尔，单选比字符串，多选比集合
func judgeCorrect(qType model.QuestionType, correctJSON string, user json.RawMessage) bool {
	switch qType {
	case model.QuestionTypeJudge:
		var correct, given bool
		if err := json.Unmarshal([]byte(correctJSON), &correct); err != nil {
			return false
		}
		if err := json.Unmarshal(user, &given); err != nil {
			return false
		}
		return correct == given

	case model.QuestionTypeSingle:
		var correct, given string
		if err := json.Unmarshal([]byte(correctJSON), &correct); err != nil {
			return false
		}
		if err := json.Unmarshal(user, &given); err != nil {
			return false
		}
		return correct == given

	case model.QuestionTypeMultiple:
		var correct, given []string
		if err := json.Unmarshal([]byte(correctJSON), &correct); err != nil {
			return false
		}
		if err := json.Unmarshal(user, &given); err != nil {
			return false
		}
		return sameStringSet(correct, given)

	default:
		return false
	}
}

func sameStringSet(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	set := make(map[string]int, len(a))
	for _, s := range a {
		set[s]++
	}
	for _, s := range b {
		if set[s] == 0 {
			return false
		}
		set[s]--
	}
	return true
}

func (s *AttemptService) List(ctx context.Context, userID int64, page, pageSize int) (*resDto.AttemptPage, error) {
	rows, total, err := s.attempts.ListByUser(ctx, userID, reqDto.AttemptFilter{Page: page, PageSize: pageSize})
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	items := make([]*resDto.AttemptItem, 0, len(rows))
	for i := range rows {
		items = append(items, &resDto.AttemptItem{
			ID:         rows[i].ID,
			PaperID:    rows[i].PaperID,
			PaperTitle: rows[i].PaperTitle,
			UserScore:  rows[i].TotalScore,
			TotalScore: rows[i].PaperTotal,
			CreatedAt:  rows[i].CreatedAt,
		})
	}

	return &resDto.AttemptPage{List: items, Total: total}, nil
}

// GetByID 用户仅能查看本人的答卷
func (s *AttemptService) GetByID(ctx context.Context, userID, attemptID int64) (*resDto.AttemptDetail, error) {
	attempt, err := s.attempts.FindByID(ctx, attemptID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, response.NewBizError(response.CodeAttemptNotFound)
	}
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}
	if attempt.UserID != userID {
		return nil, response.NewBizError(response.CodeAttemptNotFound)
	}

	rows, err := s.attempts.FindAnswers(ctx, attemptID)
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	detail := &resDto.AttemptDetail{
		ID:        attempt.ID,
		PaperID:   attempt.PaperID,
		UserScore: attempt.TotalScore,
		CreatedAt: attempt.CreatedAt,
		Answers:   make([]*resDto.AttemptAnswerView, 0, len(rows)),
	}
	for _, row := range rows {
		detail.PaperTitle = row.PaperTitle
		detail.TotalScore += row.PQScore
		view := &resDto.AttemptAnswerView{
			QuestionID:    row.QuestionID,
			Seq:           row.Seq,
			Type:          row.Type,
			Content:       row.Content,
			Score:         row.PQScore,
			UserScore:     row.Score,
			IsCorrect:     row.IsCorrect,
			CorrectAnswer: json.RawMessage(row.Answer),
			Analysis:      row.Analysis,
		}
		if row.UserAnswer != nil {
			view.UserAnswer = json.RawMessage(*row.UserAnswer)
		}
		if row.Options != nil {
			var options []model.QuestionOption
			if err := json.Unmarshal([]byte(*row.Options), &options); err == nil {
				view.Options = options
			}
		}
		detail.Answers = append(detail.Answers, view)
	}

	return detail, nil
}

// ListByPaper 管理端按试卷查看成绩列表（任意状态的试卷）
func (s *AttemptService) ListByPaper(ctx context.Context, paperID int64, page, pageSize int) (*resDto.AdminAttemptPage, error) {
	paper, err := s.papers.FindByID(ctx, paperID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, response.NewBizError(response.CodePaperNotFound)
	}
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	rows, total, err := s.attempts.ListByPaper(ctx, paperID, reqDto.AttemptFilter{Page: page, PageSize: pageSize})
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	items := make([]*resDto.AdminAttemptItem, 0, len(rows))
	for i := range rows {
		items = append(items, &resDto.AdminAttemptItem{
			ID:        rows[i].ID,
			UserID:    rows[i].UserID,
			Username:  rows[i].Username,
			UserScore: rows[i].TotalScore,
			CreatedAt: rows[i].CreatedAt,
		})
	}

	return &resDto.AdminAttemptPage{
		List:       items,
		Total:      total,
		TotalScore: paper.TotalScore,
	}, nil
}

// StatsByPaper 管理端单卷最简统计（任意状态的试卷；无作答时均值/最高为 0）
func (s *AttemptService) StatsByPaper(ctx context.Context, paperID int64) (*resDto.PaperStats, error) {
	paper, err := s.papers.FindByID(ctx, paperID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, response.NewBizError(response.CodePaperNotFound)
	}
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	stats, err := s.attempts.StatsByPaper(ctx, paperID)
	if err != nil {
		return nil, response.NewBizError(response.CodeDBError)
	}

	result := &resDto.PaperStats{
		PaperID:      paper.ID,
		Title:        paper.Title,
		TotalScore:   paper.TotalScore,
		AttemptCount: stats.AttemptCount,
	}
	if stats.AvgScore.Valid {
		result.AvgScore = stats.AvgScore.Float64
	}
	if stats.MaxScore.Valid {
		result.MaxScore = int(stats.MaxScore.Int64)
	}

	return result, nil
}
