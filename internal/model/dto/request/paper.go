package request

// PaperQuestionInput 组卷时的选题项，数组顺序即题号；Score 为空时使用题目默认分值
type PaperQuestionInput struct {
	QuestionID int64 `json:"question_id" binding:"required"`
	Score      *int  `json:"score" binding:"omitempty,min=1,max=100"`
}

type CreatePaper struct {
	Title       string               `json:"title" binding:"required,min=2,max=128"`
	Description *string              `json:"description" binding:"omitempty,max=512"`
	Questions   []PaperQuestionInput `json:"questions" binding:"required,min=1,max=200"`
}

type UpdatePaper struct {
	Title       *string              `json:"title" binding:"omitempty,min=2,max=128"`
	Description *string              `json:"description" binding:"omitempty,max=512"`
	Questions   []PaperQuestionInput `json:"questions" binding:"omitempty,min=1,max=200"`
}

type PaperFilter struct {
	Status   *string
	Keyword  string
	Page     int
	PageSize int
}
