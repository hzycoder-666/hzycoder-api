package model

import "time"

type QuestionType string

const (
	QuestionTypeSingle   QuestionType = "single"
	QuestionTypeMultiple QuestionType = "multiple"
	QuestionTypeJudge    QuestionType = "judge"
)

type QuestionStatus string

const (
	QuestionStatusEnabled  QuestionStatus = "enabled"
	QuestionStatusDisabled QuestionStatus = "disabled"
)

// QuestionOption 单个选项，Options 字段在存储层为 JSON 字符串
type QuestionOption struct {
	Key  string `json:"key"`
	Text string `json:"text"`
}

// Question options/answer 在数据库中为 JSON 字符串，由 service 层负责编解码
type Question struct {
	ID           int64          `db:"id"`
	Type         QuestionType   `db:"type"`
	Content      string         `db:"content"`
	Options      *string        `db:"options"`
	Answer       string         `db:"answer"`
	Analysis     *string        `db:"analysis"`
	Tags         *string        `db:"tags"`
	Difficulty   int            `db:"difficulty"`
	DefaultScore int            `db:"default_score"`
	Status       QuestionStatus `db:"status"`
	CreatedBy    int64          `db:"created_by"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
}
