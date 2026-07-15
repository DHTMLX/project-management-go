package main

import (
	"fmt"
	"log"
	"net/http"
	"path"
	"project-manager-go/api"
	"project-manager-go/data"
	ganttStore "project-manager-go/data/gantt"
	kanbanStore "project-manager-go/data/kanban"
	schedulerStore "project-manager-go/data/scheduler"
	todoStore "project-manager-go/data/todo"
	"project-manager-go/publisher"
	"project-manager-go/service"
	ganttService "project-manager-go/service/gantt"
	"project-manager-go/service/items"
	kanbanService "project-manager-go/service/kanban"
	schedulerService "project-manager-go/service/scheduler"
	todoService "project-manager-go/service/todo"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
	"github.com/jinzhu/configor"
)

// Config is the structure that stores the settings for this backend app
var Config AppConfig

func main() {
	configor.New(&configor.Config{ENVPrefix: "APP", Silent: true}).Load(&Config, "config.yml")

	r := initChi()
	initApp(r)
	initData()

	log.Printf("Starting webserver at port " + Config.Server.Port)
	err := http.ListenAndServe(Config.Server.Port, r)
	if err != nil {
		log.Println(err.Error())
	}
}

func initChi() *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	fmt.Println(Config.Server.Cors)
	if len(Config.Server.Cors) > 0 {
		c := cors.New(cors.Options{
			AllowedOrigins:   Config.Server.Cors,
			AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
			AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token", "Remote-Token", "X-Requested-With"},
			AllowCredentials: true,
			MaxAge:           300,
		})
		r.Use(c.Handler)
	}
	r.Use(api.AuthMiddleware)

	return r
}

func initApp(r *chi.Mux) {
	data.Init(Config.DB)

	serverApi := api.NewServerAPI(r)
	authApi := &api.AuthAPI{}
	publisherApi := publisher.NewPublisherAPI(r)

	itemsAddProcessor := data.NewItemsProcessor()
	itemsDeleteProcessor := data.NewItemsProcessor()
	itemsProcessor := data.NewItemsProcessorProvider()
	itemsProcessor.PushProcessor("add", itemsAddProcessor)
	itemsProcessor.PushProcessor("delete", itemsDeleteProcessor)

	itemsTreeStore := data.NewTreeStore(itemsProcessor)
	treeService := service.NewTreeService(itemsTreeStore)

	projectsProcessor := data.NewProjectsProcessor()
	projectsStore := data.NewProjectsStore(projectsProcessor)

	// init abstract API
	itemsPrefix := "/api/"
	itemsService := items.NewBaseItemsService(treeService, projectsStore, itemsTreeStore)
	itemsApi := api.NewBaseItemsAPI(itemsService, publisherApi, itemsPrefix)

	// init todo API
	todoPrefix := "/api/todo"
	todoStoreObj := todoStore.InitTodoStore(itemsTreeStore, projectsStore)
	todoServiceObj := todoService.NewTodoService(todoStoreObj, treeService)
	todoPublisherApiObj := publisher.NewTodoPublisher(todoStoreObj, r, path.Join(todoPrefix, "v1"), []string{"tasks", "projects"})
	todoApiObj := api.NewTodoAPI(todoServiceObj, todoPublisherApiObj, publisherApi, todoPrefix)

	// init scheduler API
	schedulerPrefix := "/api/scheduler"
	schedulerStoreObj := schedulerStore.InitSchedulerStore(itemsTreeStore, projectsStore)
	schedulerServiceObj := schedulerService.NewSchedulerService(schedulerStoreObj, treeService)
	schedulerPublisherApiObj := publisher.NewSchedulerPublisher(schedulerStoreObj, r, path.Join(schedulerPrefix, "v1"), []string{"events", "calendars"})
	schedulerApiObj := api.NewSchedulerAPI(schedulerServiceObj, schedulerPublisherApiObj, publisherApi, schedulerPrefix)

	// init kanban API
	kanbanPrefix := "/api/kanban"
	kanbanFileManager := kanbanStore.NewFileManager(Config.Binarydata, Config.Server.URL)
	kanbanStoreObj := kanbanStore.InitKanbanStore(itemsTreeStore, projectsStore)
	kanbanServiceObj := kanbanService.NewKanbanService(kanbanStoreObj, treeService)
	kanbanPublisherApiObj := publisher.NewKanbanPublisher(kanbanStoreObj, r, path.Join(kanbanPrefix, "v1"), []string{"cards", "rows", "columns", "links"})
	kanbanApiObj := api.NewKanbanAPI(kanbanServiceObj, kanbanPublisherApiObj, publisherApi, kanbanPrefix, kanbanFileManager)

	// init gantt API
	ganttPrefix := "/api/gantt"
	ganttStoreObj := ganttStore.InitGanttStore(itemsTreeStore, projectsStore)
	ganttServiceObj := ganttService.NewGanttService(ganttStoreObj, treeService)
	ganttPublisherApiObj := publisher.NewGanttPublisher(ganttServiceObj, ganttStoreObj, r, path.Join(ganttPrefix, "v1"), []string{"tasks", "links"})
	ganttApiObj := api.NewGanttAPI(ganttServiceObj, ganttPublisherApiObj, publisherApi, ganttPrefix)

	// add REST API of all widgets
	serverApi.AddAPI(itemsApi, authApi, todoApiObj, schedulerApiObj, kanbanApiObj, ganttApiObj)
	// add WS API of all widgets
	publisherApi.AddAPI(todoPublisherApiObj, schedulerPublisherApiObj, kanbanPublisherApiObj, ganttPublisherApiObj)

	// add Handlers
	itemsAddProcessor.PushHandler(
		itemsTreeStore.HandleItemAddOperation,
		kanbanStoreObj.HandleItemAddOperation,
	)
	itemsDeleteProcessor.PushHandler(
		ganttStoreObj.HandleTaskDeleteOperation,
		kanbanStoreObj.HandleTaskDeleteOperation,
	)
	projectsProcessor.PushHandler(
		projectsStore.HandleProjectAddOperation,
		schedulerStoreObj.HandleProjectAddOperation,
		kanbanStoreObj.HandleProjectAddOperation,
	)
}

func initData() {
	if !Config.DB.Reset {
		return
	}

	dbCtx := data.NewTCtx(nil)
	defer dbCtx.End(nil)

	all := data.NewAppDataProvider(Config.Demodata)
	kanbanDemo := kanbanStore.NewKanbanDataProvider(Config.Demodata)
	ganttDemo := ganttStore.NewGanttDataProvider(Config.Demodata)

	dataProvider := data.NewDemodataProvider()
	dataProvider.Down(dbCtx, all, kanbanDemo, ganttDemo)

	// initialize database with the demo data
	if Config.Demodata != "" {
		dataProvider.Up(dbCtx, all, kanbanDemo, ganttDemo)
	}
}
