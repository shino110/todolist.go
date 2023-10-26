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
	CreatedAt time.Time      `db:"created_at"`
	Memo      sql.NullString `db:"memo"`
	IsDone    bool           `db:"is_done"`
}
