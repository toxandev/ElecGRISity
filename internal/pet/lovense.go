package pet

import (
	"fmt"
)

// LovensePet implements the Pet interface for a Lovense device
// (Lovense-specific network logic will be added here)
type LovensePet struct {
	Name      string
	LovenseID string
	LovenseIP string
}

func (p *LovensePet) GetName() string {
	return p.Name
}

func (p *LovensePet) GetType() string {
	return "lovense"
}

func (p *LovensePet) SendCommand(req CommandRequest) error {
	// As an example, we only log the call intent.
	// Lovense-specific HTTP/WebSocket requests still need to be implemented.
	// Lovense generally only supports ActionVibrate, but we can map the rest.

	actionName := string(req.Action)

	// Mock example for now:
	if req.Action == ActionShock {
		fmt.Printf("Lovense %s: does not support shocks, converting to strong vibration...\n", p.Name)
		actionName = "vibrate (mapped from shock)"
	}

	fmt.Printf("Lovense %s: simulated SendCommand: Action=%s, Intensity=%d\n", p.Name, actionName, req.Intensity)

	// Lovense implementation TODO (GET/POST requests to LovenseIP)
	// Exemple: http://<LovenseIP>/command?v=1&t=<LovenseID>
	return nil
}
