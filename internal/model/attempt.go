package model

import "time"

// ExamAttempt 一次提交的答卷（练习模式：提交即完成，无进行中状态）
type ExamAttempt struct {
	ID         int64     `db:"id"`
	UserID     int64     `db:"user_id"`
	PaperID    int64     `db:"paper_id"`
	TotalScore int       `db:"total_score"`
	CreatedAt  time.Time `db:"created_at"`
}

// AttemptAnswer user_answer 为 JSON 字符串指针，未作答时为 NULL
type AttemptAnswer struct {
	ID         int64     `db:"id"`
	AttemptID  int64     `db:"attempt_id"`
	QuestionID int64     `db:"question_id"`
	UserAnswer *string   `db:"user_answer"`
	IsCorrect  bool      `db:"is_correct"`
	Score      int       `db:"score"`
	CreatedAt  time.Time `db:"created_at"`
}
