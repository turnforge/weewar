
// Generated TypeScript schemas from proto file
// DO NOT EDIT - This file is auto-generated

import { FieldType, FieldSchema, MessageSchema } from "./deserializer_schemas";


/**
 * Schema for User message
 */
export const UserSchema: MessageSchema = {
  name: "User",
  fields: [
    {
      name: "createdAt",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "updatedAt",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "id",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "name",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "description",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "tags",
      type: FieldType.REPEATED,
      id: 6,
      repeated: true,
    },
    {
      name: "imageUrl",
      type: FieldType.STRING,
      id: 7,
    },
    {
      name: "difficulty",
      type: FieldType.STRING,
      id: 8,
    },
  ],
};


/**
 * Schema for Pagination message
 */
export const PaginationSchema: MessageSchema = {
  name: "Pagination",
  fields: [
    {
      name: "pageKey",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "pageOffset",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "pageSize",
      type: FieldType.NUMBER,
      id: 3,
    },
  ],
};


/**
 * Schema for PaginationResponse message
 */
export const PaginationResponseSchema: MessageSchema = {
  name: "PaginationResponse",
  fields: [
    {
      name: "nextPageKey",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "nextPageOffset",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "hasMore",
      type: FieldType.BOOLEAN,
      id: 4,
    },
    {
      name: "totalResults",
      type: FieldType.NUMBER,
      id: 5,
    },
  ],
};


/**
 * Schema for World message
 */
export const WorldSchema: MessageSchema = {
  name: "World",
  fields: [
    {
      name: "createdAt",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "updatedAt",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "id",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "creatorId",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "name",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "description",
      type: FieldType.STRING,
      id: 6,
    },
    {
      name: "tags",
      type: FieldType.REPEATED,
      id: 7,
      repeated: true,
    },
    {
      name: "imageUrl",
      type: FieldType.STRING,
      id: 8,
    },
    {
      name: "difficulty",
      type: FieldType.STRING,
      id: 9,
    },
    {
      name: "worldData",
      type: FieldType.MESSAGE,
      id: 10,
      messageType: "weewar.v1.WorldData",
    },
  ],
};


/**
 * Schema for WorldData message
 */
export const WorldDataSchema: MessageSchema = {
  name: "WorldData",
  fields: [
    {
      name: "tiles",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.Tile",
      repeated: true,
    },
    {
      name: "units",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.Unit",
      repeated: true,
    },
  ],
};


/**
 * Schema for Tile message
 */
export const TileSchema: MessageSchema = {
  name: "Tile",
  fields: [
    {
      name: "q",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "r",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "tileType",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "player",
      type: FieldType.NUMBER,
      id: 4,
    },
  ],
};


/**
 * Schema for Unit message
 */
export const UnitSchema: MessageSchema = {
  name: "Unit",
  fields: [
    {
      name: "q",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "r",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "player",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "unitType",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "availableHealth",
      type: FieldType.NUMBER,
      id: 5,
    },
    {
      name: "distanceLeft",
      type: FieldType.NUMBER,
      id: 6,
    },
    {
      name: "turnCounter",
      type: FieldType.NUMBER,
      id: 7,
    },
  ],
};


/**
 * Schema for TerrainDefinition message
 */
export const TerrainDefinitionSchema: MessageSchema = {
  name: "TerrainDefinition",
  fields: [
    {
      name: "id",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "name",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "baseMoveCost",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "defenseBonus",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "type",
      type: FieldType.NUMBER,
      id: 5,
    },
    {
      name: "description",
      type: FieldType.STRING,
      id: 6,
    },
  ],
};


/**
 * Schema for UnitDefinition message
 */
export const UnitDefinitionSchema: MessageSchema = {
  name: "UnitDefinition",
  fields: [
    {
      name: "id",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "name",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "movementPoints",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "attackRange",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "health",
      type: FieldType.NUMBER,
      id: 5,
    },
    {
      name: "properties",
      type: FieldType.REPEATED,
      id: 6,
      repeated: true,
    },
  ],
};


/**
 * Schema for MovementMatrix message
 */
export const MovementMatrixSchema: MessageSchema = {
  name: "MovementMatrix",
  fields: [
    {
      name: "costs",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.CostsEntry",
    },
  ],
};


/**
 * Schema for TerrainCostMap message
 */
export const TerrainCostMapSchema: MessageSchema = {
  name: "TerrainCostMap",
  fields: [
    {
      name: "terrainCosts",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.TerrainCostsEntry",
    },
  ],
};


/**
 * Schema for Game message
 */
export const GameSchema: MessageSchema = {
  name: "Game",
  fields: [
    {
      name: "createdAt",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "updatedAt",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "id",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "creatorId",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "worldId",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "name",
      type: FieldType.STRING,
      id: 6,
    },
    {
      name: "description",
      type: FieldType.STRING,
      id: 7,
    },
    {
      name: "tags",
      type: FieldType.REPEATED,
      id: 8,
      repeated: true,
    },
    {
      name: "imageUrl",
      type: FieldType.STRING,
      id: 9,
    },
    {
      name: "difficulty",
      type: FieldType.STRING,
      id: 10,
    },
    {
      name: "config",
      type: FieldType.MESSAGE,
      id: 11,
      messageType: "weewar.v1.GameConfiguration",
    },
  ],
};


/**
 * Schema for GameConfiguration message
 */
export const GameConfigurationSchema: MessageSchema = {
  name: "GameConfiguration",
  fields: [
    {
      name: "players",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.GamePlayer",
      repeated: true,
    },
    {
      name: "settings",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.GameSettings",
    },
  ],
};


/**
 * Schema for GamePlayer message
 */
export const GamePlayerSchema: MessageSchema = {
  name: "GamePlayer",
  fields: [
    {
      name: "playerId",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "playerType",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "color",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "teamId",
      type: FieldType.NUMBER,
      id: 4,
    },
  ],
};


/**
 * Schema for GameSettings message
 */
export const GameSettingsSchema: MessageSchema = {
  name: "GameSettings",
  fields: [
    {
      name: "allowedUnits",
      type: FieldType.REPEATED,
      id: 1,
      repeated: true,
    },
    {
      name: "turnTimeLimit",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "teamMode",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "maxTurns",
      type: FieldType.NUMBER,
      id: 4,
    },
  ],
};


/**
 * Schema for GameState message
 */
export const GameStateSchema: MessageSchema = {
  name: "GameState",
  fields: [
    {
      name: "updatedAt",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "gameId",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "turnCounter",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "currentPlayer",
      type: FieldType.NUMBER,
      id: 5,
    },
    {
      name: "worldData",
      type: FieldType.MESSAGE,
      id: 6,
      messageType: "weewar.v1.WorldData",
    },
  ],
};


/**
 * Schema for GameMoveHistory message
 */
export const GameMoveHistorySchema: MessageSchema = {
  name: "GameMoveHistory",
  fields: [
    {
      name: "gameId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "groups",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.GameMoveGroup",
      repeated: true,
    },
  ],
};


/**
 * Schema for GameMoveGroup message
 */
export const GameMoveGroupSchema: MessageSchema = {
  name: "GameMoveGroup",
  fields: [
    {
      name: "startedAt",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "endedAt",
      type: FieldType.MESSAGE,
      id: 3,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "moves",
      type: FieldType.MESSAGE,
      id: 4,
      messageType: "weewar.v1.GameMove",
      repeated: true,
    },
    {
      name: "moveResults",
      type: FieldType.MESSAGE,
      id: 5,
      messageType: "weewar.v1.GameMoveResult",
      repeated: true,
    },
  ],
};


/**
 * Schema for GameMove message
 */
export const GameMoveSchema: MessageSchema = {
  name: "GameMove",
  fields: [
    {
      name: "player",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "timestamp",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "sequenceNum",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "moveUnit",
      type: FieldType.MESSAGE,
      id: 4,
      messageType: "weewar.v1.MoveUnitAction",
      oneofGroup: "move_type",
    },
    {
      name: "attackUnit",
      type: FieldType.MESSAGE,
      id: 5,
      messageType: "weewar.v1.AttackUnitAction",
      oneofGroup: "move_type",
    },
    {
      name: "endTurn",
      type: FieldType.MESSAGE,
      id: 6,
      messageType: "weewar.v1.EndTurnAction",
      oneofGroup: "move_type",
    },
  ],
  oneofGroups: ["move_type"],
};


/**
 * Schema for GameMoveResult message
 */
export const GameMoveResultSchema: MessageSchema = {
  name: "GameMoveResult",
  fields: [
    {
      name: "isPermanent",
      type: FieldType.BOOLEAN,
      id: 1,
    },
    {
      name: "sequenceNum",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "changes",
      type: FieldType.MESSAGE,
      id: 3,
      messageType: "weewar.v1.WorldChange",
      repeated: true,
    },
  ],
};


/**
 * Schema for MoveUnitAction message
 */
export const MoveUnitActionSchema: MessageSchema = {
  name: "MoveUnitAction",
  fields: [
    {
      name: "fromQ",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "fromR",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "toQ",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "toR",
      type: FieldType.NUMBER,
      id: 4,
    },
  ],
};


/**
 * Schema for AttackUnitAction message
 */
export const AttackUnitActionSchema: MessageSchema = {
  name: "AttackUnitAction",
  fields: [
    {
      name: "attackerQ",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "attackerR",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "defenderQ",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "defenderR",
      type: FieldType.NUMBER,
      id: 4,
    },
  ],
};


/**
 * Schema for EndTurnAction message
 */
export const EndTurnActionSchema: MessageSchema = {
  name: "EndTurnAction",
  fields: [
  ],
};


/**
 * Schema for WorldChange message
 */
export const WorldChangeSchema: MessageSchema = {
  name: "WorldChange",
  fields: [
    {
      name: "unitMoved",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.UnitMovedChange",
      oneofGroup: "change_type",
    },
    {
      name: "unitDamaged",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.UnitDamagedChange",
      oneofGroup: "change_type",
    },
    {
      name: "unitKilled",
      type: FieldType.MESSAGE,
      id: 3,
      messageType: "weewar.v1.UnitKilledChange",
      oneofGroup: "change_type",
    },
    {
      name: "playerChanged",
      type: FieldType.MESSAGE,
      id: 4,
      messageType: "weewar.v1.PlayerChangedChange",
      oneofGroup: "change_type",
    },
  ],
  oneofGroups: ["change_type"],
};


/**
 * Schema for UnitMovedChange message
 */
export const UnitMovedChangeSchema: MessageSchema = {
  name: "UnitMovedChange",
  fields: [
    {
      name: "fromQ",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "fromR",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "toQ",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "toR",
      type: FieldType.NUMBER,
      id: 5,
    },
  ],
};


/**
 * Schema for UnitDamagedChange message
 */
export const UnitDamagedChangeSchema: MessageSchema = {
  name: "UnitDamagedChange",
  fields: [
    {
      name: "previousHealth",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "newHealth",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "q",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "r",
      type: FieldType.NUMBER,
      id: 5,
    },
  ],
};


/**
 * Schema for UnitKilledChange message
 */
export const UnitKilledChangeSchema: MessageSchema = {
  name: "UnitKilledChange",
  fields: [
    {
      name: "player",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "unitType",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "q",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "r",
      type: FieldType.NUMBER,
      id: 5,
    },
  ],
};


/**
 * Schema for PlayerChangedChange message
 */
export const PlayerChangedChangeSchema: MessageSchema = {
  name: "PlayerChangedChange",
  fields: [
    {
      name: "previousPlayer",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "newPlayer",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "previousTurn",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "newTurn",
      type: FieldType.NUMBER,
      id: 4,
    },
  ],
};


/**
 * Schema for GameInfo message
 */
export const GameInfoSchema: MessageSchema = {
  name: "GameInfo",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "name",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "description",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "category",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "difficulty",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "tags",
      type: FieldType.REPEATED,
      id: 6,
      repeated: true,
    },
    {
      name: "icon",
      type: FieldType.STRING,
      id: 7,
    },
    {
      name: "lastUpdated",
      type: FieldType.STRING,
      id: 8,
    },
  ],
};


/**
 * Schema for ListGamesRequest message
 */
export const ListGamesRequestSchema: MessageSchema = {
  name: "ListGamesRequest",
  fields: [
    {
      name: "pagination",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.Pagination",
    },
    {
      name: "ownerId",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for ListGamesResponse message
 */
export const ListGamesResponseSchema: MessageSchema = {
  name: "ListGamesResponse",
  fields: [
    {
      name: "items",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.Game",
      repeated: true,
    },
    {
      name: "pagination",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.PaginationResponse",
    },
  ],
};


/**
 * Schema for GetGameRequest message
 */
export const GetGameRequestSchema: MessageSchema = {
  name: "GetGameRequest",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "version",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for GetGameResponse message
 */
export const GetGameResponseSchema: MessageSchema = {
  name: "GetGameResponse",
  fields: [
    {
      name: "game",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.Game",
    },
    {
      name: "state",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.GameState",
    },
    {
      name: "history",
      type: FieldType.MESSAGE,
      id: 3,
      messageType: "weewar.v1.GameMoveHistory",
    },
  ],
};


/**
 * Schema for GetGameContentRequest message
 */
export const GetGameContentRequestSchema: MessageSchema = {
  name: "GetGameContentRequest",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "version",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for GetGameContentResponse message
 */
export const GetGameContentResponseSchema: MessageSchema = {
  name: "GetGameContentResponse",
  fields: [
    {
      name: "weewarContent",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "recipeContent",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "readmeContent",
      type: FieldType.STRING,
      id: 3,
    },
  ],
};


/**
 * Schema for UpdateGameRequest message
 */
export const UpdateGameRequestSchema: MessageSchema = {
  name: "UpdateGameRequest",
  fields: [
    {
      name: "gameId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "newGame",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.Game",
    },
    {
      name: "newState",
      type: FieldType.MESSAGE,
      id: 3,
      messageType: "weewar.v1.GameState",
    },
    {
      name: "newHistory",
      type: FieldType.MESSAGE,
      id: 4,
      messageType: "weewar.v1.GameMoveHistory",
    },
    {
      name: "updateMask",
      type: FieldType.MESSAGE,
      id: 5,
      messageType: "google.protobuf.FieldMask",
    },
  ],
};


/**
 * Schema for UpdateGameResponse message
 */
export const UpdateGameResponseSchema: MessageSchema = {
  name: "UpdateGameResponse",
  fields: [
    {
      name: "game",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.Game",
    },
  ],
};


/**
 * Schema for DeleteGameRequest message
 */
export const DeleteGameRequestSchema: MessageSchema = {
  name: "DeleteGameRequest",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for DeleteGameResponse message
 */
export const DeleteGameResponseSchema: MessageSchema = {
  name: "DeleteGameResponse",
  fields: [
  ],
};


/**
 * Schema for GetGamesRequest message
 */
export const GetGamesRequestSchema: MessageSchema = {
  name: "GetGamesRequest",
  fields: [
    {
      name: "ids",
      type: FieldType.REPEATED,
      id: 1,
      repeated: true,
    },
  ],
};


/**
 * Schema for GetGamesResponse message
 */
export const GetGamesResponseSchema: MessageSchema = {
  name: "GetGamesResponse",
  fields: [
    {
      name: "games",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.GamesEntry",
    },
  ],
};


/**
 * Schema for CreateGameRequest message
 */
export const CreateGameRequestSchema: MessageSchema = {
  name: "CreateGameRequest",
  fields: [
    {
      name: "game",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.Game",
    },
  ],
};


/**
 * Schema for CreateGameResponse message
 */
export const CreateGameResponseSchema: MessageSchema = {
  name: "CreateGameResponse",
  fields: [
    {
      name: "game",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.Game",
    },
    {
      name: "gameState",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.GameState",
    },
    {
      name: "fieldErrors",
      type: FieldType.MESSAGE,
      id: 3,
      messageType: "weewar.v1.FieldErrorsEntry",
    },
  ],
};


/**
 * Schema for ProcessMovesRequest message
 */
export const ProcessMovesRequestSchema: MessageSchema = {
  name: "ProcessMovesRequest",
  fields: [
    {
      name: "gameId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "moves",
      type: FieldType.MESSAGE,
      id: 3,
      messageType: "weewar.v1.GameMove",
      repeated: true,
    },
  ],
};


/**
 * Schema for ProcessMovesResponse message
 */
export const ProcessMovesResponseSchema: MessageSchema = {
  name: "ProcessMovesResponse",
  fields: [
    {
      name: "moveResults",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.GameMoveResult",
      repeated: true,
    },
    {
      name: "changes",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.WorldChange",
      repeated: true,
    },
  ],
};


/**
 * Schema for GetGameStateRequest message
 */
export const GetGameStateRequestSchema: MessageSchema = {
  name: "GetGameStateRequest",
  fields: [
    {
      name: "gameId",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for GetGameStateResponse message
 */
export const GetGameStateResponseSchema: MessageSchema = {
  name: "GetGameStateResponse",
  fields: [
    {
      name: "state",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.GameState",
    },
  ],
};


/**
 * Schema for ListMovesRequest message
 */
export const ListMovesRequestSchema: MessageSchema = {
  name: "ListMovesRequest",
  fields: [
    {
      name: "gameId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "offset",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "lastN",
      type: FieldType.NUMBER,
      id: 3,
    },
  ],
};


/**
 * Schema for ListMovesResponse message
 */
export const ListMovesResponseSchema: MessageSchema = {
  name: "ListMovesResponse",
  fields: [
    {
      name: "hasMore",
      type: FieldType.BOOLEAN,
      id: 1,
    },
    {
      name: "moveGroups",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.GameMoveGroup",
      repeated: true,
    },
  ],
};


/**
 * Schema for GetOptionsAtRequest message
 */
export const GetOptionsAtRequestSchema: MessageSchema = {
  name: "GetOptionsAtRequest",
  fields: [
    {
      name: "gameId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "q",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "r",
      type: FieldType.NUMBER,
      id: 3,
    },
  ],
};


/**
 * Schema for GetOptionsAtResponse message
 */
export const GetOptionsAtResponseSchema: MessageSchema = {
  name: "GetOptionsAtResponse",
  fields: [
    {
      name: "options",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.GameOption",
      repeated: true,
    },
    {
      name: "currentPlayer",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "gameInitialized",
      type: FieldType.BOOLEAN,
      id: 3,
    },
  ],
};


/**
 * Schema for GameOption message
 */
export const GameOptionSchema: MessageSchema = {
  name: "GameOption",
  fields: [
    {
      name: "move",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.MoveOption",
      oneofGroup: "option_type",
    },
    {
      name: "attack",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.AttackOption",
      oneofGroup: "option_type",
    },
    {
      name: "endTurn",
      type: FieldType.MESSAGE,
      id: 3,
      messageType: "weewar.v1.EndTurnOption",
      oneofGroup: "option_type",
    },
    {
      name: "build",
      type: FieldType.MESSAGE,
      id: 4,
      messageType: "weewar.v1.BuildUnitOption",
      oneofGroup: "option_type",
    },
    {
      name: "capture",
      type: FieldType.MESSAGE,
      id: 5,
      messageType: "weewar.v1.CaptureBuildingOption",
      oneofGroup: "option_type",
    },
  ],
  oneofGroups: ["option_type"],
};


/**
 * Schema for EndTurnOption message
 */
export const EndTurnOptionSchema: MessageSchema = {
  name: "EndTurnOption",
  fields: [
  ],
};


/**
 * Schema for MoveOption message
 */
export const MoveOptionSchema: MessageSchema = {
  name: "MoveOption",
  fields: [
    {
      name: "q",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "r",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "movementCost",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "action",
      type: FieldType.MESSAGE,
      id: 4,
      messageType: "weewar.v1.MoveUnitAction",
    },
  ],
};


/**
 * Schema for AttackOption message
 */
export const AttackOptionSchema: MessageSchema = {
  name: "AttackOption",
  fields: [
    {
      name: "q",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "r",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "targetUnitType",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "targetUnitHealth",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "canAttack",
      type: FieldType.BOOLEAN,
      id: 5,
    },
    {
      name: "damageEstimate",
      type: FieldType.NUMBER,
      id: 6,
    },
    {
      name: "action",
      type: FieldType.MESSAGE,
      id: 7,
      messageType: "weewar.v1.AttackUnitAction",
    },
  ],
};


/**
 * Schema for BuildUnitOption message
 */
export const BuildUnitOptionSchema: MessageSchema = {
  name: "BuildUnitOption",
  fields: [
    {
      name: "q",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "r",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "tileType",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "buildCost",
      type: FieldType.NUMBER,
      id: 4,
    },
  ],
};


/**
 * Schema for CaptureBuildingOption message
 */
export const CaptureBuildingOptionSchema: MessageSchema = {
  name: "CaptureBuildingOption",
  fields: [
    {
      name: "q",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "r",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "tileType",
      type: FieldType.NUMBER,
      id: 3,
    },
  ],
};


/**
 * Schema for UserInfo message
 */
export const UserInfoSchema: MessageSchema = {
  name: "UserInfo",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "name",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "description",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "category",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "difficulty",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "tags",
      type: FieldType.REPEATED,
      id: 6,
      repeated: true,
    },
    {
      name: "icon",
      type: FieldType.STRING,
      id: 7,
    },
    {
      name: "lastUpdated",
      type: FieldType.STRING,
      id: 8,
    },
  ],
};


/**
 * Schema for ListUsersRequest message
 */
export const ListUsersRequestSchema: MessageSchema = {
  name: "ListUsersRequest",
  fields: [
    {
      name: "pagination",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.Pagination",
    },
    {
      name: "ownerId",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for ListUsersResponse message
 */
export const ListUsersResponseSchema: MessageSchema = {
  name: "ListUsersResponse",
  fields: [
    {
      name: "items",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.User",
      repeated: true,
    },
    {
      name: "pagination",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.PaginationResponse",
    },
  ],
};


/**
 * Schema for GetUserRequest message
 */
export const GetUserRequestSchema: MessageSchema = {
  name: "GetUserRequest",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "version",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for GetUserResponse message
 */
export const GetUserResponseSchema: MessageSchema = {
  name: "GetUserResponse",
  fields: [
    {
      name: "user",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.User",
    },
  ],
};


/**
 * Schema for GetUserContentRequest message
 */
export const GetUserContentRequestSchema: MessageSchema = {
  name: "GetUserContentRequest",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "version",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for GetUserContentResponse message
 */
export const GetUserContentResponseSchema: MessageSchema = {
  name: "GetUserContentResponse",
  fields: [
    {
      name: "weewarContent",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "recipeContent",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "readmeContent",
      type: FieldType.STRING,
      id: 3,
    },
  ],
};


/**
 * Schema for UpdateUserRequest message
 */
export const UpdateUserRequestSchema: MessageSchema = {
  name: "UpdateUserRequest",
  fields: [
    {
      name: "user",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.User",
    },
    {
      name: "updateMask",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "google.protobuf.FieldMask",
    },
  ],
};


/**
 * Schema for UpdateUserResponse message
 */
export const UpdateUserResponseSchema: MessageSchema = {
  name: "UpdateUserResponse",
  fields: [
    {
      name: "user",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.User",
    },
  ],
};


/**
 * Schema for DeleteUserRequest message
 */
export const DeleteUserRequestSchema: MessageSchema = {
  name: "DeleteUserRequest",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for DeleteUserResponse message
 */
export const DeleteUserResponseSchema: MessageSchema = {
  name: "DeleteUserResponse",
  fields: [
  ],
};


/**
 * Schema for GetUsersRequest message
 */
export const GetUsersRequestSchema: MessageSchema = {
  name: "GetUsersRequest",
  fields: [
    {
      name: "ids",
      type: FieldType.REPEATED,
      id: 1,
      repeated: true,
    },
  ],
};


/**
 * Schema for GetUsersResponse message
 */
export const GetUsersResponseSchema: MessageSchema = {
  name: "GetUsersResponse",
  fields: [
    {
      name: "users",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.UsersEntry",
    },
  ],
};


/**
 * Schema for CreateUserRequest message
 */
export const CreateUserRequestSchema: MessageSchema = {
  name: "CreateUserRequest",
  fields: [
    {
      name: "user",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.User",
    },
  ],
};


/**
 * Schema for CreateUserResponse message
 */
export const CreateUserResponseSchema: MessageSchema = {
  name: "CreateUserResponse",
  fields: [
    {
      name: "user",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.User",
    },
    {
      name: "fieldErrors",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.FieldErrorsEntry",
    },
  ],
};


/**
 * Schema for WorldInfo message
 */
export const WorldInfoSchema: MessageSchema = {
  name: "WorldInfo",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "name",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "description",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "category",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "difficulty",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "tags",
      type: FieldType.REPEATED,
      id: 6,
      repeated: true,
    },
    {
      name: "icon",
      type: FieldType.STRING,
      id: 7,
    },
    {
      name: "lastUpdated",
      type: FieldType.STRING,
      id: 8,
    },
  ],
};


/**
 * Schema for ListWorldsRequest message
 */
export const ListWorldsRequestSchema: MessageSchema = {
  name: "ListWorldsRequest",
  fields: [
    {
      name: "pagination",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.Pagination",
    },
    {
      name: "ownerId",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for ListWorldsResponse message
 */
export const ListWorldsResponseSchema: MessageSchema = {
  name: "ListWorldsResponse",
  fields: [
    {
      name: "items",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.World",
      repeated: true,
    },
    {
      name: "pagination",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.PaginationResponse",
    },
  ],
};


/**
 * Schema for GetWorldRequest message
 */
export const GetWorldRequestSchema: MessageSchema = {
  name: "GetWorldRequest",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "version",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for GetWorldResponse message
 */
export const GetWorldResponseSchema: MessageSchema = {
  name: "GetWorldResponse",
  fields: [
    {
      name: "world",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.World",
    },
    {
      name: "worldData",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.WorldData",
    },
  ],
};


/**
 * Schema for UpdateWorldRequest message
 */
export const UpdateWorldRequestSchema: MessageSchema = {
  name: "UpdateWorldRequest",
  fields: [
    {
      name: "world",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.World",
    },
    {
      name: "worldData",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.WorldData",
    },
    {
      name: "clearWorld",
      type: FieldType.BOOLEAN,
      id: 3,
    },
    {
      name: "updateMask",
      type: FieldType.MESSAGE,
      id: 4,
      messageType: "google.protobuf.FieldMask",
    },
  ],
};


/**
 * Schema for UpdateWorldResponse message
 */
export const UpdateWorldResponseSchema: MessageSchema = {
  name: "UpdateWorldResponse",
  fields: [
    {
      name: "world",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.World",
    },
    {
      name: "worldData",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.WorldData",
    },
  ],
};


/**
 * Schema for DeleteWorldRequest message
 */
export const DeleteWorldRequestSchema: MessageSchema = {
  name: "DeleteWorldRequest",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for DeleteWorldResponse message
 */
export const DeleteWorldResponseSchema: MessageSchema = {
  name: "DeleteWorldResponse",
  fields: [
  ],
};


/**
 * Schema for GetWorldsRequest message
 */
export const GetWorldsRequestSchema: MessageSchema = {
  name: "GetWorldsRequest",
  fields: [
    {
      name: "ids",
      type: FieldType.REPEATED,
      id: 1,
      repeated: true,
    },
  ],
};


/**
 * Schema for GetWorldsResponse message
 */
export const GetWorldsResponseSchema: MessageSchema = {
  name: "GetWorldsResponse",
  fields: [
    {
      name: "worlds",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.WorldsEntry",
    },
  ],
};


/**
 * Schema for CreateWorldRequest message
 */
export const CreateWorldRequestSchema: MessageSchema = {
  name: "CreateWorldRequest",
  fields: [
    {
      name: "world",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.World",
    },
    {
      name: "worldData",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.WorldData",
    },
  ],
};


/**
 * Schema for CreateWorldResponse message
 */
export const CreateWorldResponseSchema: MessageSchema = {
  name: "CreateWorldResponse",
  fields: [
    {
      name: "world",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.World",
    },
    {
      name: "worldData",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.WorldData",
    },
    {
      name: "fieldErrors",
      type: FieldType.MESSAGE,
      id: 3,
      messageType: "weewar.v1.FieldErrorsEntry",
    },
  ],
};



/**
 * Package-scoped schema registry for weewar.v1
 */
export const WeewarV1SchemaRegistry: Record<string, MessageSchema> = {
  "weewar.v1.User": UserSchema,
  "weewar.v1.Pagination": PaginationSchema,
  "weewar.v1.PaginationResponse": PaginationResponseSchema,
  "weewar.v1.World": WorldSchema,
  "weewar.v1.WorldData": WorldDataSchema,
  "weewar.v1.Tile": TileSchema,
  "weewar.v1.Unit": UnitSchema,
  "weewar.v1.TerrainDefinition": TerrainDefinitionSchema,
  "weewar.v1.UnitDefinition": UnitDefinitionSchema,
  "weewar.v1.MovementMatrix": MovementMatrixSchema,
  "weewar.v1.TerrainCostMap": TerrainCostMapSchema,
  "weewar.v1.Game": GameSchema,
  "weewar.v1.GameConfiguration": GameConfigurationSchema,
  "weewar.v1.GamePlayer": GamePlayerSchema,
  "weewar.v1.GameSettings": GameSettingsSchema,
  "weewar.v1.GameState": GameStateSchema,
  "weewar.v1.GameMoveHistory": GameMoveHistorySchema,
  "weewar.v1.GameMoveGroup": GameMoveGroupSchema,
  "weewar.v1.GameMove": GameMoveSchema,
  "weewar.v1.GameMoveResult": GameMoveResultSchema,
  "weewar.v1.MoveUnitAction": MoveUnitActionSchema,
  "weewar.v1.AttackUnitAction": AttackUnitActionSchema,
  "weewar.v1.EndTurnAction": EndTurnActionSchema,
  "weewar.v1.WorldChange": WorldChangeSchema,
  "weewar.v1.UnitMovedChange": UnitMovedChangeSchema,
  "weewar.v1.UnitDamagedChange": UnitDamagedChangeSchema,
  "weewar.v1.UnitKilledChange": UnitKilledChangeSchema,
  "weewar.v1.PlayerChangedChange": PlayerChangedChangeSchema,
  "weewar.v1.GameInfo": GameInfoSchema,
  "weewar.v1.ListGamesRequest": ListGamesRequestSchema,
  "weewar.v1.ListGamesResponse": ListGamesResponseSchema,
  "weewar.v1.GetGameRequest": GetGameRequestSchema,
  "weewar.v1.GetGameResponse": GetGameResponseSchema,
  "weewar.v1.GetGameContentRequest": GetGameContentRequestSchema,
  "weewar.v1.GetGameContentResponse": GetGameContentResponseSchema,
  "weewar.v1.UpdateGameRequest": UpdateGameRequestSchema,
  "weewar.v1.UpdateGameResponse": UpdateGameResponseSchema,
  "weewar.v1.DeleteGameRequest": DeleteGameRequestSchema,
  "weewar.v1.DeleteGameResponse": DeleteGameResponseSchema,
  "weewar.v1.GetGamesRequest": GetGamesRequestSchema,
  "weewar.v1.GetGamesResponse": GetGamesResponseSchema,
  "weewar.v1.CreateGameRequest": CreateGameRequestSchema,
  "weewar.v1.CreateGameResponse": CreateGameResponseSchema,
  "weewar.v1.ProcessMovesRequest": ProcessMovesRequestSchema,
  "weewar.v1.ProcessMovesResponse": ProcessMovesResponseSchema,
  "weewar.v1.GetGameStateRequest": GetGameStateRequestSchema,
  "weewar.v1.GetGameStateResponse": GetGameStateResponseSchema,
  "weewar.v1.ListMovesRequest": ListMovesRequestSchema,
  "weewar.v1.ListMovesResponse": ListMovesResponseSchema,
  "weewar.v1.GetOptionsAtRequest": GetOptionsAtRequestSchema,
  "weewar.v1.GetOptionsAtResponse": GetOptionsAtResponseSchema,
  "weewar.v1.GameOption": GameOptionSchema,
  "weewar.v1.EndTurnOption": EndTurnOptionSchema,
  "weewar.v1.MoveOption": MoveOptionSchema,
  "weewar.v1.AttackOption": AttackOptionSchema,
  "weewar.v1.BuildUnitOption": BuildUnitOptionSchema,
  "weewar.v1.CaptureBuildingOption": CaptureBuildingOptionSchema,
  "weewar.v1.UserInfo": UserInfoSchema,
  "weewar.v1.ListUsersRequest": ListUsersRequestSchema,
  "weewar.v1.ListUsersResponse": ListUsersResponseSchema,
  "weewar.v1.GetUserRequest": GetUserRequestSchema,
  "weewar.v1.GetUserResponse": GetUserResponseSchema,
  "weewar.v1.GetUserContentRequest": GetUserContentRequestSchema,
  "weewar.v1.GetUserContentResponse": GetUserContentResponseSchema,
  "weewar.v1.UpdateUserRequest": UpdateUserRequestSchema,
  "weewar.v1.UpdateUserResponse": UpdateUserResponseSchema,
  "weewar.v1.DeleteUserRequest": DeleteUserRequestSchema,
  "weewar.v1.DeleteUserResponse": DeleteUserResponseSchema,
  "weewar.v1.GetUsersRequest": GetUsersRequestSchema,
  "weewar.v1.GetUsersResponse": GetUsersResponseSchema,
  "weewar.v1.CreateUserRequest": CreateUserRequestSchema,
  "weewar.v1.CreateUserResponse": CreateUserResponseSchema,
  "weewar.v1.WorldInfo": WorldInfoSchema,
  "weewar.v1.ListWorldsRequest": ListWorldsRequestSchema,
  "weewar.v1.ListWorldsResponse": ListWorldsResponseSchema,
  "weewar.v1.GetWorldRequest": GetWorldRequestSchema,
  "weewar.v1.GetWorldResponse": GetWorldResponseSchema,
  "weewar.v1.UpdateWorldRequest": UpdateWorldRequestSchema,
  "weewar.v1.UpdateWorldResponse": UpdateWorldResponseSchema,
  "weewar.v1.DeleteWorldRequest": DeleteWorldRequestSchema,
  "weewar.v1.DeleteWorldResponse": DeleteWorldResponseSchema,
  "weewar.v1.GetWorldsRequest": GetWorldsRequestSchema,
  "weewar.v1.GetWorldsResponse": GetWorldsResponseSchema,
  "weewar.v1.CreateWorldRequest": CreateWorldRequestSchema,
  "weewar.v1.CreateWorldResponse": CreateWorldResponseSchema,
};

/**
 * Get schema for a message type from weewar.v1 package
 */
export function getSchema(messageType: string): MessageSchema | undefined {
  return WeewarV1SchemaRegistry[messageType];
}

/**
 * Get field schema by name from weewar.v1 package
 */
export function getFieldSchema(messageType: string, fieldName: string): FieldSchema | undefined {
  const schema = getSchema(messageType);
  return schema?.fields.find(field => field.name === fieldName);
}

/**
 * Get field schema by proto field ID from weewar.v1 package
 */
export function getFieldSchemaById(messageType: string, fieldId: number): FieldSchema | undefined {
  const schema = getSchema(messageType);
  return schema?.fields.find(field => field.id === fieldId);
}

/**
 * Check if field is part of a oneof group in weewar.v1 package
 */
export function isOneofField(messageType: string, fieldName: string): boolean {
  const fieldSchema = getFieldSchema(messageType, fieldName);
  return fieldSchema?.oneofGroup !== undefined;
}

/**
 * Get all fields in a oneof group from weewar.v1 package
 */
export function getOneofFields(messageType: string, oneofGroup: string): FieldSchema[] {
  const schema = getSchema(messageType);
  return schema?.fields.filter(field => field.oneofGroup === oneofGroup) || [];
}