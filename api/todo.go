package api

import (
	"net/http"
	"project-manager-go/publisher"
	"project-manager-go/service/todo"

	"github.com/go-chi/chi"
)

type TodoAPI struct {
	BaseAPI
	service *todo.TodoService
	pubTodo *publisher.TodoPublisher
	pubAll  *publisher.PublisherAPI
}

func NewTodoAPI(service *todo.TodoService, pub *publisher.TodoPublisher, pubAll *publisher.PublisherAPI, prefix string) *TodoAPI {
	return &TodoAPI{
		BaseAPI: BaseAPI{
			prefix: prefix,
		},
		service: service,
		pubTodo: pub,
		pubAll:  pubAll,
	}
}

func (api *TodoAPI) SetAPI(r *chi.Mux) {
	r.Get(api.route("/tasks"), func(w http.ResponseWriter, r *http.Request) {
		userCtx, err := parseUserContext(r)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		tasks, err := api.service.Tasks.GetAll(userCtx, nil)

		respond(w, &tasks, err)
	})

	r.Get(api.route("/tasks/projects/{id}"), func(w http.ResponseWriter, r *http.Request) {
		userCtx, err := parseUserContext(r)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		id := parseNumberParam(r, "id")
		tasks, err := api.service.Tasks.GetByProject(userCtx, nil, id)

		respond(w, &tasks, err)
	})

	r.Post(api.route("/tasks"), func(w http.ResponseWriter, r *http.Request) {
		userCtx, err := parseUserContext(r)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		task := todo.AddTask{}
		err = parseForm(w, r, &task)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		id, err := api.service.Tasks.Add(userCtx, nil, task)

		if respond(w, &ResponseID{id}, err) {
			pubCtx := publisher.PublisherContext{
				UserID:     userCtx.ID,
				DeviceID:   userCtx.DeviceID,
				FromWidget: publisher.WidgetTodo,
			}
			api.pubAll.PublishAddItem(pubCtx, id)
		}
	})

	r.Put(api.route("/tasks/{id}"), func(w http.ResponseWriter, r *http.Request) {
		userCtx, err := parseUserContext(r)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		task := todo.UpdateTask{}
		err = parseForm(w, r, &task)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		id := parseNumberParam(r, "id")

		err = api.service.Tasks.Update(userCtx, nil, id, task)

		if respond(w, &ResponseID{id}, err) {
			pubCtx := publisher.PublisherContext{
				UserID:     userCtx.ID,
				DeviceID:   userCtx.DeviceID,
				FromWidget: publisher.WidgetTodo,
			}
			api.pubAll.PublishUpdateItem(pubCtx, id)
		}
	})

	r.Delete(api.route("/tasks/{id}"), func(w http.ResponseWriter, r *http.Request) {
		userCtx, err := parseUserContext(r)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		id := parseNumberParam(r, "id")
		children, err := api.service.Tasks.Delete(userCtx, nil, id)

		if respond(w, &ResponseID{id}, err) {
			pubCtx := publisher.PublisherContext{
				UserID:     userCtx.ID,
				DeviceID:   userCtx.DeviceID,
				FromWidget: publisher.WidgetTodo,
			}
			api.pubAll.PublishDeleteItem(pubCtx, id, children)
		}
	})

	r.Delete(api.route("/tasks"), func(w http.ResponseWriter, r *http.Request) {
		userCtx, err := parseUserContext(r)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		batch := struct {
			Batch []int `json:"batch"`
		}{}
		err = parseForm(w, r, &batch)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		for _, id := range batch.Batch {
			children, err := api.service.Tasks.Delete(userCtx, nil, id)
			if err != nil {
				respondWithError(w, err.Error())
				return
			}

			pubCtx := publisher.PublisherContext{
				UserID:     userCtx.ID,
				DeviceID:   userCtx.DeviceID,
				FromWidget: publisher.WidgetTodo,
			}
			api.pubAll.PublishDeleteItem(pubCtx, id, children)
		}

		respond(w, &ResponseID{}, nil)
	})

	r.Post(api.route("/clone"), func(w http.ResponseWriter, r *http.Request) {
		userCtx, err := parseUserContext(r)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		info := todo.CloneTask{}
		err = parseForm(w, r, &info)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		pull, err := api.service.Tasks.Clone(userCtx, nil, &info)

		if respond(w, &pull, err) {
			pubCtx := publisher.PublisherContext{
				UserID:     userCtx.ID,
				DeviceID:   userCtx.DeviceID,
				FromWidget: publisher.WidgetTodo,
			}

			for tempId := range pull {
				api.pubAll.PublishAddItem(pubCtx, pull[tempId])
			}
		}
	})

	r.Put(api.route("/move/{id}"), func(w http.ResponseWriter, r *http.Request) {
		userCtx, err := parseUserContext(r)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		task := todo.MoveTask{}
		err = parseForm(w, r, &task)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		id := parseNumberParam(r, "id")
		moveId, err := api.service.Tasks.Move(userCtx, nil, id, task)

		if respond(w, &ResponseID{id}, err) {
			pubCtx := publisher.PublisherContext{
				UserID:     userCtx.ID,
				DeviceID:   userCtx.DeviceID,
				FromWidget: publisher.WidgetTodo,
			}

			params := publisher.MoveParams{
				TargetID:  int(task.TargetID),
				ParentID:  int(task.ParentID),
				ProjectID: int(task.ProjectID),
				Reverse:   task.Reverse,
			}
			api.pubAll.PublishMove(pubCtx, moveId, params)

			switch task.Operation {
			case "project":
				// publish for all clients
				api.pubAll.PublishChangeProject(pubCtx, moveId, int(task.ProjectID))
			}
		}
	})

	r.Get(api.route("/projects"), func(w http.ResponseWriter, r *http.Request) {
		userCtx, err := parseUserContext(r)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		projects, err := api.service.Projects.GetAll(userCtx, nil)

		respond(w, &projects, err)
	})

	r.Post(api.route("/projects"), func(w http.ResponseWriter, r *http.Request) {
		userCtx, err := parseUserContext(r)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		project := todo.Project{}
		err = parseForm(w, r, &project)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		id, err := api.service.Projects.Add(userCtx, nil, project)

		if respond(w, ResponseID{id}, err) {
			pubCtx := publisher.PublisherContext{
				UserID:     userCtx.ID,
				DeviceID:   userCtx.DeviceID,
				FromWidget: publisher.WidgetTodo,
			}
			api.pubAll.PublishAddProject(pubCtx, id)
		}
	})

	r.Put(api.route("/projects/{id}"), func(w http.ResponseWriter, r *http.Request) {
		userCtx, err := parseUserContext(r)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		project := todo.Project{}
		err = parseForm(w, r, &project)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		id := parseNumberParam(r, "id")
		err = api.service.Projects.Update(userCtx, nil, id, project)

		if respond(w, &ResponseID{id}, err) {
			pubCtx := publisher.PublisherContext{
				UserID:     userCtx.ID,
				DeviceID:   userCtx.DeviceID,
				FromWidget: publisher.WidgetTodo,
			}
			api.pubAll.PublishUpdateProject(pubCtx, id)
		}
	})

	r.Delete(api.route("/projects/{id}"), func(w http.ResponseWriter, r *http.Request) {
		userCtx, err := parseUserContext(r)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		id := parseNumberParam(r, "id")
		children, err := api.service.Projects.Delete(userCtx, nil, id)

		if respond(w, &ResponseID{id}, err) {
			pubCtx := publisher.PublisherContext{
				UserID:     userCtx.ID,
				DeviceID:   userCtx.DeviceID,
				FromWidget: publisher.WidgetTodo,
			}
			api.pubAll.PublishDeleteProject(pubCtx, id, children)
		}
	})

	r.Get(api.route("/users"), func(w http.ResponseWriter, r *http.Request) {
		userCtx, err := parseUserContext(r)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		users, err := api.service.Users.GetAll(userCtx, nil)

		respond(w, &users, err)
	})

	r.Get(api.route("/tags"), func(w http.ResponseWriter, r *http.Request) {
		userCtx, err := parseUserContext(r)
		if err != nil {
			respondWithError(w, err.Error())
			return
		}

		tags, err := api.service.Tags.GetAll(userCtx, nil)

		respond(w, &tags, err)
	})
}
