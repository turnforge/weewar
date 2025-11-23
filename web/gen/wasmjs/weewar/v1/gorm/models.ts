import { Any } from "@bufbuild/protobuf/wkt";


import { IndexStateGORM as IndexStateGORMInterface, IndexRecordsLROGORM as IndexRecordsLROGORMInterface, IndexInfoGORM as IndexInfoGORMInterface, TileGORM as TileGORMInterface, UnitGORM as UnitGORMInterface, AttackRecordGORM as AttackRecordGORMInterface, WorldGORM as WorldGORMInterface, WorldDataGORM as WorldDataGORMInterface, GameGORM as GameGORMInterface, GameConfigurationGORM as GameConfigurationGORMInterface, IncomeConfigGORM as IncomeConfigGORMInterface, GamePlayerGORM as GamePlayerGORMInterface, GameTeamGORM as GameTeamGORMInterface, GameSettingsGORM as GameSettingsGORMInterface, GameStateGORM as GameStateGORMInterface, GameMoveHistoryGORM as GameMoveHistoryGORMInterface, GameMoveGroupGORM as GameMoveGroupGORMInterface, GameMoveGORM as GameMoveGORMInterface } from "./interfaces";




/**
 * IndexStateGORM is the GORM representation for IndexState
 */
export class IndexStateGORM implements IndexStateGORMInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.IndexStateGORM";
  readonly __MESSAGE_TYPE = IndexStateGORM.MESSAGE_TYPE;


  
}


/**
 * IndexRecordsLROGORM is the GORM representation for IndexRecordsLRO
 */
export class IndexRecordsLROGORM implements IndexRecordsLROGORMInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.IndexRecordsLROGORM";
  readonly __MESSAGE_TYPE = IndexRecordsLROGORM.MESSAGE_TYPE;


  
}



export class IndexInfoGORM implements IndexInfoGORMInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.IndexInfoGORM";
  readonly __MESSAGE_TYPE = IndexInfoGORM.MESSAGE_TYPE;


  
}



export class TileGORM implements TileGORMInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.TileGORM";
  readonly __MESSAGE_TYPE = TileGORM.MESSAGE_TYPE;


  
}



export class UnitGORM implements UnitGORMInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UnitGORM";
  readonly __MESSAGE_TYPE = UnitGORM.MESSAGE_TYPE;


  
}



export class AttackRecordGORM implements AttackRecordGORMInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.AttackRecordGORM";
  readonly __MESSAGE_TYPE = AttackRecordGORM.MESSAGE_TYPE;


  
}



export class WorldGORM implements WorldGORMInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.WorldGORM";
  readonly __MESSAGE_TYPE = WorldGORM.MESSAGE_TYPE;

  id: string = "";
  /** Tags as JSON for cross-DB compatibility */
  tags: string[] = [];
  /** PreviewUrls as JSON for cross-DB compatibility */
  previewUrls: string[] = [];
  /** DefaultGameConfig as JSON for cross-DB compatibility
 ScreenshotIndexInfo embedded */
  screenshotIndexInfo?: IndexInfoGORM;
  /** SearchIndexInfo embedded */
  searchIndexInfo?: IndexInfoGORM;

  
}



export class WorldDataGORM implements WorldDataGORMInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.WorldDataGORM";
  readonly __MESSAGE_TYPE = WorldDataGORM.MESSAGE_TYPE;

  /** Tiles as JSON for cross-DB compatibility */
  tiles: TileGORM[] = [];
  worldId: string = "";
  /** Units as JSON for cross-DB compatibility */
  units: UnitGORM[] = [];

  
}


/**
 * Describes a game and its metadata
 */
export class GameGORM implements GameGORMInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameGORM";
  readonly __MESSAGE_TYPE = GameGORM.MESSAGE_TYPE;

  id: string = "";
  /** Tags as JSON for cross-DB compatibility */
  tags: string[] = [];
  /** PreviewUrls as JSON for cross-DB compatibility */
  previewUrls: string[] = [];
  /** ScreenshotIndexInfo embedded */
  screenshotIndexInfo?: IndexInfoGORM;
  /** SearchIndexInfo embedded */
  searchIndexInfo?: IndexInfoGORM;

  
}



export class GameConfigurationGORM implements GameConfigurationGORMInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameConfigurationGORM";
  readonly __MESSAGE_TYPE = GameConfigurationGORM.MESSAGE_TYPE;

  /** IncomeConfigs embedded */
  incomeConfigs?: IncomeConfigGORM;
  /** Settings as foreign key relationship */
  settings?: GameSettingsGORM;

  
}



export class IncomeConfigGORM implements IncomeConfigGORMInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.IncomeConfigGORM";
  readonly __MESSAGE_TYPE = IncomeConfigGORM.MESSAGE_TYPE;


  
}



export class GamePlayerGORM implements GamePlayerGORMInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GamePlayerGORM";
  readonly __MESSAGE_TYPE = GamePlayerGORM.MESSAGE_TYPE;


  
}



export class GameTeamGORM implements GameTeamGORMInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameTeamGORM";
  readonly __MESSAGE_TYPE = GameTeamGORM.MESSAGE_TYPE;


  
}



export class GameSettingsGORM implements GameSettingsGORMInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameSettingsGORM";
  readonly __MESSAGE_TYPE = GameSettingsGORM.MESSAGE_TYPE;

  /** AllowedUnits as JSON for cross-DB compatibility */
  allowedUnits: number[] = [];

  
}


/**
 * Holds the game's Active/Current state (eg world state)
 */
export class GameStateGORM implements GameStateGORMInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameStateGORM";
  readonly __MESSAGE_TYPE = GameStateGORM.MESSAGE_TYPE;

  gameId: string = "";

  
}


/**
 * Holds the game's move history (can be used as a replay log)
 */
export class GameMoveHistoryGORM implements GameMoveHistoryGORMInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameMoveHistoryGORM";
  readonly __MESSAGE_TYPE = GameMoveHistoryGORM.MESSAGE_TYPE;


  
}


/**
 * A move group - we can allow X moves in one "tick"
 */
export class GameMoveGroupGORM implements GameMoveGroupGORMInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameMoveGroupGORM";
  readonly __MESSAGE_TYPE = GameMoveGroupGORM.MESSAGE_TYPE;


  
}


/**
 * *
 Represents a single move which can be one of many actions in the game
 */
export class GameMoveGORM implements GameMoveGORMInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameMoveGORM";
  readonly __MESSAGE_TYPE = GameMoveGORM.MESSAGE_TYPE;

  gameId: string = "";
  groupNumber: string = "";
  moveNumber: number = 0;
  /** Field named "move_type" matches the oneof name in source
 This automatically skips all oneof members (move_unit, attack_unit, end_turn, build_unit) */
  moveType?: Any;
  changes: Any[] = [];

  
}


