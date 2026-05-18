package pet

import (
	"fmt"
)

// LovensePet implémente l'interface Pet pour un appareil Lovense
// (La logique réseau spécifique à Lovense sera ajoutée ici)
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
	// À titre d'exemple, on log juste l'intention d'appel.
	// Il faudra implémenter les requêtes HTTP/WebSocket spécifiques de Lovense.
	// Lovense ne gère généralement que ActionVibrate, mais on peut mapper le reste.

	actionName := string(req.Action)

	// Exemple de mock pour l'instant:
	if req.Action == ActionShock {
		fmt.Printf("Lovense %s: ne supporte pas les chocs, conversion en vibration forte...\n", p.Name)
		actionName = "vibrate (mapped from shock)"
	}

	fmt.Printf("Lovense %s: SendCommand simulé : Action=%s, Intensity=%d\n", p.Name, actionName, req.Intensity)

	// Implémentation Lovense à faire (requêtes GET/POST sur LovenseIP)
	// Exemple: http://<LovenseIP>/command?v=1&t=<LovenseID>
	return nil
}
