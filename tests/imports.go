package tests

import (
	"github.com/turnforge/weewar/lib"
	"github.com/turnforge/weewar/services/fsbe"
	"github.com/turnforge/weewar/services/singleton"
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
