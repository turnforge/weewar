package lib

import (
	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
)

// GameOptionLess compares two GameOptions for sorting
// Order: moves < attacks < end turn
// Within each type, uses type-specific sorting
func GameOptionLess(a, b *v1.GameOption) bool {
	// Get option type priorities
	aPriority := getOptionTypePriority(a)
	bPriority := getOptionTypePriority(b)

	if aPriority != bPriority {
		return aPriority < bPriority
	}

	// Same type, use type-specific comparison
	switch aOpt := a.OptionType.(type) {
	case *v1.GameOption_Move:
		if bOpt, ok := b.OptionType.(*v1.GameOption_Move); ok {
			return MoveUnitActionLess(aOpt.Move, bOpt.Move)
		}
	case *v1.GameOption_Attack:
		if bOpt, ok := b.OptionType.(*v1.GameOption_Attack); ok {
			return AttackUnitActionLess(aOpt.Attack, bOpt.Attack)
		}
	}

	return false
}

// getOptionTypePriority returns sort priority for option types
func getOptionTypePriority(opt *v1.GameOption) int {
	switch opt.OptionType.(type) {
	case *v1.GameOption_Capture:
		return 0
	case *v1.GameOption_Heal:
		return 1
	case *v1.GameOption_Move:
		return 2
	case *v1.GameOption_Attack:
		return 3
	case *v1.GameOption_Build:
		return 4
	case *v1.GameOption_EndTurn:
		return 5
	default:
		return 99
	}
}

// MoveUnitActionLess compares two MoveUnitActions
// First by movement cost, then by direction
func MoveUnitActionLess(a, b *v1.MoveUnitAction) bool {
	// Compare by cost first
	if a.MovementCost != b.MovementCost {
		return a.MovementCost < b.MovementCost
	}

	// Same cost, compare by direction
	fromA := CoordFromInt32(a.From.Q, a.From.R)
	toA := CoordFromInt32(a.To.Q, a.To.R)
	dirA := GetDirection(fromA, toA)

	fromB := CoordFromInt32(b.From.Q, b.From.R)
	toB := CoordFromInt32(b.To.Q, b.To.R)
	dirB := GetDirection(fromB, toB)

	return dirA < dirB
}

// AttackUnitActionLess compares two AttackUnitActions
// First by distance, then by target health (lowest first), then by unit type
func AttackUnitActionLess(a, b *v1.AttackUnitAction) bool {
	// Calculate distances
	attackerA := CoordFromInt32(a.Attacker.Q, a.Attacker.R)
	targetA := CoordFromInt32(a.Defender.Q, a.Defender.R)
	distA := attackerA.Distance(targetA)

	attackerB := CoordFromInt32(b.Attacker.Q, b.Attacker.R)
	targetB := CoordFromInt32(b.Defender.Q, b.Defender.R)
	distB := attackerB.Distance(targetB)

	if distA != distB {
		return distA < distB
	}

	// Same distance, compare by target health (lower health first)
	if a.TargetUnitHealth != b.TargetUnitHealth {
		return a.TargetUnitHealth < b.TargetUnitHealth
	}

	// Same health, compare by unit type
	return a.TargetUnitType < b.TargetUnitType
}

// containsAction checks if an action is in the allowed actions list
func ContainsAction(actions []string, action string) bool {
	for _, a := range actions {
		if a == action {
			return true
		}
	}
	return false
}
