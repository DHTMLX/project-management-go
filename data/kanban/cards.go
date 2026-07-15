package kanban

import (
	"errors"
	"project-manager-go/data"

	"gorm.io/gorm"
)

type cards struct {
	*data.TreeStore
}

func (s *cards) GetAll(ctx *data.DBContext, dest data.IItemsList) error {
	cards := make([]data.Item, 0)
	err := ctx.DB.
		Preload("AssignedUsers").
		Preload("Votes").
		Preload("Attached").
		Preload("Comments").
		Order("kanban_index").
		Find(&cards).
		Error
	if err != nil {
		return err
	}

	dest.PutItems(cards)

	return nil
}

func (s *cards) GetMaxIndex(ctx *data.DBContext, row, column int) (int, error) {
	item := data.Item{}
	err := ctx.DB.
		Where("project_id = ? AND kanban_column_id = ?", row, column).
		Order("kanban_index DESC").
		Take(&item).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}

	return item.Kanban_CardIndex + 1, err
}

func (s *cards) ShiftIndex(ctx *data.DBContext, from, offset int) error {
	err := ctx.DB.
		Model(data.Item{}).
		Where("kanban_index >= ?", from).
		Update("kanban_index", gorm.Expr("kanban_index + ?", offset)).
		Error
	return err
}

func (s *cards) DeleteVote(ctx *data.DBContext, cardId int, userId int, voteUser data.VoteUser) error {
	err := ctx.DB.Model(&data.VoteUser{}).Where("user_id = ? AND item_id = ?", userId, cardId).First(&voteUser).Error
	if err != nil {
		return err
	}

	err = ctx.DB.Delete(&voteUser).Error
	return err
}

type CommentInput struct {
	Date string `json:"date,omitempty"`
	Text string `json:"text"`
}

func (s *cards) AddComment(ctx *data.DBContext, userId int, id int, comment CommentInput) (int, error) {
	newComment := data.Comment{
		Text:   comment.Text,
		UserID: userId,
		ItemID: id,
		Date:   comment.Date,
	}

	err := ctx.DB.Create(&newComment).Error
	return newComment.ID, err
}

func (s *cards) DeleteComment(ctx *data.DBContext, id int) error {
	err := ctx.DB.Delete(&data.Comment{}, id).Error
	return err
}

func (s *cards) UpdateComment(ctx *data.DBContext, id int, comment CommentInput) error {
	commentToUpdate := data.Comment{}
	err := ctx.DB.First(&commentToUpdate, id).Update("text", comment.Text).Error
	return err
}

func (s *cards) Attachments(ctx *data.DBContext, id int, upd []data.File) error {
	if len(upd) == 0 {
		return ctx.DB.
			Model(data.File{}).
			Where("item_id = ?", id).
			Update("item_id", nil).
			Error
	}

	idx := make([]int, len(upd))
	coverID := 0
	for i, u := range upd {
		idx[i] = u.ID
		if u.IsCover {
			coverID = u.ID
		}
	}

	err := ctx.DB.
		Model(data.File{}).
		Where("item_id = ? and id not in (?)", id, idx).
		Update("item_id", nil).
		Error
	if err != nil {
		return err
	}

	err = ctx.DB.
		Model(data.File{}).
		Where("item_id = ? and is_cover = 1", id).
		Update("is_cover", 0).
		Error
	if err != nil {
		return err
	}

	if coverID != 0 {
		err = ctx.DB.
			Model(data.File{}).
			Where("id = ?", coverID).
			Update("is_cover", 1).
			Error
	}
	return err
}
