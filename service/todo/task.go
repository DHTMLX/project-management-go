package todo

import (
	"project-manager-go/common"
	"project-manager-go/data"
	"project-manager-go/service"
	"time"
)

type Task struct {
	ID             common.TID        `json:"id"`
	ParentID       common.TID        `json:"parent,omitempty"`
	ProjectID      common.TID        `json:"project,omitempty"`
	Text           string            `json:"text"`
	Checked        bool              `json:"checked"`
	Users          []common.FuzzyInt `json:"assigned"`
	DueDate        *time.Time        `json:"due_date,omitempty"`
	CompletionDate *time.Time        `json:"completion_date,omitempty"`
	CreationDate   *time.Time        `json:"creation_date,omitempty"`
	EditedDate     *time.Time        `json:"edited_date,omitempty"`
	Priority       common.FuzzyInt   `json:"priority"`

	Index int `json:"-"`
}

type TempTask struct {
	Task
	ID       string        `json:"id"`
	ParentID common.TempID `json:"parent"`
}

type Meta struct {
	ProjectID common.TID `json:"project"`
	ParentID  common.TID `json:"parent,omitempty"`
	TargetID  common.TID `json:"targetId"`
	Reverse   bool       `json:"reverse"`
}

type AddTask struct {
	Meta
	Task Task `json:"task"`
}

type UpdateTask struct {
	Task
	Batch []Task `json:"batch"`
}

type MoveTask struct {
	Meta
	ID        common.TID `json:"id"`
	IDs       []int      `json:"batch,omitempty"`
	Operation string     `json:"operation"`
}

type CloneTask struct {
	Meta
	Batch []TempTask `json:"batch"`
}

func (t *Task) PutItem(item data.Item) {
	t.ID = common.TID(item.ID)
	t.ParentID = common.TID(item.ParentID)
	t.ProjectID = common.TID(item.ProjectID)
	t.Text = item.Text
	t.Checked = item.Checked
	t.DueDate = item.EndDate
	t.CompletionDate = item.CompletionDate
	t.EditedDate = item.EditedDate
	t.CreationDate = item.CreationDate
	t.Priority = common.FuzzyInt(item.Priority)
	t.Index = item.Index
	t.Users = service.UsersToIDs[common.FuzzyInt](item.AssignedUsers)
}

func (t Task) FillItem(item *data.Item) {
	item.ParentID = int(t.ParentID)
	item.ProjectID = int(t.ProjectID)
	item.Text = t.Text
	item.Checked = t.Checked
	item.EndDate = t.DueDate
	item.CompletionDate = t.CompletionDate
	item.EditedDate = t.EditedDate
	item.CreationDate = t.CreationDate
	item.Priority = int(t.Priority)
	item.Index = t.Index
	item.AssignedUsers = service.IDsToUsers(t.Users)
}
