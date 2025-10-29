
// Generated TypeScript schemas from proto file
// DO NOT EDIT - This file is auto-generated

import { FieldType, FieldSchema, MessageSchema, BaseSchemaRegistry } from "@protoc-gen-go-wasmjs/runtime";


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
    {
      name: "shortcut",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "lastActedTurn",
      type: FieldType.NUMBER,
      id: 6,
    },
    {
      name: "lastToppedupTurn",
      type: FieldType.NUMBER,
      id: 7,
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
      name: "shortcut",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "availableHealth",
      type: FieldType.NUMBER,
      id: 6,
    },
    {
      name: "distanceLeft",
      type: FieldType.NUMBER,
      id: 7,
    },
    {
      name: "lastActedTurn",
      type: FieldType.NUMBER,
      id: 8,
    },
    {
      name: "lastToppedupTurn",
      type: FieldType.NUMBER,
      id: 9,
    },
    {
      name: "attacksReceivedThisTurn",
      type: FieldType.NUMBER,
      id: 10,
    },
    {
      name: "attackHistory",
      type: FieldType.MESSAGE,
      id: 11,
      messageType: "weewar.v1.AttackRecord",
      repeated: true,
    },
  ],
};


/**
 * Schema for AttackRecord message
 */
export const AttackRecordSchema: MessageSchema = {
  name: "AttackRecord",
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
      name: "isRanged",
      type: FieldType.BOOLEAN,
      id: 3,
    },
    {
      name: "turnNumber",
      type: FieldType.NUMBER,
      id: 4,
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
      name: "type",
      type: FieldType.NUMBER,
      id: 5,
    },
    {
      name: "description",
      type: FieldType.STRING,
      id: 6,
    },
    {
      name: "unitProperties",
      type: FieldType.STRING,
      id: 7,
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
      name: "description",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "health",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "coins",
      type: FieldType.NUMBER,
      id: 5,
    },
    {
      name: "movementPoints",
      type: FieldType.NUMBER,
      id: 6,
    },
    {
      name: "retreatPoints",
      type: FieldType.NUMBER,
      id: 7,
    },
    {
      name: "defense",
      type: FieldType.NUMBER,
      id: 8,
    },
    {
      name: "attackRange",
      type: FieldType.NUMBER,
      id: 9,
    },
    {
      name: "minAttackRange",
      type: FieldType.NUMBER,
      id: 10,
    },
    {
      name: "splashDamage",
      type: FieldType.NUMBER,
      id: 11,
    },
    {
      name: "terrainProperties",
      type: FieldType.STRING,
      id: 12,
    },
    {
      name: "properties",
      type: FieldType.REPEATED,
      id: 13,
      repeated: true,
    },
    {
      name: "unitClass",
      type: FieldType.STRING,
      id: 14,
    },
    {
      name: "unitTerrain",
      type: FieldType.STRING,
      id: 15,
    },
    {
      name: "attackVsClass",
      type: FieldType.STRING,
      id: 16,
    },
    {
      name: "actionOrder",
      type: FieldType.REPEATED,
      id: 17,
      repeated: true,
    },
    {
      name: "actionLimits",
      type: FieldType.STRING,
      id: 18,
    },
  ],
};


/**
 * Schema for TerrainUnitProperties message
 */
export const TerrainUnitPropertiesSchema: MessageSchema = {
  name: "TerrainUnitProperties",
  fields: [
    {
      name: "terrainId",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "unitId",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "movementCost",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "healingBonus",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "canBuild",
      type: FieldType.BOOLEAN,
      id: 5,
    },
    {
      name: "canCapture",
      type: FieldType.BOOLEAN,
      id: 6,
    },
    {
      name: "attackBonus",
      type: FieldType.NUMBER,
      id: 7,
    },
    {
      name: "defenseBonus",
      type: FieldType.NUMBER,
      id: 8,
    },
    {
      name: "attackRange",
      type: FieldType.NUMBER,
      id: 9,
    },
    {
      name: "minAttackRange",
      type: FieldType.NUMBER,
      id: 10,
    },
  ],
};


/**
 * Schema for UnitUnitProperties message
 */
export const UnitUnitPropertiesSchema: MessageSchema = {
  name: "UnitUnitProperties",
  fields: [
    {
      name: "attackerId",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "defenderId",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "attackOverride",
      type: FieldType.STRING,
      id: 3,
      oneofGroup: "_attack_override",
      optional: true,
    },
    {
      name: "defenseOverride",
      type: FieldType.STRING,
      id: 4,
      oneofGroup: "_defense_override",
      optional: true,
    },
    {
      name: "damage",
      type: FieldType.MESSAGE,
      id: 5,
      messageType: "weewar.v1.DamageDistribution",
    },
  ],
  oneofGroups: ["_attack_override", "_defense_override"],
};


/**
 * Schema for DamageDistribution message
 */
export const DamageDistributionSchema: MessageSchema = {
  name: "DamageDistribution",
  fields: [
    {
      name: "minDamage",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "maxDamage",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "expectedDamage",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "ranges",
      type: FieldType.MESSAGE,
      id: 4,
      messageType: "weewar.v1.DamageRange",
      repeated: true,
    },
  ],
};


/**
 * Schema for DamageRange message
 */
export const DamageRangeSchema: MessageSchema = {
  name: "DamageRange",
  fields: [
    {
      name: "minValue",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "maxValue",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "probability",
      type: FieldType.NUMBER,
      id: 3,
    },
  ],
};


/**
 * Schema for RulesEngine message
 */
export const RulesEngineSchema: MessageSchema = {
  name: "RulesEngine",
  fields: [
    {
      name: "units",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "terrains",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "terrainUnitProperties",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "unitUnitProperties",
      type: FieldType.STRING,
      id: 4,
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
      name: "teams",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.GameTeam",
      repeated: true,
    },
    {
      name: "settings",
      type: FieldType.MESSAGE,
      id: 3,
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
    {
      name: "name",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "isActive",
      type: FieldType.BOOLEAN,
      id: 6,
    },
  ],
};


/**
 * Schema for GameTeam message
 */
export const GameTeamSchema: MessageSchema = {
  name: "GameTeam",
  fields: [
    {
      name: "teamId",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "name",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "color",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "isActive",
      type: FieldType.BOOLEAN,
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
    {
      name: "stateHash",
      type: FieldType.STRING,
      id: 8,
    },
    {
      name: "version",
      type: FieldType.NUMBER,
      id: 9,
    },
    {
      name: "status",
      type: FieldType.STRING,
      id: 10,
    },
    {
      name: "finished",
      type: FieldType.BOOLEAN,
      id: 11,
    },
    {
      name: "winningPlayer",
      type: FieldType.NUMBER,
      id: 12,
    },
    {
      name: "winningTeam",
      type: FieldType.NUMBER,
      id: 13,
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
      name: "previousUnit",
      type: FieldType.MESSAGE,
      id: 6,
      messageType: "weewar.v1.Unit",
    },
    {
      name: "updatedUnit",
      type: FieldType.MESSAGE,
      id: 7,
      messageType: "weewar.v1.Unit",
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
      name: "previousUnit",
      type: FieldType.MESSAGE,
      id: 6,
      messageType: "weewar.v1.Unit",
    },
    {
      name: "updatedUnit",
      type: FieldType.MESSAGE,
      id: 7,
      messageType: "weewar.v1.Unit",
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
      name: "previousUnit",
      type: FieldType.MESSAGE,
      id: 6,
      messageType: "weewar.v1.Unit",
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
    {
      name: "resetUnits",
      type: FieldType.MESSAGE,
      id: 5,
      messageType: "weewar.v1.Unit",
      repeated: true,
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
      type: FieldType.STRING,
      id: 1,
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
      type: FieldType.STRING,
      id: 3,
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
    {
      name: "expectedResponse",
      type: FieldType.MESSAGE,
      id: 2,
      messageType: "weewar.v1.ProcessMovesResponse",
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
    {
      name: "allPaths",
      type: FieldType.MESSAGE,
      id: 5,
      messageType: "weewar.v1.AllPaths",
    },
  ],
};


/**
 * Schema for AllPaths message
 */
export const AllPathsSchema: MessageSchema = {
  name: "AllPaths",
  fields: [
    {
      name: "sourceQ",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "sourceR",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "edges",
      type: FieldType.STRING,
      id: 3,
    },
  ],
};


/**
 * Schema for PathEdge message
 */
export const PathEdgeSchema: MessageSchema = {
  name: "PathEdge",
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
    {
      name: "movementCost",
      type: FieldType.NUMBER,
      id: 5,
    },
    {
      name: "totalCost",
      type: FieldType.NUMBER,
      id: 6,
    },
    {
      name: "terrainType",
      type: FieldType.STRING,
      id: 7,
    },
    {
      name: "explanation",
      type: FieldType.STRING,
      id: 8,
    },
  ],
};


/**
 * Schema for Path message
 */
export const PathSchema: MessageSchema = {
  name: "Path",
  fields: [
    {
      name: "edges",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.PathEdge",
      repeated: true,
    },
    {
      name: "directions",
      type: FieldType.REPEATED,
      id: 2,
      repeated: true,
    },
    {
      name: "totalCost",
      type: FieldType.NUMBER,
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
    {
      name: "reconstructedPath",
      type: FieldType.MESSAGE,
      id: 5,
      messageType: "weewar.v1.Path",
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
      name: "unitType",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "cost",
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
 * Schema for SimulateAttackRequest message
 */
export const SimulateAttackRequestSchema: MessageSchema = {
  name: "SimulateAttackRequest",
  fields: [
    {
      name: "attackerUnitType",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "attackerTerrain",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "attackerHealth",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "defenderUnitType",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "defenderTerrain",
      type: FieldType.NUMBER,
      id: 5,
    },
    {
      name: "defenderHealth",
      type: FieldType.NUMBER,
      id: 6,
    },
    {
      name: "woundBonus",
      type: FieldType.NUMBER,
      id: 7,
    },
    {
      name: "numSimulations",
      type: FieldType.NUMBER,
      id: 8,
    },
  ],
};


/**
 * Schema for SimulateAttackResponse message
 */
export const SimulateAttackResponseSchema: MessageSchema = {
  name: "SimulateAttackResponse",
  fields: [
    {
      name: "attackerDamageDistribution",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "defenderDamageDistribution",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "attackerMeanDamage",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "defenderMeanDamage",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "attackerKillProbability",
      type: FieldType.NUMBER,
      id: 5,
    },
    {
      name: "defenderKillProbability",
      type: FieldType.NUMBER,
      id: 6,
    },
  ],
};


/**
 * Schema for EmptyRequest message
 */
export const EmptyRequestSchema: MessageSchema = {
  name: "EmptyRequest",
  fields: [
  ],
};


/**
 * Schema for EmptyResponse message
 */
export const EmptyResponseSchema: MessageSchema = {
  name: "EmptyResponse",
  fields: [
  ],
};


/**
 * Schema for SetContentRequest message
 */
export const SetContentRequestSchema: MessageSchema = {
  name: "SetContentRequest",
  fields: [
    {
      name: "innerHtml",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for SetContentResponse message
 */
export const SetContentResponseSchema: MessageSchema = {
  name: "SetContentResponse",
  fields: [
  ],
};


/**
 * Schema for LogMessageRequest message
 */
export const LogMessageRequestSchema: MessageSchema = {
  name: "LogMessageRequest",
  fields: [
    {
      name: "message",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for LogMessageResponse message
 */
export const LogMessageResponseSchema: MessageSchema = {
  name: "LogMessageResponse",
  fields: [
  ],
};


/**
 * Schema for SetGameStateRequest message
 */
export const SetGameStateRequestSchema: MessageSchema = {
  name: "SetGameStateRequest",
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
  ],
};


/**
 * Schema for SetGameStateResponse message
 */
export const SetGameStateResponseSchema: MessageSchema = {
  name: "SetGameStateResponse",
  fields: [
  ],
};


/**
 * Schema for UpdateGameStatusRequest message
 */
export const UpdateGameStatusRequestSchema: MessageSchema = {
  name: "UpdateGameStatusRequest",
  fields: [
    {
      name: "currentPlayer",
      type: FieldType.NUMBER,
      id: 1,
    },
    {
      name: "turnCounter",
      type: FieldType.NUMBER,
      id: 2,
    },
  ],
};


/**
 * Schema for UpdateGameStatusResponse message
 */
export const UpdateGameStatusResponseSchema: MessageSchema = {
  name: "UpdateGameStatusResponse",
  fields: [
  ],
};


/**
 * Schema for SetTileAtRequest message
 */
export const SetTileAtRequestSchema: MessageSchema = {
  name: "SetTileAtRequest",
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
      name: "tile",
      type: FieldType.MESSAGE,
      id: 3,
      messageType: "weewar.v1.Tile",
    },
  ],
};


/**
 * Schema for SetTileAtResponse message
 */
export const SetTileAtResponseSchema: MessageSchema = {
  name: "SetTileAtResponse",
  fields: [
  ],
};


/**
 * Schema for SetUnitAtRequest message
 */
export const SetUnitAtRequestSchema: MessageSchema = {
  name: "SetUnitAtRequest",
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
      name: "unit",
      type: FieldType.MESSAGE,
      id: 3,
      messageType: "weewar.v1.Unit",
    },
  ],
};


/**
 * Schema for SetUnitAtResponse message
 */
export const SetUnitAtResponseSchema: MessageSchema = {
  name: "SetUnitAtResponse",
  fields: [
  ],
};


/**
 * Schema for RemoveTileAtRequest message
 */
export const RemoveTileAtRequestSchema: MessageSchema = {
  name: "RemoveTileAtRequest",
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
  ],
};


/**
 * Schema for RemoveTileAtResponse message
 */
export const RemoveTileAtResponseSchema: MessageSchema = {
  name: "RemoveTileAtResponse",
  fields: [
  ],
};


/**
 * Schema for RemoveUnitAtRequest message
 */
export const RemoveUnitAtRequestSchema: MessageSchema = {
  name: "RemoveUnitAtRequest",
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
  ],
};


/**
 * Schema for RemoveUnitAtResponse message
 */
export const RemoveUnitAtResponseSchema: MessageSchema = {
  name: "RemoveUnitAtResponse",
  fields: [
  ],
};


/**
 * Schema for ShowHighlightsRequest message
 */
export const ShowHighlightsRequestSchema: MessageSchema = {
  name: "ShowHighlightsRequest",
  fields: [
    {
      name: "highlights",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.HighlightSpec",
      repeated: true,
    },
  ],
};


/**
 * Schema for ShowHighlightsResponse message
 */
export const ShowHighlightsResponseSchema: MessageSchema = {
  name: "ShowHighlightsResponse",
  fields: [
  ],
};


/**
 * Schema for HighlightSpec message
 */
export const HighlightSpecSchema: MessageSchema = {
  name: "HighlightSpec",
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
      name: "type",
      type: FieldType.STRING,
      id: 3,
    },
  ],
};


/**
 * Schema for ClearHighlightsRequest message
 */
export const ClearHighlightsRequestSchema: MessageSchema = {
  name: "ClearHighlightsRequest",
  fields: [
    {
      name: "types",
      type: FieldType.REPEATED,
      id: 1,
      repeated: true,
    },
  ],
};


/**
 * Schema for ClearHighlightsResponse message
 */
export const ClearHighlightsResponseSchema: MessageSchema = {
  name: "ClearHighlightsResponse",
  fields: [
  ],
};


/**
 * Schema for ShowPathRequest message
 */
export const ShowPathRequestSchema: MessageSchema = {
  name: "ShowPathRequest",
  fields: [
    {
      name: "coords",
      type: FieldType.REPEATED,
      id: 1,
      repeated: true,
    },
    {
      name: "color",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "thickness",
      type: FieldType.NUMBER,
      id: 3,
    },
  ],
};


/**
 * Schema for ShowPathResponse message
 */
export const ShowPathResponseSchema: MessageSchema = {
  name: "ShowPathResponse",
  fields: [
  ],
};


/**
 * Schema for ClearPathsRequest message
 */
export const ClearPathsRequestSchema: MessageSchema = {
  name: "ClearPathsRequest",
  fields: [
  ],
};


/**
 * Schema for ClearPathsResponse message
 */
export const ClearPathsResponseSchema: MessageSchema = {
  name: "ClearPathsResponse",
  fields: [
  ],
};


/**
 * Schema for TurnOptionClickedRequest message
 */
export const TurnOptionClickedRequestSchema: MessageSchema = {
  name: "TurnOptionClickedRequest",
  fields: [
    {
      name: "gameId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "optionIndex",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "optionType",
      type: FieldType.STRING,
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
 * Schema for TurnOptionClickedResponse message
 */
export const TurnOptionClickedResponseSchema: MessageSchema = {
  name: "TurnOptionClickedResponse",
  fields: [
    {
      name: "gameId",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for SceneClickedRequest message
 */
export const SceneClickedRequestSchema: MessageSchema = {
  name: "SceneClickedRequest",
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
    {
      name: "layer",
      type: FieldType.STRING,
      id: 4,
    },
  ],
};


/**
 * Schema for SceneClickedResponse message
 */
export const SceneClickedResponseSchema: MessageSchema = {
  name: "SceneClickedResponse",
  fields: [
    {
      name: "gameId",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for EndTurnButtonClickedRequest message
 */
export const EndTurnButtonClickedRequestSchema: MessageSchema = {
  name: "EndTurnButtonClickedRequest",
  fields: [
    {
      name: "gameId",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for EndTurnButtonClickedResponse message
 */
export const EndTurnButtonClickedResponseSchema: MessageSchema = {
  name: "EndTurnButtonClickedResponse",
  fields: [
    {
      name: "gameId",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for InitializeGameRequest message
 */
export const InitializeGameRequestSchema: MessageSchema = {
  name: "InitializeGameRequest",
  fields: [
    {
      name: "gameData",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "gameState",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "moveHistory",
      type: FieldType.STRING,
      id: 3,
    },
  ],
};


/**
 * Schema for InitializeGameResponse message
 */
export const InitializeGameResponseSchema: MessageSchema = {
  name: "InitializeGameResponse",
  fields: [
    {
      name: "success",
      type: FieldType.BOOLEAN,
      id: 1,
    },
    {
      name: "error",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "currentPlayer",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "turnCounter",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "gameName",
      type: FieldType.STRING,
      id: 5,
    },
  ],
};


/**
 * Schema for ThemeInfo message
 */
export const ThemeInfoSchema: MessageSchema = {
  name: "ThemeInfo",
  fields: [
    {
      name: "name",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "version",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "basePath",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "assetType",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "needsPostProcessing",
      type: FieldType.BOOLEAN,
      id: 5,
    },
  ],
};


/**
 * Schema for UnitMapping message
 */
export const UnitMappingSchema: MessageSchema = {
  name: "UnitMapping",
  fields: [
    {
      name: "old",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "name",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "image",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "description",
      type: FieldType.STRING,
      id: 4,
    },
  ],
};


/**
 * Schema for TerrainMapping message
 */
export const TerrainMappingSchema: MessageSchema = {
  name: "TerrainMapping",
  fields: [
    {
      name: "old",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "name",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "image",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "description",
      type: FieldType.STRING,
      id: 4,
    },
  ],
};


/**
 * Schema for ThemeManifest message
 */
export const ThemeManifestSchema: MessageSchema = {
  name: "ThemeManifest",
  fields: [
    {
      name: "themeInfo",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "weewar.v1.ThemeInfo",
    },
    {
      name: "units",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "terrains",
      type: FieldType.STRING,
      id: 3,
    },
  ],
};


/**
 * Schema for PlayerColor message
 */
export const PlayerColorSchema: MessageSchema = {
  name: "PlayerColor",
  fields: [
    {
      name: "primary",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "secondary",
      type: FieldType.STRING,
      id: 2,
    },
  ],
};


/**
 * Schema for AssetResult message
 */
export const AssetResultSchema: MessageSchema = {
  name: "AssetResult",
  fields: [
    {
      name: "type",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "data",
      type: FieldType.STRING,
      id: 2,
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
      type: FieldType.STRING,
      id: 1,
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
      type: FieldType.STRING,
      id: 2,
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
      type: FieldType.STRING,
      id: 1,
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
      type: FieldType.STRING,
      id: 3,
    },
  ],
};



/**
 * Package-scoped schema registry for weewar.v1
 */
export const weewar_v1SchemaRegistry: Record<string, MessageSchema> = {
  "weewar.v1.User": UserSchema,
  "weewar.v1.Pagination": PaginationSchema,
  "weewar.v1.PaginationResponse": PaginationResponseSchema,
  "weewar.v1.World": WorldSchema,
  "weewar.v1.WorldData": WorldDataSchema,
  "weewar.v1.Tile": TileSchema,
  "weewar.v1.Unit": UnitSchema,
  "weewar.v1.AttackRecord": AttackRecordSchema,
  "weewar.v1.TerrainDefinition": TerrainDefinitionSchema,
  "weewar.v1.UnitDefinition": UnitDefinitionSchema,
  "weewar.v1.TerrainUnitProperties": TerrainUnitPropertiesSchema,
  "weewar.v1.UnitUnitProperties": UnitUnitPropertiesSchema,
  "weewar.v1.DamageDistribution": DamageDistributionSchema,
  "weewar.v1.DamageRange": DamageRangeSchema,
  "weewar.v1.RulesEngine": RulesEngineSchema,
  "weewar.v1.Game": GameSchema,
  "weewar.v1.GameConfiguration": GameConfigurationSchema,
  "weewar.v1.GamePlayer": GamePlayerSchema,
  "weewar.v1.GameTeam": GameTeamSchema,
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
  "weewar.v1.AllPaths": AllPathsSchema,
  "weewar.v1.PathEdge": PathEdgeSchema,
  "weewar.v1.Path": PathSchema,
  "weewar.v1.GameOption": GameOptionSchema,
  "weewar.v1.EndTurnOption": EndTurnOptionSchema,
  "weewar.v1.MoveOption": MoveOptionSchema,
  "weewar.v1.AttackOption": AttackOptionSchema,
  "weewar.v1.BuildUnitOption": BuildUnitOptionSchema,
  "weewar.v1.CaptureBuildingOption": CaptureBuildingOptionSchema,
  "weewar.v1.SimulateAttackRequest": SimulateAttackRequestSchema,
  "weewar.v1.SimulateAttackResponse": SimulateAttackResponseSchema,
  "weewar.v1.EmptyRequest": EmptyRequestSchema,
  "weewar.v1.EmptyResponse": EmptyResponseSchema,
  "weewar.v1.SetContentRequest": SetContentRequestSchema,
  "weewar.v1.SetContentResponse": SetContentResponseSchema,
  "weewar.v1.LogMessageRequest": LogMessageRequestSchema,
  "weewar.v1.LogMessageResponse": LogMessageResponseSchema,
  "weewar.v1.SetGameStateRequest": SetGameStateRequestSchema,
  "weewar.v1.SetGameStateResponse": SetGameStateResponseSchema,
  "weewar.v1.UpdateGameStatusRequest": UpdateGameStatusRequestSchema,
  "weewar.v1.UpdateGameStatusResponse": UpdateGameStatusResponseSchema,
  "weewar.v1.SetTileAtRequest": SetTileAtRequestSchema,
  "weewar.v1.SetTileAtResponse": SetTileAtResponseSchema,
  "weewar.v1.SetUnitAtRequest": SetUnitAtRequestSchema,
  "weewar.v1.SetUnitAtResponse": SetUnitAtResponseSchema,
  "weewar.v1.RemoveTileAtRequest": RemoveTileAtRequestSchema,
  "weewar.v1.RemoveTileAtResponse": RemoveTileAtResponseSchema,
  "weewar.v1.RemoveUnitAtRequest": RemoveUnitAtRequestSchema,
  "weewar.v1.RemoveUnitAtResponse": RemoveUnitAtResponseSchema,
  "weewar.v1.ShowHighlightsRequest": ShowHighlightsRequestSchema,
  "weewar.v1.ShowHighlightsResponse": ShowHighlightsResponseSchema,
  "weewar.v1.HighlightSpec": HighlightSpecSchema,
  "weewar.v1.ClearHighlightsRequest": ClearHighlightsRequestSchema,
  "weewar.v1.ClearHighlightsResponse": ClearHighlightsResponseSchema,
  "weewar.v1.ShowPathRequest": ShowPathRequestSchema,
  "weewar.v1.ShowPathResponse": ShowPathResponseSchema,
  "weewar.v1.ClearPathsRequest": ClearPathsRequestSchema,
  "weewar.v1.ClearPathsResponse": ClearPathsResponseSchema,
  "weewar.v1.TurnOptionClickedRequest": TurnOptionClickedRequestSchema,
  "weewar.v1.TurnOptionClickedResponse": TurnOptionClickedResponseSchema,
  "weewar.v1.SceneClickedRequest": SceneClickedRequestSchema,
  "weewar.v1.SceneClickedResponse": SceneClickedResponseSchema,
  "weewar.v1.EndTurnButtonClickedRequest": EndTurnButtonClickedRequestSchema,
  "weewar.v1.EndTurnButtonClickedResponse": EndTurnButtonClickedResponseSchema,
  "weewar.v1.InitializeGameRequest": InitializeGameRequestSchema,
  "weewar.v1.InitializeGameResponse": InitializeGameResponseSchema,
  "weewar.v1.ThemeInfo": ThemeInfoSchema,
  "weewar.v1.UnitMapping": UnitMappingSchema,
  "weewar.v1.TerrainMapping": TerrainMappingSchema,
  "weewar.v1.ThemeManifest": ThemeManifestSchema,
  "weewar.v1.PlayerColor": PlayerColorSchema,
  "weewar.v1.AssetResult": AssetResultSchema,
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
 * Schema registry instance for weewar.v1 package with utility methods
 * Extends BaseSchemaRegistry with package-specific schema data
 */
// Schema utility functions (now inherited from BaseSchemaRegistry in runtime package)
// Creating instance with package-specific schema registry
const registryInstance = new BaseSchemaRegistry(weewar_v1SchemaRegistry);

export const getSchema = registryInstance.getSchema.bind(registryInstance);
export const getFieldSchema = registryInstance.getFieldSchema.bind(registryInstance);
export const getFieldSchemaById = registryInstance.getFieldSchemaById.bind(registryInstance);
export const isOneofField = registryInstance.isOneofField.bind(registryInstance);
export const getOneofFields = registryInstance.getOneofFields.bind(registryInstance);