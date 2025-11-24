import { Any, FieldMask, Timestamp } from "@bufbuild/protobuf/wkt";


import { IndexInfo as IndexInfoInterface, User as UserInterface, Pagination as PaginationInterface, PaginationResponse as PaginationResponseInterface, World as WorldInterface, WorldData as WorldDataInterface, Tile as TileInterface, Unit as UnitInterface, AttackRecord as AttackRecordInterface, TerrainDefinition as TerrainDefinitionInterface, UnitDefinition as UnitDefinitionInterface, TerrainUnitProperties as TerrainUnitPropertiesInterface, UnitUnitProperties as UnitUnitPropertiesInterface, DamageDistribution as DamageDistributionInterface, DamageRange as DamageRangeInterface, RulesEngine as RulesEngineInterface, Game as GameInterface, GameConfiguration as GameConfigurationInterface, IncomeConfig as IncomeConfigInterface, GamePlayer as GamePlayerInterface, GameTeam as GameTeamInterface, GameSettings as GameSettingsInterface, GameState as GameStateInterface, GameMoveHistory as GameMoveHistoryInterface, GameMoveGroup as GameMoveGroupInterface, GameMove as GameMoveInterface, MoveUnitAction as MoveUnitActionInterface, AttackUnitAction as AttackUnitActionInterface, BuildUnitAction as BuildUnitActionInterface, CaptureBuildingAction as CaptureBuildingActionInterface, EndTurnAction as EndTurnActionInterface, WorldChange as WorldChangeInterface, UnitMovedChange as UnitMovedChangeInterface, UnitDamagedChange as UnitDamagedChangeInterface, UnitKilledChange as UnitKilledChangeInterface, PlayerChangedChange as PlayerChangedChangeInterface, UnitBuiltChange as UnitBuiltChangeInterface, CoinsChangedChange as CoinsChangedChangeInterface, AllPaths as AllPathsInterface, PathEdge as PathEdgeInterface, Path as PathInterface, ListGamesRequest as ListGamesRequestInterface, ListGamesResponse as ListGamesResponseInterface, GetGameRequest as GetGameRequestInterface, GetGameResponse as GetGameResponseInterface, GetGameContentRequest as GetGameContentRequestInterface, GetGameContentResponse as GetGameContentResponseInterface, UpdateGameRequest as UpdateGameRequestInterface, UpdateGameResponse as UpdateGameResponseInterface, DeleteGameRequest as DeleteGameRequestInterface, DeleteGameResponse as DeleteGameResponseInterface, GetGamesRequest as GetGamesRequestInterface, GetGamesResponse as GetGamesResponseInterface, CreateGameRequest as CreateGameRequestInterface, CreateGameResponse as CreateGameResponseInterface, ProcessMovesRequest as ProcessMovesRequestInterface, ProcessMovesResponse as ProcessMovesResponseInterface, GetGameStateRequest as GetGameStateRequestInterface, GetGameStateResponse as GetGameStateResponseInterface, ListMovesRequest as ListMovesRequestInterface, ListMovesResponse as ListMovesResponseInterface, GetOptionsAtRequest as GetOptionsAtRequestInterface, GetOptionsAtResponse as GetOptionsAtResponseInterface, GameOption as GameOptionInterface, SimulateAttackRequest as SimulateAttackRequestInterface, SimulateAttackResponse as SimulateAttackResponseInterface, EmptyRequest as EmptyRequestInterface, EmptyResponse as EmptyResponseInterface, SetContentRequest as SetContentRequestInterface, SetContentResponse as SetContentResponseInterface, ShowBuildOptionsRequest as ShowBuildOptionsRequestInterface, ShowBuildOptionsResponse as ShowBuildOptionsResponseInterface, LogMessageRequest as LogMessageRequestInterface, LogMessageResponse as LogMessageResponseInterface, SetGameStateRequest as SetGameStateRequestInterface, SetGameStateResponse as SetGameStateResponseInterface, UpdateGameStatusRequest as UpdateGameStatusRequestInterface, UpdateGameStatusResponse as UpdateGameStatusResponseInterface, SetTileAtRequest as SetTileAtRequestInterface, SetTileAtResponse as SetTileAtResponseInterface, SetUnitAtRequest as SetUnitAtRequestInterface, SetUnitAtResponse as SetUnitAtResponseInterface, RemoveTileAtRequest as RemoveTileAtRequestInterface, RemoveTileAtResponse as RemoveTileAtResponseInterface, RemoveUnitAtRequest as RemoveUnitAtRequestInterface, RemoveUnitAtResponse as RemoveUnitAtResponseInterface, ShowHighlightsRequest as ShowHighlightsRequestInterface, ShowHighlightsResponse as ShowHighlightsResponseInterface, HighlightSpec as HighlightSpecInterface, ClearHighlightsRequest as ClearHighlightsRequestInterface, ClearHighlightsResponse as ClearHighlightsResponseInterface, ShowPathRequest as ShowPathRequestInterface, ShowPathResponse as ShowPathResponseInterface, ClearPathsRequest as ClearPathsRequestInterface, ClearPathsResponse as ClearPathsResponseInterface, MoveUnitRequest as MoveUnitRequestInterface, MoveUnitResponse as MoveUnitResponseInterface, HexCoord as HexCoordInterface, ShowAttackEffectRequest as ShowAttackEffectRequestInterface, SplashTarget as SplashTargetInterface, ShowAttackEffectResponse as ShowAttackEffectResponseInterface, ShowHealEffectRequest as ShowHealEffectRequestInterface, ShowHealEffectResponse as ShowHealEffectResponseInterface, ShowCaptureEffectRequest as ShowCaptureEffectRequestInterface, ShowCaptureEffectResponse as ShowCaptureEffectResponseInterface, SetAllowedPanelsRequest as SetAllowedPanelsRequestInterface, SetAllowedPanelsResponse as SetAllowedPanelsResponseInterface, IndexState as IndexStateInterface, EnsureIndexStateRequest as EnsureIndexStateRequestInterface, EnsureIndexStateResponse as EnsureIndexStateResponseInterface, GetIndexStatesRequest as GetIndexStatesRequestInterface, IndexStateList as IndexStateListInterface, GetIndexStatesResponse as GetIndexStatesResponseInterface, ListIndexStatesRequest as ListIndexStatesRequestInterface, ListIndexStatesResponse as ListIndexStatesResponseInterface, DeleteIndexStatesRequest as DeleteIndexStatesRequestInterface, DeleteIndexStatesResponse as DeleteIndexStatesResponseInterface, IndexRecord as IndexRecordInterface, IndexRecordsLRO as IndexRecordsLROInterface, CreateIndexRecordsLRORequest as CreateIndexRecordsLRORequestInterface, CreateIndexRecordsLROResponse as CreateIndexRecordsLROResponseInterface, UpdateIndexRecordsLRORequest as UpdateIndexRecordsLRORequestInterface, UpdateIndexRecordsLROResponse as UpdateIndexRecordsLROResponseInterface, GetIndexRecordsLRORequest as GetIndexRecordsLRORequestInterface, GetIndexRecordsLROResponse as GetIndexRecordsLROResponseInterface, Job as JobInterface, RepeatInfo as RepeatInfoInterface, Run as RunInterface, InitializeSingletonRequest as InitializeSingletonRequestInterface, InitializeSingletonResponse as InitializeSingletonResponseInterface, TurnOptionClickedRequest as TurnOptionClickedRequestInterface, TurnOptionClickedResponse as TurnOptionClickedResponseInterface, SceneClickedRequest as SceneClickedRequestInterface, SceneClickedResponse as SceneClickedResponseInterface, EndTurnButtonClickedRequest as EndTurnButtonClickedRequestInterface, EndTurnButtonClickedResponse as EndTurnButtonClickedResponseInterface, BuildOptionClickedRequest as BuildOptionClickedRequestInterface, BuildOptionClickedResponse as BuildOptionClickedResponseInterface, InitializeGameRequest as InitializeGameRequestInterface, InitializeGameResponse as InitializeGameResponseInterface, ThemeInfo as ThemeInfoInterface, UnitMapping as UnitMappingInterface, TerrainMapping as TerrainMappingInterface, ThemeManifest as ThemeManifestInterface, PlayerColor as PlayerColorInterface, AssetResult as AssetResultInterface, UserInfo as UserInfoInterface, ListUsersRequest as ListUsersRequestInterface, ListUsersResponse as ListUsersResponseInterface, GetUserRequest as GetUserRequestInterface, GetUserResponse as GetUserResponseInterface, GetUserContentRequest as GetUserContentRequestInterface, GetUserContentResponse as GetUserContentResponseInterface, UpdateUserRequest as UpdateUserRequestInterface, UpdateUserResponse as UpdateUserResponseInterface, DeleteUserRequest as DeleteUserRequestInterface, DeleteUserResponse as DeleteUserResponseInterface, GetUsersRequest as GetUsersRequestInterface, GetUsersResponse as GetUsersResponseInterface, CreateUserRequest as CreateUserRequestInterface, CreateUserResponse as CreateUserResponseInterface, WorldInfo as WorldInfoInterface, ListWorldsRequest as ListWorldsRequestInterface, ListWorldsResponse as ListWorldsResponseInterface, GetWorldRequest as GetWorldRequestInterface, GetWorldResponse as GetWorldResponseInterface, UpdateWorldRequest as UpdateWorldRequestInterface, UpdateWorldResponse as UpdateWorldResponseInterface, DeleteWorldRequest as DeleteWorldRequestInterface, DeleteWorldResponse as DeleteWorldResponseInterface, GetWorldsRequest as GetWorldsRequestInterface, GetWorldsResponse as GetWorldsResponseInterface, CreateWorldRequest as CreateWorldRequestInterface, CreateWorldResponse as CreateWorldResponseInterface, GameStatus, PathDirection, IndexStatus, RunState, Type } from "./interfaces";





export class IndexInfo implements IndexInfoInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.IndexInfo";
  readonly __MESSAGE_TYPE = IndexInfo.MESSAGE_TYPE;

  /** We maintain an IndexInfo for each type of "indexing" operation needed
 For example one update may change the keywords (so we need to update indexes for search)
 Another might update the "units" so we may need a new screenshot
 Each one's updates - updated and indexed timestamps separately so they can be tracked sepately */
  lastUpdatedAt?: Timestamp;
  lastIndexedAt?: Timestamp;
  needsIndexing: boolean = false;

  
}



export class User implements UserInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.User";
  readonly __MESSAGE_TYPE = User.MESSAGE_TYPE;

  createdAt?: Timestamp;
  updatedAt?: Timestamp;
  /** Unique ID for the user */
  id: string = "";
  /** Name if items have names */
  name: string = "";
  /** Description if user has a description */
  description: string = "";
  /** Some tags */
  tags: string[] = [];
  /** A possible image url */
  imageUrl: string = "";
  /** Difficulty - example attribute */
  difficulty: string = "";

  
}



export class Pagination implements PaginationInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.Pagination";
  readonly __MESSAGE_TYPE = Pagination.MESSAGE_TYPE;

  /** *
 Instead of an offset an abstract  "page" key is provided that offers
 an opaque "pointer" into some offset in a result set. */
  pageKey: string = "";
  /** *
 If a pagekey is not supported we can also support a direct integer offset
 for cases where it makes sense. */
  pageOffset: number = 0;
  /** *
 Number of results to return. */
  pageSize: number = 0;

  
}



export class PaginationResponse implements PaginationResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.PaginationResponse";
  readonly __MESSAGE_TYPE = PaginationResponse.MESSAGE_TYPE;

  /** *
 The key/pointer string that subsequent List requests should pass to
 continue the pagination. */
  nextPageKey: string = "";
  /** *
 Also support an integer offset if possible */
  nextPageOffset: number = 0;
  /** *
 Whether theere are more results. */
  hasMore: boolean = false;
  /** *
 Total number of results. */
  totalResults: number = 0;

  
}



export class World implements WorldInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.World";
  readonly __MESSAGE_TYPE = World.MESSAGE_TYPE;

  createdAt?: Timestamp;
  updatedAt?: Timestamp;
  /** Unique ID for the world */
  id: string = "";
  /** User that created the world */
  creatorId: string = "";
  /** Name if items have names */
  name: string = "";
  /** Description if world has a description */
  description: string = "";
  /** Some tags */
  tags: string[] = [];
  /** A possible image url */
  imageUrl: string = "";
  /** Difficulty - example attribute */
  difficulty: string = "";
  /** URL to screenshot/preview image (defaults to /worlds/{id}/screenshots/{screenshotName})
 Can be overridden to point to CDN or external hosting */
  previewUrls: string[] = [];
  /** Default game configs */
  defaultGameConfig?: GameConfiguration;
  searchIndexInfo?: IndexInfo;

  
}



export class WorldData implements WorldDataInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.WorldData";
  readonly __MESSAGE_TYPE = WorldData.MESSAGE_TYPE;

  /** JSON-fied tile data about what units and terrains are at each location */
  tiles: Tile[] = [];
  /** All units on the world and who they belong to */
  units: Unit[] = [];
  /** When this world data was updated (may have happened without world updating) */
  screenshotIndexInfo?: IndexInfo;
  /** We will only update if hash's are different */
  contentHash: string = "";

  
}



export class Tile implements TileInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.Tile";
  readonly __MESSAGE_TYPE = Tile.MESSAGE_TYPE;

  /** Q and R in Cubed coordinates */
  q: number = 0;
  r: number = 0;
  tileType: number = 0;
  /** Whether the tile itself belongs to a player */
  player: number = 0;
  shortcut: string = "";
  /** Keep track of turns when the move was last made and when a "top up" was last done on/for this tile.
 This helps us not having to "top up" or "reset" the stats at the end
 of each turn.  Instead as the game turn is incremented we can do a 
 lazy reset for any unit or tile where unit_or_tile.last_toppedup_turn < game.curent_turn
 So we just have to increment the game_turn and the unit is automaticaly flagged as
 needing a top up of its health/balance/movement etc */
  lastActedTurn: number = 0;
  lastToppedupTurn: number = 0;

  
}



export class Unit implements UnitInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.Unit";
  readonly __MESSAGE_TYPE = Unit.MESSAGE_TYPE;

  /** Q and R in Cubed coordinates */
  q: number = 0;
  r: number = 0;
  player: number = 0;
  unitType: number = 0;
  shortcut: string = "";
  /** Runtime state fields */
  availableHealth: number = 0;
  distanceLeft: number = 0;
  /** Keep track of turns when the move was last made and when a "top up" was last done on/for this tile.
 This helps us not having to "top up" or "reset" the stats at the end
 of each turn.  Instead as the game turn is incremented we can do a 
 lazy reset for any unit or tile where unit_or_tile.last_toppedup_turn < game.curent_turn
 So we just have to increment the game_turn and the unit is automaticaly flagged as
 needing a top up of its health/balance/movement etc */
  lastActedTurn: number = 0;
  lastToppedupTurn: number = 0;
  /** Details around wound bonus tracking for this turn */
  attacksReceivedThisTurn: number = 0;
  attackHistory: AttackRecord[] = [];
  /** Action progression tracking - index into UnitDefinition.action_order
 Indicates which step in the action sequence the unit is currently on
 Reset to 0 at turn start via TopUpUnitIfNeeded() */
  progressionStep: number = 0;
  /** When current step has pipe-separated alternatives (e.g., "attack|capture"),
 this tracks which alternative the user chose, preventing switching mid-step
 Cleared when advancing to next step */
  chosenAlternative: string = "";

  
}



export class AttackRecord implements AttackRecordInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.AttackRecord";
  readonly __MESSAGE_TYPE = AttackRecord.MESSAGE_TYPE;

  q: number = 0;
  r: number = 0;
  isRanged: boolean = false;
  turnNumber: number = 0;

  
}


/**
 * Rules engine terrain definition
 */
export class TerrainDefinition implements TerrainDefinitionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.TerrainDefinition";
  readonly __MESSAGE_TYPE = TerrainDefinition.MESSAGE_TYPE;

  id: number = 0;
  name: string = "";
  /** double base_move_cost = 3;     // Base movement cost
 double defense_bonus = 4;      // Defense bonus multiplier (0.0 to 1.0) */
  type: number = 0;
  description: string = "";
  /** How this terrain impacts */
  unitProperties: Record<number, TerrainUnitProperties> = {};
  /** List of units that can be built on this terrain */
  buildableUnitIds: number[] = [];
  incomePerTurn: number = 0;

  
}


/**
 * Rules engine unit definition
 */
export class UnitDefinition implements UnitDefinitionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UnitDefinition";
  readonly __MESSAGE_TYPE = UnitDefinition.MESSAGE_TYPE;

  id: number = 0;
  name: string = "";
  description: string = "";
  health: number = 0;
  coins: number = 0;
  movementPoints: number = 0;
  retreatPoints: number = 0;
  defense: number = 0;
  attackRange: number = 0;
  minAttackRange: number = 0;
  splashDamage: number = 0;
  terrainProperties: Record<number, TerrainUnitProperties> = {};
  properties: string[] = [];
  /** Unit classification for attack calculations */
  unitClass: string = "";
  unitTerrain: string = "";
  /** Attack table: base attack values against different unit classes
 Key format: "Light:Air", "Heavy:Land", "Stealth:Water", etc.
 Value 0 or missing key means "n/a" (cannot attack) */
  attackVsClass: Record<string, number> = {};
  /** Ordered list of allowed actions this turn
 Examples:
   ["move", "attack"] - can move then attack
   ["move", "attack|capture"] - can move then either attack or capture
   ["attack"] - can only attack (no movement)
 Default if empty: ["move", "attack|capture"] */
  actionOrder: string[] = [];
  /** How many times each action type can be performed per turn
 Key: action name, Value: max count
 Example: {"attack": 2} means can attack twice
 Default if not specified: 1 per action type */
  actionLimits: Record<string, number> = {};

  
}


/**
 * Properties that are specific to unit on a particular terrain
 */
export class TerrainUnitProperties implements TerrainUnitPropertiesInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.TerrainUnitProperties";
  readonly __MESSAGE_TYPE = TerrainUnitProperties.MESSAGE_TYPE;

  terrainId: number = 0;
  unitId: number = 0;
  movementCost: number = 0;
  healingBonus: number = 0;
  canBuild: boolean = false;
  canCapture: boolean = false;
  attackBonus: number = 0;
  defenseBonus: number = 0;
  attackRange: number = 0;
  minAttackRange: number = 0;

  
}


/**
 * Properties for unit-vs-unit combat interactions
 */
export class UnitUnitProperties implements UnitUnitPropertiesInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UnitUnitProperties";
  readonly __MESSAGE_TYPE = UnitUnitProperties.MESSAGE_TYPE;

  attackerId: number = 0;
  defenderId: number = 0;
  attackOverride?: number | undefined;
  defenseOverride?: number | undefined;
  damage?: DamageDistribution;

  
}


/**
 * Damage distribution for combat calculations
 */
export class DamageDistribution implements DamageDistributionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.DamageDistribution";
  readonly __MESSAGE_TYPE = DamageDistribution.MESSAGE_TYPE;

  minDamage: number = 0;
  maxDamage: number = 0;
  expectedDamage: number = 0;
  ranges: DamageRange[] = [];

  
}



export class DamageRange implements DamageRangeInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.DamageRange";
  readonly __MESSAGE_TYPE = DamageRange.MESSAGE_TYPE;

  minValue: number = 0;
  maxValue: number = 0;
  probability: number = 0;

  
}


/**
 * Main rules engine definition - centralized source of truth
 */
export class RulesEngine implements RulesEngineInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.RulesEngine";
  readonly __MESSAGE_TYPE = RulesEngine.MESSAGE_TYPE;

  /** Core entity definitions */
  units: Record<number, UnitDefinition> = {};
  terrains: Record<number, TerrainDefinition> = {};
  /** Centralized property definitions (source of truth)
 Key format: "terrain_id:unit_id" (e.g., "1:3" for terrain 1, unit 3) */
  terrainUnitProperties: Record<string, TerrainUnitProperties> = {};
  /** Key format: "attacker_id:defender_id" (e.g., "1:2" for unit 1 attacking unit 2) */
  unitUnitProperties: Record<string, UnitUnitProperties> = {};

  
}


/**
 * Describes a game and its metadata
 */
export class Game implements GameInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.Game";
  readonly __MESSAGE_TYPE = Game.MESSAGE_TYPE;

  createdAt?: Timestamp;
  updatedAt?: Timestamp;
  /** Unique ID for the game */
  id: string = "";
  /** User who started/created the game */
  creatorId: string = "";
  /** The world this game was created from */
  worldId: string = "";
  /** Name if items have names */
  name: string = "";
  /** Description if game has a description */
  description: string = "";
  /** Some tags */
  tags: string[] = [];
  /** A possible image url */
  imageUrl: string = "";
  /** Difficulty - example attribute */
  difficulty: string = "";
  /** Game configuration */
  config?: GameConfiguration;
  /** URL to screenshot/preview image (defaults to /games/{id}/screenshots/{screenshotName})
 Can be overridden to point to CDN or external hosting */
  previewUrls: string[] = [];
  screenshotIndexInfo?: IndexInfo;
  searchIndexInfo?: IndexInfo;

  
}



export class GameConfiguration implements GameConfigurationInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameConfiguration";
  readonly __MESSAGE_TYPE = GameConfiguration.MESSAGE_TYPE;

  /** Player configuration */
  players: GamePlayer[] = [];
  /** Team configuration */
  teams: GameTeam[] = [];
  /** Various kinds of per turn income configs */
  incomeConfigs?: IncomeConfig;
  /** Game settings */
  settings?: GameSettings;

  
}



export class IncomeConfig implements IncomeConfigInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.IncomeConfig";
  readonly __MESSAGE_TYPE = IncomeConfig.MESSAGE_TYPE;

  /** How much starting coins to give each player */
  startingCoins: number = 0;
  /** Income each player just for being in the game */
  gameIncome: number = 0;
  /** Income from each landbase per turn */
  landbaseIncome: number = 0;
  /** Income from each navalbase per turn */
  navalbaseIncome: number = 0;
  /** Income from each airport base per turn */
  airportbaseIncome: number = 0;
  /** Income from each missile silo per turn */
  missilesiloIncome: number = 0;
  /** Income from each mine per turn */
  minesIncome: number = 0;

  
}



export class GamePlayer implements GamePlayerInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GamePlayer";
  readonly __MESSAGE_TYPE = GamePlayer.MESSAGE_TYPE;

  /** Player ID (1-based) */
  playerId: number = 0;
  /** Player type */
  playerType: string = "";
  /** Player color */
  color: string = "";
  /** Team ID (0 = no team, 1+ = team number) */
  teamId: number = 0;
  /** Nickname for the player in this game */
  name: string = "";
  /** Whether play is still in the game - can this just be inferred? */
  isActive: boolean = false;
  /** How many coins the player started off with */
  startingCoins: number = 0;
  /** Player's current money/coins balance for building units */
  coins: number = 0;

  
}



export class GameTeam implements GameTeamInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameTeam";
  readonly __MESSAGE_TYPE = GameTeam.MESSAGE_TYPE;

  /** ID of the team within the game (unique to the game) */
  teamId: number = 0;
  /** Name of the team - in a game */
  name: string = "";
  /** Just a color for this team */
  color: string = "";
  /** Whether team has active players - can also be inferred */
  isActive: boolean = false;

  
}



export class GameSettings implements GameSettingsInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameSettings";
  readonly __MESSAGE_TYPE = GameSettings.MESSAGE_TYPE;

  /** List of allowed unit type IDs */
  allowedUnits: number[] = [];
  /** Turn time limit in seconds (0 = no limit) */
  turnTimeLimit: number = 0;
  /** Team mode */
  teamMode: string = "";
  /** Maximum number of turns (0 = unlimited) */
  maxTurns: number = 0;

  
}


/**
 * Holds the game's Active/Current state (eg world state)
 */
export class GameState implements GameStateInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameState";
  readonly __MESSAGE_TYPE = GameState.MESSAGE_TYPE;

  updatedAt?: Timestamp;
  /** ID of the game whos state is being tracked */
  gameId: string = "";
  turnCounter: number = 0;
  currentPlayer: number = 0;
  /** Current world state */
  worldData?: WorldData;
  /** Current state hash for validation */
  stateHash: string = "";
  /** Version number for optimistic locking */
  version: number = 0;
  status: GameStatus = GameStatus.GAME_STATUS_UNSPECIFIED;
  /** Only set after a win has been possible */
  finished: boolean = false;
  winningPlayer: number = 0;
  winningTeam: number = 0;

  
}


/**
 * Holds the game's move history (can be used as a replay log)
 */
export class GameMoveHistory implements GameMoveHistoryInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameMoveHistory";
  readonly __MESSAGE_TYPE = GameMoveHistory.MESSAGE_TYPE;

  /** Move history for the game */
  gameId: string = "";
  /** Each entry in our history is a "group" of moves */
  groups: GameMoveGroup[] = [];

  
}


/**
 * A move group - we can allow X moves in one "tick"
 */
export class GameMoveGroup implements GameMoveGroupInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameMoveGroup";
  readonly __MESSAGE_TYPE = GameMoveGroup.MESSAGE_TYPE;

  /** When the moves happened (or were submitted) */
  startedAt?: Timestamp;
  endedAt?: Timestamp;
  /** *
 List of moves to add - */
  moves: GameMove[] = [];

  
}


/**
 * *
 Represents a single move which can be one of many actions in the game
 */
export class GameMove implements GameMoveInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameMove";
  readonly __MESSAGE_TYPE = GameMove.MESSAGE_TYPE;

  player: number = 0;
  timestamp?: Timestamp;
  moveUnit?: MoveUnitAction;
  attackUnit?: AttackUnitAction;
  endTurn?: EndTurnAction;
  buildUnit?: BuildUnitAction;
  /** A monotonically increasing and unique (within the game) sequence number for the move
 This is generated by the server */
  sequenceNum: number = 0;
  /** Whether the result is permenant and can be undone.
 Just moving a unit for example is not permanent, but attacking a unit
 would be (ie a player cannot undo it).  This is also determined by the server/validator */
  isPermanent: boolean = false;
  /** The corresponding "result" for the move.  This can be "proposed" or can be evaluated.
 Keeping this colocated with the Move for consistency and simplicity */
  changes: WorldChange[] = [];

  
}


/**
 * *
 Move unit from one position to another
 */
export class MoveUnitAction implements MoveUnitActionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.MoveUnitAction";
  readonly __MESSAGE_TYPE = MoveUnitAction.MESSAGE_TYPE;

  fromQ: number = 0;
  fromR: number = 0;
  toQ: number = 0;
  toR: number = 0;
  /** Optional fields that can be used for showing move options as well as debugging */
  movementCost: number = 0;
  /** Debug fields */
  reconstructedPath?: Path;

  
}


/**
 * *
 Attack with one unit against another
 */
export class AttackUnitAction implements AttackUnitActionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.AttackUnitAction";
  readonly __MESSAGE_TYPE = AttackUnitAction.MESSAGE_TYPE;

  attackerQ: number = 0;
  attackerR: number = 0;
  defenderQ: number = 0;
  defenderR: number = 0;
  /** Optional fields for presenting during "options" and debugging */
  targetUnitType: number = 0;
  targetUnitHealth: number = 0;
  canAttack: boolean = false;
  damageEstimate: number = 0;

  
}


/**
 * *
 An action to build a unit (at a city tile)
 */
export class BuildUnitAction implements BuildUnitActionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.BuildUnitAction";
  readonly __MESSAGE_TYPE = BuildUnitAction.MESSAGE_TYPE;

  q: number = 0;
  r: number = 0;
  unitType: number = 0;
  cost: number = 0;

  
}


/**
 * *
 A move where a unit can capture a building
 */
export class CaptureBuildingAction implements CaptureBuildingActionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.CaptureBuildingAction";
  readonly __MESSAGE_TYPE = CaptureBuildingAction.MESSAGE_TYPE;

  q: number = 0;
  r: number = 0;
  tileType: number = 0;

  
}


/**
 * *
 End current player's turn
 */
export class EndTurnAction implements EndTurnActionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.EndTurnAction";
  readonly __MESSAGE_TYPE = EndTurnAction.MESSAGE_TYPE;


  
}


/**
 * *
 Represents a change to the game world
 */
export class WorldChange implements WorldChangeInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.WorldChange";
  readonly __MESSAGE_TYPE = WorldChange.MESSAGE_TYPE;

  unitMoved?: UnitMovedChange;
  unitDamaged?: UnitDamagedChange;
  unitKilled?: UnitKilledChange;
  playerChanged?: PlayerChangedChange;
  unitBuilt?: UnitBuiltChange;
  coinsChanged?: CoinsChangedChange;

  
}


/**
 * *
 A unit moved from one position to another
 */
export class UnitMovedChange implements UnitMovedChangeInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UnitMovedChange";
  readonly __MESSAGE_TYPE = UnitMovedChange.MESSAGE_TYPE;

  /** Complete unit state before the move */
  previousUnit?: Unit;
  /** Complete unit state after the move (includes updated position, distanceLeft, etc.) */
  updatedUnit?: Unit;

  
}


/**
 * *
 A unit took damage
 */
export class UnitDamagedChange implements UnitDamagedChangeInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UnitDamagedChange";
  readonly __MESSAGE_TYPE = UnitDamagedChange.MESSAGE_TYPE;

  /** Complete unit state before taking damage */
  previousUnit?: Unit;
  /** Complete unit state after taking damage */
  updatedUnit?: Unit;

  
}


/**
 * *
 A unit was killed
 */
export class UnitKilledChange implements UnitKilledChangeInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UnitKilledChange";
  readonly __MESSAGE_TYPE = UnitKilledChange.MESSAGE_TYPE;

  /** Complete unit state before being killed */
  previousUnit?: Unit;

  
}


/**
 * *
 Active player changed
 */
export class PlayerChangedChange implements PlayerChangedChangeInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.PlayerChangedChange";
  readonly __MESSAGE_TYPE = PlayerChangedChange.MESSAGE_TYPE;

  previousPlayer: number = 0;
  newPlayer: number = 0;
  previousTurn: number = 0;
  newTurn: number = 0;
  /** Units that had their movement/health reset for the new turn */
  resetUnits: Unit[] = [];

  
}


/**
 * *
 A new unit was built at a tile
 */
export class UnitBuiltChange implements UnitBuiltChangeInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UnitBuiltChange";
  readonly __MESSAGE_TYPE = UnitBuiltChange.MESSAGE_TYPE;

  /** The newly created unit */
  unit?: Unit;
  /** Tile coordinates where unit was built */
  tileQ: number = 0;
  tileR: number = 0;
  /** Cost in coins */
  coinsCost: number = 0;
  /** Player's remaining coins after build */
  playerCoins: number = 0;

  
}


/**
 * *
 A player's coin balance changed
 */
export class CoinsChangedChange implements CoinsChangedChangeInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.CoinsChangedChange";
  readonly __MESSAGE_TYPE = CoinsChangedChange.MESSAGE_TYPE;

  /** Which player's coins changed */
  playerId: number = 0;
  /** Previous coin balance */
  previousCoins: number = 0;
  /** New coin balance */
  newCoins: number = 0;
  /** Reason for change: "build", "income", "repair", etc. */
  reason: string = "";

  
}


/**
 * Compact representation of all reachable paths from a source
 */
export class AllPaths implements AllPathsInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.AllPaths";
  readonly __MESSAGE_TYPE = AllPaths.MESSAGE_TYPE;

  /** Starting coordinate for all paths */
  sourceQ: number = 0;
  sourceR: number = 0;
  /** Map of edges: key is "toQ,toR" for quick parent lookup
 Each edge represents the optimal way to reach 'to' from its parent */
  edges: Record<string, PathEdge> = {};

  
}


/**
 * A single edge in a path with movement details
 */
export class PathEdge implements PathEdgeInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.PathEdge";
  readonly __MESSAGE_TYPE = PathEdge.MESSAGE_TYPE;

  fromQ: number = 0;
  fromR: number = 0;
  toQ: number = 0;
  toR: number = 0;
  movementCost: number = 0;
  totalCost: number = 0;
  terrainType: string = "";
  explanation: string = "";

  
}


/**
 * Full path from source to destination (constructed on-demand from AllPaths)
 */
export class Path implements PathInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.Path";
  readonly __MESSAGE_TYPE = Path.MESSAGE_TYPE;

  /** Edges in order from source to destination */
  edges: PathEdge[] = [];
  /** len(directions) = len(edges) - 1
 and directions[i] = direction from edge[i - 1] -> edge[i] */
  directions: PathDirection[] = [];
  /** Sum of all edge costs */
  totalCost: number = 0;

  
}


/**
 * Request messages
 */
export class ListGamesRequest implements ListGamesRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ListGamesRequest";
  readonly __MESSAGE_TYPE = ListGamesRequest.MESSAGE_TYPE;

  /** Pagination info */
  pagination?: Pagination;
  /** May be filter by owner id */
  ownerId: string = "";

  
}



export class ListGamesResponse implements ListGamesResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ListGamesResponse";
  readonly __MESSAGE_TYPE = ListGamesResponse.MESSAGE_TYPE;

  items: Game[] = [];
  pagination?: PaginationResponse;

  
}



export class GetGameRequest implements GetGameRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetGameRequest";
  readonly __MESSAGE_TYPE = GetGameRequest.MESSAGE_TYPE;

  id: string = "";
  version: string = "";

  
}



export class GetGameResponse implements GetGameResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetGameResponse";
  readonly __MESSAGE_TYPE = GetGameResponse.MESSAGE_TYPE;

  game?: Game;
  state?: GameState;
  history?: GameMoveHistory;

  
}



export class GetGameContentRequest implements GetGameContentRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetGameContentRequest";
  readonly __MESSAGE_TYPE = GetGameContentRequest.MESSAGE_TYPE;

  id: string = "";
  version: string = "";

  
}



export class GetGameContentResponse implements GetGameContentResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetGameContentResponse";
  readonly __MESSAGE_TYPE = GetGameContentResponse.MESSAGE_TYPE;

  weewarContent: string = "";
  recipeContent: string = "";
  readmeContent: string = "";

  
}



export class UpdateGameRequest implements UpdateGameRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UpdateGameRequest";
  readonly __MESSAGE_TYPE = UpdateGameRequest.MESSAGE_TYPE;

  /** Game id to modify */
  gameId: string = "";
  /** *
 Game being updated */
  newGame?: Game;
  /** New world state to save */
  newState?: GameState;
  /** History to save */
  newHistory?: GameMoveHistory;
  /** *
 Mask of fields being updated in this Game to make partial changes. */
  updateMask?: FieldMask;

  
}


/**
 * *
 The request for (partially) updating an Game.
 */
export class UpdateGameResponse implements UpdateGameResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UpdateGameResponse";
  readonly __MESSAGE_TYPE = UpdateGameResponse.MESSAGE_TYPE;

  /** *
 Game being updated */
  game?: Game;

  
}


/**
 * *
 Request to delete an game.
 */
export class DeleteGameRequest implements DeleteGameRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.DeleteGameRequest";
  readonly __MESSAGE_TYPE = DeleteGameRequest.MESSAGE_TYPE;

  /** *
 ID of the game to be deleted. */
  id: string = "";

  
}


/**
 * *
 Game deletion response
 */
export class DeleteGameResponse implements DeleteGameResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.DeleteGameResponse";
  readonly __MESSAGE_TYPE = DeleteGameResponse.MESSAGE_TYPE;


  
}


/**
 * *
 Request to batch get games
 */
export class GetGamesRequest implements GetGamesRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetGamesRequest";
  readonly __MESSAGE_TYPE = GetGamesRequest.MESSAGE_TYPE;

  /** *
 IDs of the game to be fetched */
  ids: string[] = [];

  
}


/**
 * *
 Game batch-get response
 */
export class GetGamesResponse implements GetGamesResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetGamesResponse";
  readonly __MESSAGE_TYPE = GetGamesResponse.MESSAGE_TYPE;

  games: Record<string, Game> = {};

  
}


/**
 * *
 Game creation request object
 */
export class CreateGameRequest implements CreateGameRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.CreateGameRequest";
  readonly __MESSAGE_TYPE = CreateGameRequest.MESSAGE_TYPE;

  /** *
 Game being updated */
  game?: Game;

  
}


/**
 * *
 Response of an game creation.
 */
export class CreateGameResponse implements CreateGameResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.CreateGameResponse";
  readonly __MESSAGE_TYPE = CreateGameResponse.MESSAGE_TYPE;

  /** *
 Game being created */
  game?: Game;
  /** The starting game state */
  gameState?: GameState;
  /** *
 Error specific to a field if there are any errors. */
  fieldErrors: Record<string, string> = {};

  
}


/**
 * *
 Request to add moves to a game
 The model is that a game in each "tick" can handle multiple moves (by possibly various players).
 It is upto the move manager/processor in the game to ensure the "transaction" of moves is handled
 atomically.

 For example we may have 3 moves where first two units are moved to a common location
 and then they attack another unit.  Here If we treat it as a single unit attacking it
 will have different outcomes than a "combined" attack.
 */
export class ProcessMovesRequest implements ProcessMovesRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ProcessMovesRequest";
  readonly __MESSAGE_TYPE = ProcessMovesRequest.MESSAGE_TYPE;

  /** *
 Game ID to add moves to */
  gameId: string = "";
  /** *
 List of moves to add */
  moves: GameMove[] = [];
  /** *
 The player can submit a list of "Expected" changes when in local-first mode
 If this is list provided the server will validate it - either via the coordinator
 or by itself.  If it is not provided then the server will validate it and return
 the changes. */
  expectedResponse?: ProcessMovesResponse;

  
}


/**
 * *
 Response after adding moves to game.
 */
export class ProcessMovesResponse implements ProcessMovesResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ProcessMovesResponse";
  readonly __MESSAGE_TYPE = ProcessMovesResponse.MESSAGE_TYPE;

  /** *
 Returns the moves that were passed in along wth changes and other data filled in. */
  moves: GameMove[] = [];

  
}


/**
 * *
 Request to get the game's latest state
 */
export class GetGameStateRequest implements GetGameStateRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetGameStateRequest";
  readonly __MESSAGE_TYPE = GetGameStateRequest.MESSAGE_TYPE;

  /** *
 Game ID to add moves to */
  gameId: string = "";

  
}


/**
 * *
 Response holding latest game state
 */
export class GetGameStateResponse implements GetGameStateResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetGameStateResponse";
  readonly __MESSAGE_TYPE = GetGameStateResponse.MESSAGE_TYPE;

  state?: GameState;

  
}


/**
 * *
 Request to list moves for a game
 */
export class ListMovesRequest implements ListMovesRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ListMovesRequest";
  readonly __MESSAGE_TYPE = ListMovesRequest.MESSAGE_TYPE;

  /** *
 Game ID to add moves to */
  gameId: string = "";
  /** Offset of the move to begin fetching from in reverse order from "latest".
 0 => start from now */
  offset: number = 0;
  /** *
 Limit to last N moves (from offset).  if <= 0 return all moves */
  lastN: number = 0;

  
}


/**
 * *
 Response after adding moves to game.
 */
export class ListMovesResponse implements ListMovesResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ListMovesResponse";
  readonly __MESSAGE_TYPE = ListMovesResponse.MESSAGE_TYPE;

  /** Whether there are more moves before this */
  hasMore: boolean = false;
  moveGroups: GameMoveGroup[] = [];

  
}


/**
 * *
 Request to get all available options at a position
 */
export class GetOptionsAtRequest implements GetOptionsAtRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetOptionsAtRequest";
  readonly __MESSAGE_TYPE = GetOptionsAtRequest.MESSAGE_TYPE;

  gameId: string = "";
  q: number = 0;
  r: number = 0;

  
}


/**
 * *
 Response with all available options at a position
 */
export class GetOptionsAtResponse implements GetOptionsAtResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetOptionsAtResponse";
  readonly __MESSAGE_TYPE = GetOptionsAtResponse.MESSAGE_TYPE;

  options: GameOption[] = [];
  currentPlayer: number = 0;
  gameInitialized: boolean = false;
  /** A Path from source to dest along with cost on each tile for tracking */
  allPaths?: AllPaths;

  
}


/**
 * *
 A single game option available at a position
 */
export class GameOption implements GameOptionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameOption";
  readonly __MESSAGE_TYPE = GameOption.MESSAGE_TYPE;

  move?: MoveUnitAction;
  attack?: AttackUnitAction;
  build?: BuildUnitAction;
  capture?: CaptureBuildingAction;
  endTurn?: EndTurnAction;

  
}


/**
 * *
 Request for simulating combat between two units
 */
export class SimulateAttackRequest implements SimulateAttackRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SimulateAttackRequest";
  readonly __MESSAGE_TYPE = SimulateAttackRequest.MESSAGE_TYPE;

  attackerUnitType: number = 0;
  attackerTerrain: number = 0;
  attackerHealth: number = 0;
  defenderUnitType: number = 0;
  defenderTerrain: number = 0;
  defenderHealth: number = 0;
  woundBonus: number = 0;
  numSimulations: number = 0;

  
}


/**
 * *
 Response containing damage distribution statistics
 */
export class SimulateAttackResponse implements SimulateAttackResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SimulateAttackResponse";
  readonly __MESSAGE_TYPE = SimulateAttackResponse.MESSAGE_TYPE;

  /** Damage distributions: damage_value -> number_of_occurrences */
  attackerDamageDistribution: Record<number, number> = {};
  defenderDamageDistribution: Record<number, number> = {};
  /** Statistical summary */
  attackerMeanDamage: number = 0;
  defenderMeanDamage: number = 0;
  attackerKillProbability: number = 0;
  defenderKillProbability: number = 0;

  
}



export class EmptyRequest implements EmptyRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.EmptyRequest";
  readonly __MESSAGE_TYPE = EmptyRequest.MESSAGE_TYPE;


  
}



export class EmptyResponse implements EmptyResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.EmptyResponse";
  readonly __MESSAGE_TYPE = EmptyResponse.MESSAGE_TYPE;


  
}


/**
 * Request to fetch data from a URL
 */
export class SetContentRequest implements SetContentRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SetContentRequest";
  readonly __MESSAGE_TYPE = SetContentRequest.MESSAGE_TYPE;

  innerHtml: string = "";

  
}


/**
 * Response from fetch
 */
export class SetContentResponse implements SetContentResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SetContentResponse";
  readonly __MESSAGE_TYPE = SetContentResponse.MESSAGE_TYPE;


  
}



export class ShowBuildOptionsRequest implements ShowBuildOptionsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ShowBuildOptionsRequest";
  readonly __MESSAGE_TYPE = ShowBuildOptionsRequest.MESSAGE_TYPE;

  innerHtml: string = "";
  hide: boolean = false;
  q: number = 0;
  r: number = 0;

  
}



export class ShowBuildOptionsResponse implements ShowBuildOptionsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ShowBuildOptionsResponse";
  readonly __MESSAGE_TYPE = ShowBuildOptionsResponse.MESSAGE_TYPE;


  
}


/**
 * Request to fetch data from a URL
 */
export class LogMessageRequest implements LogMessageRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.LogMessageRequest";
  readonly __MESSAGE_TYPE = LogMessageRequest.MESSAGE_TYPE;

  message: string = "";

  
}


/**
 * Response from fetch
 */
export class LogMessageResponse implements LogMessageResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.LogMessageResponse";
  readonly __MESSAGE_TYPE = LogMessageResponse.MESSAGE_TYPE;


  
}


/**
 * Request to fetch data from a URL
 */
export class SetGameStateRequest implements SetGameStateRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SetGameStateRequest";
  readonly __MESSAGE_TYPE = SetGameStateRequest.MESSAGE_TYPE;

  game?: Game;
  state?: GameState;

  
}


/**
 * Response from fetch
 */
export class SetGameStateResponse implements SetGameStateResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SetGameStateResponse";
  readonly __MESSAGE_TYPE = SetGameStateResponse.MESSAGE_TYPE;


  
}


/**
 * Request to update game UI status (current player, turn counter)
 */
export class UpdateGameStatusRequest implements UpdateGameStatusRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UpdateGameStatusRequest";
  readonly __MESSAGE_TYPE = UpdateGameStatusRequest.MESSAGE_TYPE;

  currentPlayer: number = 0;
  turnCounter: number = 0;

  
}



export class UpdateGameStatusResponse implements UpdateGameStatusResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UpdateGameStatusResponse";
  readonly __MESSAGE_TYPE = UpdateGameStatusResponse.MESSAGE_TYPE;


  
}


/**
 * Request to set a tile at a specific coordinate
 */
export class SetTileAtRequest implements SetTileAtRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SetTileAtRequest";
  readonly __MESSAGE_TYPE = SetTileAtRequest.MESSAGE_TYPE;

  q: number = 0;
  r: number = 0;
  tile?: Tile;

  
}



export class SetTileAtResponse implements SetTileAtResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SetTileAtResponse";
  readonly __MESSAGE_TYPE = SetTileAtResponse.MESSAGE_TYPE;


  
}


/**
 * Request to set a unit at a specific coordinate
 */
export class SetUnitAtRequest implements SetUnitAtRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SetUnitAtRequest";
  readonly __MESSAGE_TYPE = SetUnitAtRequest.MESSAGE_TYPE;

  q: number = 0;
  r: number = 0;
  unit?: Unit;
  flash: boolean = false;
  appear: boolean = false;

  
}



export class SetUnitAtResponse implements SetUnitAtResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SetUnitAtResponse";
  readonly __MESSAGE_TYPE = SetUnitAtResponse.MESSAGE_TYPE;


  
}


/**
 * Request to remove a tile at a specific coordinate
 */
export class RemoveTileAtRequest implements RemoveTileAtRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.RemoveTileAtRequest";
  readonly __MESSAGE_TYPE = RemoveTileAtRequest.MESSAGE_TYPE;

  q: number = 0;
  r: number = 0;
  animate: boolean = false;

  
}



export class RemoveTileAtResponse implements RemoveTileAtResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.RemoveTileAtResponse";
  readonly __MESSAGE_TYPE = RemoveTileAtResponse.MESSAGE_TYPE;


  
}


/**
 * Request to remove a unit at a specific coordinate
 */
export class RemoveUnitAtRequest implements RemoveUnitAtRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.RemoveUnitAtRequest";
  readonly __MESSAGE_TYPE = RemoveUnitAtRequest.MESSAGE_TYPE;

  q: number = 0;
  r: number = 0;
  animate: boolean = false;

  
}



export class RemoveUnitAtResponse implements RemoveUnitAtResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.RemoveUnitAtResponse";
  readonly __MESSAGE_TYPE = RemoveUnitAtResponse.MESSAGE_TYPE;


  
}


/**
 * Request to show highlights on the game board
 */
export class ShowHighlightsRequest implements ShowHighlightsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ShowHighlightsRequest";
  readonly __MESSAGE_TYPE = ShowHighlightsRequest.MESSAGE_TYPE;

  highlights: HighlightSpec[] = [];

  
}



export class ShowHighlightsResponse implements ShowHighlightsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ShowHighlightsResponse";
  readonly __MESSAGE_TYPE = ShowHighlightsResponse.MESSAGE_TYPE;


  
}


/**
 * Specification for a single highlight
 */
export class HighlightSpec implements HighlightSpecInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.HighlightSpec";
  readonly __MESSAGE_TYPE = HighlightSpec.MESSAGE_TYPE;

  q: number = 0;
  r: number = 0;
  type: string = "";
  move?: MoveUnitAction;
  attack?: AttackUnitAction;
  build?: BuildUnitAction;
  capture?: CaptureBuildingAction;

  
}


/**
 * Request to clear highlights
 */
export class ClearHighlightsRequest implements ClearHighlightsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ClearHighlightsRequest";
  readonly __MESSAGE_TYPE = ClearHighlightsRequest.MESSAGE_TYPE;

  types: string[] = [];

  
}



export class ClearHighlightsResponse implements ClearHighlightsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ClearHighlightsResponse";
  readonly __MESSAGE_TYPE = ClearHighlightsResponse.MESSAGE_TYPE;


  
}


/**
 * Request to show a path on the game board
 */
export class ShowPathRequest implements ShowPathRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ShowPathRequest";
  readonly __MESSAGE_TYPE = ShowPathRequest.MESSAGE_TYPE;

  coords: number[] = [];
  color: number = 0;
  thickness: number = 0;

  
}



export class ShowPathResponse implements ShowPathResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ShowPathResponse";
  readonly __MESSAGE_TYPE = ShowPathResponse.MESSAGE_TYPE;


  
}


/**
 * Request to clear paths
 */
export class ClearPathsRequest implements ClearPathsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ClearPathsRequest";
  readonly __MESSAGE_TYPE = ClearPathsRequest.MESSAGE_TYPE;


  
}



export class ClearPathsResponse implements ClearPathsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ClearPathsResponse";
  readonly __MESSAGE_TYPE = ClearPathsResponse.MESSAGE_TYPE;


  
}


/**
 * Request to animate unit movement along a path
 */
export class MoveUnitRequest implements MoveUnitRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.MoveUnitRequest";
  readonly __MESSAGE_TYPE = MoveUnitRequest.MESSAGE_TYPE;

  unit?: Unit;
  path: HexCoord[] = [];

  
}



export class MoveUnitResponse implements MoveUnitResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.MoveUnitResponse";
  readonly __MESSAGE_TYPE = MoveUnitResponse.MESSAGE_TYPE;


  
}


/**
 * Hex coordinate for paths
 */
export class HexCoord implements HexCoordInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.HexCoord";
  readonly __MESSAGE_TYPE = HexCoord.MESSAGE_TYPE;

  q: number = 0;
  r: number = 0;

  
}


/**
 * Request to show attack effect animation
 */
export class ShowAttackEffectRequest implements ShowAttackEffectRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ShowAttackEffectRequest";
  readonly __MESSAGE_TYPE = ShowAttackEffectRequest.MESSAGE_TYPE;

  fromQ: number = 0;
  fromR: number = 0;
  toQ: number = 0;
  toR: number = 0;
  damage: number = 0;
  splashTargets: SplashTarget[] = [];

  
}



export class SplashTarget implements SplashTargetInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SplashTarget";
  readonly __MESSAGE_TYPE = SplashTarget.MESSAGE_TYPE;

  q: number = 0;
  r: number = 0;
  damage: number = 0;

  
}



export class ShowAttackEffectResponse implements ShowAttackEffectResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ShowAttackEffectResponse";
  readonly __MESSAGE_TYPE = ShowAttackEffectResponse.MESSAGE_TYPE;


  
}


/**
 * Request to show heal effect animation
 */
export class ShowHealEffectRequest implements ShowHealEffectRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ShowHealEffectRequest";
  readonly __MESSAGE_TYPE = ShowHealEffectRequest.MESSAGE_TYPE;

  q: number = 0;
  r: number = 0;
  amount: number = 0;

  
}



export class ShowHealEffectResponse implements ShowHealEffectResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ShowHealEffectResponse";
  readonly __MESSAGE_TYPE = ShowHealEffectResponse.MESSAGE_TYPE;


  
}


/**
 * Request to show capture effect animation
 */
export class ShowCaptureEffectRequest implements ShowCaptureEffectRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ShowCaptureEffectRequest";
  readonly __MESSAGE_TYPE = ShowCaptureEffectRequest.MESSAGE_TYPE;

  q: number = 0;
  r: number = 0;

  
}



export class ShowCaptureEffectResponse implements ShowCaptureEffectResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ShowCaptureEffectResponse";
  readonly __MESSAGE_TYPE = ShowCaptureEffectResponse.MESSAGE_TYPE;


  
}


/**
 * Request to set allowed panels and their order
 */
export class SetAllowedPanelsRequest implements SetAllowedPanelsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SetAllowedPanelsRequest";
  readonly __MESSAGE_TYPE = SetAllowedPanelsRequest.MESSAGE_TYPE;

  panelIds: string[] = [];

  
}



export class SetAllowedPanelsResponse implements SetAllowedPanelsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SetAllowedPanelsResponse";
  readonly __MESSAGE_TYPE = SetAllowedPanelsResponse.MESSAGE_TYPE;


  
}



export class IndexState implements IndexStateInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.IndexState";
  readonly __MESSAGE_TYPE = IndexState.MESSAGE_TYPE;

  entityType: string = "";
  entityId: string = "";
  /** eg "screenshots", "keywords" etc
 EntityType + EntityId + IndexType should be  unique */
  indexType: string = "";
  /** When the last indexing was queued */
  createdAt?: Timestamp;
  /** when the last time the entity was recorded for an update (means it is eligible for a re-indexing) */
  updatedAt?: Timestamp;
  /** When did the last indexing finish */
  indexedAt?: Timestamp;
  /** Whether indexing is needed or not */
  needsIndexing: boolean = false;
  /** "queued/pending", "indexing", "completed", "failed" */
  status: IndexStatus = IndexStatus.INDEX_STATUS_UNSPECIFIED;
  /** If there was an error in the last indexing */
  lastError: string = "";
  /** A way to ignore multiple requests if they are updates but
 nothing has changed */
  idempotencyKey: string = "";
  retryCount: number = 0;

  
}



export class EnsureIndexStateRequest implements EnsureIndexStateRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.EnsureIndexStateRequest";
  readonly __MESSAGE_TYPE = EnsureIndexStateRequest.MESSAGE_TYPE;

  indexState?: IndexState;
  /** *
 Mask of fields being updated in this Game to make partial changes. */
  updateMask?: FieldMask;

  
}



export class EnsureIndexStateResponse implements EnsureIndexStateResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.EnsureIndexStateResponse";
  readonly __MESSAGE_TYPE = EnsureIndexStateResponse.MESSAGE_TYPE;

  indexState?: IndexState;

  
}



export class GetIndexStatesRequest implements GetIndexStatesRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetIndexStatesRequest";
  readonly __MESSAGE_TYPE = GetIndexStatesRequest.MESSAGE_TYPE;

  entityType: string = "";
  entityId: string = "";
  /** Optional - can be used to get "all" indexer states or just once specified here */
  indexTypes: string[] = [];

  
}



export class IndexStateList implements IndexStateListInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.IndexStateList";
  readonly __MESSAGE_TYPE = IndexStateList.MESSAGE_TYPE;

  states: IndexState[] = [];

  
}



export class GetIndexStatesResponse implements GetIndexStatesResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetIndexStatesResponse";
  readonly __MESSAGE_TYPE = GetIndexStatesResponse.MESSAGE_TYPE;

  states: Record<string, IndexState> = {};

  
}



export class ListIndexStatesRequest implements ListIndexStatesRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ListIndexStatesRequest";
  readonly __MESSAGE_TYPE = ListIndexStatesRequest.MESSAGE_TYPE;

  entityType: string = "";
  /** Get records indexed "before" this time */
  updatedBefore?: Timestamp | undefined;
  /** Get records updated "after" this time */
  updatedAfter?: Timestamp | undefined;
  /** Filter by index types or get all */
  indexTypes: string[] = [];
  /** "id" or "indexed_at" */
  orderBy: string = "";
  /** limit to max items */
  count: number = 0;

  
}



export class ListIndexStatesResponse implements ListIndexStatesResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ListIndexStatesResponse";
  readonly __MESSAGE_TYPE = ListIndexStatesResponse.MESSAGE_TYPE;

  items: IndexState[] = [];
  /** How to identify the next "page" in this list */
  nextPageKey: string = "";

  
}



export class DeleteIndexStatesRequest implements DeleteIndexStatesRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.DeleteIndexStatesRequest";
  readonly __MESSAGE_TYPE = DeleteIndexStatesRequest.MESSAGE_TYPE;

  entityType: string = "";
  entityId: string = "";
  /** Optional - can be used to get "all" indexer states or just once specified here */
  indexTypes: string[] = [];

  
}



export class DeleteIndexStatesResponse implements DeleteIndexStatesResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.DeleteIndexStatesResponse";
  readonly __MESSAGE_TYPE = DeleteIndexStatesResponse.MESSAGE_TYPE;


  
}


/**
 * Request messages
 */
export class IndexRecord implements IndexRecordInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.IndexRecord";
  readonly __MESSAGE_TYPE = IndexRecord.MESSAGE_TYPE;

  entityId: string = "";
  updatedAt?: Timestamp;
  entityData?: Any;
  indexerTypes: string[] = [];

  
}


/**
 * Each IndexRecords gets its own "long running operation" 
 so we can track how things are
 */
export class IndexRecordsLRO implements IndexRecordsLROInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.IndexRecordsLRO";
  readonly __MESSAGE_TYPE = IndexRecordsLRO.MESSAGE_TYPE;

  lroId: string = "";
  /** Single entity type in a index request */
  entityType: string = "";
  /** When this request was created */
  createdAt?: Timestamp;
  updatedAt?: Timestamp;
  /** Callback url for this request (if any)
 All records indexed in this operation will be notified via this url */
  callbackUrl: string = "";
  /** Records to enqueue - note that indexing by its nature is asynchronous
 so we wont wait for all the indexing to finish */
  records: IndexRecord[] = [];

  
}



export class CreateIndexRecordsLRORequest implements CreateIndexRecordsLRORequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.CreateIndexRecordsLRORequest";
  readonly __MESSAGE_TYPE = CreateIndexRecordsLRORequest.MESSAGE_TYPE;

  lro?: IndexRecordsLRO;

  
}



export class CreateIndexRecordsLROResponse implements CreateIndexRecordsLROResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.CreateIndexRecordsLROResponse";
  readonly __MESSAGE_TYPE = CreateIndexRecordsLROResponse.MESSAGE_TYPE;

  lro?: IndexRecordsLRO;

  
}



export class UpdateIndexRecordsLRORequest implements UpdateIndexRecordsLRORequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UpdateIndexRecordsLRORequest";
  readonly __MESSAGE_TYPE = UpdateIndexRecordsLRORequest.MESSAGE_TYPE;

  lro?: IndexRecordsLRO;
  /** *
 Mask of fields being updated in this Game to make partial changes. */
  updateMask?: FieldMask;

  
}



export class UpdateIndexRecordsLROResponse implements UpdateIndexRecordsLROResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UpdateIndexRecordsLROResponse";
  readonly __MESSAGE_TYPE = UpdateIndexRecordsLROResponse.MESSAGE_TYPE;

  lro?: IndexRecordsLRO;

  
}



export class GetIndexRecordsLRORequest implements GetIndexRecordsLRORequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetIndexRecordsLRORequest";
  readonly __MESSAGE_TYPE = GetIndexRecordsLRORequest.MESSAGE_TYPE;

  lroId: string = "";

  
}



export class GetIndexRecordsLROResponse implements GetIndexRecordsLROResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetIndexRecordsLROResponse";
  readonly __MESSAGE_TYPE = GetIndexRecordsLROResponse.MESSAGE_TYPE;

  lro?: IndexRecordsLRO;

  
}


/**
 * Job describes the work that needs to be done.
 */
export class Job implements JobInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.Job";
  readonly __MESSAGE_TYPE = Job.MESSAGE_TYPE;

  entityType: string = "";
  entityId: string = "";
  jobType: string = "";
  /** When the last indexing was queued */
  createdAt?: Timestamp;
  /** when the last indexing was completed */
  updatedAt?: Timestamp;
  /** Job specific data */
  jobData?: Any;
  /** Debounce so we dont run it too many time within this many seconds */
  debounceWindowSeconds: number = 0;
  /** Whether the job is a oneoff or can repeat */
  repeatInfo?: RepeatInfo;

  
}



export class RepeatInfo implements RepeatInfoInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.RepeatInfo";
  readonly __MESSAGE_TYPE = RepeatInfo.MESSAGE_TYPE;


  
}



export class Run implements RunInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.Run";
  readonly __MESSAGE_TYPE = Run.MESSAGE_TYPE;

  jobId: string = "";
  /** A unique run_id */
  runId: string = "";
  createdAt?: Timestamp;
  startedAt?: Timestamp;
  updatedAt?: Timestamp;
  state: RunState = RunState.RUN_STATE_UNSPECIFIED;
  /** Run specific data */
  runData?: Any;
  /** If there was an error in the last indexing */
  lastError: string = "";
  /** Keep a hash of the contents for quick check to check updated
 (not sure if needed) - This should be provided by the source */
  lastContentHash: string = "";
  /** If there were retries */
  retryCount: number = 0;

  
}


/**
 * Called when the end turn button was clicked
 */
export class InitializeSingletonRequest implements InitializeSingletonRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.InitializeSingletonRequest";
  readonly __MESSAGE_TYPE = InitializeSingletonRequest.MESSAGE_TYPE;

  gameId: string = "";
  gameData: string = "";
  gameState: string = "";
  moveHistory: string = "";

  
}


/**
 * Response of a turn option click
 */
export class InitializeSingletonResponse implements InitializeSingletonResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.InitializeSingletonResponse";
  readonly __MESSAGE_TYPE = InitializeSingletonResponse.MESSAGE_TYPE;

  response?: InitializeGameResponse;

  
}


/**
 * Called when a turn option is clicked in TurnOptionsPanel
 */
export class TurnOptionClickedRequest implements TurnOptionClickedRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.TurnOptionClickedRequest";
  readonly __MESSAGE_TYPE = TurnOptionClickedRequest.MESSAGE_TYPE;

  gameId: string = "";
  optionIndex: number = 0;
  optionType: string = "";
  q: number = 0;
  r: number = 0;

  
}


/**
 * Response of a turn option click
 */
export class TurnOptionClickedResponse implements TurnOptionClickedResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.TurnOptionClickedResponse";
  readonly __MESSAGE_TYPE = TurnOptionClickedResponse.MESSAGE_TYPE;

  gameId: string = "";

  
}


/**
 * Called when the scene was clicked
 */
export class SceneClickedRequest implements SceneClickedRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SceneClickedRequest";
  readonly __MESSAGE_TYPE = SceneClickedRequest.MESSAGE_TYPE;

  gameId: string = "";
  q: number = 0;
  r: number = 0;
  layer: string = "";

  
}


/**
 * Response of a turn option click
 */
export class SceneClickedResponse implements SceneClickedResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SceneClickedResponse";
  readonly __MESSAGE_TYPE = SceneClickedResponse.MESSAGE_TYPE;

  gameId: string = "";

  
}


/**
 * Called when the end turn button was clicked
 */
export class EndTurnButtonClickedRequest implements EndTurnButtonClickedRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.EndTurnButtonClickedRequest";
  readonly __MESSAGE_TYPE = EndTurnButtonClickedRequest.MESSAGE_TYPE;

  gameId: string = "";

  
}


/**
 * Response of a turn option click
 */
export class EndTurnButtonClickedResponse implements EndTurnButtonClickedResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.EndTurnButtonClickedResponse";
  readonly __MESSAGE_TYPE = EndTurnButtonClickedResponse.MESSAGE_TYPE;

  gameId: string = "";

  
}


/**
 * Called when a build option is clicked in BuildOptionsModal
 */
export class BuildOptionClickedRequest implements BuildOptionClickedRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.BuildOptionClickedRequest";
  readonly __MESSAGE_TYPE = BuildOptionClickedRequest.MESSAGE_TYPE;

  gameId: string = "";
  q: number = 0;
  r: number = 0;
  unitType: number = 0;

  
}


/**
 * Response of a build option click
 */
export class BuildOptionClickedResponse implements BuildOptionClickedResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.BuildOptionClickedResponse";
  readonly __MESSAGE_TYPE = BuildOptionClickedResponse.MESSAGE_TYPE;


  
}


/**
 * Called when the end turn button was clicked
 */
export class InitializeGameRequest implements InitializeGameRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.InitializeGameRequest";
  readonly __MESSAGE_TYPE = InitializeGameRequest.MESSAGE_TYPE;

  gameId: string = "";

  
}


/**
 * Response of a turn option click
 */
export class InitializeGameResponse implements InitializeGameResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.InitializeGameResponse";
  readonly __MESSAGE_TYPE = InitializeGameResponse.MESSAGE_TYPE;

  success: boolean = false;
  error: string = "";
  /** Initial UI state information */
  currentPlayer: number = 0;
  turnCounter: number = 0;
  gameName: string = "";

  
}


/**
 * ThemeInfo contains metadata about a theme
 */
export class ThemeInfo implements ThemeInfoInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ThemeInfo";
  readonly __MESSAGE_TYPE = ThemeInfo.MESSAGE_TYPE;

  name: string = "";
  version: string = "";
  basePath: string = "";
  assetType: string = "";
  needsPostProcessing: boolean = false;

  
}


/**
 * UnitMapping maps a unit ID to its theme-specific representation
 */
export class UnitMapping implements UnitMappingInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UnitMapping";
  readonly __MESSAGE_TYPE = UnitMapping.MESSAGE_TYPE;

  old: string = "";
  name: string = "";
  image: string = "";
  description: string = "";

  
}


/**
 * TerrainMapping maps a terrain ID to its theme-specific representation
 */
export class TerrainMapping implements TerrainMappingInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.TerrainMapping";
  readonly __MESSAGE_TYPE = TerrainMapping.MESSAGE_TYPE;

  old: string = "";
  name: string = "";
  image: string = "";
  description: string = "";

  
}


/**
 * ThemeManifest represents the full theme configuration
 This matches the structure of mapping.json files
 */
export class ThemeManifest implements ThemeManifestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ThemeManifest";
  readonly __MESSAGE_TYPE = ThemeManifest.MESSAGE_TYPE;

  themeInfo?: ThemeInfo;
  units: Record<number, UnitMapping> = {};
  terrains: Record<number, TerrainMapping> = {};

  
}


/**
 * PlayerColor defines the color scheme for a player
 */
export class PlayerColor implements PlayerColorInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.PlayerColor";
  readonly __MESSAGE_TYPE = PlayerColor.MESSAGE_TYPE;

  primary: string = "";
  secondary: string = "";

  
}


/**
 * AssetResult represents a rendered asset
 */
export class AssetResult implements AssetResultInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.AssetResult";
  readonly __MESSAGE_TYPE = AssetResult.MESSAGE_TYPE;

  type: Type = Type.TYPE_UNSPECIFIED;
  data: string = "";

  
}


/**
 * UserInfo represents a user in the catalog
 */
export class UserInfo implements UserInfoInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UserInfo";
  readonly __MESSAGE_TYPE = UserInfo.MESSAGE_TYPE;

  id: string = "";
  name: string = "";
  description: string = "";
  category: string = "";
  difficulty: string = "";
  tags: string[] = [];
  icon: string = "";
  lastUpdated: string = "";

  
}


/**
 * Request messages
 */
export class ListUsersRequest implements ListUsersRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ListUsersRequest";
  readonly __MESSAGE_TYPE = ListUsersRequest.MESSAGE_TYPE;

  /** Pagination info */
  pagination?: Pagination;
  /** May be filter by owner id */
  ownerId: string = "";

  
}



export class ListUsersResponse implements ListUsersResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ListUsersResponse";
  readonly __MESSAGE_TYPE = ListUsersResponse.MESSAGE_TYPE;

  items: User[] = [];
  pagination?: PaginationResponse;

  
}



export class GetUserRequest implements GetUserRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetUserRequest";
  readonly __MESSAGE_TYPE = GetUserRequest.MESSAGE_TYPE;

  id: string = "";
  version: string = "";

  
}



export class GetUserResponse implements GetUserResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetUserResponse";
  readonly __MESSAGE_TYPE = GetUserResponse.MESSAGE_TYPE;

  user?: User;

  
}



export class GetUserContentRequest implements GetUserContentRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetUserContentRequest";
  readonly __MESSAGE_TYPE = GetUserContentRequest.MESSAGE_TYPE;

  id: string = "";
  version: string = "";

  
}



export class GetUserContentResponse implements GetUserContentResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetUserContentResponse";
  readonly __MESSAGE_TYPE = GetUserContentResponse.MESSAGE_TYPE;

  weewarContent: string = "";
  recipeContent: string = "";
  readmeContent: string = "";

  
}



export class UpdateUserRequest implements UpdateUserRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UpdateUserRequest";
  readonly __MESSAGE_TYPE = UpdateUserRequest.MESSAGE_TYPE;

  /** *
 User being updated */
  user?: User;
  /** *
 Mask of fields being updated in this User to make partial changes. */
  updateMask?: FieldMask;

  
}


/**
 * *
 The request for (partially) updating an User.
 */
export class UpdateUserResponse implements UpdateUserResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UpdateUserResponse";
  readonly __MESSAGE_TYPE = UpdateUserResponse.MESSAGE_TYPE;

  /** *
 User being updated */
  user?: User;

  
}


/**
 * *
 Request to delete an user.
 */
export class DeleteUserRequest implements DeleteUserRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.DeleteUserRequest";
  readonly __MESSAGE_TYPE = DeleteUserRequest.MESSAGE_TYPE;

  /** *
 ID of the user to be deleted. */
  id: string = "";

  
}


/**
 * *
 User deletion response
 */
export class DeleteUserResponse implements DeleteUserResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.DeleteUserResponse";
  readonly __MESSAGE_TYPE = DeleteUserResponse.MESSAGE_TYPE;


  
}


/**
 * *
 Request to batch get users
 */
export class GetUsersRequest implements GetUsersRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetUsersRequest";
  readonly __MESSAGE_TYPE = GetUsersRequest.MESSAGE_TYPE;

  /** *
 IDs of the user to be fetched */
  ids: string[] = [];

  
}


/**
 * *
 User batch-get response
 */
export class GetUsersResponse implements GetUsersResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetUsersResponse";
  readonly __MESSAGE_TYPE = GetUsersResponse.MESSAGE_TYPE;

  users: Record<string, User> = {};

  
}


/**
 * *
 User creation request object
 */
export class CreateUserRequest implements CreateUserRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.CreateUserRequest";
  readonly __MESSAGE_TYPE = CreateUserRequest.MESSAGE_TYPE;

  /** *
 User being updated */
  user?: User;

  
}


/**
 * *
 Response of an user creation.
 */
export class CreateUserResponse implements CreateUserResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.CreateUserResponse";
  readonly __MESSAGE_TYPE = CreateUserResponse.MESSAGE_TYPE;

  /** *
 User being created */
  user?: User;
  /** *
 Error specific to a field if there are any errors. */
  fieldErrors: Record<string, string> = {};

  
}


/**
 * WorldInfo represents a world in the catalog
 */
export class WorldInfo implements WorldInfoInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.WorldInfo";
  readonly __MESSAGE_TYPE = WorldInfo.MESSAGE_TYPE;

  id: string = "";
  name: string = "";
  description: string = "";
  category: string = "";
  difficulty: string = "";
  tags: string[] = [];
  icon: string = "";
  lastUpdated: string = "";

  
}


/**
 * Request messages
 */
export class ListWorldsRequest implements ListWorldsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ListWorldsRequest";
  readonly __MESSAGE_TYPE = ListWorldsRequest.MESSAGE_TYPE;

  /** Pagination info */
  pagination?: Pagination;
  /** May be filter by owner id */
  ownerId: string = "";

  
}



export class ListWorldsResponse implements ListWorldsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ListWorldsResponse";
  readonly __MESSAGE_TYPE = ListWorldsResponse.MESSAGE_TYPE;

  items: World[] = [];
  pagination?: PaginationResponse;

  
}



export class GetWorldRequest implements GetWorldRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetWorldRequest";
  readonly __MESSAGE_TYPE = GetWorldRequest.MESSAGE_TYPE;

  id: string = "";
  version: string = "";

  
}



export class GetWorldResponse implements GetWorldResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetWorldResponse";
  readonly __MESSAGE_TYPE = GetWorldResponse.MESSAGE_TYPE;

  world?: World;
  worldData?: WorldData;

  
}



export class UpdateWorldRequest implements UpdateWorldRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UpdateWorldRequest";
  readonly __MESSAGE_TYPE = UpdateWorldRequest.MESSAGE_TYPE;

  /** *
 World being updated */
  world?: World;
  worldData?: WorldData;
  clearWorld: boolean = false;
  /** *
 Mask of fields being updated in this World to make partial changes. */
  updateMask?: FieldMask;

  
}


/**
 * *
 The request for (partially) updating an World.
 */
export class UpdateWorldResponse implements UpdateWorldResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UpdateWorldResponse";
  readonly __MESSAGE_TYPE = UpdateWorldResponse.MESSAGE_TYPE;

  /** *
 World being updated */
  world?: World;
  worldData?: WorldData;

  
}


/**
 * *
 Request to delete an world.
 */
export class DeleteWorldRequest implements DeleteWorldRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.DeleteWorldRequest";
  readonly __MESSAGE_TYPE = DeleteWorldRequest.MESSAGE_TYPE;

  /** *
 ID of the world to be deleted. */
  id: string = "";

  
}


/**
 * *
 World deletion response
 */
export class DeleteWorldResponse implements DeleteWorldResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.DeleteWorldResponse";
  readonly __MESSAGE_TYPE = DeleteWorldResponse.MESSAGE_TYPE;


  
}


/**
 * *
 Request to batch get worlds
 */
export class GetWorldsRequest implements GetWorldsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetWorldsRequest";
  readonly __MESSAGE_TYPE = GetWorldsRequest.MESSAGE_TYPE;

  /** *
 IDs of the world to be fetched */
  ids: string[] = [];

  
}


/**
 * *
 World batch-get response
 */
export class GetWorldsResponse implements GetWorldsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetWorldsResponse";
  readonly __MESSAGE_TYPE = GetWorldsResponse.MESSAGE_TYPE;

  worlds: Record<string, World> = {};

  
}


/**
 * *
 World creation request object
 */
export class CreateWorldRequest implements CreateWorldRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.CreateWorldRequest";
  readonly __MESSAGE_TYPE = CreateWorldRequest.MESSAGE_TYPE;

  /** *
 World being updated */
  world?: World;
  worldData?: WorldData;

  
}


/**
 * *
 Response of an world creation.
 */
export class CreateWorldResponse implements CreateWorldResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.CreateWorldResponse";
  readonly __MESSAGE_TYPE = CreateWorldResponse.MESSAGE_TYPE;

  /** *
 World being created */
  world?: World;
  worldData?: WorldData;
  /** *
 Error specific to a field if there are any errors. */
  fieldErrors: Record<string, string> = {};

  
}


