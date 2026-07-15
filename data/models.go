package data

import (
	"project-manager-go/data/gantt/gantt_props"
	"project-manager-go/data/kanban/kanban_props"
	"project-manager-go/data/scheduler/scheduler_props"
	"project-manager-go/data/todo/todo_props"
	"time"
)

// item is the most common model that contains all properties for all widgets
type Item struct {
	todo_props.TodoItemProps
	kanban_props.KanbanItemProps
	gantt_props.GanttItemProps
	scheduler_props.SchedulerItemProps

	ID             int        `json:"id"`
	ProjectID      int        `json:"project"`
	ParentID       int        `json:"parent"`
	Priority       int        `json:"priority"`
	Progress       float32    `json:"progress"`
	Text           string     `json:"text" gorm:"default:Untitled"`
	Description    string     `json:"description"`
	Color          string     `json:"color"`
	Checked        bool       `json:"checked"`
	AllDay         bool       `json:"allday"`
	StartDate      *time.Time `json:"start_date"`
	EndDate        *time.Time `json:"end_date"`
	CreationDate   *time.Time `json:"creation_date"`
	EditedDate     *time.Time `json:"edited_date"`
	CompletionDate *time.Time `json:"completion_date"`
	Attached       []File     `gorm:"foreignKey:ItemID" json:"attached"`
	Comments       []Comment  `gorm:"foreignKey:ItemID" json:"comments"`
	Votes          []User     `gorm:"many2many:vote_user;" json:"votes"`

	AssignedUsers []User `gorm:"many2many:item_user;" json:"assigned_users"`

	Index int `json:"index"`
}

type Project struct {
	todo_props.TodoProjectProps
	kanban_props.KanbanProjectProps
	gantt_props.GanttProjectProps
	scheduler_props.SchedulerProjectProps

	ID    int    `json:"id"`
	Label string `json:"label" gorm:"default:Untitled"`
	Index int    `json:"index"`
}

type User struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`

	AssignedItems []Item `gorm:"many2many:item_user" json:"-"`
}

type ItemUser struct {
	ItemID int `gorm:"column:item_id;primaryKey"`
	UserID int `gorm:"column:user_id;primaryKey"`
}

type VoteUser struct {
	ItemID int `gorm:"column:item_id;primaryKey"`
	UserID int `gorm:"column:user_id;primaryKey"`
}

func (VoteUser) TableName() string {
	return "vote_user"
}

func (ItemUser) TableName() string {
	return "item_user"
}

type File struct {
	ID      int    `gorm:"primaryKey" json:"id"`
	Name    string `json:"name"`
	URL     string `json:"url"`
	IsCover bool   `json:"isCover"`
	Path    string `json:"-"`

	ItemID int `json:"-"`
}

type Comment struct {
	ID     int    `json:"id"`
	Text   string `json:"text"`
	Date   string `json:"date"`
	UserID int    `json:"userId"`
	ItemID int    `json:"cardId"`
}
