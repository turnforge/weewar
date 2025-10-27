import { FieldMask, Timestamp } from "@bufbuild/protobuf/wkt";


import { User as UserInterface, Pagination as PaginationInterface, PaginationResponse as PaginationResponseInterface, World as WorldInterface, WorldData as WorldDataInterface, Tile as TileInterface, Unit as UnitInterface, TerrainDefinition as TerrainDefinitionInterface, UnitDefinition as UnitDefinitionInterface, TerrainUnitProperties as TerrainUnitPropertiesInterface, UnitUnitProperties as UnitUnitPropertiesInterface, DamageDistribution as DamageDistributionInterface, DamageRange as DamageRangeInterface, RulesEngine as RulesEngineInterface, Game as GameInterface, GameConfiguration as GameConfigurationInterface, GamePlayer as GamePlayerInterface, GameTeam as GameTeamInterface, GameSettings as GameSettingsInterface, GameState as GameStateInterface, GameMoveHistory as GameMoveHistoryInterface, GameMoveGroup as GameMoveGroupInterface, GameMove as GameMoveInterface, GameMoveResult as GameMoveResultInterface, MoveUnitAction as MoveUnitActionInterface, AttackUnitAction as AttackUnitActionInterface, EndTurnAction as EndTurnActionInterface, WorldChange as WorldChangeInterface, UnitMovedChange as UnitMovedChangeInterface, UnitDamagedChange as UnitDamagedChangeInterface, UnitKilledChange as UnitKilledChangeInterface, PlayerChangedChange as PlayerChangedChangeInterface, GameInfo as GameInfoInterface, ListGamesRequest as ListGamesRequestInterface, ListGamesResponse as ListGamesResponseInterface, GetGameRequest as GetGameRequestInterface, GetGameResponse as GetGameResponseInterface, GetGameContentRequest as GetGameContentRequestInterface, GetGameContentResponse as GetGameContentResponseInterface, UpdateGameRequest as UpdateGameRequestInterface, UpdateGameResponse as UpdateGameResponseInterface, DeleteGameRequest as DeleteGameRequestInterface, DeleteGameResponse as DeleteGameResponseInterface, GetGamesRequest as GetGamesRequestInterface, GetGamesResponse as GetGamesResponseInterface, CreateGameRequest as CreateGameRequestInterface, CreateGameResponse as CreateGameResponseInterface, ProcessMovesRequest as ProcessMovesRequestInterface, ProcessMovesResponse as ProcessMovesResponseInterface, GetGameStateRequest as GetGameStateRequestInterface, GetGameStateResponse as GetGameStateResponseInterface, ListMovesRequest as ListMovesRequestInterface, ListMovesResponse as ListMovesResponseInterface, GetOptionsAtRequest as GetOptionsAtRequestInterface, GetOptionsAtResponse as GetOptionsAtResponseInterface, AllPaths as AllPathsInterface, PathEdge as PathEdgeInterface, Path as PathInterface, GameOption as GameOptionInterface, EndTurnOption as EndTurnOptionInterface, MoveOption as MoveOptionInterface, AttackOption as AttackOptionInterface, BuildUnitOption as BuildUnitOptionInterface, CaptureBuildingOption as CaptureBuildingOptionInterface, EmptyRequest as EmptyRequestInterface, EmptyResponse as EmptyResponseInterface, SetContentRequest as SetContentRequestInterface, SetContentResponse as SetContentResponseInterface, LogMessageRequest as LogMessageRequestInterface, LogMessageResponse as LogMessageResponseInterface, SetGameStateRequest as SetGameStateRequestInterface, SetGameStateResponse as SetGameStateResponseInterface, UpdateGameStatusRequest as UpdateGameStatusRequestInterface, UpdateGameStatusResponse as UpdateGameStatusResponseInterface, SetTileAtRequest as SetTileAtRequestInterface, SetTileAtResponse as SetTileAtResponseInterface, SetUnitAtRequest as SetUnitAtRequestInterface, SetUnitAtResponse as SetUnitAtResponseInterface, RemoveTileAtRequest as RemoveTileAtRequestInterface, RemoveTileAtResponse as RemoveTileAtResponseInterface, RemoveUnitAtRequest as RemoveUnitAtRequestInterface, RemoveUnitAtResponse as RemoveUnitAtResponseInterface, ShowHighlightsRequest as ShowHighlightsRequestInterface, ShowHighlightsResponse as ShowHighlightsResponseInterface, HighlightSpec as HighlightSpecInterface, ClearHighlightsRequest as ClearHighlightsRequestInterface, ClearHighlightsResponse as ClearHighlightsResponseInterface, ShowPathRequest as ShowPathRequestInterface, ShowPathResponse as ShowPathResponseInterface, ClearPathsRequest as ClearPathsRequestInterface, ClearPathsResponse as ClearPathsResponseInterface, TurnOptionClickedRequest as TurnOptionClickedRequestInterface, TurnOptionClickedResponse as TurnOptionClickedResponseInterface, SceneClickedRequest as SceneClickedRequestInterface, SceneClickedResponse as SceneClickedResponseInterface, EndTurnButtonClickedRequest as EndTurnButtonClickedRequestInterface, EndTurnButtonClickedResponse as EndTurnButtonClickedResponseInterface, InitializeGameRequest as InitializeGameRequestInterface, InitializeGameResponse as InitializeGameResponseInterface, ThemeInfo as ThemeInfoInterface, UnitMapping as UnitMappingInterface, TerrainMapping as TerrainMappingInterface, ThemeManifest as ThemeManifestInterface, PlayerColor as PlayerColorInterface, AssetResult as AssetResultInterface, UserInfo as UserInfoInterface, ListUsersRequest as ListUsersRequestInterface, ListUsersResponse as ListUsersResponseInterface, GetUserRequest as GetUserRequestInterface, GetUserResponse as GetUserResponseInterface, GetUserContentRequest as GetUserContentRequestInterface, GetUserContentResponse as GetUserContentResponseInterface, UpdateUserRequest as UpdateUserRequestInterface, UpdateUserResponse as UpdateUserResponseInterface, DeleteUserRequest as DeleteUserRequestInterface, DeleteUserResponse as DeleteUserResponseInterface, GetUsersRequest as GetUsersRequestInterface, GetUsersResponse as GetUsersResponseInterface, CreateUserRequest as CreateUserRequestInterface, CreateUserResponse as CreateUserResponseInterface, WorldInfo as WorldInfoInterface, ListWorldsRequest as ListWorldsRequestInterface, ListWorldsResponse as ListWorldsResponseInterface, GetWorldRequest as GetWorldRequestInterface, GetWorldResponse as GetWorldResponseInterface, UpdateWorldRequest as UpdateWorldRequestInterface, UpdateWorldResponse as UpdateWorldResponseInterface, DeleteWorldRequest as DeleteWorldRequestInterface, DeleteWorldResponse as DeleteWorldResponseInterface, GetWorldsRequest as GetWorldsRequestInterface, GetWorldsResponse as GetWorldsResponseInterface, CreateWorldRequest as CreateWorldRequestInterface, CreateWorldResponse as CreateWorldResponseInterface, GameStatus, PathDirection, Type } from "./interfaces";
import { Weewar_v1Deserializer } from "./deserializer";



export class User implements UserInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.User";

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

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized User instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<User>(User.MESSAGE_TYPE, data);
  }
}



export class Pagination implements PaginationInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.Pagination";

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

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized Pagination instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<Pagination>(Pagination.MESSAGE_TYPE, data);
  }
}



export class PaginationResponse implements PaginationResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.PaginationResponse";

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

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized PaginationResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<PaginationResponse>(PaginationResponse.MESSAGE_TYPE, data);
  }
}



export class World implements WorldInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.World";

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
  /** The actual world contents/data */
  worldData?: WorldData;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized World instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<World>(World.MESSAGE_TYPE, data);
  }
}



export class WorldData implements WorldDataInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.WorldData";

  /** JSON-fied tile data about what units and terrains are at each location */
  tiles: Tile[] = [];
  /** All units on the world and who they belong to */
  units: Unit[] = [];

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized WorldData instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<WorldData>(WorldData.MESSAGE_TYPE, data);
  }
}



export class Tile implements TileInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.Tile";

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

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized Tile instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<Tile>(Tile.MESSAGE_TYPE, data);
  }
}



export class Unit implements UnitInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.Unit";

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

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized Unit instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<Unit>(Unit.MESSAGE_TYPE, data);
  }
}


/**
 * Rules engine terrain definition
 */
export class TerrainDefinition implements TerrainDefinitionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.TerrainDefinition";

  id: number = 0;
  name: string = "";
  /** double base_move_cost = 3;     // Base movement cost
 double defense_bonus = 4;      // Defense bonus multiplier (0.0 to 1.0) */
  type: number = 0;
  description: string = "";
  /** How this terrain impacts */
  unitProperties: Record<number, TerrainUnitProperties> = {};

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized TerrainDefinition instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<TerrainDefinition>(TerrainDefinition.MESSAGE_TYPE, data);
  }
}


/**
 * Rules engine unit definition
 */
export class UnitDefinition implements UnitDefinitionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UnitDefinition";

  id: number = 0;
  name: string = "";
  description: string = "";
  health: number = 0;
  coins: number = 0;
  movementPoints: number = 0;
  attackRange: number = 0;
  minAttackRange: number = 0;
  splashDamage: number = 0;
  terrainProperties: Record<number, TerrainUnitProperties> = {};
  properties: string[] = [];

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized UnitDefinition instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<UnitDefinition>(UnitDefinition.MESSAGE_TYPE, data);
  }
}


/**
 * Properties that are specific to unit on a particular terrain
 */
export class TerrainUnitProperties implements TerrainUnitPropertiesInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.TerrainUnitProperties";

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

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized TerrainUnitProperties instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<TerrainUnitProperties>(TerrainUnitProperties.MESSAGE_TYPE, data);
  }
}


/**
 * Properties for unit-vs-unit combat interactions
 */
export class UnitUnitProperties implements UnitUnitPropertiesInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UnitUnitProperties";

  attackerId: number = 0;
  defenderId: number = 0;
  damage?: DamageDistribution;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized UnitUnitProperties instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<UnitUnitProperties>(UnitUnitProperties.MESSAGE_TYPE, data);
  }
}


/**
 * Damage distribution for combat calculations
 */
export class DamageDistribution implements DamageDistributionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.DamageDistribution";

  minDamage: number = 0;
  maxDamage: number = 0;
  expectedDamage: number = 0;
  ranges: DamageRange[] = [];

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized DamageDistribution instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<DamageDistribution>(DamageDistribution.MESSAGE_TYPE, data);
  }
}



export class DamageRange implements DamageRangeInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.DamageRange";

  minValue: number = 0;
  maxValue: number = 0;
  probability: number = 0;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized DamageRange instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<DamageRange>(DamageRange.MESSAGE_TYPE, data);
  }
}


/**
 * Main rules engine definition - centralized source of truth
 */
export class RulesEngine implements RulesEngineInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.RulesEngine";

  /** Core entity definitions */
  units: Record<number, UnitDefinition> = {};
  terrains: Record<number, TerrainDefinition> = {};
  /** Centralized property definitions (source of truth)
 Key format: "terrain_id:unit_id" (e.g., "1:3" for terrain 1, unit 3) */
  terrainUnitProperties: Record<string, TerrainUnitProperties> = {};
  /** Key format: "attacker_id:defender_id" (e.g., "1:2" for unit 1 attacking unit 2) */
  unitUnitProperties: Record<string, UnitUnitProperties> = {};

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized RulesEngine instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<RulesEngine>(RulesEngine.MESSAGE_TYPE, data);
  }
}


/**
 * Describes a game and its metadata
 */
export class Game implements GameInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.Game";

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

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized Game instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<Game>(Game.MESSAGE_TYPE, data);
  }
}



export class GameConfiguration implements GameConfigurationInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameConfiguration";

  /** Player configuration */
  players: GamePlayer[] = [];
  /** Team configuration */
  teams: GameTeam[] = [];
  /** Game settings */
  settings?: GameSettings;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GameConfiguration instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GameConfiguration>(GameConfiguration.MESSAGE_TYPE, data);
  }
}



export class GamePlayer implements GamePlayerInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GamePlayer";

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

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GamePlayer instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GamePlayer>(GamePlayer.MESSAGE_TYPE, data);
  }
}



export class GameTeam implements GameTeamInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameTeam";

  /** ID of the team within the game (unique to the game) */
  teamId: number = 0;
  /** Name of the team - in a game */
  name: string = "";
  /** Just a color for this team */
  color: string = "";
  /** Whether team has active players - can also be inferred */
  isActive: boolean = false;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GameTeam instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GameTeam>(GameTeam.MESSAGE_TYPE, data);
  }
}



export class GameSettings implements GameSettingsInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameSettings";

  /** List of allowed unit type IDs */
  allowedUnits: number[] = [];
  /** Turn time limit in seconds (0 = no limit) */
  turnTimeLimit: number = 0;
  /** Team mode */
  teamMode: string = "";
  /** Maximum number of turns (0 = unlimited) */
  maxTurns: number = 0;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GameSettings instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GameSettings>(GameSettings.MESSAGE_TYPE, data);
  }
}


/**
 * Holds the game's Active/Current state (eg world state)
 */
export class GameState implements GameStateInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameState";

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

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GameState instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GameState>(GameState.MESSAGE_TYPE, data);
  }
}


/**
 * Holds the game's move history (can be used as a replay log)
 */
export class GameMoveHistory implements GameMoveHistoryInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameMoveHistory";

  /** Move history for the game */
  gameId: string = "";
  /** Each entry in our history is a "group" of moves */
  groups: GameMoveGroup[] = [];

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GameMoveHistory instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GameMoveHistory>(GameMoveHistory.MESSAGE_TYPE, data);
  }
}


/**
 * A move group - we can allow X moves in one "tick"
 */
export class GameMoveGroup implements GameMoveGroupInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameMoveGroup";

  /** When the moves happened (or were submitted) */
  startedAt?: Timestamp;
  endedAt?: Timestamp;
  /** *
 List of moves to add - */
  moves: GameMove[] = [];
  /** Each game move result stores the result of the individual Move in the request.
 ie move_results[i] = ResultOfProcessing(ProcessMoveRequest.moves[i]) */
  moveResults: GameMoveResult[] = [];

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GameMoveGroup instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GameMoveGroup>(GameMoveGroup.MESSAGE_TYPE, data);
  }
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

  player: number = 0;
  timestamp?: Timestamp;
  /** A monotonically increasing and unique (within the game) sequence number for the move
 This is generated by the server */
  sequenceNum: number = 0;
  moveUnit?: MoveUnitAction;
  attackUnit?: AttackUnitAction;
  endTurn?: EndTurnAction;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GameMove instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GameMove>(GameMove.MESSAGE_TYPE, data);
  }
}


/**
 * *
 Represents the result of executing a move
 */
export class GameMoveResult implements GameMoveResultInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameMoveResult";

  /** Whether the result is permenant and can be undone.
 Just moving a unit for example is not permanent, but attacking a unit
 would be (ie a player cannot undo it). */
  isPermanent: boolean = false;
  /** A monotonically increasing and unique (within the game) sequence number for the move */
  sequenceNum: number = 0;
  /** A set of changes to the world as a result of making this move */
  changes: WorldChange[] = [];

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GameMoveResult instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GameMoveResult>(GameMoveResult.MESSAGE_TYPE, data);
  }
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

  fromQ: number = 0;
  fromR: number = 0;
  toQ: number = 0;
  toR: number = 0;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized MoveUnitAction instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<MoveUnitAction>(MoveUnitAction.MESSAGE_TYPE, data);
  }
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

  attackerQ: number = 0;
  attackerR: number = 0;
  defenderQ: number = 0;
  defenderR: number = 0;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized AttackUnitAction instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<AttackUnitAction>(AttackUnitAction.MESSAGE_TYPE, data);
  }
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


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized EndTurnAction instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<EndTurnAction>(EndTurnAction.MESSAGE_TYPE, data);
  }
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

  unitMoved?: UnitMovedChange;
  unitDamaged?: UnitDamagedChange;
  unitKilled?: UnitKilledChange;
  playerChanged?: PlayerChangedChange;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized WorldChange instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<WorldChange>(WorldChange.MESSAGE_TYPE, data);
  }
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

  /** Complete unit state before the move */
  previousUnit?: Unit;
  /** Complete unit state after the move (includes updated position, distanceLeft, etc.) */
  updatedUnit?: Unit;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized UnitMovedChange instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<UnitMovedChange>(UnitMovedChange.MESSAGE_TYPE, data);
  }
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

  /** Complete unit state before taking damage */
  previousUnit?: Unit;
  /** Complete unit state after taking damage */
  updatedUnit?: Unit;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized UnitDamagedChange instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<UnitDamagedChange>(UnitDamagedChange.MESSAGE_TYPE, data);
  }
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

  /** Complete unit state before being killed */
  previousUnit?: Unit;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized UnitKilledChange instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<UnitKilledChange>(UnitKilledChange.MESSAGE_TYPE, data);
  }
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

  previousPlayer: number = 0;
  newPlayer: number = 0;
  previousTurn: number = 0;
  newTurn: number = 0;
  /** Units that had their movement/health reset for the new turn */
  resetUnits: Unit[] = [];

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized PlayerChangedChange instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<PlayerChangedChange>(PlayerChangedChange.MESSAGE_TYPE, data);
  }
}


/**
 * GameInfo represents a game in the catalog
 */
export class GameInfo implements GameInfoInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GameInfo";

  id: string = "";
  name: string = "";
  description: string = "";
  category: string = "";
  difficulty: string = "";
  tags: string[] = [];
  icon: string = "";
  lastUpdated: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GameInfo instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GameInfo>(GameInfo.MESSAGE_TYPE, data);
  }
}


/**
 * Request messages
 */
export class ListGamesRequest implements ListGamesRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ListGamesRequest";

  /** Pagination info */
  pagination?: Pagination;
  /** May be filter by owner id */
  ownerId: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ListGamesRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ListGamesRequest>(ListGamesRequest.MESSAGE_TYPE, data);
  }
}



export class ListGamesResponse implements ListGamesResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ListGamesResponse";

  items: Game[] = [];
  pagination?: PaginationResponse;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ListGamesResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ListGamesResponse>(ListGamesResponse.MESSAGE_TYPE, data);
  }
}



export class GetGameRequest implements GetGameRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetGameRequest";

  id: string = "";
  version: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetGameRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetGameRequest>(GetGameRequest.MESSAGE_TYPE, data);
  }
}



export class GetGameResponse implements GetGameResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetGameResponse";

  game?: Game;
  state?: GameState;
  history?: GameMoveHistory;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetGameResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetGameResponse>(GetGameResponse.MESSAGE_TYPE, data);
  }
}



export class GetGameContentRequest implements GetGameContentRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetGameContentRequest";

  id: string = "";
  version: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetGameContentRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetGameContentRequest>(GetGameContentRequest.MESSAGE_TYPE, data);
  }
}



export class GetGameContentResponse implements GetGameContentResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetGameContentResponse";

  weewarContent: string = "";
  recipeContent: string = "";
  readmeContent: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetGameContentResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetGameContentResponse>(GetGameContentResponse.MESSAGE_TYPE, data);
  }
}



export class UpdateGameRequest implements UpdateGameRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UpdateGameRequest";

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

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized UpdateGameRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<UpdateGameRequest>(UpdateGameRequest.MESSAGE_TYPE, data);
  }
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

  /** *
 Game being updated */
  game?: Game;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized UpdateGameResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<UpdateGameResponse>(UpdateGameResponse.MESSAGE_TYPE, data);
  }
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

  /** *
 ID of the game to be deleted. */
  id: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized DeleteGameRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<DeleteGameRequest>(DeleteGameRequest.MESSAGE_TYPE, data);
  }
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


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized DeleteGameResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<DeleteGameResponse>(DeleteGameResponse.MESSAGE_TYPE, data);
  }
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

  /** *
 IDs of the game to be fetched */
  ids: string[] = [];

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetGamesRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetGamesRequest>(GetGamesRequest.MESSAGE_TYPE, data);
  }
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

  games: Record<string, Game> = {};

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetGamesResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetGamesResponse>(GetGamesResponse.MESSAGE_TYPE, data);
  }
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

  /** *
 Game being updated */
  game?: Game;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized CreateGameRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<CreateGameRequest>(CreateGameRequest.MESSAGE_TYPE, data);
  }
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

  /** *
 Game being created */
  game?: Game;
  /** The starting game state */
  gameState?: GameState;
  /** *
 Error specific to a field if there are any errors. */
  fieldErrors: Record<string, string> = {};

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized CreateGameResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<CreateGameResponse>(CreateGameResponse.MESSAGE_TYPE, data);
  }
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

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ProcessMovesRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ProcessMovesRequest>(ProcessMovesRequest.MESSAGE_TYPE, data);
  }
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

  /** *
 Each game move result stores the result of the individual Move in the request.
 ie move_results[i] = ResultOfProcessing(ProcessMoveRequest.moves[i]) */
  moveResults: GameMoveResult[] = [];
  /** *
 List of changes that resulted from the moves on the game state as a whole
 For example 10 moves could have resulted in 2 unit creations and 4 city changes

 It is not clear if this is needed.  For example concatenating all changes from all the move_results *may* suffice
 as long as the MoveProcessor is making sure that updates are atomic and snapshots the world state before 
 starting a snapshot (and not just a move) */
  changes: WorldChange[] = [];

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ProcessMovesResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ProcessMovesResponse>(ProcessMovesResponse.MESSAGE_TYPE, data);
  }
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

  /** *
 Game ID to add moves to */
  gameId: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetGameStateRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetGameStateRequest>(GetGameStateRequest.MESSAGE_TYPE, data);
  }
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

  state?: GameState;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetGameStateResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetGameStateResponse>(GetGameStateResponse.MESSAGE_TYPE, data);
  }
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

  /** *
 Game ID to add moves to */
  gameId: string = "";
  /** Offset of the move to begin fetching from in reverse order from "latest".
 0 => start from now */
  offset: number = 0;
  /** *
 Limit to last N moves (from offset).  if <= 0 return all moves */
  lastN: number = 0;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ListMovesRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ListMovesRequest>(ListMovesRequest.MESSAGE_TYPE, data);
  }
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

  /** Whether there are more moves before this */
  hasMore: boolean = false;
  moveGroups: GameMoveGroup[] = [];

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ListMovesResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ListMovesResponse>(ListMovesResponse.MESSAGE_TYPE, data);
  }
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

  gameId: string = "";
  q: number = 0;
  r: number = 0;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetOptionsAtRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetOptionsAtRequest>(GetOptionsAtRequest.MESSAGE_TYPE, data);
  }
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

  options: GameOption[] = [];
  currentPlayer: number = 0;
  gameInitialized: boolean = false;
  /** A Path from source to dest along with cost on each tile for tracking */
  allPaths?: AllPaths;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetOptionsAtResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetOptionsAtResponse>(GetOptionsAtResponse.MESSAGE_TYPE, data);
  }
}


/**
 * Compact representation of all reachable paths from a source
 */
export class AllPaths implements AllPathsInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.AllPaths";

  /** Starting coordinate for all paths */
  sourceQ: number = 0;
  sourceR: number = 0;
  /** Map of edges: key is "toQ,toR" for quick parent lookup
 Each edge represents the optimal way to reach 'to' from its parent */
  edges: Record<string, PathEdge> = {};

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized AllPaths instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<AllPaths>(AllPaths.MESSAGE_TYPE, data);
  }
}


/**
 * A single edge in a path with movement details
 */
export class PathEdge implements PathEdgeInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.PathEdge";

  fromQ: number = 0;
  fromR: number = 0;
  toQ: number = 0;
  toR: number = 0;
  movementCost: number = 0;
  totalCost: number = 0;
  terrainType: string = "";
  explanation: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized PathEdge instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<PathEdge>(PathEdge.MESSAGE_TYPE, data);
  }
}


/**
 * Full path from source to destination (constructed on-demand from AllPaths)
 */
export class Path implements PathInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.Path";

  /** Edges in order from source to destination */
  edges: PathEdge[] = [];
  /** len(directions) = len(edges) - 1
 and directions[i] = direction from edge[i - 1] -> edge[i] */
  directions: PathDirection[] = [];
  /** Sum of all edge costs */
  totalCost: number = 0;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized Path instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<Path>(Path.MESSAGE_TYPE, data);
  }
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

  move?: MoveOption;
  attack?: AttackOption;
  endTurn?: EndTurnOption;
  build?: BuildUnitOption;
  capture?: CaptureBuildingOption;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GameOption instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GameOption>(GameOption.MESSAGE_TYPE, data);
  }
}


/**
 * *
 Option to end the current turn
 */
export class EndTurnOption implements EndTurnOptionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.EndTurnOption";


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized EndTurnOption instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<EndTurnOption>(EndTurnOption.MESSAGE_TYPE, data);
  }
}


/**
 * *
 Option to move to a specific coordinate
 */
export class MoveOption implements MoveOptionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.MoveOption";

  movementCost: number = 0;
  /** Ready-to-use action object for ProcessMoves */
  action?: MoveUnitAction;
  /** Debug fields */
  reconstructedPath?: Path;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized MoveOption instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<MoveOption>(MoveOption.MESSAGE_TYPE, data);
  }
}


/**
 * *
 A possible attack target
 */
export class AttackOption implements AttackOptionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.AttackOption";

  /** Target unit type and health */
  targetUnitType: number = 0;
  targetUnitHealth: number = 0;
  canAttack: boolean = false;
  damageEstimate: number = 0;
  /** Ready-to-use action object for ProcessMoves */
  action?: AttackUnitAction;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized AttackOption instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<AttackOption>(AttackOption.MESSAGE_TYPE, data);
  }
}


/**
 * *
 An option to build a unit (at a city tile)
 */
export class BuildUnitOption implements BuildUnitOptionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.BuildUnitOption";

  q: number = 0;
  r: number = 0;
  unitType: number = 0;
  cost: number = 0;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized BuildUnitOption instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<BuildUnitOption>(BuildUnitOption.MESSAGE_TYPE, data);
  }
}


/**
 * *
 A move where a unit can capture a building
 */
export class CaptureBuildingOption implements CaptureBuildingOptionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.CaptureBuildingOption";

  q: number = 0;
  r: number = 0;
  tileType: number = 0;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized CaptureBuildingOption instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<CaptureBuildingOption>(CaptureBuildingOption.MESSAGE_TYPE, data);
  }
}



export class EmptyRequest implements EmptyRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.EmptyRequest";


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized EmptyRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<EmptyRequest>(EmptyRequest.MESSAGE_TYPE, data);
  }
}



export class EmptyResponse implements EmptyResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.EmptyResponse";


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized EmptyResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<EmptyResponse>(EmptyResponse.MESSAGE_TYPE, data);
  }
}


/**
 * Request to fetch data from a URL
 */
export class SetContentRequest implements SetContentRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SetContentRequest";

  innerHtml: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized SetContentRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<SetContentRequest>(SetContentRequest.MESSAGE_TYPE, data);
  }
}


/**
 * Response from fetch
 */
export class SetContentResponse implements SetContentResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SetContentResponse";


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized SetContentResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<SetContentResponse>(SetContentResponse.MESSAGE_TYPE, data);
  }
}


/**
 * Request to fetch data from a URL
 */
export class LogMessageRequest implements LogMessageRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.LogMessageRequest";

  message: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized LogMessageRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<LogMessageRequest>(LogMessageRequest.MESSAGE_TYPE, data);
  }
}


/**
 * Response from fetch
 */
export class LogMessageResponse implements LogMessageResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.LogMessageResponse";


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized LogMessageResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<LogMessageResponse>(LogMessageResponse.MESSAGE_TYPE, data);
  }
}


/**
 * Request to fetch data from a URL
 */
export class SetGameStateRequest implements SetGameStateRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SetGameStateRequest";

  game?: Game;
  state?: GameState;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized SetGameStateRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<SetGameStateRequest>(SetGameStateRequest.MESSAGE_TYPE, data);
  }
}


/**
 * Response from fetch
 */
export class SetGameStateResponse implements SetGameStateResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SetGameStateResponse";


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized SetGameStateResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<SetGameStateResponse>(SetGameStateResponse.MESSAGE_TYPE, data);
  }
}


/**
 * Request to update game UI status (current player, turn counter)
 */
export class UpdateGameStatusRequest implements UpdateGameStatusRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UpdateGameStatusRequest";

  currentPlayer: number = 0;
  turnCounter: number = 0;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized UpdateGameStatusRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<UpdateGameStatusRequest>(UpdateGameStatusRequest.MESSAGE_TYPE, data);
  }
}



export class UpdateGameStatusResponse implements UpdateGameStatusResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UpdateGameStatusResponse";


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized UpdateGameStatusResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<UpdateGameStatusResponse>(UpdateGameStatusResponse.MESSAGE_TYPE, data);
  }
}


/**
 * Request to set a tile at a specific coordinate
 */
export class SetTileAtRequest implements SetTileAtRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SetTileAtRequest";

  q: number = 0;
  r: number = 0;
  tile?: Tile;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized SetTileAtRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<SetTileAtRequest>(SetTileAtRequest.MESSAGE_TYPE, data);
  }
}



export class SetTileAtResponse implements SetTileAtResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SetTileAtResponse";


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized SetTileAtResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<SetTileAtResponse>(SetTileAtResponse.MESSAGE_TYPE, data);
  }
}


/**
 * Request to set a unit at a specific coordinate
 */
export class SetUnitAtRequest implements SetUnitAtRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SetUnitAtRequest";

  q: number = 0;
  r: number = 0;
  unit?: Unit;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized SetUnitAtRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<SetUnitAtRequest>(SetUnitAtRequest.MESSAGE_TYPE, data);
  }
}



export class SetUnitAtResponse implements SetUnitAtResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SetUnitAtResponse";


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized SetUnitAtResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<SetUnitAtResponse>(SetUnitAtResponse.MESSAGE_TYPE, data);
  }
}


/**
 * Request to remove a tile at a specific coordinate
 */
export class RemoveTileAtRequest implements RemoveTileAtRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.RemoveTileAtRequest";

  q: number = 0;
  r: number = 0;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized RemoveTileAtRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<RemoveTileAtRequest>(RemoveTileAtRequest.MESSAGE_TYPE, data);
  }
}



export class RemoveTileAtResponse implements RemoveTileAtResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.RemoveTileAtResponse";


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized RemoveTileAtResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<RemoveTileAtResponse>(RemoveTileAtResponse.MESSAGE_TYPE, data);
  }
}


/**
 * Request to remove a unit at a specific coordinate
 */
export class RemoveUnitAtRequest implements RemoveUnitAtRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.RemoveUnitAtRequest";

  q: number = 0;
  r: number = 0;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized RemoveUnitAtRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<RemoveUnitAtRequest>(RemoveUnitAtRequest.MESSAGE_TYPE, data);
  }
}



export class RemoveUnitAtResponse implements RemoveUnitAtResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.RemoveUnitAtResponse";


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized RemoveUnitAtResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<RemoveUnitAtResponse>(RemoveUnitAtResponse.MESSAGE_TYPE, data);
  }
}


/**
 * Request to show highlights on the game board
 */
export class ShowHighlightsRequest implements ShowHighlightsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ShowHighlightsRequest";

  highlights: HighlightSpec[] = [];

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ShowHighlightsRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ShowHighlightsRequest>(ShowHighlightsRequest.MESSAGE_TYPE, data);
  }
}



export class ShowHighlightsResponse implements ShowHighlightsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ShowHighlightsResponse";


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ShowHighlightsResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ShowHighlightsResponse>(ShowHighlightsResponse.MESSAGE_TYPE, data);
  }
}


/**
 * Specification for a single highlight
 */
export class HighlightSpec implements HighlightSpecInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.HighlightSpec";

  q: number = 0;
  r: number = 0;
  type: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized HighlightSpec instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<HighlightSpec>(HighlightSpec.MESSAGE_TYPE, data);
  }
}


/**
 * Request to clear highlights
 */
export class ClearHighlightsRequest implements ClearHighlightsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ClearHighlightsRequest";

  types: string[] = [];

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ClearHighlightsRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ClearHighlightsRequest>(ClearHighlightsRequest.MESSAGE_TYPE, data);
  }
}



export class ClearHighlightsResponse implements ClearHighlightsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ClearHighlightsResponse";


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ClearHighlightsResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ClearHighlightsResponse>(ClearHighlightsResponse.MESSAGE_TYPE, data);
  }
}


/**
 * Request to show a path on the game board
 */
export class ShowPathRequest implements ShowPathRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ShowPathRequest";

  coords: number[] = [];
  color: number = 0;
  thickness: number = 0;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ShowPathRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ShowPathRequest>(ShowPathRequest.MESSAGE_TYPE, data);
  }
}



export class ShowPathResponse implements ShowPathResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ShowPathResponse";


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ShowPathResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ShowPathResponse>(ShowPathResponse.MESSAGE_TYPE, data);
  }
}


/**
 * Request to clear paths
 */
export class ClearPathsRequest implements ClearPathsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ClearPathsRequest";


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ClearPathsRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ClearPathsRequest>(ClearPathsRequest.MESSAGE_TYPE, data);
  }
}



export class ClearPathsResponse implements ClearPathsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ClearPathsResponse";


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ClearPathsResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ClearPathsResponse>(ClearPathsResponse.MESSAGE_TYPE, data);
  }
}


/**
 * Called when a turn option is clicked in TurnOptionsPanel
 */
export class TurnOptionClickedRequest implements TurnOptionClickedRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.TurnOptionClickedRequest";

  gameId: string = "";
  optionIndex: number = 0;
  optionType: string = "";
  q: number = 0;
  r: number = 0;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized TurnOptionClickedRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<TurnOptionClickedRequest>(TurnOptionClickedRequest.MESSAGE_TYPE, data);
  }
}


/**
 * Response of a turn option click
 */
export class TurnOptionClickedResponse implements TurnOptionClickedResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.TurnOptionClickedResponse";

  gameId: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized TurnOptionClickedResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<TurnOptionClickedResponse>(TurnOptionClickedResponse.MESSAGE_TYPE, data);
  }
}


/**
 * Called when the scene was clicked
 */
export class SceneClickedRequest implements SceneClickedRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SceneClickedRequest";

  gameId: string = "";
  q: number = 0;
  r: number = 0;
  layer: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized SceneClickedRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<SceneClickedRequest>(SceneClickedRequest.MESSAGE_TYPE, data);
  }
}


/**
 * Response of a turn option click
 */
export class SceneClickedResponse implements SceneClickedResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.SceneClickedResponse";

  gameId: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized SceneClickedResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<SceneClickedResponse>(SceneClickedResponse.MESSAGE_TYPE, data);
  }
}


/**
 * Called when the end turn button was clicked
 */
export class EndTurnButtonClickedRequest implements EndTurnButtonClickedRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.EndTurnButtonClickedRequest";

  gameId: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized EndTurnButtonClickedRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<EndTurnButtonClickedRequest>(EndTurnButtonClickedRequest.MESSAGE_TYPE, data);
  }
}


/**
 * Response of a turn option click
 */
export class EndTurnButtonClickedResponse implements EndTurnButtonClickedResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.EndTurnButtonClickedResponse";

  gameId: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized EndTurnButtonClickedResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<EndTurnButtonClickedResponse>(EndTurnButtonClickedResponse.MESSAGE_TYPE, data);
  }
}


/**
 * Called when the end turn button was clicked
 */
export class InitializeGameRequest implements InitializeGameRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.InitializeGameRequest";

  gameData: string = "";
  gameState: string = "";
  moveHistory: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized InitializeGameRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<InitializeGameRequest>(InitializeGameRequest.MESSAGE_TYPE, data);
  }
}


/**
 * Response of a turn option click
 */
export class InitializeGameResponse implements InitializeGameResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.InitializeGameResponse";

  success: boolean = false;
  error: string = "";
  /** Initial UI state information */
  currentPlayer: number = 0;
  turnCounter: number = 0;
  gameName: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized InitializeGameResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<InitializeGameResponse>(InitializeGameResponse.MESSAGE_TYPE, data);
  }
}


/**
 * ThemeInfo contains metadata about a theme
 */
export class ThemeInfo implements ThemeInfoInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ThemeInfo";

  name: string = "";
  version: string = "";
  basePath: string = "";
  assetType: string = "";
  needsPostProcessing: boolean = false;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ThemeInfo instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ThemeInfo>(ThemeInfo.MESSAGE_TYPE, data);
  }
}


/**
 * UnitMapping maps a unit ID to its theme-specific representation
 */
export class UnitMapping implements UnitMappingInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UnitMapping";

  old: string = "";
  name: string = "";
  image: string = "";
  description: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized UnitMapping instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<UnitMapping>(UnitMapping.MESSAGE_TYPE, data);
  }
}


/**
 * TerrainMapping maps a terrain ID to its theme-specific representation
 */
export class TerrainMapping implements TerrainMappingInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.TerrainMapping";

  old: string = "";
  name: string = "";
  image: string = "";
  description: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized TerrainMapping instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<TerrainMapping>(TerrainMapping.MESSAGE_TYPE, data);
  }
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

  themeInfo?: ThemeInfo;
  units: Record<number, UnitMapping> = {};
  terrains: Record<number, TerrainMapping> = {};

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ThemeManifest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ThemeManifest>(ThemeManifest.MESSAGE_TYPE, data);
  }
}


/**
 * PlayerColor defines the color scheme for a player
 */
export class PlayerColor implements PlayerColorInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.PlayerColor";

  primary: string = "";
  secondary: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized PlayerColor instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<PlayerColor>(PlayerColor.MESSAGE_TYPE, data);
  }
}


/**
 * AssetResult represents a rendered asset
 */
export class AssetResult implements AssetResultInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.AssetResult";

  type: Type = Type.TYPE_UNSPECIFIED;
  data: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized AssetResult instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<AssetResult>(AssetResult.MESSAGE_TYPE, data);
  }
}


/**
 * UserInfo represents a user in the catalog
 */
export class UserInfo implements UserInfoInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UserInfo";

  id: string = "";
  name: string = "";
  description: string = "";
  category: string = "";
  difficulty: string = "";
  tags: string[] = [];
  icon: string = "";
  lastUpdated: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized UserInfo instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<UserInfo>(UserInfo.MESSAGE_TYPE, data);
  }
}


/**
 * Request messages
 */
export class ListUsersRequest implements ListUsersRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ListUsersRequest";

  /** Pagination info */
  pagination?: Pagination;
  /** May be filter by owner id */
  ownerId: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ListUsersRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ListUsersRequest>(ListUsersRequest.MESSAGE_TYPE, data);
  }
}



export class ListUsersResponse implements ListUsersResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ListUsersResponse";

  items: User[] = [];
  pagination?: PaginationResponse;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ListUsersResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ListUsersResponse>(ListUsersResponse.MESSAGE_TYPE, data);
  }
}



export class GetUserRequest implements GetUserRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetUserRequest";

  id: string = "";
  version: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetUserRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetUserRequest>(GetUserRequest.MESSAGE_TYPE, data);
  }
}



export class GetUserResponse implements GetUserResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetUserResponse";

  user?: User;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetUserResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetUserResponse>(GetUserResponse.MESSAGE_TYPE, data);
  }
}



export class GetUserContentRequest implements GetUserContentRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetUserContentRequest";

  id: string = "";
  version: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetUserContentRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetUserContentRequest>(GetUserContentRequest.MESSAGE_TYPE, data);
  }
}



export class GetUserContentResponse implements GetUserContentResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetUserContentResponse";

  weewarContent: string = "";
  recipeContent: string = "";
  readmeContent: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetUserContentResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetUserContentResponse>(GetUserContentResponse.MESSAGE_TYPE, data);
  }
}



export class UpdateUserRequest implements UpdateUserRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UpdateUserRequest";

  /** *
 User being updated */
  user?: User;
  /** *
 Mask of fields being updated in this User to make partial changes. */
  updateMask?: FieldMask;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized UpdateUserRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<UpdateUserRequest>(UpdateUserRequest.MESSAGE_TYPE, data);
  }
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

  /** *
 User being updated */
  user?: User;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized UpdateUserResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<UpdateUserResponse>(UpdateUserResponse.MESSAGE_TYPE, data);
  }
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

  /** *
 ID of the user to be deleted. */
  id: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized DeleteUserRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<DeleteUserRequest>(DeleteUserRequest.MESSAGE_TYPE, data);
  }
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


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized DeleteUserResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<DeleteUserResponse>(DeleteUserResponse.MESSAGE_TYPE, data);
  }
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

  /** *
 IDs of the user to be fetched */
  ids: string[] = [];

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetUsersRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetUsersRequest>(GetUsersRequest.MESSAGE_TYPE, data);
  }
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

  users: Record<string, User> = {};

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetUsersResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetUsersResponse>(GetUsersResponse.MESSAGE_TYPE, data);
  }
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

  /** *
 User being updated */
  user?: User;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized CreateUserRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<CreateUserRequest>(CreateUserRequest.MESSAGE_TYPE, data);
  }
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

  /** *
 User being created */
  user?: User;
  /** *
 Error specific to a field if there are any errors. */
  fieldErrors: Record<string, string> = {};

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized CreateUserResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<CreateUserResponse>(CreateUserResponse.MESSAGE_TYPE, data);
  }
}


/**
 * WorldInfo represents a world in the catalog
 */
export class WorldInfo implements WorldInfoInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.WorldInfo";

  id: string = "";
  name: string = "";
  description: string = "";
  category: string = "";
  difficulty: string = "";
  tags: string[] = [];
  icon: string = "";
  lastUpdated: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized WorldInfo instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<WorldInfo>(WorldInfo.MESSAGE_TYPE, data);
  }
}


/**
 * Request messages
 */
export class ListWorldsRequest implements ListWorldsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ListWorldsRequest";

  /** Pagination info */
  pagination?: Pagination;
  /** May be filter by owner id */
  ownerId: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ListWorldsRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ListWorldsRequest>(ListWorldsRequest.MESSAGE_TYPE, data);
  }
}



export class ListWorldsResponse implements ListWorldsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.ListWorldsResponse";

  items: World[] = [];
  pagination?: PaginationResponse;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ListWorldsResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<ListWorldsResponse>(ListWorldsResponse.MESSAGE_TYPE, data);
  }
}



export class GetWorldRequest implements GetWorldRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetWorldRequest";

  id: string = "";
  version: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetWorldRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetWorldRequest>(GetWorldRequest.MESSAGE_TYPE, data);
  }
}



export class GetWorldResponse implements GetWorldResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.GetWorldResponse";

  world?: World;
  worldData?: WorldData;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetWorldResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetWorldResponse>(GetWorldResponse.MESSAGE_TYPE, data);
  }
}



export class UpdateWorldRequest implements UpdateWorldRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "weewar.v1.UpdateWorldRequest";

  /** *
 World being updated */
  world?: World;
  worldData?: WorldData;
  clearWorld: boolean = false;
  /** *
 Mask of fields being updated in this World to make partial changes. */
  updateMask?: FieldMask;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized UpdateWorldRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<UpdateWorldRequest>(UpdateWorldRequest.MESSAGE_TYPE, data);
  }
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

  /** *
 World being updated */
  world?: World;
  worldData?: WorldData;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized UpdateWorldResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<UpdateWorldResponse>(UpdateWorldResponse.MESSAGE_TYPE, data);
  }
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

  /** *
 ID of the world to be deleted. */
  id: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized DeleteWorldRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<DeleteWorldRequest>(DeleteWorldRequest.MESSAGE_TYPE, data);
  }
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


  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized DeleteWorldResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<DeleteWorldResponse>(DeleteWorldResponse.MESSAGE_TYPE, data);
  }
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

  /** *
 IDs of the world to be fetched */
  ids: string[] = [];

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetWorldsRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetWorldsRequest>(GetWorldsRequest.MESSAGE_TYPE, data);
  }
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

  worlds: Record<string, World> = {};

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetWorldsResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<GetWorldsResponse>(GetWorldsResponse.MESSAGE_TYPE, data);
  }
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

  /** *
 World being updated */
  world?: World;
  worldData?: WorldData;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized CreateWorldRequest instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<CreateWorldRequest>(CreateWorldRequest.MESSAGE_TYPE, data);
  }
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

  /** *
 World being created */
  world?: World;
  worldData?: WorldData;
  /** *
 Error specific to a field if there are any errors. */
  fieldErrors: Record<string, string> = {};

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized CreateWorldResponse instance or null if creation failed
   */
  static from(data: any) {
    return Weewar_v1Deserializer.from<CreateWorldResponse>(CreateWorldResponse.MESSAGE_TYPE, data);
  }
}


