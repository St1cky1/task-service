package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/St1cky1/task-service/internal/models"
	"github.com/St1cky1/task-service/internal/service"
	"github.com/go-chi/chi/v5"
)

type TaskHandler struct {
	taskService *service.TaskService
}

func NewTaskHandler(taskService *service.TaskService) *TaskHandler {
	return &TaskHandler{
		taskService: taskService,
	}
}

// создаем новую задачу
func (h *TaskHandler) CreateTask(w http.ResponseWriter, r *http.Request) {

	var req models.CreateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil { // распарсиваем ответ в структурку
		http.Error(w, "Invalid JSON", http.StatusBadRequest) // если ошибка, 400 выводим
		return
	}

	userId := 1

	task, err := h.taskService.CreateTask(r.Context(), &req, userId)
	if err != nil {
		switch err {

		case models.ErrUserNotFound:
			http.Error(w, "user not found", http.StatusNotFound) // 404

		case models.ErrInvalidTaskData:
			http.Error(w, "invalid task data", http.StatusBadRequest) // 400
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError) // 500
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) GetTask(w http.ResponseWriter, r *http.Request) {
	taskIdStr := chi.URLParam(r, "id") // id задачи
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		http.Error(w, "Invalid task Id", http.StatusBadRequest)
	}

	userId := 1

	task, err := h.taskService.GetTask(r.Context(), taskId, userId)

	if err != nil {
		switch err {
		case models.ErrTaskNotFound:
			http.Error(w, "tack not found", http.StatusNotFound)
		case models.ErrForbidden:
			http.Error(w, "Access denied", http.StatusForbidden)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Context-Type", "aplication/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	taskIdStr := chi.URLParam(r, "id")
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		http.Error(w, "Invalid task Id", http.StatusBadRequest)
		return
	}

	var req models.UpdateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	userId := 1

	task, err := h.taskService.UpdateTask(r.Context(), taskId, userId, &req)
	if err != nil {
		switch err {
		case models.ErrTaskNotFound:
			http.Error(w, "task not found", http.StatusNotFound)
		case models.ErrNoFieldsToUpdate:
			http.Error(w, "no fields to update", http.StatusBadRequest)
		case models.ErrForbidden:
			http.Error(w, "access denied", http.StatusForbidden)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Context-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

func (h *TaskHandler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	taskIdStr := chi.URLParam(r, "id")
	taskId, err := strconv.Atoi(taskIdStr)
	if err != nil {
		http.Error(w, "Invalid task Id", http.StatusBadRequest)
	}

	userId := 1

	err = h.taskService.DeleteTask(r.Context(), taskId, userId)
	if err != nil {
		switch err {
		case models.ErrTaskNotFound:
			http.Error(w, "task not found", http.StatusNotFound)
		case models.ErrForbidden:
			http.Error(w, "Access denied", http.StatusForbidden)
		default:
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *TaskHandler) ListTasks(w http.ResponseWriter, r *http.Request) {
	userId := 1

	status := r.URL.Query().Get("status")

	task, err := h.taskService.ListTasks(r.Context(), userId, status)
	if err != nil {
		http.Error(w, "Internal server", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}
