package kanban

import (
	"errors"
	"fmt"
	uCtx "project-manager-go/api/context"
	"project-manager-go/common"
	"project-manager-go/data"
	"project-manager-go/data/kanban"
	"project-manager-go/service"
)

type MoveCards struct {
	MoveParams
	Batch []MoveParams `json:"batch"`
}

type MoveCardResponse struct {
	// contains the cards ids where the project id has been changed
	UpdatedProject []int
	// contains the cards ids where the status has been changed
	UpdatedStatus []int
}

type cards struct {
	tree  *service.TreeService
	store *kanban.KanbanStore
}

func (s *cards) GetAll(userCtx uCtx.UserContext, dbCtx *data.DBContext) (arr []Card, err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	list := data.NewItemsList[Card]()
	err = s.store.Cards.GetAll(dbCtx, list)

	return list.GetArray(), err
}

func (s *cards) Add(userCtx uCtx.UserContext, dbCtx *data.DBContext, card Card) (id int, err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	index, err := s.store.Cards.GetMaxIndex(dbCtx, int(card.RowID), int(card.ColumnID))
	if err != nil {
		return 0, err
	}
	card.Index = index

	id, err = s.store.Cards.Add(dbCtx, &card)

	proj := int(card.RowID)
	maxIndex, err := s.store.Cards.MaxBranchIndex(dbCtx, proj, 0)
	if err != nil {
		return 0, err
	}
	err = s.tree.Move(dbCtx, id, proj, 0, maxIndex+1)

	return id, err
}

func (s *cards) Update(userCtx uCtx.UserContext, dbCtx *data.DBContext, id int, upd Card) (err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	card := Card{}
	err = s.store.Cards.GetOne(dbCtx, id, &card)
	if err != nil {
		return err
	}
	upd.Index = card.Index

	err = s.store.Cards.DeleteAssociations(dbCtx, id)
	if err != nil {
		return err
	}

	err = s.store.Cards.Update(dbCtx, id, &upd)
	if err != nil {
		return err
	}

	err = s.store.Cards.Attachments(dbCtx, id, upd.Attached)
	return err
}

func (s *cards) Delete(userCtx uCtx.UserContext, dbCtx *data.DBContext, id int) (children []int, err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	children, err = s.store.Cards.DeleteCascade(dbCtx, id)

	return children, err
}

func (s *cards) Move(userCtx uCtx.UserContext, dbCtx *data.DBContext, id int, batch []MoveParams) (res MoveCardResponse, err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	res = MoveCardResponse{
		UpdatedProject: make([]int, 0),
		UpdatedStatus:  make([]int, 0),
	}

	if len(batch) == 0 {
		return res, errors.New("invalid move params")
	}

	moveIDs := make([]int, len(batch))
	for i := range batch {
		moveIDs[i] = int(batch[i].ID)
	}

	before := int(batch[0].Before)
	row := int(batch[0].RowID)
	column := int(batch[0].ColumnID)
	index := 0

	if before == 0 {
		// move cards to the end
		index, err = s.store.Cards.GetMaxIndex(dbCtx, row, column)
		if err != nil {
			return res, err
		}
	} else {
		cardBefore := Card{}
		err := s.store.Cards.GetOne(dbCtx, before, &cardBefore)
		if err != nil {
			return res, err
		}
		index = cardBefore.Index

		// shift index to insert moved cards
		err = s.store.Cards.ShiftIndex(dbCtx, index, len(batch))
		if err != nil {
			return res, err
		}
	}

	for _, id := range moveIDs {
		// update card index
		card := Card{}
		err := s.store.Cards.GetOne(dbCtx, id, &card)
		if err != nil {
			return res, err
		}

		if int(card.RowID) != row {
			// project changed
			res.UpdatedProject = append(res.UpdatedProject, id)

			// contains cached cards index for the "row:column" key
			cache := make(map[string]int)
			s.tree.ChangeNodeProject(dbCtx, id, row, func(ctx *data.DBContext, childId int, fields map[string]any) error {
				ctx = data.NewTCtx(ctx)
				defer func() { err = ctx.End(err) }()

				child := Card{}
				err := s.store.Cards.GetOne(dbCtx, id, &child)
				if err != nil {
					return err
				}

				key := fmt.Sprintf("%d:%d", child.RowID, child.ColumnID)

				// find the index of the card in context of row/column
				index, ok := cache[key]
				if !ok {
					index, err = s.store.Cards.GetMaxIndex(dbCtx, row, column)
					if err != nil {
						return err
					}
				} else {
					index++
				}
				cache[key] = index

				fields["kanban_index"] = index

				return nil
			})
		}

		if int(card.ColumnID) != column {
			// status changed
			res.UpdatedStatus = append(res.UpdatedStatus, id)
		}

		card.Index = index
		card.RowID = common.TID(row)
		card.ColumnID = common.TID(column)

		err = s.store.Cards.Update(dbCtx, id, &card)
		if err != nil {
			return res, err
		}

		index++
	}

	return res, err
}

func (s *cards) Vote(userCtx uCtx.UserContext, dbCtx *data.DBContext, id int, vote bool) (err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	card := Card{}
	err = s.store.Cards.GetOne(dbCtx, id, &card)
	if err != nil {
		return err
	}
	if vote {
		card.Votes = append(card.Votes, userCtx.ID)
	} else {
		var voteUser data.VoteUser
		err = s.store.Cards.DeleteVote(dbCtx, id, userCtx.ID, voteUser)
		if err != nil {
			return err
		}
		card.Votes = remove(card.Votes, userCtx.ID)
	}

	err = s.store.Cards.Update(dbCtx, id, &card)
	return err
}

func (s *cards) AddComment(userCtx uCtx.UserContext, dbCtx *data.DBContext, id int, comment kanban.CommentInput) (commentID int, err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	commentID, err = s.store.Cards.AddComment(dbCtx, userCtx.ID, id, comment)
	return
}

func (s *cards) DeleteComment(userCtx uCtx.UserContext, dbCtx *data.DBContext, id int) (err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	err = s.store.Cards.DeleteComment(dbCtx, id)
	if err != nil {
		return err
	}

	return nil
}

func (s *cards) UpdateComment(userCtx uCtx.UserContext, dbCtx *data.DBContext, id int, comment kanban.CommentInput) (err error) {
	dbCtx = data.NewTCtx(dbCtx)
	defer func() { err = dbCtx.End(err) }()

	err = s.store.Cards.UpdateComment(dbCtx, id, comment)
	if err != nil {
		return err
	}

	return nil
}

func remove(slice []int, s int) []int {
	for i, v := range slice {
		if v == s {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
