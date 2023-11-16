package db

// schema.go provides data models in DB
import (
	"database/sql"
	"time"
)

// Task corresponds to a row in `tasks` table
type Task struct {
	ID        uint64         `db:"id"`
	Title     string         `db:"title"`
	DeadLine  sql.NullTime   `db:"deadline"`
	CreatedAt time.Time      `db:"created_at"`
	Memo      sql.NullString `db:"memo"`
	IsDone    bool           `db:"is_done"`
}

type User struct {
	ID       uint64 `db:"id"`
	Name     string `db:"name"`
	Password []byte `db:"password"`
}

type Owner struct {
	UserID uint64 `db:"user_id"`
	TaskID uint64 `db:"task_id"`
}
