package response

import (
	"encoding/json"
	"time"

	"hzycoder.com/lion/internal/model"
)

type QuestionItem struct {
	ID           int64                  `json:"id"`
	Type         model.QuestionType     `json:"type"`
	Content      string                 `json:"content"`
	Options      []model.QuestionOption `json:"options"`
	Answer       json.RawMessage        `json:"answer"`
	Analysis     *string                `json:"analysis"`
	Tags         *string                `json:"tags"`
	Difficulty   int                    `json:"difficulty"`
	DefaultScore int                    `json:"default_score"`
	Status       model.QuestionStatus   `json:"status"`
	CreatedAt    time.Time              `json:"created_at"`
}

type QuestionPage struct {
	List  []*QuestionItem `json:"list"`
	Total int64           `json:"total"`
}
