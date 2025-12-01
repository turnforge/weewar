// Generated TypeScript interfaces from proto file
// DO NOT EDIT - This file is auto-generated

import { Any } from "@bufbuild/protobuf/wkt";



/**
 * IndexStateGORM is the GORM representation for IndexState
 */
export interface IndexStateGORM {
}


/**
 * IndexRecordsLROGORM is the GORM representation for IndexRecordsLRO
 */
export interface IndexRecordsLROGORM {
}



export interface IndexInfoGORM {
}



export interface TileGORM {
}



export interface CrossingGORM {
}



export interface UnitGORM {
}



export interface AttackRecordGORM {
}



export interface WorldGORM {
  id: string;
  /** Tags as JSON for cross-DB compatibility */
  tags: string[];
  /** PreviewUrls as JSON for cross-DB compatibility */
  previewUrls: string[];
  /** DefaultGameConfig as JSON for cross-DB compatibility
 SearchIndexInfo embedded */
  searchIndexInfo?: IndexInfoGORM;
}



export interface WorldDataGORM {
  worldId: string;
  /** Units as JSON for cross-DB compatibility */
  crossings: Record<string, CrossingGORM>;
  /** ScreenshotIndexInfo embedded */
  screenshotIndexInfo?: IndexInfoGORM;
  /** Tiles as JSON for cross-DB compatibility */
  tilesMap: Record<string, TileGORM>;
  /** Units as JSON for cross-DB compatibility */
  unitsMap: Record<string, UnitGORM>;
}


/**
 * Describes a game and its metadata
 */
export interface GameGORM {
  id: string;
  /** Tags as JSON for cross-DB compatibility */
  tags: string[];
  /** PreviewUrls as JSON for cross-DB compatibility */
  previewUrls: string[];
  /** SearchIndexInfo embedded */
  searchIndexInfo?: IndexInfoGORM;
}



export interface GameConfigurationGORM {
  /** IncomeConfigs embedded */
  incomeConfigs?: IncomeConfigGORM;
  /** Settings as foreign key relationship */
  settings?: GameSettingsGORM;
}



export interface IncomeConfigGORM {
}



export interface GamePlayerGORM {
}



export interface GameTeamGORM {
}



export interface GameSettingsGORM {
  /** AllowedUnits as JSON for cross-DB compatibility */
  allowedUnits: number[];
}


/**
 * GameWorldDataGORM is same as WorldDataGORM but without the
 primary key so it can be embedded
 */
export interface GameWorldDataGORM {
  /** ScreenshotIndexInfo embedded */
  screenshotIndexInfo?: IndexInfoGORM;
  /** Units as JSON for cross-DB compatibility */
  crossings: Record<string, CrossingGORM>;
  /** Tiles as JSON for cross-DB compatibility */
  tilesMap: Record<string, TileGORM>;
  /** Units as JSON for cross-DB compatibility */
  unitsMap: Record<string, UnitGORM>;
}


/**
 * Holds the game's Active/Current state (eg world state)
 */
export interface GameStateGORM {
  gameId: string;
  /** ScreenshotIndexInfo embedded */
  worldData?: GameWorldDataGORM;
}


/**
 * Holds the game's move history (can be used as a replay log)
 */
export interface GameMoveHistoryGORM {
}


/**
 * A move group - we can allow X moves in one "tick"
 */
export interface GameMoveGroupGORM {
}


/**
 * *
 Represents a single move which can be one of many actions in the game
 */
export interface GameMoveGORM {
  gameId: string;
  groupNumber: string;
  moveNumber: number;
  /** Version number for optimistic locking */
  version: number;
  /** Field named "move_type" matches the oneof name in source
 This automatically skips all oneof members (move_unit, attack_unit, end_turn, build_unit) */
  moveType?: Any;
  changes?: Any[];
}

