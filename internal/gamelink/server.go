package gamelink

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sync"
	"telemetry-server/internal/pet"
	"time"
)

type GameEvent struct {
	Event string `json:"event"`
	Value int    `json:"value"`
}

type Server struct {
	port                int
	pets                map[string]pet.Pet
	logChannel          chan<- string
	httpServer          *http.Server
	mu                  sync.RWMutex
	baseIntensity       float64
	clickCounter        int
	lastItemBuy         int
	lastMoneyAddRequest pet.CommandRequest
	shopOpenCounter     int
	fearCancel          context.CancelFunc
}

func NewServer(port int, pets map[string]pet.Pet, logChannel chan<- string) *Server {
	return &Server{
		port:            port,
		pets:            pets,
		logChannel:      logChannel,
		baseIntensity:   100,
		clickCounter:    0,
		shopOpenCounter: 0,
	}
}

func (s *Server) setLastItemBuy(itemID int, req pet.CommandRequest) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastItemBuy = itemID
	s.lastMoneyAddRequest = req
}

func (s *Server) getLastMoneyAddRequest() pet.CommandRequest {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.lastMoneyAddRequest.Action == "" {
		// default
		return pet.CommandRequest{Action: pet.ActionVibrate, Intensity: int(s.baseIntensity * 0.1), Duration: 300}
	}
	return s.lastMoneyAddRequest
}

func moneyAddRequestForItem(s *Server, itemID int) pet.CommandRequest {
	switch itemID {
	case 7: // feather purchase
		return pet.CommandRequest{Action: pet.ActionVibrate, Intensity: int(s.baseIntensity * 0.2), Duration: 300}
	case 8: // needle purchase
		return pet.CommandRequest{Action: pet.ActionShock, Intensity: int(s.baseIntensity * 0.1), Duration: 300}
	case 9: // hammer purchase
		return pet.CommandRequest{Action: pet.ActionShock, Intensity: int(s.baseIntensity * 0.25), Duration: 500}
	case 10: // scissors purchase
		return pet.CommandRequest{Action: pet.ActionShock, Intensity: int(s.baseIntensity * 0.4), Duration: 1000}
	case 11: // match purchase
		return pet.CommandRequest{Action: pet.ActionBeep, Intensity: int(s.baseIntensity * 0.6), Duration: 2000}
	case 12: // knife purchase
		return pet.CommandRequest{Action: pet.ActionShock, Intensity: int(s.baseIntensity * 0.75), Duration: 3000}
	case 13: // gun purchase
		return pet.CommandRequest{Action: pet.ActionShock, Intensity: int(s.baseIntensity * 1.0), Duration: 8000}
	default:
		return pet.CommandRequest{Action: pet.ActionVibrate, Intensity: int(s.baseIntensity * 0.1), Duration: 300}
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

	ln, err := net.Listen("tcp", addr)
	if err != nil {
		s.logChannel <- fmt.Sprintf("❌ Failed to bind %s: %v", addr, err)
		return err
	}

	s.logChannel <- fmt.Sprintf("Listening for game events on %s", addr)

	if err := srv.Serve(ln); err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (s *Server) handleEvent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost && r.Method != http.MethodPut {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var event GameEvent
	if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	s.logChannel <- fmt.Sprintf("Event received: %s (Value: %d)", event.Event, event.Value)

	if event.Event == "item_buy" {
		itemRequest := moneyAddRequestForItem(s, event.Value)
		s.setLastItemBuy(event.Value, itemRequest)

		switch event.Value {
		case 7: // feather purchase
			s.logChannel <- "🛒 Detected purchase of Item Feather!"
		case 8: // needle purchase
			s.logChannel <- "🛒 Detected purchase of Needle!"
		case 9: // hammer purchase
			s.logChannel <- "🛒 Detected purchase of Hammer!"
		case 10: // scissors purchase
			s.logChannel <- "🛒 Detected purchase of Scissors!"
		case 11: // match purchase
			s.logChannel <- "🛒 Detected purchase of Match!"
		case 12: // knife purchase
			s.logChannel <- "🛒 Detected purchase of Knife!"
		case 13: // gun purchase
			s.logChannel <- "🛒 Detected purchase of Gun!"
		default:
			s.logChannel <- fmt.Sprintf("🛒 Detected purchase of unknown item (ID: %d)", event.Value)
		}

		s.logChannel <- fmt.Sprintf("📌 money_add will use Action=%s, Intensity=%d, Duration=%d", itemRequest.Action, itemRequest.Intensity, itemRequest.Duration)
		s.shopOpenCounter = 0 // reset counter after money_add event
	}

	// safe-word system to turn down intensity by a bit :)
	if event.Event == "shop_open" {
		s.shopOpenCounter++
		if s.shopOpenCounter >= 3 {
			s.baseIntensity -= 10
			s.logChannel <- fmt.Sprintf("⚠️ Shop opened %d times, reducing base intensity to %f", s.shopOpenCounter, s.baseIntensity)
		}
		s.startFearAura()
	}

	if event.Event == "shop_close" {
		s.stopFearAura()
	}

	if event.Event == "money_add" {
		s.clickCounter++
		request := s.getLastMoneyAddRequest()

		// trigger shock every 25 clicks, or if the user has gun.
		if s.clickCounter%25 == 0 || s.lastItemBuy == 13 {
			for _, p := range s.pets {
				s.logChannel <- fmt.Sprintf("⚡ Triggering action on %s! Action=%s Intensity=%d Duration=%d", p.GetName(), request.Action, request.Intensity, request.Duration)

				err := p.SendCommand(request)
				if err != nil {
					s.logChannel <- fmt.Sprintf("❌ Failed to command %s: %v", p.GetName(), err)
				} else {
					s.logChannel <- fmt.Sprintf("✅ Successfully commanded %s", p.GetName())
				}
				break
			}
		}
		s.shopOpenCounter = 0 // reset counter after money_add event
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) startFearAura() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.fearCancel != nil {
		s.fearCancel()
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.fearCancel = cancel

	s.logChannel <- "👻 Fear aura started — light vibration while shop is open"

	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()

		sendFear := func() {
			req := pet.CommandRequest{Action: pet.ActionVibrate, Intensity: int(s.baseIntensity * 0.1), Duration: 1000}
			for _, p := range s.pets {
				err := p.SendCommand(req)
				if err != nil {
					s.logChannel <- fmt.Sprintf("❌ Fear aura failed on %s: %v", p.GetName(), err)
				}
				break
			}
		}

		sendFear()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				sendFear()
			}
		}
	}()
}

func (s *Server) stopFearAura() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.fearCancel != nil {
		s.fearCancel()
		s.fearCancel = nil
		s.logChannel <- "👻 Fear aura stopped — shop closed"
	}
}
