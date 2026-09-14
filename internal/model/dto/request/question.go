package request

import (
	"encoding/json"

	"hzycoder.com/lion/internal/model"
)

// CreateQuestion 新建题目。Answer 为 JSON 原始值，具体结构由题型决定：
// single: "A"，multiple: ["A","C"]，judge: true
type CreateQuestion struct {
	Type         model.QuestionType     `json:"type" binding:"required,oneof=single multiple judge"`
	Content      string                 `json:"content" binding:"required,min=2,max=2000"`
	Options      []model.QuestionOption `json:"options" binding:"max=10"`
	Answer       json.RawMessage        `json:"answer" binding:"required"`
	Analysis     *string                `json:"analysis" binding:"omitempty,max=2000"`
	Tags         *string                `json:"tags" binding:"omitempty,max=255"`
	Difficulty   *int                   `json:"difficulty" binding:"omitempty,min=1,max=5"`
	DefaultScore *int                   `json:"default_score" binding:"omitempty,min=1,max=100"`
}

// UpdateQuestion 全量更新题目，nil 字段表示不修改
type UpdateQuestion struct {
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

type QuestionFilter struct {
	Type       *model.QuestionType
	Status     *model.QuestionStatus
	Difficulty *int
	Tag        string
	Keyword    string
	Page       int
	PageSize   int
}
