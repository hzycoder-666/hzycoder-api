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

type QuestionRepository interface {
	Create(ctx context.Context, question model.Question) (*model.Question, error)
	Update(ctx context.Context, question model.Question) error
	FindByID(ctx context.Context, id int64) (*model.Question, error)
	FindByIDs(ctx context.Context, ids []int64) ([]model.Question, error)
	Delete(ctx context.Context, id int64) error
	CountPaperRefs(ctx context.Context, id int64) (int, error)
	List(ctx context.Context, filter reqDto.QuestionFilter) ([]model.Question, int64, error)
}

type SQLQuestionRepository struct {
	db *sqlx.DB
}

func NewQuestionRepository(db *sqlx.DB) *SQLQuestionRepository {
	return &SQLQuestionRepository{db: db}
}

func (r *SQLQuestionRepository) Create(ctx context.Context, question model.Question) (*model.Question, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	now := time.Now()
	question.CreatedAt = now
	question.UpdatedAt = now

	result, err := r.db.NamedExecContext(ctx, `
		INSERT INTO question (type, content, options, answer, analysis, tags, difficulty, default_score, status, created_by, created_at, updated_at)
		VALUES (:type, :content, :options, :answer, :analysis, :tags, :difficulty, :default_score, :status, :created_by, :created_at, :updated_at)
	`, question)
	if err != nil {
		return nil, fmt.Errorf("create question: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get inserted question id: %w", err)
	}
	question.ID = id

	return &question, nil
}

func (r *SQLQuestionRepository) Update(ctx context.Context, question model.Question) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	question.UpdatedAt = time.Now()

	_, err := r.db.NamedExecContext(ctx, `
		UPDATE question
		SET type = :type, content = :content, options = :options, answer = :answer,
			analysis = :analysis, tags = :tags, difficulty = :difficulty,
			default_score = :default_score, status = :status, updated_at = :updated_at
		WHERE id = :id
	`, question)
	if err != nil {
		return fmt.Errorf("update question: %w", err)
	}

	return nil
}

func (r *SQLQuestionRepository) FindByID(ctx context.Context, id int64) (*model.Question, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var question model.Question
	err := r.db.GetContext(ctx, &question, `
		SELECT id, type, content, options, answer, analysis, tags, difficulty, default_score, status, created_by, created_at, updated_at
		FROM question
		WHERE id = ?
	`, id)
	if err != nil {
		return nil, err
	}

	return &question, nil
}

func (r *SQLQuestionRepository) FindByIDs(ctx context.Context, ids []int64) ([]model.Question, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	query, args, err := sqlx.In(`
		SELECT id, type, content, options, answer, analysis, tags, difficulty, default_score, status, created_by, created_at, updated_at
		FROM question
		WHERE id IN (?)
	`, ids)
	if err != nil {
		return nil, fmt.Errorf("build find questions by ids query: %w", err)
	}

	var questions []model.Question
	if err := r.db.SelectContext(ctx, &questions, r.db.Rebind(query), args...); err != nil {
		return nil, fmt.Errorf("find questions by ids: %w", err)
	}

	return questions, nil
}

func (r *SQLQuestionRepository) Delete(ctx context.Context, id int64) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	result, err := r.db.ExecContext(ctx, `DELETE FROM question WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete question: %w", err)
	}
	if affected, _ := result.RowsAffected(); affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func (r *SQLQuestionRepository) CountPaperRefs(ctx context.Context, id int64) (int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var count int
	err := r.db.GetContext(ctx, &count, `SELECT COUNT(1) FROM paper_question WHERE question_id = ?`, id)
	if err != nil {
		return 0, fmt.Errorf("count question paper refs: %w", err)
	}

	return count, nil
}

func (r *SQLQuestionRepository) List(ctx context.Context, filter reqDto.QuestionFilter) ([]model.Question, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	where, args := buildQuestionWhere(filter)

	var total int64
	if err := r.db.GetContext(ctx, &total, `SELECT COUNT(1) FROM question WHERE `+where, args...); err != nil {
		return nil, 0, fmt.Errorf("count questions: %w", err)
	}

	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)
	var questions []model.Question
	err := r.db.SelectContext(ctx, &questions, `
		SELECT id, type, content, options, answer, analysis, tags, difficulty, default_score, status, created_by, created_at, updated_at
		FROM question
		WHERE `+where+`
		ORDER BY id DESC
		LIMIT ? OFFSET ?
	`, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list questions: %w", err)
	}

	return questions, total, nil
}

func buildQuestionWhere(filter reqDto.QuestionFilter) (string, []interface{}) {
	var conds []string
	var args []interface{}

	if filter.Type != nil {
		conds = append(conds, "type = ?")
		args = append(args, *filter.Type)
	}
	if filter.Status != nil {
		conds = append(conds, "status = ?")
		args = append(args, *filter.Status)
	}
	if filter.Difficulty != nil {
		conds = append(conds, "difficulty = ?")
		args = append(args, *filter.Difficulty)
	}
	if filter.Tag != "" {
		conds = append(conds, "FIND_IN_SET(?, tags)")
		args = append(args, filter.Tag)
	}
	if filter.Keyword != "" {
		conds = append(conds, "content LIKE ?")
		args = append(args, "%"+filter.Keyword+"%")
	}
	if len(conds) == 0 {
		return "1 = 1", args
	}

	return strings.Join(conds, " AND "), args
}
