package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"hzycoder.com/lion/internal/model"
	reqDto "hzycoder.com/lion/internal/model/dto/request"
)

type AttemptRepository interface {
	Create(ctx context.Context, attempt model.ExamAttempt, answers []model.AttemptAnswer) (*model.ExamAttempt, error)
	FindByID(ctx context.Context, id int64) (*model.ExamAttempt, error)
	ListByUser(ctx context.Context, userID int64, filter reqDto.AttemptFilter) ([]AttemptItemRow, int64, error)
	FindAnswers(ctx context.Context, attemptID int64) ([]AttemptAnswerRow, error)
	ListByPaper(ctx context.Context, paperID int64, filter reqDto.AttemptFilter) ([]AdminAttemptItemRow, int64, error)
	StatsByPaper(ctx context.Context, paperID int64) (*PaperAttemptStatsRow, error)
}

// AttemptItemRow exam_attempt 联查 paper 标题与总分
type AttemptItemRow struct {
	model.ExamAttempt
	PaperTitle string `db:"paper_title"`
	PaperTotal int    `db:"paper_total_score"`
}

// AttemptAnswerRow attempt_answer 联查 paper_question（题号/卷内分值）与 question
type AttemptAnswerRow struct {
	model.AttemptAnswer
	Seq        int                `db:"seq"`
	PQScore    int                `db:"pq_score"`
	Type       model.QuestionType `db:"type"`
	Content    string             `db:"content"`
	Options    *string            `db:"options"`
	Answer     string             `db:"answer"`
	Analysis   *string            `db:"analysis"`
	PaperTitle string             `db:"paper_title"`
}

// AdminAttemptItemRow 管理端按试卷查看成绩，联查用户名
type AdminAttemptItemRow struct {
	model.ExamAttempt
	Username string `db:"username"`
}

// PaperAttemptStatsRow 单卷最简统计
type PaperAttemptStatsRow struct {
	AttemptCount int             `db:"attempt_count"`
	AvgScore     sql.NullFloat64 `db:"avg_score"`
	MaxScore     sql.NullInt64   `db:"max_score"`
}

type SQLAttemptRepository struct {
	db *sqlx.DB
}

func NewAttemptRepository(db *sqlx.DB) *SQLAttemptRepository {
	return &SQLAttemptRepository{db: db}
}

func (r *SQLAttemptRepository) Create(ctx context.Context, attempt model.ExamAttempt, answers []model.AttemptAnswer) (*model.ExamAttempt, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	now := time.Now()
	attempt.CreatedAt = now
	for i := range answers {
		answers[i].CreatedAt = now
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create attempt tx: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.NamedExecContext(ctx, `
		INSERT INTO exam_attempt (user_id, paper_id, total_score, created_at)
		VALUES (:user_id, :paper_id, :total_score, :created_at)
	`, attempt)
	if err != nil {
		return nil, fmt.Errorf("create attempt: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted attempt id: %w", err)
	}
	attempt.ID = id

	for i := range answers {
		answers[i].AttemptID = id
		if _, err := tx.NamedExecContext(ctx, `
			INSERT INTO attempt_answer (attempt_id, question_id, user_answer, is_correct, score, created_at)
			VALUES (:attempt_id, :question_id, :user_answer, :is_correct, :score, :created_at)
		`, answers[i]); err != nil {
			return nil, fmt.Errorf("create attempt answer: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create attempt tx: %w", err)
	}

	return &attempt, nil
}

func (r *SQLAttemptRepository) FindByID(ctx context.Context, id int64) (*model.ExamAttempt, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var attempt model.ExamAttempt
	err := r.db.GetContext(ctx, &attempt, `
		SELECT id, user_id, paper_id, total_score, created_at
		FROM exam_attempt
		WHERE id = ?
	`, id)
	if err != nil {
		return nil, err
	}

	return &attempt, nil
}

func (r *SQLAttemptRepository) ListByUser(ctx context.Context, userID int64, filter reqDto.AttemptFilter) ([]AttemptItemRow, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var total int64
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(1) FROM exam_attempt WHERE user_id = ?`, userID); err != nil {
		return nil, 0, fmt.Errorf("count attempts: %w", err)
	}

	var rows []AttemptItemRow
	err := r.db.SelectContext(ctx, &rows, `
		SELECT a.id, a.user_id, a.paper_id, a.total_score, a.created_at,
			p.title AS paper_title, p.total_score AS paper_total_score
		FROM exam_attempt a
		JOIN paper p ON p.id = a.paper_id
		WHERE a.user_id = ?
		ORDER BY a.id DESC
		LIMIT ? OFFSET ?
	`, userID, filter.PageSize, (filter.Page-1)*filter.PageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list attempts: %w", err)
	}

	return rows, total, nil
}

func (r *SQLAttemptRepository) FindAnswers(ctx context.Context, attemptID int64) ([]AttemptAnswerRow, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var rows []AttemptAnswerRow
	err := r.db.SelectContext(ctx, &rows, `
		SELECT aa.id, aa.attempt_id, aa.question_id, aa.user_answer, aa.is_correct, aa.score, aa.created_at,
			pq.seq, pq.score AS pq_score, q.type, q.content, q.options, q.answer, q.analysis,
			p.title AS paper_title
		FROM attempt_answer aa
		JOIN exam_attempt a ON a.id = aa.attempt_id
		JOIN paper p ON p.id = a.paper_id
		JOIN paper_question pq ON pq.paper_id = a.paper_id AND pq.question_id = aa.question_id
		JOIN question q ON q.id = aa.question_id
		WHERE aa.attempt_id = ?
		ORDER BY pq.seq
	`, attemptID)
	if err != nil {
		return nil, fmt.Errorf("find attempt answers: %w", err)
	}

	return rows, nil
}

func (r *SQLAttemptRepository) ListByPaper(ctx context.Context, paperID int64, filter reqDto.AttemptFilter) ([]AdminAttemptItemRow, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var total int64
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(1) FROM exam_attempt WHERE paper_id = ?`, paperID); err != nil {
		return nil, 0, fmt.Errorf("count paper attempts: %w", err)
	}

	var rows []AdminAttemptItemRow
	err := r.db.SelectContext(ctx, &rows, `
		SELECT a.id, a.user_id, a.paper_id, a.total_score, a.created_at,
			u.username
		FROM exam_attempt a
		JOIN sys_user u ON u.id = a.user_id
		WHERE a.paper_id = ?
		ORDER BY a.id DESC
		LIMIT ? OFFSET ?
	`, paperID, filter.PageSize, (filter.Page-1)*filter.PageSize)
	if err != nil {
		return nil, 0, fmt.Errorf("list paper attempts: %w", err)
	}

	return rows, total, nil
}

func (r *SQLAttemptRepository) StatsByPaper(ctx context.Context, paperID int64) (*PaperAttemptStatsRow, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var stats PaperAttemptStatsRow
	err := r.db.GetContext(ctx, &stats, `
		SELECT COUNT(1) AS attempt_count,
			AVG(total_score) AS avg_score,
			MAX(total_score) AS max_score
		FROM exam_attempt
		WHERE paper_id = ?
	`, paperID)
	if err != nil {
		return nil, fmt.Errorf("stats paper attempts: %w", err)
	}

	return &stats, nil
}
