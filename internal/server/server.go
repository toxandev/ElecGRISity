package server

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"telemetry-server/internal/pet"
)

type GameEvent struct {
	Event string `json:"event"`
	Value int    `json:"value"`
}

type Server struct {
	port       int
	pets       map[string]pet.Pet
	logChannel chan<- string
	httpServer *http.Server
}

func NewServer(port int, pets map[string]pet.Pet, logChannel chan<- string) *Server {
	return &Server{
		port:       port,
		pets:       pets,
		logChannel: logChannel,
	}
}

func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/event", s.handleEvent)

	addr := fmt.Sprintf(":%d", s.port)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// Goroutine to gracefully shut down the server when context is canceled
	go func() {
		<-ctx.Done()
		srv.Close()
	}()

	s.logChannel <- fmt.Sprintf("Listening for game telemetry on %s", addr)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) handleEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event GameEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	s.logChannel <- fmt.Sprintf("Event received: %s (Value: %d)", event.Event, event.Value)

	if event.Event == "damage" {
		for _, p := range s.pets {
			s.logChannel <- fmt.Sprintf("⚡ Triggering action on %s! (Intensity: %d)", p.GetName(), event.Value)

			req := pet.CommandRequest{
				Action:    pet.ActionShock,
				Intensity: event.Value,
				Duration:  1,
			}

			err := p.SendCommand(req)

			if err != nil {
				s.logChannel <- fmt.Sprintf("❌ Failed to command %s: %v", p.GetName(), err)
			} else {
				s.logChannel <- fmt.Sprintf("✅ Successfully commanded %s", p.GetName())
			}
			break
		}
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
