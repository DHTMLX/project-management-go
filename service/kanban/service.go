package kanban

import (
	"project-manager-go/common"
	"project-manager-go/data/kanban"
	"project-manager-go/service"
)

type MoveParams struct {
	ID       common.TID      `json:"id"`
	Before   common.FuzzyInt `json:"before"`
	ColumnID common.FuzzyInt `json:"columnId"`
	RowID    common.FuzzyInt `json:"rowId"`
}

type KanbanService struct {
	Cards *cards
	Rows  *rows
	Cols  *cols
	Links *links
	Users *users
}

func NewKanbanService(store *kanban.KanbanStore, tree *service.TreeService) *KanbanService {
	return &KanbanService{
		Cards: &cards{tree, store},
		Rows:  &rows{store},
		Cols:  &cols{store},
		Links: &links{store},
		Users: &users{store},
	}
}
