package kanban

import (
	"errors"
	"project-manager-go/data"

	"gorm.io/gorm"
)

type KanbanStore struct {
	Cards   *cards
	Rows    *rows
	Columns *columns
	Links   *links
	Users   *users
}

func InitKanbanStore(itemsTreeStore *data.TreeStore, projectsStore *data.ProjectsStore) *KanbanStore {
	db := data.GetDB()

	// Add a custom model to the db. This model only relates to Kanban
	err := db.AutoMigrate(&KanbanColumn{})
	if err != nil {
		panic(err)
	}

	// Add a custom model to the db. This model only relates to Kanban
	err = db.AutoMigrate(&KanbanLink{})
	if err != nil {
		panic(err)
	}

	return &KanbanStore{
		Cards: &cards{
			TreeStore: itemsTreeStore,
		},
		Rows: &rows{
			ProjectsStore: projectsStore,
		},
		Columns: &columns{},
		Links:   &links{},
		Users:   &users{},
	}
}

func (s *KanbanStore) HandleItemAddOperation(ctx *data.DBContext, obj *data.Item) error {
	// the Handler is called before the item is created, default values can be defined here

	if obj.Kanban_ColumnID == 0 {
		column := KanbanColumn{}
		err := ctx.DB.Table("kanban_columns").Order("`index`").Take(&column).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		obj.Kanban_ColumnID = column.ID
	}

	if obj.Kanban_CardIndex == 0 {
		maxIndex, err := s.Cards.GetMaxIndex(ctx, obj.ProjectID, obj.Kanban_ColumnID)
		if err != nil {
			return err
		}
		obj.Kanban_CardIndex = maxIndex
	}

	return nil
}

func (s *KanbanStore) HandleProjectAddOperation(ctx *data.DBContext, obj *data.Project) error {
	// the Handler is called before the project is created, default values can be defined here

	if obj.Kanban_RowIndex == 0 {
		maxIndex, err := s.Rows.GetMaxIndex(ctx)
		if err != nil {
			return err
		}
		obj.Kanban_RowIndex = maxIndex
	}

	return nil
}

func (s *KanbanStore) HandleTaskDeleteOperation(ctx *data.DBContext, obj *data.Item) error {
	// the Handler is called before the item is deleted, some relations can be cleared here

	err := ctx.DB.Where("source = ? OR target = ?", obj.ID, obj.ID).Delete(&KanbanLink{}).Error

	return err
}
