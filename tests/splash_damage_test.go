package tests

import (
	"math/rand"
	"testing"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	"github.com/turnforge/lilbattle/lib"
)

// TestSplashDamageBasic tests that splash damage is calculated and applied correctly
func TestSplashDamageBasic(t *testing.T) {
	rulesEngine := lib.DefaultRulesEngine()
	if rulesEngine == nil {
		t.Fatal("Failed to load default rules engine")
	}

	// Use Artillery (Mega) which has splash_damage = 1
	attackerType := int32(25) // Artillery (Mega)
	attacker := &v1.Unit{
		Q:               0,
		R:               0,
		Player:          1,
		UnitType:        attackerType,
		AvailableHealth: 10,
		DistanceLeft:    3.0,
	}

	attackerTile := &v1.Tile{
		Q:        0,
		R:        0,
		TileType: 5, // Grass
	}

	// Defender at (2, 0)
	defenderCoord := lib.AxialCoord{Q: 2, R: 0}

	// Create adjacent units around defender
	// Position (3, 0) - adjacent to defender on the right
	adjacentUnit1 := &v1.Unit{
		Q:               3,
		R:               0,
		Player:          2, // Enemy
		UnitType:        1, // Soldier
		AvailableHealth: 10,
	}

	// Position (2, -1) - adjacent to defender on top-right
	adjacentUnit2 := &v1.Unit{
		Q:               2,
		R:               -1,
		Player:          1, // Friendly fire!
		UnitType:        1, // Soldier
		AvailableHealth: 10,
	}

	adjacentUnits := []*v1.Unit{adjacentUnit1, adjacentUnit2}

	// Create a simple world
	world := lib.NewWorld("test", nil)
	world.AddTile(&v1.Tile{Q: 0, R: 0, TileType: 5})
	world.AddTile(&v1.Tile{Q: 2, R: 0, TileType: 5})
	world.AddTile(&v1.Tile{Q: 3, R: 0, TileType: 5})
	world.AddTile(&v1.Tile{Q: 2, R: -1, TileType: 5})
	world.AddUnit(attacker)
	world.AddUnit(adjacentUnit1)
	world.AddUnit(adjacentUnit2)

	// Calculate splash damage
	rng := rand.New(rand.NewSource(42))
	splashTargets, err := rulesEngine.CalculateSplashDamage(
		attacker,
		attackerTile,
		defenderCoord,
		adjacentUnits,
		world,
		rng,
	)

	if err != nil {
		t.Fatalf("Failed to calculate splash damage: %v", err)
	}

	// Should have some targets (depends on RNG, but with high HP units should hit something)
	t.Logf("Splash targets: %d", len(splashTargets))
	for i, target := range splashTargets {
		targetDef, _ := rulesEngine.GetUnitData(target.Unit.UnitType)
		t.Logf("Target %d: %s at (%d, %d) - Damage: %d HP",
			i+1, targetDef.Name, target.Unit.Q, target.Unit.R, target.Damage)
	}

	// Verify that damage is > 4 for all targets (per rules)
	for _, target := range splashTargets {
		if target.Damage <= 4 {
			t.Errorf("Splash damage %d should be > 4", target.Damage)
		}
	}
}

// TestSplashDamageAirImmunity tests that air units are immune to splash damage
func TestSplashDamageAirImmunity(t *testing.T) {
	rulesEngine := lib.DefaultRulesEngine()
	if rulesEngine == nil {
		t.Fatal("Failed to load default rules engine")
	}

	// Use Nuclear Missile which has splash_damage = 6
	attackerType := int32(22) // Missile (Nuclear)
	attacker := &v1.Unit{
		Q:               0,
		R:               0,
		Player:          1,
		UnitType:        attackerType,
		AvailableHealth: 10,
	}

	attackerTile := &v1.Tile{
		Q:        0,
		R:        0,
		TileType: 5, // Grass
	}

	defenderCoord := lib.AxialCoord{Q: 2, R: 0}

	// Create ground unit adjacent to defender
	groundUnit := &v1.Unit{
		Q:               3,
		R:               0,
		Player:          2,
		UnitType:        1, // Soldier (Land unit)
		AvailableHealth: 10,
	}

	// Create air unit adjacent to defender
	airUnit := &v1.Unit{
		Q:               2,
		R:               -1,
		Player:          2,
		UnitType:        17, // Helicopter (Air unit)
		AvailableHealth: 10,
	}

	adjacentUnits := []*v1.Unit{groundUnit, airUnit}

	// Create world
	world := lib.NewWorld("test", nil)
	world.AddTile(&v1.Tile{Q: 0, R: 0, TileType: 5})
	world.AddTile(&v1.Tile{Q: 2, R: 0, TileType: 5})
	world.AddTile(&v1.Tile{Q: 3, R: 0, TileType: 5})
	world.AddTile(&v1.Tile{Q: 2, R: -1, TileType: 5})
	world.AddUnit(attacker)
	world.AddUnit(groundUnit)
	world.AddUnit(airUnit)

	// Calculate splash damage
	rng := rand.New(rand.NewSource(42))
	splashTargets, err := rulesEngine.CalculateSplashDamage(
		attacker,
		attackerTile,
		defenderCoord,
		adjacentUnits,
		world,
		rng,
	)

	if err != nil {
		t.Fatalf("Failed to calculate splash damage: %v", err)
	}

	// Verify air unit is not in splash targets
	for _, target := range splashTargets {
		if target.Unit.UnitType == 17 { // Helicopter
			t.Errorf("Air unit should be immune to splash damage but was hit")
		}
	}

	t.Logf("Splash targets (air units should be excluded): %d", len(splashTargets))
}

// TestNoSplashDamageUnit tests that units without splash_damage don't cause splash
func TestNoSplashDamageUnit(t *testing.T) {
	rulesEngine := lib.DefaultRulesEngine()
	if rulesEngine == nil {
		t.Fatal("Failed to load default rules engine")
	}

	// Use regular Soldier which has splash_damage = 0
	attacker := &v1.Unit{
		Q:               0,
		R:               0,
		Player:          1,
		UnitType:        1, // Soldier
		AvailableHealth: 10,
	}

	attackerTile := &v1.Tile{
		Q:        0,
		R:        0,
		TileType: 5,
	}

	defenderCoord := lib.AxialCoord{Q: 1, R: 0}

	// Create adjacent unit
	adjacentUnit := &v1.Unit{
		Q:               2,
		R:               0,
		Player:          2,
		UnitType:        1,
		AvailableHealth: 10,
	}

	adjacentUnits := []*v1.Unit{adjacentUnit}

	world := lib.NewWorld("test", nil)
	world.AddTile(&v1.Tile{Q: 0, R: 0, TileType: 5})
	world.AddTile(&v1.Tile{Q: 1, R: 0, TileType: 5})
	world.AddTile(&v1.Tile{Q: 2, R: 0, TileType: 5})
	world.AddUnit(attacker)
	world.AddUnit(adjacentUnit)

	rng := rand.New(rand.NewSource(42))
	splashTargets, err := rulesEngine.CalculateSplashDamage(
		attacker,
		attackerTile,
		defenderCoord,
		adjacentUnits,
		world,
		rng,
	)

	if err != nil {
		t.Fatalf("Failed to calculate splash damage: %v", err)
	}

	// Should have no splash targets
	if len(splashTargets) > 0 {
		t.Errorf("Soldier should not cause splash damage, but hit %d units", len(splashTargets))
	}
}
