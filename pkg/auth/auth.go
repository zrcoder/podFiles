package auth

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/zrcoder/amisgo/util"
	"github.com/zrcoder/podFiles/pkg/state"
	"github.com/zrcoder/podFiles/pkg/util/log"
)

func Auth(c *gin.Context) {
	slog.Debug("auth begin")
	s, err := c.Cookie(state.SessionKey)
	if err != nil {
		slog.Error("auth", log.Error(err))
		slog.Info("generate session for new user")
		s = uuid.NewString()
		c.SetCookie(state.SessionKey, s, state.SessionMinutes*60, "/", "", false, true)
		slog.Debug("set cookie", slog.String("session", s))
		c.Abort()
		return
	}
	if state.Get(s) == nil {
		state.Add(s)
	}
	c.Set(state.SessionKey, s)
	c.Next()
}

func K8s(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s, err := r.Cookie(state.SessionKey)
		if err != nil {
			slog.Error("auth: get session for new user", log.Error(err))
			util.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		st := state.Get(s.Value)
		if st == nil || st.Namespace == "" || st.Pod == "" || st.Container == "" {
			slog.Error("auth", slog.String("error", "namespace, pod or container is required"))
			util.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		next.ServeHTTP(w, r)
	})
}
