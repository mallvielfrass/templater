package internal

import (
	"github.com/mallvielfrass/templater/internal/config"
	filestorage "github.com/mallvielfrass/templater/internal/fileStorage"
	"github.com/mallvielfrass/templater/internal/router"
)

func NewApp() *App {
	return &App{}
}

func (app *App) Run() error {
	cfg, err := config.NewConfig(".env")
	if err != nil {
		return err
	}
	if err := cfg.Validate(); err != nil {
		return err
	}
	store, err := filestorage.NewStorage(cfg.BadgerDBPath)
	if err != nil {
		return err
	}
	r := router.NewRouter(store, store, store, cfg.JWTSecret, cfg.PublicBaseURL, cfg.OnlyOfficeJWTSecret, cfg.OnlyOfficeURL)
	r.SetCORSOrigins(cfg.CORSOrigins)
	r.SetStaticDir(cfg.StaticDir)
	r.Mount()
	return r.Run(cfg.HttpPort)
}
