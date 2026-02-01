package app

import (
	"database/sql"
	"log/slog"
	"social-network/internal/services"
)

type App struct {
	DB     *sql.DB
	Logger *slog.Logger
	File   *services.FileService
	Hub    *services.Hub
}
