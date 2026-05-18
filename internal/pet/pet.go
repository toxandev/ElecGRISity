package pet

// ActionType defines the action type (shock, vibration, beep)
type ActionType string

const (
	ActionShock   ActionType = "shock"
	ActionVibrate ActionType = "vibrate"
	ActionBeep    ActionType = "beep"
)

// CommandRequest standardizes a request for any pet
type CommandRequest struct {
	Action    ActionType
	Intensity int // Generally from 1 to 100
	Duration  int // In seconds or milliseconds, adjusted by the implementation
}

// Pet is the common interface for all devices
type Pet interface {
	SendCommand(req CommandRequest) error
	GetName() string
	GetType() string
}
