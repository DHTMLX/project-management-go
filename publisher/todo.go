package publisher

import (
	"project-manager-go/common"
	"project-manager-go/data"
	todoStore "project-manager-go/data/todo"
	todoService "project-manager-go/service/todo"

	"github.com/go-chi/chi"
)

type TodoEvent struct {
	EventBase
	ID   int `json:"id"`
	Data any `json:"data,omitempty"`
}

type TodoParams struct {
	ID        int  `json:"id"`
	ParentID  int  `json:"parent,omitempty"`
	TargetID  int  `json:"targetId,omitempty"`
	ProjectID int  `json:"project,omitempty"`
	Reverse   bool `json:"reverse,omitempty"`
}

type TodoPublisher struct {
	BasePublisher
	store *todoStore.TodoStore
}

func NewTodoPublisher(store *todoStore.TodoStore, r *chi.Mux, prefix string, routes []string) *TodoPublisher {
	t := WidgetTodo
	api := newRemoteAPI(r, prefix, routes, t)
	return &TodoPublisher{
		BasePublisher: BasePublisher{
			widgetType: t,
			api:        api,
		},
		store: store,
	}
}

func (p *TodoPublisher) AddItem(ctx PublisherContext, id int) (err error) {
	dbCtx := data.NewCtx(nil)

	task := todoService.Task{}
	err = p.store.Tasks.GetOne(dbCtx, id, &task)
	if err != nil {
		return err
	}

	prevTask := todoService.Task{}
	err = p.store.Tasks.GetByIndex(dbCtx, int(task.ProjectID), int(task.ParentID), task.Index-1, &prevTask)
	if err != nil {
		return err
	}

	p.api.Events.Publish(
		"tasks",
		&TodoEvent{
			EventBase: EventBase{
				Widget: ctx.FromWidget,
				From:   ctx.DeviceID,
				Type:   "add-task",
			},
			ID: id,
			Data: todoService.AddTask{
				Task: task,
				Meta: todoService.Meta{
					TargetID:  prevTask.ID,
					ParentID:  task.ParentID,
					ProjectID: task.ProjectID,
					Reverse:   task.Index == 0 && task.ParentID == 0,
				},
			},
		},
	)

	return nil
}

func (p *TodoPublisher) UpdateItem(ctx PublisherContext, id int) (err error) {
	dbCtx := data.NewCtx(nil)

	task := todoService.Task{}
	err = p.store.Tasks.GetOne(dbCtx, id, &task)
	if err != nil {
		return err
	}

	p.api.Events.Publish(
		"tasks",
		&TodoEvent{
			EventBase: EventBase{
				Widget: ctx.FromWidget,
				From:   ctx.DeviceID,
				Type:   "update-task",
			},
			ID:   id,
			Data: &task,
		},
	)

	return nil
}

func (p *TodoPublisher) DeleteItem(ctx PublisherContext, id int, children []int) error {
	p.api.Events.Publish(
		"tasks",
		&TodoEvent{
			EventBase: EventBase{
				Widget: ctx.FromWidget,
				From:   ctx.DeviceID,
				Type:   "delete-task",
			},
			ID: id,
		},
	)
	return nil
}

func (p *TodoPublisher) AddProject(ctx PublisherContext, id int) (err error) {
	dbCtx := data.NewCtx(nil)

	project := todoService.Project{}
	err = p.store.Projects.GetOne(dbCtx, id, &project)
	if err != nil {
		return err
	}

	p.api.Events.Publish(
		"projects",
		&TodoEvent{
			EventBase: EventBase{
				Widget: ctx.FromWidget,
				From:   ctx.DeviceID,
				Type:   "add-project",
			},
			ID:   id,
			Data: &project,
		},
	)

	return nil
}

func (p *TodoPublisher) UpdateProject(ctx PublisherContext, id int) (err error) {
	dbCtx := data.NewCtx(nil)

	project := todoService.Project{}
	err = p.store.Projects.GetOne(dbCtx, id, &project)
	if err != nil {
		return err
	}

	p.api.Events.Publish(
		"projects",
		&TodoEvent{
			EventBase: EventBase{
				Widget: ctx.FromWidget,
				From:   ctx.DeviceID,
				Type:   "update-project",
			},
			ID:   id,
			Data: &project,
		},
	)

	return nil
}

func (p *TodoPublisher) DeleteProject(ctx PublisherContext, id int, children []int) error {
	p.api.Events.Publish(
		"projects",
		&TodoEvent{
			EventBase: EventBase{
				Widget: ctx.FromWidget,
				From:   ctx.DeviceID,
				Type:   "delete-project",
			},
			Data: todoService.Project{ID: common.TID(id)},
		},
	)

	return nil
}

func (p *TodoPublisher) ChangeProject(ctx PublisherContext, id int, projectId int) error {
	params := MoveParams{
		ID:        id,
		ProjectID: projectId,
	}
	return p.Move(ctx, id, params)
}

func (p *TodoPublisher) Move(ctx PublisherContext, id int, params MoveParams) error {
	params.ID = id
	p.api.Events.Publish(
		"tasks",
		&TodoEvent{
			EventBase: EventBase{
				Widget: ctx.FromWidget,
				From:   ctx.DeviceID,
				Type:   "move-task",
			},
			Data: &params,
			ID:   id,
		},
	)
	return nil
}

func (p *TodoPublisher) CloneTask(ctx PublisherContext, params todoService.CloneTask) error {
	p.api.Events.Publish(
		"tasks",
		&TodoEvent{
			EventBase: EventBase{
				Widget: ctx.FromWidget,
				From:   ctx.DeviceID,
				Type:   "clone-task",
			},
			Data: &params,
		},
	)
	return nil
}
