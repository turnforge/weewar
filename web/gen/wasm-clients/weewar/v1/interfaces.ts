// Generated TypeScript interfaces from proto file
// DO NOT EDIT - This file is auto-generated





export interface User {
  createdAt?: Date;
  updatedAt?: Date;
  /** Unique ID for the user */
  id: string;
  /** Name if items have names */
  name: string;
  /** Description if user has a description */
  description: string;
  /** Some tags */
  tags: string[];
  /** A possible image url */
  imageUrl: string;
  /** Difficulty - example attribute */
  difficulty: string;
}



export interface Pagination {
  /** *
 Instead of an offset an abstract  "page" key is provided that offers
 an opaque "pointer" into some offset in a result set. */
  pageKey: string;
  /** *
 If a pagekey is not supported we can also support a direct integer offset
 for cases where it makes sense. */
  pageOffset: number;
  /** *
 Number of results to return. */
  pageSize: number;
}



export interface PaginationResponse {
  /** *
 The key/pointer string that subsequent List requests should pass to
 continue the pagination. */
  nextPageKey: string;
  /** *
 Also support an integer offset if possible */
  nextPageOffset: number;
  /** *
 Whether theere are more results. */
  hasMore: boolean;
  /** *
 Total number of results. */
  totalResults: number;
}



export interface World {
  createdAt?: Date;
  updatedAt?: Date;
  /** Unique ID for the world */
  id: string;
  /** User that created the world */
  creatorId: string;
  /** Name if items have names */
  name: string;
  /** Description if world has a description */
  description: string;
  /** Some tags */
  tags: string[];
  /** A possible image url */
  imageUrl: string;
  /** Difficulty - example attribute */
  difficulty: string;
  /** The actual world contents/data */
  worldData?: WorldData;
}



export interface WorldData {
  /** JSON-fied tile data about what units and terrains are at each location */
  tiles?: Tile[];
  /** All units on the world and who they belong to */
  units?: Unit[];
}



export interface Tile {
  /** Q and R in Cubed coordinates */
  q: number;
  r: number;
  tileType: number;
  /** Whether the tile itself belongs to a player */
  player: number;
}



export interface Unit {
  /** Q and R in Cubed coordinates */
  q: number;
  r: number;
  player: number;
  unitType: number;
  /** Runtime state fields */
  availableHealth: number;
  distanceLeft: number;
  turnCounter: number;
}


/**
 * Rules engine terrain definition
 */
export interface TerrainDefinition {
  id: number;
  name: string;
  baseMoveCost: number;
  defenseBonus: number;
  type: number;
  description: string;
}


/**
 * Rules engine unit definition
 */
export interface UnitDefinition {
  id: number;
  name: string;
  movementPoints: number;
  attackRange: number;
  health: number;
  properties: string[];
}


/**
 * Movement cost matrix for unit types on terrain types
 */
export interface MovementMatrix {
  /** Map of unit_id -> (terrain_id -> movement_cost) */
  costs?: Map<number, TerrainCostMap>;
}



export interface TerrainCostMap {
  /** Map of terrain_id -> movement_cost */
  terrainCosts?: Map<number, number>;
}


/**
 * Describes a game and its metadata
 */
export interface Game {
  createdAt?: Date;
  updatedAt?: Date;
  /** Unique ID for the game */
  id: string;
  /** User who started/created the game */
  creatorId: string;
  /** The world this game was created from */
  worldId: string;
  /** Name if items have names */
  name: string;
  /** Description if game has a description */
  description: string;
  /** Some tags */
  tags: string[];
  /** A possible image url */
  imageUrl: string;
  /** Difficulty - example attribute */
  difficulty: string;
  /** Game configuration */
  config?: GameConfiguration;
}



export interface GameConfiguration {
  /** Player configuration */
  players?: GamePlayer[];
  /** Game settings */
  settings?: GameSettings;
}



export interface GamePlayer {
  /** Player ID (1-based) */
  playerId: number;
  /** Player type */
  playerType: string;
  /** Player color */
  color: string;
  /** Team ID (0 = no team, 1+ = team number) */
  teamId: number;
}



export interface GameSettings {
  /** List of allowed unit type IDs */
  allowedUnits: number[];
  /** Turn time limit in seconds (0 = no limit) */
  turnTimeLimit: number;
  /** Team mode */
  teamMode: string;
  /** Maximum number of turns (0 = unlimited) */
  maxTurns: number;
}


/**
 * Holds the game's Active/Current state (eg world state)
 */
export interface GameState {
  updatedAt?: Date;
  /** ID of the game whos state is being tracked */
  gameId: string;
  turnCounter: number;
  currentPlayer: number;
  /** Current world state */
  worldData?: WorldData;
}


/**
 * Holds the game's move history (can be used as a replay log)
 */
export interface GameMoveHistory {
  /** Move history for the game */
  gameId: string;
  /** Each entry in our history is a "group" of moves */
  groups?: GameMoveGroup[];
}


/**
 * A move group - we can allow X moves in one "tick"
 */
export interface GameMoveGroup {
  /** When the moves happened (or were submitted) */
  startedAt?: Date;
  endedAt?: Date;
  /** *
 List of moves to add - */
  moves?: GameMove[];
  /** Each game move result stores the result of the individual Move in the request.
 ie move_results[i] = ResultOfProcessing(ProcessMoveRequest.moves[i]) */
  moveResults?: GameMoveResult[];
}


/**
 * *
 Represents a single move which can be one of many actions in the game
 */
export interface GameMove {
  player: number;
  timestamp?: Date;
  /** A monotonically increasing and unique (within the game) sequence number for the move
 This is generated by the server */
  sequenceNum: number;
  moveUnit?: MoveUnitAction;
  attackUnit?: AttackUnitAction;
  endTurn?: EndTurnAction;
}


/**
 * *
 Represents the result of executing a move
 */
export interface GameMoveResult {
  /** Whether the result is permenant and can be undone.
 Just moving a unit for example is not permanent, but attacking a unit
 would be (ie a player cannot undo it). */
  isPermanent: boolean;
  /** A monotonically increasing and unique (within the game) sequence number for the move */
  sequenceNum: number;
  /** A set of changes to the world as a result of making this move */
  changes?: WorldChange[];
}


/**
 * *
 Move unit from one position to another
 */
export interface MoveUnitAction {
  fromQ: number;
  fromR: number;
  toQ: number;
  toR: number;
}


/**
 * *
 Attack with one unit against another
 */
export interface AttackUnitAction {
  attackerQ: number;
  attackerR: number;
  defenderQ: number;
  defenderR: number;
}


/**
 * *
 End current player's turn
 */
export interface EndTurnAction {
}


/**
 * *
 Represents a change to the game world
 */
export interface WorldChange {
  unitMoved?: UnitMovedChange;
  unitDamaged?: UnitDamagedChange;
  unitKilled?: UnitKilledChange;
  playerChanged?: PlayerChangedChange;
}


/**
 * *
 A unit moved from one position to another
 */
export interface UnitMovedChange {
  fromQ: number;
  fromR: number;
  toQ: number;
  toR: number;
}


/**
 * *
 A unit took damage
 */
export interface UnitDamagedChange {
  previousHealth: number;
  newHealth: number;
  q: number;
  r: number;
}


/**
 * *
 A unit was killed
 */
export interface UnitKilledChange {
  player: number;
  unitType: number;
  q: number;
  r: number;
}


/**
 * *
 Active player changed
 */
export interface PlayerChangedChange {
  previousPlayer: number;
  newPlayer: number;
  previousTurn: number;
  newTurn: number;
}


/**
 * GameInfo represents a game in the catalog
 */
export interface GameInfo {
  id: string;
  name: string;
  description: string;
  category: string;
  difficulty: string;
  tags: string[];
  icon: string;
  lastUpdated: string;
}


/**
 * Request messages
 */
export interface ListGamesRequest {
  /** Pagination info */
  pagination?: Pagination;
  /** May be filter by owner id */
  ownerId: string;
}



export interface ListGamesResponse {
  items?: Game[];
  pagination?: PaginationResponse;
}



export interface GetGameRequest {
  id: string;
  version: string;
}



export interface GetGameResponse {
  game?: Game;
  state?: GameState;
  history?: GameMoveHistory;
}



export interface GetGameContentRequest {
  id: string;
  version: string;
}



export interface GetGameContentResponse {
  weewarContent: string;
  recipeContent: string;
  readmeContent: string;
}



export interface UpdateGameRequest {
  /** Game id to modify */
  gameId: string;
  /** *
 Game being updated */
  newGame?: Game;
  /** New world state to save */
  newState?: GameState;
  /** History to save */
  newHistory?: GameMoveHistory;
  /** *
 Mask of fields being updated in this Game to make partial changes. */
  updateMask?: string[];
}


/**
 * *
 The request for (partially) updating an Game.
 */
export interface UpdateGameResponse {
  /** *
 Game being updated */
  game?: Game;
}


/**
 * *
 Request to delete an game.
 */
export interface DeleteGameRequest {
  /** *
 ID of the game to be deleted. */
  id: string;
}


/**
 * *
 Game deletion response
 */
export interface DeleteGameResponse {
}


/**
 * *
 Request to batch get games
 */
export interface GetGamesRequest {
  /** *
 IDs of the game to be fetched */
  ids: string[];
}


/**
 * *
 Game batch-get response
 */
export interface GetGamesResponse {
  games?: Map<string, Game>;
}


/**
 * *
 Game creation request object
 */
export interface CreateGameRequest {
  /** *
 Game being updated */
  game?: Game;
}


/**
 * *
 Response of an game creation.
 */
export interface CreateGameResponse {
  /** *
 Game being created */
  game?: Game;
  /** The starting game state */
  gameState?: GameState;
  /** *
 Error specific to a field if there are any errors. */
  fieldErrors?: Map<string, string>;
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
export interface ProcessMovesRequest {
  /** *
 Game ID to add moves to */
  gameId: string;
  /** *
 List of moves to add */
  moves?: GameMove[];
}


/**
 * *
 Response after adding moves to game.
 */
export interface ProcessMovesResponse {
  /** *
 Each game move result stores the result of the individual Move in the request.
 ie move_results[i] = ResultOfProcessing(ProcessMoveRequest.moves[i]) */
  moveResults?: GameMoveResult[];
  /** *
 List of changes that resulted from the moves on the game state as a whole
 For example 10 moves could have resulted in 2 unit creations and 4 city changes

 It is not clear if this is needed.  For example concatenating all changes from all the move_results *may* suffice
 as long as the MoveProcessor is making sure that updates are atomic and snapshots the world state before 
 starting a snapshot (and not just a move) */
  changes?: WorldChange[];
}


/**
 * *
 Request to get the game's latest state
 */
export interface GetGameStateRequest {
  /** *
 Game ID to add moves to */
  gameId: string;
}


/**
 * *
 Response holding latest game state
 */
export interface GetGameStateResponse {
  state?: GameState;
}


/**
 * *
 Request to list moves for a game
 */
export interface ListMovesRequest {
  /** *
 Game ID to add moves to */
  gameId: string;
  /** Offset of the move to begin fetching from in reverse order from "latest".
 0 => start from now */
  offset: number;
  /** *
 Limit to last N moves (from offset).  if <= 0 return all moves */
  lastN: number;
}


/**
 * *
 Response after adding moves to game.
 */
export interface ListMovesResponse {
  /** Whether there are more moves before this */
  hasMore: boolean;
  moveGroups?: GameMoveGroup[];
}


/**
 * *
 Request to get all available options at a position
 */
export interface GetOptionsAtRequest {
  gameId: string;
  q: number;
  r: number;
}


/**
 * *
 Response with all available options at a position
 */
export interface GetOptionsAtResponse {
  options?: GameOption[];
  currentPlayer: number;
  gameInitialized: boolean;
}


/**
 * *
 A single game option available at a position
 */
export interface GameOption {
  move?: MoveOption;
  attack?: AttackOption;
  endTurn?: EndTurnOption;
  build?: BuildUnitOption;
  capture?: CaptureBuildingOption;
}


/**
 * *
 Option to end the current turn
 */
export interface EndTurnOption {
}


/**
 * *
 Option to move to a specific coordinate
 */
export interface MoveOption {
  q: number;
  r: number;
  movementCost: number;
  /** Ready-to-use action object for ProcessMoves */
  action?: MoveUnitAction;
}


/**
 * *
 A possible attack target
 */
export interface AttackOption {
  q: number;
  r: number;
  /** Target unit type and health */
  targetUnitType: number;
  targetUnitHealth: number;
  canAttack: boolean;
  damageEstimate: number;
  /** Ready-to-use action object for ProcessMoves */
  action?: AttackUnitAction;
}


/**
 * *
 An option to build a unit (at a city tile)
 */
export interface BuildUnitOption {
  q: number;
  r: number;
  tileType: number;
  buildCost: number;
}


/**
 * *
 A move where a unit can capture a building
 */
export interface CaptureBuildingOption {
  q: number;
  r: number;
  tileType: number;
}


/**
 * UserInfo represents a user in the catalog
 */
export interface UserInfo {
  id: string;
  name: string;
  description: string;
  category: string;
  difficulty: string;
  tags: string[];
  icon: string;
  lastUpdated: string;
}


/**
 * Request messages
 */
export interface ListUsersRequest {
  /** Pagination info */
  pagination?: Pagination;
  /** May be filter by owner id */
  ownerId: string;
}



export interface ListUsersResponse {
  items?: User[];
  pagination?: PaginationResponse;
}



export interface GetUserRequest {
  id: string;
  version: string;
}



export interface GetUserResponse {
  user?: User;
}



export interface GetUserContentRequest {
  id: string;
  version: string;
}



export interface GetUserContentResponse {
  weewarContent: string;
  recipeContent: string;
  readmeContent: string;
}



export interface UpdateUserRequest {
  /** *
 User being updated */
  user?: User;
  /** *
 Mask of fields being updated in this User to make partial changes. */
  updateMask?: string[];
}


/**
 * *
 The request for (partially) updating an User.
 */
export interface UpdateUserResponse {
  /** *
 User being updated */
  user?: User;
}


/**
 * *
 Request to delete an user.
 */
export interface DeleteUserRequest {
  /** *
 ID of the user to be deleted. */
  id: string;
}


/**
 * *
 User deletion response
 */
export interface DeleteUserResponse {
}


/**
 * *
 Request to batch get users
 */
export interface GetUsersRequest {
  /** *
 IDs of the user to be fetched */
  ids: string[];
}


/**
 * *
 User batch-get response
 */
export interface GetUsersResponse {
  users?: Map<string, User>;
}


/**
 * *
 User creation request object
 */
export interface CreateUserRequest {
  /** *
 User being updated */
  user?: User;
}


/**
 * *
 Response of an user creation.
 */
export interface CreateUserResponse {
  /** *
 User being created */
  user?: User;
  /** *
 Error specific to a field if there are any errors. */
  fieldErrors?: Map<string, string>;
}


/**
 * WorldInfo represents a world in the catalog
 */
export interface WorldInfo {
  id: string;
  name: string;
  description: string;
  category: string;
  difficulty: string;
  tags: string[];
  icon: string;
  lastUpdated: string;
}


/**
 * Request messages
 */
export interface ListWorldsRequest {
  /** Pagination info */
  pagination?: Pagination;
  /** May be filter by owner id */
  ownerId: string;
}



export interface ListWorldsResponse {
  items?: World[];
  pagination?: PaginationResponse;
}



export interface GetWorldRequest {
  id: string;
  version: string;
}



export interface GetWorldResponse {
  world?: World;
  worldData?: WorldData;
}



export interface UpdateWorldRequest {
  /** *
 World being updated */
  world?: World;
  worldData?: WorldData;
  clearWorld: boolean;
  /** *
 Mask of fields being updated in this World to make partial changes. */
  updateMask?: string[];
}


/**
 * *
 The request for (partially) updating an World.
 */
export interface UpdateWorldResponse {
  /** *
 World being updated */
  world?: World;
  worldData?: WorldData;
}


/**
 * *
 Request to delete an world.
 */
export interface DeleteWorldRequest {
  /** *
 ID of the world to be deleted. */
  id: string;
}


/**
 * *
 World deletion response
 */
export interface DeleteWorldResponse {
}


/**
 * *
 Request to batch get worlds
 */
export interface GetWorldsRequest {
  /** *
 IDs of the world to be fetched */
  ids: string[];
}


/**
 * *
 World batch-get response
 */
export interface GetWorldsResponse {
  worlds?: Map<string, World>;
}


/**
 * *
 World creation request object
 */
export interface CreateWorldRequest {
  /** *
 World being updated */
  world?: World;
  worldData?: WorldData;
}


/**
 * *
 Response of an world creation.
 */
export interface CreateWorldResponse {
  /** *
 World being created */
  world?: World;
  worldData?: WorldData;
  /** *
 Error specific to a field if there are any errors. */
  fieldErrors?: Map<string, string>;
}

