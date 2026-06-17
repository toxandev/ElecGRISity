package pet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// PiShock op codes
const (
	OpShock   = 0
	OpVibrate = 1
	OpBeep    = 2
)

// PiShockPet implements the Pet interface for a PiShock device
type PiShockPet struct {
	Name      string
	APIKey    string
	ShockerID string
}

type apiPayload struct {
	AgentName             string `json:"AgentName"`
	Operation             int    `json:"Operation"`
	Duration              int    `json:"Duration"`
	Intensity             int    `json:"Intensity"`
	MinimumDuration       int    `json:"MinimumDuration"`
	MinimumIntensity      int    `json:"MinimumIntensity"`
	IntensityAsPercentage bool   `json:"IntensityAsPercentage"`
}

func (p *PiShockPet) GetName() string {
	return p.Name
}

func (p *PiShockPet) GetType() string {
	return "pishock"
}

func (p *PiShockPet) SendCommand(req CommandRequest) error {
	op := OpVibrate // Default
	switch req.Action {
	case ActionShock:
		op = OpShock
	case ActionVibrate:
		op = OpVibrate
	case ActionBeep:
		op = OpBeep
	}

	durationMs := req.Duration
	if durationMs <= 0 {
		durationMs = 1000
	} else if durationMs < 16 {
		durationMs *= 1000
	}

	intensity := req.Intensity
	if intensity <= 0 {
		intensity = 10
	}

	payload := apiPayload{
		AgentName:             p.Name,
		Operation:             op,
		Duration:              durationMs,
		Intensity:             intensity,
		MinimumDuration:       durationMs,
		MinimumIntensity:      intensity,
		IntensityAsPercentage: false,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	url := "https://api.pishock.com/Shockers/" + p.ShockerID
	postReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	postReq.Header.Set("Content-Type", "application/json;charset=UTF-8")
	postReq.Header.Set("X-PiShock-Api-Key", p.APIKey)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}
	resp, err := client.Do(postReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNoContent {
		fmt.Printf("PiShock %s: Successfully sent %s signal, intensity: %d, duration: %d ms\n", p.Name, req.Action, intensity, durationMs)
	} else {
		return fmt.Errorf("PiShock %s: Failed to send signal. Status: %d, body: %s", p.Name, resp.StatusCode, string(body))
	}

	return nil
}
