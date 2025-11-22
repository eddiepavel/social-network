package app

import (
	"database/sql"
	"log/slog"
)

type App struct {
	DB     *sql.DB
	Logger *slog.Logger
}
