package kanban

import (
	"project-manager-go/api/context"
	"project-manager-go/common"
	"project-manager-go/data"
	"project-manager-go/data/kanban"
)

type Link struct {
	ID       common.TID      `json:"id"`
	Source   common.FuzzyInt `json:"source"`
	Target   common.FuzzyInt `json:"target"`
	Relation string          `json:"relation"`
}

type links struct {
	store *kanban.KanbanStore
}

func (s *links) GetAll(userCtx context.UserContext, dbCtx *data.DBContext) (arr []kanban.KanbanLink, err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	arr, err = s.store.Links.GetAll(dbCtx)

	return arr, err
}

func (s *links) Add(userCtx context.UserContext, dbCtx *data.DBContext, link Link) (id int, err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	upd := kanban.KanbanLink{
		ID:       int(link.ID),
		Source:   int(link.Source),
		Target:   int(link.Target),
		Relation: link.Relation,
	}

	id, err = s.store.Links.Add(dbCtx, upd)

	return id, err
}

func (s *links) Delete(userCtx context.UserContext, dbCtx *data.DBContext, id int) (err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	err = s.store.Links.Delete(dbCtx, id)

	return err
}
