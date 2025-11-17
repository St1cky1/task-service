package api

import (
	"github.com/St1cky1/task-service/internal/api/handlers"
	"github.com/St1cky1/task-service/internal/usecase"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(taskService *usecase.TaskService) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)

	taskHandler := handlers.NewTaskHandler(taskService)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/tasks", func(r chi.Router) {
			r.Get("/", taskHandler.ListTasks)
			r.Post("/", taskHandler.CreateTask)
			r.Route("/{id}", func(r chi.Router) {
				r.Get("/", taskHandler.GetTask)
				r.Put("/", taskHandler.UpdateTask)
				r.Delete("/", taskHandler.DeleteTask)
			})

		})
	})

	return r
}
