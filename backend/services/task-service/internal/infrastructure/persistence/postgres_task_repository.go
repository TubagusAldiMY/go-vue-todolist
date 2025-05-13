// file: backend/services/task-service/internal/infrastructure/persistence/postgres_task_repository.go
package persistence

import (
	"context"
	"errors" // Pastikan ini diimpor
	"fmt"    // Untuk error wrapping

	"github.com/TubagusAldiMY/go-vue-todolist/backend/services/task-service/internal/domain" // Sesuaikan path module Anda
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresTaskRepository adalah implementasi dari domain.TaskRepository menggunakan PostgreSQL.
type PostgresTaskRepository struct {
	dbpool *pgxpool.Pool
}

// NewPostgresTaskRepository adalah constructor untuk PostgresTaskRepository.
func NewPostgresTaskRepository(dbpool *pgxpool.Pool) domain.TaskRepository {
	return &PostgresTaskRepository{
		dbpool: dbpool,
	}
}

// Save menyimpan task baru ke dalam database.
func (r *PostgresTaskRepository) Save(ctx context.Context, task *domain.Task) error {
	// Generate ID baru jika belum ada (best practice: biarkan DB generate jika memungkinkan,
	// atau generate di aplikasi sebelum insert untuk konsistensi)
	if task.ID == "" {
		task.ID = uuid.NewString()
	}

	query := `INSERT INTO tasks (id, user_id, title, description, completed, created_at, updated_at)
	           VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err := r.dbpool.Exec(ctx, query,
		task.ID,
		task.UserID,
		task.Title,
		task.Description,
		task.Completed,
		task.CreatedAt,
		task.UpdatedAt,
	)

	if err != nil {
		// Cek apakah ada error duplikasi Primary Key (jika ID sudah ada)
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 adalah kode error unique_violation
			return fmt.Errorf("error saving task: id %s already exists: %w", task.ID, err)
		}
		return fmt.Errorf("error saving task: %w", err)
	}
	return nil
}

// FindByID mencari task berdasarkan ID uniknya.
func (r *PostgresTaskRepository) FindByID(ctx context.Context, id string) (*domain.Task, error) {
	query := `SELECT id, user_id, title, description, completed, created_at, updated_at
	           FROM tasks WHERE id = $1`
	task := &domain.Task{}
	err := r.dbpool.QueryRow(ctx, query, id).Scan(
		&task.ID,
		&task.UserID,
		&task.Title,
		&task.Description,
		&task.Completed,
		&task.CreatedAt,
		&task.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTaskNotFound
		}
		return nil, fmt.Errorf("error finding task by id %s: %w", id, err)
	}
	return task, nil
}

// FindByUserID mencari semua task yang dimiliki oleh pengguna tertentu.
func (r *PostgresTaskRepository) FindByUserID(ctx context.Context, userID domain.UserID) ([]*domain.Task, error) {
	query := `SELECT id, user_id, title, description, completed, created_at, updated_at
	           FROM tasks WHERE user_id = $1 ORDER BY created_at DESC` // Urutkan berdasarkan terbaru
	rows, err := r.dbpool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("error finding tasks by user_id %s: %w", userID, err)
	}
	defer rows.Close()

	var tasks []*domain.Task
	for rows.Next() {
		task := &domain.Task{}
		err := rows.Scan(
			&task.ID,
			&task.UserID,
			&task.Title,
			&task.Description,
			&task.Completed,
			&task.CreatedAt,
			&task.UpdatedAt,
		)
		if err != nil {
			// Sebaiknya log error ini dan mungkin skip task yang error, atau batalkan semua
			return nil, fmt.Errorf("error scanning task row: %w", err)
		}
		tasks = append(tasks, task)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating task rows: %w", err)
	}

	return tasks, nil
}

// Update memperbarui data task yang sudah ada di penyimpanan.
func (r *PostgresTaskRepository) Update(ctx context.Context, task *domain.Task) error {
	query := `UPDATE tasks
	           SET title = $1, description = $2, completed = $3, updated_at = $4
	           WHERE id = $5 AND user_id = $6` // Pastikan hanya pemilik yang bisa update
	cmdTag, err := r.dbpool.Exec(ctx, query,
		task.Title,
		task.Description,
		task.Completed,
		task.UpdatedAt,
		task.ID,
		task.UserID, // Penting untuk otorisasi di level DB (tambahan selain di app layer)
	)

	if err != nil {
		return fmt.Errorf("error updating task %s: %w", task.ID, err)
	}
	if cmdTag.RowsAffected() == 0 {
		// Ini bisa berarti task tidak ditemukan atau user_id tidak cocok.
		// Kita bisa cek dulu apakah task ada untuk memberikan error yang lebih spesifik,
		// tapi untuk sekarang ErrTaskNotFound sudah cukup.
		return domain.ErrTaskNotFound
	}
	return nil
}

// Delete menghapus task berdasarkan ID uniknya dari penyimpanan.
func (r *PostgresTaskRepository) Delete(ctx context.Context, id string) error {
	// Untuk keamanan, idealnya kita juga butuh UserID di sini untuk memastikan
	// hanya pemilik yang bisa menghapus, atau logika ini sepenuhnya di application layer.
	// Karena Delete di application layer sudah mengambil UserID dan TaskID,
	// dan melakukan pengecekan kepemilikan sebelum memanggil repo.Delete(id),
	// maka query ini cukup berdasarkan ID.
	query := `DELETE FROM tasks WHERE id = $1`
	cmdTag, err := r.dbpool.Exec(ctx, query, id)

	if err != nil {
		return fmt.Errorf("error deleting task %s: %w", id, err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrTaskNotFound
	}
	return nil
}
