package handler

import (
	"errors"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"demodesk/neko/internal/types"
	"demodesk/neko/internal/types/event"
	"demodesk/neko/internal/types/message"
)

func (h *MessageHandlerCtx) systemInit(session types.Session) error {
	host := h.sessions.GetHost()

	controlHost := message.ControlHost{
		HasHost: host != nil,
	}

	if controlHost.HasHost {
		controlHost.HostID = host.ID()
	}

	size := h.desktop.GetScreenSize()
	if size == nil {
		return errors.New("could not get screen size")
	}

	sessions := map[string]message.SessionData{}
	for _, session := range h.sessions.List() {
		sessionId := session.ID()
		sessions[sessionId] = message.SessionData{
			ID:      sessionId,
			Profile: session.Profile(),
			State:   session.State(),
		}
	}

	session.Send(
		event.SYSTEM_INIT,
		message.SystemInit{
			SessionId:         session.ID(),
			ControlHost:       controlHost,
			ScreenSize:        message.ScreenSize(*size),
			Sessions:          sessions,
			ImplicitHosting:   h.sessions.ImplicitHosting(),
			InactiveCursors:   h.sessions.InactiveCursors(),
			ScreencastEnabled: h.capture.Screencast().Enabled(),
			WebRTC: message.SystemWebRTC{
				Videos: h.capture.VideoIDs(),
			},
		})

	return nil
}

func (h *MessageHandlerCtx) systemAdmin(session types.Session) error {
	screenSizesList := []message.ScreenSize{}
	for _, size := range h.desktop.ScreenConfigurations() {
		for _, rate := range size.Rates {
			screenSizesList = append(screenSizesList, message.ScreenSize{
				Width:  size.Width,
				Height: size.Height,
				Rate:   rate,
			})
		}
	}

	broadcast := h.capture.Broadcast()
	session.Send(
		event.SYSTEM_ADMIN,
		message.SystemAdmin{
			ScreenSizesList: screenSizesList,
			BroadcastStatus: message.BroadcastStatus{
				IsActive: broadcast.Started(),
				URL:      broadcast.Url(),
			},
		})

	return nil
}

func (h *MessageHandlerCtx) systemLogs(session types.Session, payload *message.SystemLogs) error {
	for _, msg := range *payload {
		level, _ := zerolog.ParseLevel(msg.Level)

		if level < zerolog.DebugLevel || level > zerolog.ErrorLevel {
			level = zerolog.NoLevel
		}

		// do not use handler logger context
		log.WithLevel(level).
			Fields(msg.Fields).
			Str("module", "client").
			Str("session_id", session.ID()).
			Msg(msg.Message)
	}

	return nil
}
