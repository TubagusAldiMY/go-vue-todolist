package domain

import (
	"context"
	"errors"
	"time"
)

// UserID adalah tipe custom untuk ID pengguna.
// Menggunakan tipe string untuk saat ini, diasumsikan ID dari Supabase Auth (sub claim) adalah string.
// Pertimbangkan menggunakan uuid.UUID jika ID pengguna selalu UUID.
type UserID string

// Task merepresentasikan entitas tugas dalam sistem.
type Task struct {
	ID          string    `json:"id"`          // ID unik untuk task (misalnya, UUID)
	UserID      UserID    `json:"user_id"`     // ID pengguna yang memiliki task ini
	Title       string    `json:"title"`       // Judul task
	Description string    `json:"description"` // Deskripsi task (opsional)
	Completed   bool      `json:"completed"`   // Status selesai task
	CreatedAt   time.Time `json:"created_at"`  // Waktu pembuatan task
	UpdatedAt   time.Time `json:"updated_at"`  // Waktu pembaruan terakhir task
}

// Definisikan error domain yang umum
var (
	ErrTaskNotFound       = errors.New("task not found")
	ErrTaskUpdateConflict = errors.New("task update conflict") // Contoh jika ada pemeriksaan versi
	// Tambahkan error domain lain jika diperlukan
)

// TaskRepository mendefinisikan kontrak untuk operasi data Task.
// Layer infrastructure (persistence) akan mengimplementasikan interface ini.
type TaskRepository interface {
	// Save menyimpan task baru ke dalam penyimpanan.
	Save(ctx context.Context, task *Task) error

	// FindByID mencari task berdasarkan ID uniknya.
	// Mengembalikan ErrTaskNotFound jika tidak ditemukan.
	FindByID(ctx context.Context, id string) (*Task, error)

	// FindByUserID mencari semua task yang dimiliki oleh pengguna tertentu.
	FindByUserID(ctx context.Context, userID UserID) ([]*Task, error)

	// Update memperbarui data task yang sudah ada di penyimpanan.
	// Sebaiknya hanya field yang relevan (Title, Description, Completed, UpdatedAt) yang diupdate.
	// Mengembalikan ErrTaskNotFound jika task tidak ada.
	Update(ctx context.Context, task *Task) error

	// Delete menghapus task berdasarkan ID uniknya dari penyimpanan.
	// Mengembalikan ErrTaskNotFound jika task tidak ada.
	Delete(ctx context.Context, id string) error
}
