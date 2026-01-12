package services

import "github.com/turnforge/lilbattle/lib"

type AxialCoord = lib.AxialCoord
type Position = lib.Position
type NeighborDirection = lib.NeighborDirection

var ReconstructPath = lib.ReconstructPath
var CoordFromInt32 = lib.CoordFromInt32
var CubeDistance = lib.CubeDistance
var GetDirection = lib.GetDirection
var DirectionToString = lib.DirectionToString
var DirectionToLongString = lib.DirectionToLongString
var RowColToHex = lib.RowColToHex
var HexToRowCol = lib.HexToRowCol
var ParseDirection = lib.ParseDirection
var ExtractPathCoords = lib.ExtractPathCoords
var FormatPathDetailed = lib.FormatPathDetailed
var FormatPathCompact = lib.FormatPathCompact
