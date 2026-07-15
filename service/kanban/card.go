package kanban

import (
	"project-manager-go/common"
	"project-manager-go/data"
	"project-manager-go/service"
	"time"
)

type Card struct {
	ID          common.TID      `json:"id"`
	ColumnID    common.TID      `json:"column"`
	RowID       common.TID      `json:"row"`
	Label       string          `json:"label"`
	Description string          `json:"description"`
	Color       string          `json:"color"`
	StartDate   *time.Time      `json:"start_date"`
	EndDate     *time.Time      `json:"end_date"`
	Progress    common.FuzzyInt `json:"progress"`
	Priority    common.FuzzyInt `json:"priority"`
	Users       []int           `json:"users"`
	Attached    []data.File     `json:"attached"`
	Comments    []data.Comment  `json:"comments"`
	Votes       []int           `json:"votes"`

	Index int `json:"-"`
}

func (c Card) FillItem(item *data.Item) {
	item.ProjectID = int(c.RowID)
	item.Kanban_ColumnID = int(c.ColumnID)
	item.Text = c.Label
	item.Description = c.Description
	item.Color = c.Color
	item.Priority = int(c.Priority)
	item.Progress = float32(c.Progress) / 100.0
	item.StartDate = c.StartDate
	item.Attached = c.Attached
	item.EndDate = c.EndDate
	item.Kanban_CardIndex = c.Index
	item.Votes = service.IDsToUsers(c.Votes)
	item.AssignedUsers = service.IDsToUsers(c.Users)
}

func (c *Card) PutItem(item data.Item) {
	c.ID = common.TID(item.ID)
	c.RowID = common.TID(item.ProjectID)
	c.ColumnID = common.TID(item.Kanban_ColumnID)
	c.Label = item.Text
	c.Description = item.Description
	c.Color = item.Color
	c.Priority = common.FuzzyInt(item.Priority)
	c.Progress = common.FuzzyInt(item.Progress * 100)
	c.StartDate = item.StartDate
	c.EndDate = item.EndDate
	c.Attached = item.Attached
	c.Comments = item.Comments
	c.Index = item.Kanban_CardIndex
	c.Votes = service.UsersToIDs[int](item.Votes)
	c.Users = service.UsersToIDs[int](item.AssignedUsers)
}
