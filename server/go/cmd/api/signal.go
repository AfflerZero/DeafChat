package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"regexp"

	"github.com/AfflerZero/DeafChat/internal/data"
	"github.com/AfflerZero/DeafChat/internal/signal"
	"github.com/AfflerZero/DeafChat/internal/validator"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			return false
		}
		for _, trusted := range trustedOrigins {
			if origin == trusted {
				return true
			}
		}
		return false
	},
}

var trustedOrigins []string
var roomNameRegexp = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9_-]*[A-Za-z0-9])?$`)

func SetTrustedOrigins(origins []string) {
	trustedOrigins = origins
}

func (app *application) signalHandler(hub *signal.Hub) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		room := r.URL.Query().Get("room")
		if room == "" {
			app.badRequestResponse(w, r, errors.New("room query parameter is required"))
			return
		}
		if app.config.ws.minRoomLength > 0 && len(room) < app.config.ws.minRoomLength {
			app.badRequestResponse(w, r, fmt.Errorf("room must be at least %d characters long", app.config.ws.minRoomLength))
			return
		}
		if app.config.ws.maxRoomLength > 0 && len(room) > app.config.ws.maxRoomLength {
			app.badRequestResponse(w, r, fmt.Errorf("room must be no more than %d characters long", app.config.ws.maxRoomLength))
			return
		}
		if !roomNameRegexp.MatchString(room) {
			app.badRequestResponse(w, r, errors.New("room contains invalid characters"))
			return
		}
		ok, reason := hub.CanAccept(room)
		if !ok {
			app.errorResponse(w, r, http.StatusTooManyRequests, reason)
			return
		}

		cookie, err := r.Cookie("session_token")
		if err != nil || cookie.Value == "" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := cookie.Value

		v := validator.New()
		if data.ValidateTokenPlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := app.models.Users.GetForToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		if user.IsAnonymous() || !user.Activated {
			app.authenticationRequiredResponse(w, r)
			return
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			app.logError(r, err)
			return
		}

		clientID, err := generateClientID()
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}
		client := signal.NewClient(hub, conn, room, clientID, app.logger)

		hub.Register <- client

		go client.WritePump()
		go client.ReadPump()
	}
}

func generateClientID() (string, error) {
	b := make([]byte, 8)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("generate client id: %w", err)
	}
	return hex.EncodeToString(b), nil
}
