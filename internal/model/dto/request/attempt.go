package request

import "encoding/json"

type SubmitAnswer struct {
	QuestionID int64           `json:"question_id" binding:"required"`
	Answer     json.RawMessage `json:"answer"` // null 或缺省表示未作答
}

type SubmitPaper struct {
	Answers []SubmitAnswer `json:"answers" binding:"required,min=1,max=200"`
}

type AttemptFilter struct {
	Page     int
	PageSize int
}
