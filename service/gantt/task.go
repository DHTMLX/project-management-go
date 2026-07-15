package gantt

import (
	"project-manager-go/common"
	"project-manager-go/data"
	"time"
)

type Task struct {
	ID        int             `json:"id"`
	Type      string          `json:"type"`
	ParentID  common.FuzzyInt `json:"parent"`
	Text      string          `json:"text"`
	Duration  int             `json:"duration"`
	Progress  float32         `json:"progress"`
	StartDate *common.JDate   `json:"start_date"`

	Index int `json:"-"`
}

func (t *Task) PutItem(item data.Item) {
	t.ID = item.ID
	t.Type = item.GanttTaskType
	t.ParentID = common.FuzzyInt(item.ParentID)
	t.Text = item.Text
	t.Progress = item.Progress
	t.StartDate = (*common.JDate)(item.StartDate)
	t.Index = item.Index

	if item.EndDate != nil && item.StartDate != nil {
		duration := item.EndDate.Sub(*item.StartDate) / time.Hour / 24
		t.Duration = int(duration)
	} else {
		item.StartDate = nil
		t.StartDate = nil
		t.Duration = 0
	}
}

func (t Task) FillItem(item *data.Item) {
	item.GanttTaskType = t.Type
	item.ParentID = int(t.ParentID)
	item.Text = t.Text
	item.Progress = t.Progress
	item.StartDate = (*time.Time)(t.StartDate)
	item.Index = t.Index

	if !t.StartDate.IsEmpty() && t.Duration > 0 {
		calculatedEndDate := item.StartDate.Add(time.Hour * 24 * time.Duration(t.Duration))
		item.EndDate = &calculatedEndDate
	} else {
		item.StartDate = nil
	}
}
