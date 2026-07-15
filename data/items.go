package data

import (
	"errors"

	"gorm.io/gorm"
)

// Very base store, that contains primitive methods
type ItemsStore struct {
	Processor *ItemsProcessorProvider
}

func NewItemsStore(provider *ItemsProcessorProvider) *ItemsStore {
	return &ItemsStore{
		Processor: provider,
	}
}

func (s *ItemsStore) GetOne(ctx *DBContext, id int, item IItem) error {
	obj, err := s.getById(ctx, id)
	if err != nil {
		return err
	}

	item.PutItem(obj)

	return nil
}

func (s *ItemsStore) GetAll(ctx *DBContext, list IItemsList) error {
	items := make([]Item, 0)
	err := ctx.DB.Find(&items).Error
	if err != nil {
		return err
	}

	list.PutItems(items)

	return nil
}

func (s *ItemsStore) Add(ctx *DBContext, item IItem) (int, error) {
	obj := Item{}
	item.FillItem(&obj)

	err := s.Processor.Handle("add", ctx, &obj)
	if err != nil {
		return 0, err
	}

	err = ctx.DB.Create(&obj).Error
	return obj.ID, err
}

func (s *ItemsStore) Update(ctx *DBContext, id int, item IItem) error {
	obj, err := s.getById(ctx, id)
	if err != nil {
		return err
	}

	item.FillItem(&obj)
	err = ctx.DB.Save(&obj).Error

	return err
}

func (s *ItemsStore) UpdateFields(ctx *DBContext, id int, fields map[string]any) error {
	_, err := s.getById(ctx, id)
	if err != nil {
		return err
	}

	if len(fields) == 0 {
		return errors.New("fields set must be not empty")
	}

	err = ctx.DB.Model(&Item{ID: id}).Updates(fields).Error

	return err
}

func (s *ItemsStore) Delete(ctx *DBContext, ids ...int) error {
	for _, id := range ids {
		item, err := s.getById(ctx, id)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}

		err = s.Processor.Handle("delete", ctx, &item)
		if err != nil {
			return err
		}
	}

	err := s.DeleteAssociations(ctx, ids...)
	if err != nil {
		return err
	}

	err = ctx.DB.Delete(&Item{}, "id IN (?)", ids).Error

	return err
}

func (s *ItemsStore) DeleteAssociations(ctx *DBContext, ids ...int) error {
	for _, id := range ids {
		err := ctx.DB.
			Model(&Item{ID: id}).
			Association("AssignedUsers").Clear()
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ItemsStore) getById(ctx *DBContext, id int) (Item, error) {
	obj := Item{}
	err := ctx.DB.
		Preload("AssignedUsers").
		Preload("Attached").
		Preload("Comments").
		Preload("Votes").
		Take(&obj, id).Error
	return obj, err
}

func (s *ItemsStore) HandleItemAddOperation(ctx *DBContext, item *Item) error {
	// the Handler is called before the item is created, default values can be defined here

	if item.ProjectID == 0 {
		project := Project{}
		err := ctx.DB.
			Order("`index`").
			Take(&project).Error
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		}
		item.ProjectID = project.ID
	}

	// if item.StartDate == nil {
	// 	if item.EndDate != nil {
	// 		item.StartDate = item.EndDate
	// 	} else {
	// 		now := time.Now()
	// 		item.StartDate = &now
	// 	}
	// }

	// if item.EndDate == nil {
	// 	end := item.StartDate.Add(time.Hour * 24)
	// 	item.EndDate = &end
	// }

	return nil
}
