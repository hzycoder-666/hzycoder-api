package model

import "time"

type PaperStatus string

const (
	PaperStatusDraft     PaperStatus = "draft"
	PaperStatusPublished PaperStatus = "published"
	PaperStatusOffline   PaperStatus = "offline"
)

type Paper struct {
	ID            int64       `db:"id"`
	Title         string      `db:"title"`
	Description   *string     `db:"description"`
	Status        PaperStatus `db:"status"`
	TotalScore    int         `db:"total_score"`
	QuestionCount int         `db:"question_count"`
	CreatedBy     int64       `db:"created_by"`
	CreatedAt     time.Time   `db:"created_at"`
	UpdatedAt     time.Time   `db:"updated_at"`
}

type PaperQuestion struct {
	ID         int64     `db:"id"`
	PaperID    int64     `db:"paper_id"`
	QuestionID int64     `db:"question_id"`
	Seq        int       `db:"seq"`
	Score      int       `db:"score"`
	CreatedAt  time.Time `db:"created_at"`
}
