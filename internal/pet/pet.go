package pet

// ActionType définit le type d'action (choc, vibration, bip)
type ActionType string

const (
	ActionShock   ActionType = "shock"
	ActionVibrate ActionType = "vibrate"
	ActionBeep    ActionType = "beep"
)

// CommandRequest standardise une requête pour n'importe quel pet
type CommandRequest struct {
	Action    ActionType
	Intensity int // Généralement de 1 à 100
	Duration  int // En secondes ou millisecondes, adapté par l'implémentation
}

// Pet est l'interface commune pour tous les appareils
type Pet interface {
	SendCommand(req CommandRequest) error
	GetName() string
	GetType() string
}
