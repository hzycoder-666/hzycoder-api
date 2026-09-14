package response

import (
	"encoding/json"
	"time"

	"hzycoder.com/lion/internal/model"
)

type PaperItem struct {
	ID            int64             `json:"id"`
	Title         string            `json:"title"`
	Description   *string           `json:"description"`
	Status        model.PaperStatus `json:"status"`
	TotalScore    int               `json:"total_score"`
	QuestionCount int               `json:"question_count"`
	CreatedAt     time.Time         `json:"created_at"`
}

type PaperPage struct {
	List  []*PaperItem `json:"list"`
	Total int64        `json:"total"`
}

// PaperQuestionView 试卷中的单题视图；答题视图不带 Answer 和 Analysis
type PaperQuestionView struct {
	QuestionID int64                  `json:"question_id"`
	Seq        int                    `json:"seq"`
	Score      int                    `json:"score"`
	Type       model.QuestionType     `json:"type"`
	Content    string                 `json:"content"`
	Options    []model.QuestionOption `json:"options"`
	Answer     json.RawMessage        `json:"answer,omitempty"`
	Analysis   *string                `json:"analysis,omitempty"`
}

type PaperDetail struct {
	ID            int64                `json:"id"`
	Title         string               `json:"title"`
	Description   *string              `json:"description"`
	Status        model.PaperStatus    `json:"status"`
	TotalScore    int                  `json:"total_score"`
	QuestionCount int                  `json:"question_count"`
	Questions     []*PaperQuestionView `json:"questions"`
	CreatedAt     time.Time            `json:"created_at"`
}
