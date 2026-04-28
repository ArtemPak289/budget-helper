package app

import (
	"context"
	"database/sql"
	"errors"
	"os/signal"
	"syscall"
	"time"

	"budget-helper/internal/backup"
	"budget-helper/internal/config"
	"budget-helper/internal/logger"
	"budget-helper/internal/repository"
	"budget-helper/internal/service"
	"budget-helper/internal/tui"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	tea "github.com/charmbracelet/bubbletea"
)

type App struct {
	cfg    config.Config
	log    zerolog.Logger
	db     *sql.DB
	backup *backup.Service
	svc    *service.LedgerService
}

func Run() error {
	var (
		app App
		err error
	)
	app, err = buildApp()
	if err == nil {
		err = app.run()
	}
	return err
}

func buildApp() (App, error) {
	var (
		app App
		err error
	)
	app.cfg, err = config.Load()
	if err == nil {
		app.log = logger.New(app.cfg.Debug)
		err = app.openDB()
	}
	if err == nil {
		err = repository.Migrate(app.db)
	}
	if err == nil {
		app.svc = buildService(app.db, app.log)
		app.backup = backup.NewService(app.cfg.DBPath, app.cfg.BackupDir, app.log, time.Now)
	}
	return app, err
}

func (a *App) run() error {
	var err error
	if a.svc == nil || a.backup == nil || a.db == nil {
		err = errors.New("app is not initialized")
	}
	if err == nil {
		a.log.Info().Msg("starting app")
		a.backupIfPossible(context.Background(), "startup")
		ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer stop()
		a.backup.StartCleanup(ctx, time.Hour)
		err = a.runTUI(ctx)
		a.backupIfPossible(context.Background(), "shutdown")
		err = joinErr(err, a.db.Close())
	}
	return err
}

func (a *App) openDB() error {
	var err error
	a.db, err = repository.OpenSQLite(a.cfg.DBPath)
	if err == nil {
		a.log.Info().Str("db", a.cfg.DBPath).Msg("db opened")
	}
	if err != nil {
		a.log.Error().Err(err).Msg("db open failed")
	}
	return err
}

func buildService(db *sql.DB, log zerolog.Logger) *service.LedgerService {
	repo := repository.NewSQLiteRepository(db, log)
	idGen := func() (string, error) {
		return uuid.NewString(), nil
	}
	return service.NewLedgerService(repo, repo, time.Now, idGen)
}

func (a *App) runTUI(ctx context.Context) error {
	var err error
	model := tui.NewModel(a.svc, a.cfg, a.log)
	program := tea.NewProgram(model, tea.WithContext(ctx))
	_, err = program.Run()
	return err
}

func (a *App) backupIfPossible(ctx context.Context, stage string) {
	err := a.backup.Backup(ctx)
	if err != nil {
		a.log.Error().Err(err).Str("stage", stage).Msg("backup failed")
	}
	if err == nil {
		a.log.Info().Str("stage", stage).Msg("backup completed")
	}
}

func joinErr(primary error, secondary error) error {
	var err error
	err = primary
	if err == nil {
		err = secondary
	}
	return err
}
