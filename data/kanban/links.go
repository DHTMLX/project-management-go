package kanban

import (
	"project-manager-go/data"
)

type KanbanLink struct {
	ID       int    `json:"id"`
	Source   int    `json:"source"`
	Target   int    `json:"target"`
	Relation string `json:"relation"`
}

type links struct {
	data.ItemsStore
}

func (s *links) GetAll(ctx *data.DBContext) ([]KanbanLink, error) {
	links := make([]KanbanLink, 0)
	err := ctx.DB.
		Find(&links).
		Error

	return links, err
}

func (d *links) GetOne(ctx *data.DBContext, id int) (KanbanLink, error) {
	link := KanbanLink{}
	err := ctx.DB.Take(&link, id).Error
	return link, err
}

func (d *links) Add(ctx *data.DBContext, link KanbanLink) (int, error) {
	err := ctx.DB.Create(&link).Error
	return link.ID, err
}

func (d *links) Delete(ctx *data.DBContext, id int) error {
	err := ctx.DB.Delete(&KanbanLink{}, id).Error
	return err
}
