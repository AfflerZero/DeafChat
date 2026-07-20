package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"

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

		clientID := generateClientID()
		client := signal.NewClient(hub, conn, room, clientID, app.logger)

		hub.Register <- client

		go client.WritePump()
		go client.ReadPump()
	}
}

func generateClientID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
