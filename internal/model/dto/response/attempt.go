package response

import (
	"encoding/json"
	"time"

	"hzycoder.com/lion/internal/model"
)

// AnswerResult 提交后立即返回的逐题评价
type AnswerResult struct {
	QuestionID    int64           `json:"question_id"`
	Seq           int             `json:"seq"`
	Score         int             `json:"score"`      // 题目分值
	UserScore     int             `json:"user_score"` // 本题得分
	IsCorrect     bool            `json:"is_correct"`
	UserAnswer    json.RawMessage `json:"user_answer"`
	CorrectAnswer json.RawMessage `json:"correct_answer"`
	Analysis      *string         `json:"analysis,omitempty"`
}

type SubmitResult struct {
	AttemptID  int64           `json:"attempt_id"`
	PaperID    int64           `json:"paper_id"`
	Title      string          `json:"title"`
	UserScore  int             `json:"user_score"`
	TotalScore int             `json:"total_score"`
	Answers    []*AnswerResult `json:"answers"`
}

type AttemptItem struct {
	ID         int64     `json:"id"`
	PaperID    int64     `json:"paper_id"`
	PaperTitle string    `json:"paper_title"`
	UserScore  int       `json:"user_score"`
	TotalScore int       `json:"total_score"`
	CreatedAt  time.Time `json:"created_at"`
}

type AttemptPage struct {
	List  []*AttemptItem `json:"list"`
	Total int64          `json:"total"`
}

// AttemptAnswerView 单次答卷详情里的逐题记录
type AttemptAnswerView struct {
	QuestionID    int64                  `json:"question_id"`
	Seq           int                    `json:"seq"`
	Type          model.QuestionType     `json:"type"`
	Content       string                 `json:"content"`
	Options       []model.QuestionOption `json:"options"`
	Score         int                    `json:"score"`
	UserScore     int                    `json:"user_score"`
	IsCorrect     bool                   `json:"is_correct"`
	UserAnswer    json.RawMessage        `json:"user_answer"`
	CorrectAnswer json.RawMessage        `json:"correct_answer"`
	Analysis      *string                `json:"analysis,omitempty"`
}

type AttemptDetail struct {
	ID         int64                `json:"id"`
	PaperID    int64                `json:"paper_id"`
	PaperTitle string               `json:"paper_title"`
	UserScore  int                  `json:"user_score"`
	TotalScore int                  `json:"total_score"`
	CreatedAt  time.Time            `json:"created_at"`
	Answers    []*AttemptAnswerView `json:"answers"`
}

// AdminAttemptItem 管理端成绩列表项：含作答用户
type AdminAttemptItem struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Username  string    `json:"username"`
	UserScore int       `json:"user_score"`
	CreatedAt time.Time `json:"created_at"`
}

type AdminAttemptPage struct {
	List       []*AdminAttemptItem `json:"list"`
	Total      int64               `json:"total"`
	TotalScore int                 `json:"total_score"` // 试卷满分，便于前端算得分率
}

// PaperStats 单卷最简统计
type PaperStats struct {
	PaperID      int64   `json:"paper_id"`
	Title        string  `json:"title"`
	TotalScore   int     `json:"total_score"`
	AttemptCount int     `json:"attempt_count"`
	AvgScore     float64 `json:"avg_score"`
	MaxScore     int     `json:"max_score"`
}
