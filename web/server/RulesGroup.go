package server

import (
	"net/http"

	goal "github.com/panyam/goapplib"
)

// RulesGroup implements goal.PageGroup for /rules routes.
type RulesGroup struct{}

// RegisterRoutes registers all rules-related routes using goal.Register.
func (g *RulesGroup) RegisterRoutes(app *goal.App[*LilBattleApp]) *http.ServeMux {
	mux := http.NewServeMux()

	// Attack simulator page
	goal.Register[*AttackSimulatorPage](app, mux, "/attacksim")

	// Fix (repair) simulator page
	goal.Register[*FixSimulatorPage](app, mux, "/fixsim")

	return mux
}
