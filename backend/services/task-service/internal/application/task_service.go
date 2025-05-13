// file: backend/services/task-service/internal/application/task_service.go
package application

import (
	"context"
	"errors"
	"time"

	"github.com/TubagusAldiMY/go-vue-todolist/backend/services/task-service/internal/domain" // Sesuaikan dengan path module Anda
)

// TaskInput adalah struct untuk data input pembuatan atau pembaruan task.
// Kita bisa menggunakan DTO (Data Transfer Object) yang lebih spesifik nanti jika diperlukan,
// terutama jika input dari API berbeda signifikan dengan struktur domain.
type CreateTaskInput struct {
	Title       string
	Description string
}

type UpdateTaskInput struct {
	Title       *string // Pointer untuk menandakan field mana yang ingin diupdate
	Description *string
	Completed   *bool
}

// TaskApplicationService mendefinisikan interface untuk service aplikasi Task.
// Ini adalah kontrak untuk use cases yang berhubungan dengan Task.
type TaskApplicationService interface {
	CreateTask(ctx context.Context, userID domain.UserID, input CreateTaskInput) (*domain.Task, error)
	GetTaskByID(ctx context.Context, userID domain.UserID, taskID string) (*domain.Task, error)
	GetTasksByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Task, error)
	UpdateTask(ctx context.Context, userID domain.UserID, taskID string, input UpdateTaskInput) (*domain.Task, error)
	DeleteTask(ctx context.Context, userID domain.UserID, taskID string) error
}

// taskService adalah implementasi dari TaskApplicationService.
type taskService struct {
	taskRepo domain.TaskRepository // Dependensi ke TaskRepository dari domain layer
}

// NewTaskService adalah constructor untuk taskService.
// Ini menerapkan dependency injection untuk TaskRepository.
func NewTaskService(repo domain.TaskRepository) TaskApplicationService {
	return &taskService{
		taskRepo: repo,
	}
}

// CreateTask menghandle logika bisnis untuk membuat task baru.
func (s *taskService) CreateTask(ctx context.Context, userID domain.UserID, input CreateTaskInput) (*domain.Task, error) {
	// Di sini bisa ada validasi input tambahan jika diperlukan
	if input.Title == "" {
		// Sebaiknya gunakan error yang lebih spesifik atau error package
		return nil, errors.New("title cannot be empty")
	}

	newTask := &domain.Task{
		// ID akan di-generate oleh persistence layer atau database (misalnya, UUID)
		UserID:      userID,
		Title:       input.Title,
		Description: input.Description,
		Completed:   false, // Default saat pembuatan
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err := s.taskRepo.Save(ctx, newTask)
	if err != nil {
		// Log error di sini jika perlu
		return nil, err
	}
	return newTask, nil
}

// GetTaskByID mengambil task berdasarkan ID, memastikan pengguna memiliki akses.
func (s *taskService) GetTaskByID(ctx context.Context, userID domain.UserID, taskID string) (*domain.Task, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, err // Bisa jadi domain.ErrTaskNotFound
	}

	// Otorisasi: Pastikan task milik pengguna yang meminta
	if task.UserID != userID {
		return nil, domain.ErrTaskNotFound // Atau error Forbidden yang lebih spesifik
	}

	return task, nil
}

// GetTasksByUserID mengambil semua task milik pengguna tertentu.
func (s *taskService) GetTasksByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Task, error) {
	return s.taskRepo.FindByUserID(ctx, userID)
}

// UpdateTask menghandle logika bisnis untuk memperbarui task.
func (s *taskService) UpdateTask(ctx context.Context, userID domain.UserID, taskID string, input UpdateTaskInput) (*domain.Task, error) {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return nil, err
	}

	// Otorisasi: Pastikan task milik pengguna yang meminta
	if task.UserID != userID {
		return nil, domain.ErrTaskNotFound // Atau error Forbidden
	}

	// Terapkan perubahan jika ada inputnya
	if input.Title != nil {
		task.Title = *input.Title
	}
	if input.Description != nil {
		task.Description = *input.Description
	}
	if input.Completed != nil {
		task.Completed = *input.Completed
	}
	task.UpdatedAt = time.Now()

	err = s.taskRepo.Update(ctx, task)
	if err != nil {
		return nil, err
	}
	return task, nil
}

// DeleteTask menghandle logika bisnis untuk menghapus task.
func (s *taskService) DeleteTask(ctx context.Context, userID domain.UserID, taskID string) error {
	task, err := s.taskRepo.FindByID(ctx, taskID)
	if err != nil {
		return err
	}

	// Otorisasi: Pastikan task milik pengguna yang meminta
	if task.UserID != userID {
		return domain.ErrTaskNotFound // Atau error Forbidden
	}

	return s.taskRepo.Delete(ctx, taskID)
}
