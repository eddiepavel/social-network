package app

import (
	"database/sql"
	"log/slog"
	"social-network/internal/services"
	"social-network/internal/websocket"
)

type App struct {
	DB        *sql.DB
	Logger    *slog.Logger
	File      *services.FileService
	WsManager *websocket.Manager
}
