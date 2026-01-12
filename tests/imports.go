package tests

import (
	"github.com/turnforge/lilbattle/lib"
	"github.com/turnforge/lilbattle/services/fsbe"
	"github.com/turnforge/lilbattle/services/singleton"
)

type AxialCoord = lib.AxialCoord
type CombatContext = lib.CombatContext
type Game = lib.Game
type World = lib.World
type SingletonGamesService = singleton.SingletonGamesService

var DevDataPath = fsbe.DevDataPath
var UnitSetCoord = lib.UnitSetCoord
var NewWorld = lib.NewWorld
var NewGame = lib.NewGame
var NewTile = lib.NewTile
var NewUnit = lib.NewUnit
var CubeDistance = lib.CubeDistance
var NewSingletonGamesService = singleton.NewSingletonGamesService
var DefaultRulesEngine = lib.DefaultRulesEngine
var ParseActionAlternatives = lib.ParseActionAlternatives
var LoadRulesEngineFromFile = lib.LoadRulesEngineFromFile

// Tile type constants
const (
	TileTypeLandBase    = lib.TileTypeLandBase
	TileTypeNavalBase   = lib.TileTypeNavalBase
	TileTypeAirport     = lib.TileTypeAirport
	TileTypeGrass       = lib.TileTypeGrass
	TileTypeMissileSilo = lib.TileTypeMissileSilo
	TileTypeMines       = lib.TileTypeMines
)

// Unit type constants
const (
	UnitTypeSoldier = lib.UnitTypeSoldier
)
