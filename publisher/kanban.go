package publisher

import (
	"project-manager-go/common"
	"project-manager-go/data"
	kanbanStore "project-manager-go/data/kanban"
	kanbanService "project-manager-go/service/kanban"

	"github.com/go-chi/chi"
)

type KanbanMove struct {
	ID       int `json:"id"`
	RowID    int `json:"row"`
	ColumnID int `json:"column"`
}

type KanbanEvent struct {
	EventBase
	Card   *kanbanService.Card       `json:"card,omitempty"`
	Row    *kanbanService.Row        `json:"row,omitempty"`
	Column *kanbanStore.KanbanColumn `json:"column,omitempty"`
	Link   *kanbanStore.KanbanLink   `json:"link,omitempty"`
	Before int                       `json:"before"`
}

type KanbanMoveEvent struct {
	EventBase
	Move   *KanbanMove `json:"card"`
	Before int         `json:"before,omitempty"`
}

type KanbanPublisher struct {
	BasePublisher
	store *kanbanStore.KanbanStore
}

func NewKanbanPublisher(store *kanbanStore.KanbanStore, r *chi.Mux, prefix string, routes []string) *KanbanPublisher {
	t := WidgetKanban
	api := newRemoteAPI(r, prefix, routes, t)
	return &KanbanPublisher{
		BasePublisher: BasePublisher{
			widgetType: t,
			api:        api,
		},
		store: store,
	}
}

func (p *KanbanPublisher) AddItem(ctx PublisherContext, id int) (err error) {
	dbCtx := data.NewCtx(nil)

	card := kanbanService.Card{}
	err = p.store.Cards.GetOne(dbCtx, id, &card)
	if err != nil {
		return err
	}
	if card.RowID == 0 {
		return
	}

	if card.ColumnID == 0 {
		cid, err := p.store.Columns.Add(dbCtx, kanbanStore.KanbanColumn{
			Label: "Untitled",
		})
		if err != nil {
			return err
		}
		err = p.AddColumn(ctx, cid)
		if err != nil {
			return err
		}
		card.ColumnID = common.TID(cid)
		err = p.store.Cards.UpdateFields(dbCtx, id, map[string]any{
			"kanban_column_id": cid,
		})
		if err != nil {
			return err
		}
	}

	p.api.Events.Publish(
		"cards",
		&KanbanEvent{
			EventBase: EventBase{
				Widget: ctx.FromWidget,
				From:   ctx.DeviceID,
				Type:   "add-card",
			},
			Card: &card,
		},
	)

	return nil
}

func (p *KanbanPublisher) UpdateItem(ctx PublisherContext, id int) (err error) {
	dbCtx := data.NewCtx(nil)

	card := kanbanService.Card{}
	err = p.store.Cards.GetOne(dbCtx, id, &card)
	if err != nil {
		return err
	}

	p.api.Events.Publish(
		"cards",
		&KanbanEvent{
			EventBase: EventBase{
				Widget: ctx.FromWidget,
				From:   ctx.DeviceID,
				Type:   "update-card",
			},
			Card: &card,
		},
	)

	return nil
}

func (p *KanbanPublisher) DeleteItem(ctx PublisherContext, id int, children []int) error {
	ids := append([]int{id}, children...)
	for i, id := range ids {
		p.api.Events.Publish(
			"cards",
			&KanbanEvent{
				EventBase: EventBase{
					Widget: ctx.FromWidget,
					From:   ctx.DeviceID,
					Type:   "delete-card",
					Self:   i > 0,
				},
				Card: &kanbanService.Card{ID: common.TID(id)},
			},
		)
	}

	return nil
}

func (p *KanbanPublisher) AddProject(ctx PublisherContext, id int) (err error) {
	dbCtx := data.NewCtx(nil)

	c := kanbanService.Row{}
	err = p.store.Rows.GetOne(dbCtx, id, &c)
	if err != nil {
		return err
	}

	p.api.Events.Publish(
		"rows",
		&KanbanEvent{
			EventBase: EventBase{
				Widget: ctx.FromWidget,
				From:   ctx.DeviceID,
				Type:   "add-row",
			},
			Row: &c,
		},
	)
	return nil
}

func (p *KanbanPublisher) UpdateProject(ctx PublisherContext, id int) (err error) {
	dbCtx := data.NewCtx(nil)

	c := kanbanService.Row{}
	err = p.store.Rows.GetOne(dbCtx, id, &c)
	if err != nil {
		return err
	}

	p.api.Events.Publish(
		"rows",
		&KanbanEvent{
			EventBase: EventBase{
				Widget: ctx.FromWidget,
				From:   ctx.DeviceID,
				Type:   "update-row",
			},
			Row: &c,
		},
	)
	return nil
}

func (p *KanbanPublisher) DeleteProject(ctx PublisherContext, id int, children []int) error {
	p.api.Events.Publish(
		"rows",
		&KanbanEvent{
			EventBase: EventBase{
				Widget: ctx.FromWidget,
				From:   ctx.DeviceID,
				Type:   "delete-row",
			},
			Row: &kanbanService.Row{ID: common.TID(id)},
		},
	)
	return nil
}

// Custom methods:

func (p *KanbanPublisher) ChangeProject(ctx PublisherContext, id, projectId int) (err error) {
	dbCtx := data.NewCtx(nil)

	children, err := p.store.Cards.GetAllChildrenIDs(dbCtx, projectId, id)
	if err != nil {
		return err
	}

	ids := append([]int{id}, children...)
	for i, cardId := range ids {
		card := kanbanService.Card{}
		err := p.store.Cards.GetOne(dbCtx, cardId, &card)
		if err != nil {
			return err
		}

		p.api.Events.Publish(
			"cards",
			&KanbanEvent{
				EventBase: EventBase{
					Widget: ctx.FromWidget,
					From:   ctx.DeviceID,
					Type:   "update-card",
					Self:   i > 0,
				},
				Card: &card,
			},
		)
	}

	return nil
}

func (p *KanbanPublisher) AddColumn(ctx PublisherContext, id int) (err error) {
	dbCtx := data.NewCtx(nil)

	column, err := p.store.Columns.GetOne(dbCtx, id)
	if err != nil {
		return err
	}

	p.api.Events.Publish(
		"columns",
		&KanbanEvent{
			EventBase: EventBase{
				Widget: ctx.FromWidget,
				From:   ctx.DeviceID,
				Type:   "add-column",
			},
			Column: &column,
		},
	)

	return nil
}

func (p *KanbanPublisher) UpdateColumn(ctx PublisherContext, id int) (err error) {
	dbCtx := data.NewCtx(nil)

	column, err := p.store.Columns.GetOne(dbCtx, id)
	if err != nil {
		return err
	}

	p.api.Events.Publish(
		"columns",
		&KanbanEvent{
			EventBase: EventBase{
				Widget: ctx.FromWidget,
				From:   ctx.DeviceID,
				Type:   "update-column",
			},
			Column: &column,
		},
	)

	return nil
}

func (p *KanbanPublisher) DeleteColumn(ctx PublisherContext, id int, children []int) error {
	ids := append([]int{id}, children...)
	for i, id := range ids {
		p.api.Events.Publish(
			"columns",
			&KanbanEvent{
				EventBase: EventBase{
					Widget: ctx.FromWidget,
					From:   ctx.DeviceID,
					Type:   "delete-column",
					Self:   i > 0,
				},
				Column: &kanbanStore.KanbanColumn{ID: id},
			},
		)
	}

	return nil
}

func (p *KanbanPublisher) MoveCard(ctx PublisherContext, id, before, row, column int) error {
	p.api.Events.Publish("cards", KanbanMoveEvent{
		EventBase: EventBase{
			Widget: ctx.FromWidget,
			Type:   "move-card",
			From:   ctx.DeviceID,
		},
		Move: &KanbanMove{
			ID:       id,
			ColumnID: column,
			RowID:    row,
		},
		Before: before,
	})

	return nil
}

func (p *KanbanPublisher) MoveRow(ctx PublisherContext, id, before int) error {
	p.api.Events.Publish("rows", KanbanEvent{
		EventBase: EventBase{
			Widget: ctx.FromWidget,
			Type:   "move-row",
			From:   ctx.DeviceID,
		},
		Row: &kanbanService.Row{
			ID: common.TID(id),
		},
		Before: before,
	})

	return nil
}

func (p *KanbanPublisher) MoveColumn(ctx PublisherContext, id, before int) error {
	p.api.Events.Publish("columns", KanbanEvent{
		EventBase: EventBase{
			Widget: ctx.FromWidget,
			Type:   "move-column",
			From:   ctx.DeviceID,
		},
		Column: &kanbanStore.KanbanColumn{
			ID: id,
		},
		Before: before,
	})

	return nil
}

func (p *KanbanPublisher) AddLink(ctx PublisherContext, id int) error {
	dbCtx := data.NewCtx(nil)

	link, err := p.store.Links.GetOne(dbCtx, id)
	if err != nil {
		return err
	}

	p.api.Events.Publish(
		"links",
		&KanbanEvent{
			EventBase: EventBase{
				From:   ctx.DeviceID,
				Type:   "add-link",
				Widget: ctx.FromWidget,
			},
			Link: &link,
		},
	)

	return nil
}
func (p *KanbanPublisher) UpdateLink(ctx PublisherContext, id int) error {
	dbCtx := data.NewCtx(nil)

	link, err := p.store.Links.GetOne(dbCtx, id)
	if err != nil {
		return err
	}

	p.api.Events.Publish(
		"links",
		&KanbanEvent{
			EventBase: EventBase{
				From:   ctx.DeviceID,
				Type:   "update-link",
				Widget: ctx.FromWidget,
			},
			Link: &link,
		},
	)

	return nil
}

func (p *KanbanPublisher) DeleteLink(ctx PublisherContext, id int) error {
	p.api.Events.Publish(
		"links",
		&KanbanEvent{
			EventBase: EventBase{
				From:   ctx.DeviceID,
				Type:   "delete-link",
				Widget: ctx.FromWidget,
			},
			Link: &kanbanStore.KanbanLink{
				ID: id,
			},
		},
	)

	return nil
}
