
// Generated TypeScript schemas from proto file
// DO NOT EDIT - This file is auto-generated

import { FieldType, FieldSchema, MessageSchema, BaseSchemaRegistry } from "@protoc-gen-go-wasmjs/runtime";


/**
 * Schema for IndexStateGORM message
 */
export const IndexStateGORMSchema: MessageSchema = {
  name: "IndexStateGORM",
  fields: [
  ],
};


/**
 * Schema for IndexRecordsLROGORM message
 */
export const IndexRecordsLROGORMSchema: MessageSchema = {
  name: "IndexRecordsLROGORM",
  fields: [
  ],
};


/**
 * Schema for IndexInfoGORM message
 */
export const IndexInfoGORMSchema: MessageSchema = {
  name: "IndexInfoGORM",
  fields: [
  ],
};


/**
 * Schema for TileGORM message
 */
export const TileGORMSchema: MessageSchema = {
  name: "TileGORM",
  fields: [
  ],
};


/**
 * Schema for CrossingGORM message
 */
export const CrossingGORMSchema: MessageSchema = {
  name: "CrossingGORM",
  fields: [
  ],
};


/**
 * Schema for UnitGORM message
 */
export const UnitGORMSchema: MessageSchema = {
  name: "UnitGORM",
  fields: [
  ],
};


/**
 * Schema for AttackRecordGORM message
 */
export const AttackRecordGORMSchema: MessageSchema = {
  name: "AttackRecordGORM",
  fields: [
  ],
};


/**
 * Schema for WorldGORM message
 */
export const WorldGORMSchema: MessageSchema = {
  name: "WorldGORM",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "tags",
      type: FieldType.REPEATED,
      id: 7,
      repeated: true,
    },
    {
      name: "previewUrls",
      type: FieldType.REPEATED,
      id: 11,
      repeated: true,
    },
    {
      name: "searchIndexInfo",
      type: FieldType.MESSAGE,
      id: 13,
      messageType: "weewar.v1.IndexInfoGORM",
    },
  ],
};


/**
 * Schema for WorldDataGORM message
 */
export const WorldDataGORMSchema: MessageSchema = {
  name: "WorldDataGORM",
  fields: [
    {
      name: "worldId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "crossings",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "screenshotIndexInfo",
      type: FieldType.MESSAGE,
      id: 5,
      messageType: "weewar.v1.IndexInfoGORM",
    },
    {
      name: "tilesMap",
      type: FieldType.STRING,
      id: 6,
    },
    {
      name: "unitsMap",
      type: FieldType.STRING,
      id: 7,
    },
  ],
};


/**
 * Schema for GameGORM message
 */
export const GameGORMSchema: MessageSchema = {
  name: "GameGORM",
  fields: [
    {
      name: "id",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "tags",
      type: FieldType.REPEATED,
      id: 7,
      repeated: true,
    },
    {
      name: "previewUrls",
      type: FieldType.REPEATED,
      id: 11,
      repeated: true,
    },
    {
      name: "searchIndexInfo",
      type: FieldType.MESSAGE,
      id: 13,
      messageType: "weewar.v1.IndexInfoGORM",
    },
  ],
};


/**
 * Schema for GameConfigurationGORM message
 */
export const GameConfigurationGORMSchema: MessageSchema = {
  name: "GameConfigurationGORM",
  fields: [
    {
      name: "incomeConfigs",
      type: FieldType.MESSAGE,
      id: 3,
      messageType: "weewar.v1.IncomeConfigGORM",
    },
    {
      name: "settings",
      type: FieldType.MESSAGE,
      id: 4,
      messageType: "weewar.v1.GameSettingsGORM",
    },
  ],
};


/**
 * Schema for IncomeConfigGORM message
 */
export const IncomeConfigGORMSchema: MessageSchema = {
  name: "IncomeConfigGORM",
  fields: [
  ],
};


/**
 * Schema for GamePlayerGORM message
 */
export const GamePlayerGORMSchema: MessageSchema = {
  name: "GamePlayerGORM",
  fields: [
  ],
};


/**
 * Schema for GameTeamGORM message
 */
export const GameTeamGORMSchema: MessageSchema = {
  name: "GameTeamGORM",
  fields: [
  ],
};


/**
 * Schema for GameSettingsGORM message
 */
export const GameSettingsGORMSchema: MessageSchema = {
  name: "GameSettingsGORM",
  fields: [
    {
      name: "allowedUnits",
      type: FieldType.REPEATED,
      id: 1,
      repeated: true,
    },
  ],
};


/**
 * Schema for GameWorldDataGORM message
 */
export const GameWorldDataGORMSchema: MessageSchema = {
  name: "GameWorldDataGORM",
  fields: [
    {
      name: "screenshotIndexInfo",
      type: FieldType.MESSAGE,
      id: 4,
      messageType: "weewar.v1.IndexInfoGORM",
    },
    {
      name: "crossings",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "tilesMap",
      type: FieldType.STRING,
      id: 6,
    },
    {
      name: "unitsMap",
      type: FieldType.STRING,
      id: 7,
    },
  ],
};


/**
 * Schema for GameStateGORM message
 */
export const GameStateGORMSchema: MessageSchema = {
  name: "GameStateGORM",
  fields: [
    {
      name: "gameId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "worldData",
      type: FieldType.MESSAGE,
      id: 4,
      messageType: "weewar.v1.GameWorldDataGORM",
    },
  ],
};


/**
 * Schema for GameMoveHistoryGORM message
 */
export const GameMoveHistoryGORMSchema: MessageSchema = {
  name: "GameMoveHistoryGORM",
  fields: [
  ],
};


/**
 * Schema for GameMoveGroupGORM message
 */
export const GameMoveGroupGORMSchema: MessageSchema = {
  name: "GameMoveGroupGORM",
  fields: [
  ],
};


/**
 * Schema for GameMoveGORM message
 */
export const GameMoveGORMSchema: MessageSchema = {
  name: "GameMoveGORM",
  fields: [
    {
      name: "gameId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "groupNumber",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "moveNumber",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "version",
      type: FieldType.NUMBER,
      id: 4,
    },
    {
      name: "moveType",
      type: FieldType.MESSAGE,
      id: 5,
      messageType: "google.protobuf.Any",
    },
    {
      name: "changes",
      type: FieldType.MESSAGE,
      id: 6,
      messageType: "google.protobuf.Any",
      repeated: true,
    },
  ],
};



/**
 * Package-scoped schema registry for weewar.v1
 */
export const weewar_v1SchemaRegistry: Record<string, MessageSchema> = {
  "weewar.v1.IndexStateGORM": IndexStateGORMSchema,
  "weewar.v1.IndexRecordsLROGORM": IndexRecordsLROGORMSchema,
  "weewar.v1.IndexInfoGORM": IndexInfoGORMSchema,
  "weewar.v1.TileGORM": TileGORMSchema,
  "weewar.v1.CrossingGORM": CrossingGORMSchema,
  "weewar.v1.UnitGORM": UnitGORMSchema,
  "weewar.v1.AttackRecordGORM": AttackRecordGORMSchema,
  "weewar.v1.WorldGORM": WorldGORMSchema,
  "weewar.v1.WorldDataGORM": WorldDataGORMSchema,
  "weewar.v1.GameGORM": GameGORMSchema,
  "weewar.v1.GameConfigurationGORM": GameConfigurationGORMSchema,
  "weewar.v1.IncomeConfigGORM": IncomeConfigGORMSchema,
  "weewar.v1.GamePlayerGORM": GamePlayerGORMSchema,
  "weewar.v1.GameTeamGORM": GameTeamGORMSchema,
  "weewar.v1.GameSettingsGORM": GameSettingsGORMSchema,
  "weewar.v1.GameWorldDataGORM": GameWorldDataGORMSchema,
  "weewar.v1.GameStateGORM": GameStateGORMSchema,
  "weewar.v1.GameMoveHistoryGORM": GameMoveHistoryGORMSchema,
  "weewar.v1.GameMoveGroupGORM": GameMoveGroupGORMSchema,
  "weewar.v1.GameMoveGORM": GameMoveGORMSchema,
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