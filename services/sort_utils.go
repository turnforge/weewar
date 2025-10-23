package services

import (
	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
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
			return MoveOptionLess(aOpt.Move, bOpt.Move)
		}
	case *v1.GameOption_Attack:
		if bOpt, ok := b.OptionType.(*v1.GameOption_Attack); ok {
			return AttackOptionLess(aOpt.Attack, bOpt.Attack)
		}
	}

	return false
}

// getOptionTypePriority returns sort priority for option types
func getOptionTypePriority(opt *v1.GameOption) int {
	switch opt.OptionType.(type) {
	case *v1.GameOption_Move:
		return 0
	case *v1.GameOption_Attack:
		return 1
	case *v1.GameOption_EndTurn:
		return 2
	case *v1.GameOption_Build:
		return 3
	case *v1.GameOption_Capture:
		return 4
	default:
		return 99
	}
}

// MoveOptionLess compares two MoveOptions
// First by movement cost, then by direction
func MoveOptionLess(a, b *v1.MoveOption) bool {
	// Compare by cost first
	if a.MovementCost != b.MovementCost {
		return a.MovementCost < b.MovementCost
	}

	// Same cost, compare by direction
	fromA := CoordFromInt32(a.Action.FromQ, a.Action.FromR)
	toA := CoordFromInt32(a.Action.ToQ, a.Action.ToR)
	dirA := GetDirection(fromA, toA)

	fromB := CoordFromInt32(b.Action.FromQ, b.Action.FromR)
	toB := CoordFromInt32(b.Action.ToQ, b.Action.ToR)
	dirB := GetDirection(fromB, toB)

	return dirA < dirB
}

// AttackOptionLess compares two AttackOptions
// First by distance, then by target health (lowest first), then by unit type
func AttackOptionLess(a, b *v1.AttackOption) bool {
	// Calculate distances
	attackerA := CoordFromInt32(a.Action.AttackerQ, a.Action.AttackerR)
	targetA := CoordFromInt32(a.Action.DefenderQ, a.Action.DefenderR)
	distA := attackerA.Distance(targetA)

	attackerB := CoordFromInt32(b.Action.AttackerQ, b.Action.AttackerR)
	targetB := CoordFromInt32(b.Action.DefenderQ, b.Action.DefenderR)
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
