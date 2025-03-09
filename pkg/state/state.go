package state

import (
	"log/slog"
	"time"

	"github.com/zrcoder/podFiles/pkg/models"

	"github.com/patrickmn/go-cache"
)

const (
	SessionKey     = "podfiles_session"
	SessionMinutes = 30
	sessionLife    = SessionMinutes * time.Minute
)

var sessins = cache.New(sessionLife, 5*time.Minute)

func Add(session string) {
	slog.Debug("add session", slog.String("session", session))
	sessins.Add(session, &models.State{}, sessionLife)
}

func Get(session string) *models.State {
	slog.Debug("get session", slog.String("session", session))
	s, ok := sessins.Get(session)
	if !ok {
		return nil
	}
	return s.(*models.State)
}

func Remove(session string) {
	sessins.Delete(session)
}
