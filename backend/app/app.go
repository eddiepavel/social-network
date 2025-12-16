package app

import (
	"database/sql"
	"log/slog"
	"social-network/internal/services"
	"social-network/pkg/ws"
)

type App struct {
	DB     *sql.DB
	Logger *slog.Logger
	File   *services.FileService
	WsManager *ws.Manager
}
