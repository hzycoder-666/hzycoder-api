package repository

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"hzycoder.com/lion/internal/model"
	reqDto "hzycoder.com/lion/internal/model/dto/request"
)

type PaperRepository interface {
	Create(ctx context.Context, paper model.Paper, items []model.PaperQuestion) (*model.Paper, error)
	Update(ctx context.Context, paper model.Paper, items []model.PaperQuestion) error
	FindByID(ctx context.Context, id int64) (*model.Paper, error)
	List(ctx context.Context, filter reqDto.PaperFilter) ([]model.Paper, int64, error)
	UpdateStatus(ctx context.Context, id int64, status model.PaperStatus) error
	FindQuestions(ctx context.Context, paperID int64) ([]PaperQuestionRow, error)
}

// PaperQuestionRow paper_question 联查 question 的结果
type PaperQuestionRow struct {
	model.PaperQuestion
	Type         model.QuestionType   `db:"type"`
	Content      string               `db:"content"`
	Options      *string              `db:"options"`
	Answer       string               `db:"answer"`
	Analysis     *string              `db:"analysis"`
	Tags         *string              `db:"tags"`
	Difficulty   int                  `db:"difficulty"`
	DefaultScore int                  `db:"default_score"`
	Status       model.QuestionStatus `db:"status"`
}

type SQLPaperRepository struct {
	db *sqlx.DB
}

func NewPaperRepository(db *sqlx.DB) *SQLPaperRepository {
	return &SQLPaperRepository{db: db}
}

func (r *SQLPaperRepository) Create(ctx context.Context, paper model.Paper, items []model.PaperQuestion) (*model.Paper, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	now := time.Now()
	paper.CreatedAt = now
	paper.UpdatedAt = now
	for i := range items {
		items[i].CreatedAt = now
	}

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin create paper tx: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.NamedExecContext(ctx, `
		INSERT INTO paper (title, description, status, total_score, question_count, created_by, created_at, updated_at)
		VALUES (:title, :description, :status, :total_score, :question_count, :created_by, :created_at, :updated_at)
	`, paper)
	if err != nil {
		return nil, fmt.Errorf("create paper: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted paper id: %w", err)
	}
	paper.ID = id

	if err := insertPaperQuestions(ctx, tx, id, items); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit create paper tx: %w", err)
	}

	return &paper, nil
}

func (r *SQLPaperRepository) Update(ctx context.Context, paper model.Paper, items []model.PaperQuestion) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	now := time.Now()
	paper.UpdatedAt = now

	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin update paper tx: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.NamedExecContext(ctx, `
		UPDATE paper
		SET title = :title, description = :description, total_score = :total_score,
			question_count = :question_count, updated_at = :updated_at
		WHERE id = :id
	`, paper)
	if err != nil {
		return fmt.Errorf("update paper: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}

	if _, err := tx.ExecContext(ctx, `DELETE FROM paper_question WHERE paper_id = ?`, paper.ID); err != nil {
		return fmt.Errorf("replace paper questions: %w", err)
	}

	if err := insertPaperQuestions(ctx, tx, paper.ID, items); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit update paper tx: %w", err)
	}

	return nil
}

func insertPaperQuestions(ctx context.Context, tx *sqlx.Tx, paperID int64, items []model.PaperQuestion) error {
	for i := range items {
		items[i].PaperID = paperID
		if _, err := tx.NamedExecContext(ctx, `
			INSERT INTO paper_question (paper_id, question_id, seq, score, created_at)
			VALUES (:paper_id, :question_id, :seq, :score, :created_at)
		`, items[i]); err != nil {
			return fmt.Errorf("insert paper question: %w", err)
		}
	}

	return nil
}

func (r *SQLPaperRepository) FindByID(ctx context.Context, id int64) (*model.Paper, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var paper model.Paper
	err := r.db.GetContext(ctx, &paper, `
		SELECT id, title, description, status, total_score, question_count, created_by, created_at, updated_at
		FROM paper
		WHERE id = ?
	`, id)
	if err != nil {
		return nil, err
	}

	return &paper, nil
}

func (r *SQLPaperRepository) List(ctx context.Context, filter reqDto.PaperFilter) ([]model.Paper, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	where, args := buildPaperWhere(filter)

	var total int64
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(1) FROM paper WHERE `+where, args...); err != nil {
		return nil, 0, fmt.Errorf("count papers: %w", err)
	}

	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)
	var papers []model.Paper
	err := r.db.SelectContext(ctx, &papers, `
		SELECT id, title, description, status, total_score, question_count, created_by, created_at, updated_at
		FROM paper
		WHERE `+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list papers: %w", err)
	}

	return papers, total, nil
}

func (r *SQLPaperRepository) UpdateStatus(ctx context.Context, id int64, status model.PaperStatus) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	_, err := r.db.ExecContext(ctx, `UPDATE paper SET status = ?, updated_at = ? WHERE id = ?`, status, time.Now(), id)
	if err != nil {
		return fmt.Errorf("update paper status: %w", err)
	}

	return nil
}

func (r *SQLPaperRepository) FindQuestions(ctx context.Context, paperID int64) ([]PaperQuestionRow, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var rows []PaperQuestionRow
	err := r.db.SelectContext(ctx, &rows, `
		SELECT pq.id, pq.paper_id, pq.question_id, pq.seq, pq.score, pq.created_at,
			q.type, q.content, q.options, q.answer, q.analysis, q.tags, q.difficulty, q.default_score, q.status
		FROM paper_question pq
		JOIN question q ON q.id = pq.question_id
		WHERE pq.paper_id = ?
		ORDER BY pq.seq
	`, paperID)
	if err != nil {
		return nil, fmt.Errorf("find paper questions: %w", err)
	}

	return rows, nil
}

func buildPaperWhere(filter reqDto.PaperFilter) (string, []interface{}) {
	var conds []string
	var args []interface{}

	if filter.Status != nil {
		conds = append(conds, "status = ?")
		args = append(args, *filter.Status)
	}
	if filter.Keyword != "" {
		conds = append(conds, "title LIKE ?")
		args = append(args, "%"+filter.Keyword+"%")
	}
	if len(conds) == 0 {
		return "1 = 1", args
	}

	return strings.Join(conds, " AND "), args
}
