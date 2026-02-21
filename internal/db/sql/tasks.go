package dbsql

import "database/sql"

// InitSchema ensures the required tables exist.
func InitSchema(db *sql.DB) error {
	// Simple schema for Task Planner
	const ddl = `
CREATE TABLE IF NOT EXISTS tasks (
  id TEXT PRIMARY KEY,
  session_id TEXT NOT NULL,
  title TEXT NOT NULL,
  status TEXT NOT NULL,
  current_step_index INTEGER NOT NULL,
  steps TEXT NOT NULL,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL
);
`
	_, err := db.Exec(ddl)
	return err
}
