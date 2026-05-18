package pet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

// PiShock op codes
const (
	OpShock   = 0
	OpVibrate = 1
	OpBeep    = 2
)

// PiShockPet implémente l'interface Pet pour un appareil PiShock
type PiShockPet struct {
	Name      string
	ShareCode string
	Username  string
	APIKey    string
	AppName   string
}

// apiPayload est le format JSON requis par PiShock
type apiPayload struct {
	Username  string `json:"Username"`
	APIKey    string `json:"APIKey"`
	Code      string `json:"Code"`
	Name      string `json:"Name"`
	Op        int    `json:"Op"`
	Intensity int    `json:"Intensity"`
	Duration  int    `json:"Duration"`
}

func (p *PiShockPet) GetName() string {
	return p.Name
}

func (p *PiShockPet) GetType() string {
	return "pishock"
}

func (p *PiShockPet) SendCommand(req CommandRequest) error {
	op := OpVibrate // Défaut
	switch req.Action {
	case ActionShock:
		op = OpShock
	case ActionVibrate:
		op = OpVibrate
	case ActionBeep:
		op = OpBeep
	}

	payload := apiPayload{
		Username:  p.Username,
		APIKey:    p.APIKey,
		Code:      p.ShareCode,
		Name:      p.AppName,
		Op:        op,
		Intensity: req.Intensity,
		Duration:  req.Duration,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	resp, err := http.Post("https://do.pishock.com/api/apioperate/", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Printf("PiShock %s: Successfully sent %s signal\n", p.Name, req.Action)
	} else {
		return fmt.Errorf("PiShock %s: Failed to send signal. Status: %d", p.Name, resp.StatusCode)
	}

	return nil
}
