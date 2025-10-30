// Generated TypeScript interfaces from proto file
// DO NOT EDIT - This file is auto-generated

import { FieldMask, Timestamp } from "@bufbuild/protobuf/wkt";


/**
 * /////// Game related models
 */
export enum GameStatus {
  GAME_STATUS_UNSPECIFIED = 0,
  GAME_STATUS_PLAYING = 1,
  GAME_STATUS_PAUSED = 2,
  GAME_STATUS_ENDED = 3,
}


export enum PathDirection {
  PATH_DIRECTION_UNSPECIFIED = 0,
  PATH_DIRECTION_LEFT = 1,
  PATH_DIRECTION_TOP_LEFT = 2,
  PATH_DIRECTION_TOP_RIGHT = 3,
  PATH_DIRECTION_RIGHT = 4,
  PATH_DIRECTION_BOTTOM_RIGHT = 5,
  PATH_DIRECTION_BOTTOM_LEFT = 6,
}


export enum Type {
  TYPE_UNSPECIFIED = 0,
  TYPE_PATH = 1,
  TYPE_SVG = 2,
  TYPE_DATA_URL = 3,
}



export interface User {
  createdAt?: Timestamp;
  updatedAt?: Timestamp;
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
  createdAt?: Timestamp;
  updatedAt?: Timestamp;
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
  /** URL to screenshot/preview image (defaults to /worlds/{id}/screenshot)
 Can be overridden to point to CDN or external hosting */
  screenshotUrl: string;
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
  shortcut: string;
  /** Keep track of turns when the move was last made and when a "top up" was last done on/for this tile.
 This helps us not having to "top up" or "reset" the stats at the end
 of each turn.  Instead as the game turn is incremented we can do a 
 lazy reset for any unit or tile where unit_or_tile.last_toppedup_turn < game.curent_turn
 So we just have to increment the game_turn and the unit is automaticaly flagged as
 needing a top up of its health/balance/movement etc */
  lastActedTurn: number;
  lastToppedupTurn: number;
}



export interface Unit {
  /** Q and R in Cubed coordinates */
  q: number;
  r: number;
  player: number;
  unitType: number;
  shortcut: string;
  /** Runtime state fields */
  availableHealth: number;
  distanceLeft: number;
  /** Keep track of turns when the move was last made and when a "top up" was last done on/for this tile.
 This helps us not having to "top up" or "reset" the stats at the end
 of each turn.  Instead as the game turn is incremented we can do a 
 lazy reset for any unit or tile where unit_or_tile.last_toppedup_turn < game.curent_turn
 So we just have to increment the game_turn and the unit is automaticaly flagged as
 needing a top up of its health/balance/movement etc */
  lastActedTurn: number;
  lastToppedupTurn: number;
  /** Details around wound bonus tracking for this turn */
  attacksReceivedThisTurn: number;
  attackHistory?: AttackRecord[];
  /** Action progression tracking - index into UnitDefinition.action_order
 Indicates which step in the action sequence the unit is currently on
 Reset to 0 at turn start via TopUpUnitIfNeeded() */
  progressionStep: number;
  /** When current step has pipe-separated alternatives (e.g., "attack|capture"),
 this tracks which alternative the user chose, preventing switching mid-step
 Cleared when advancing to next step */
  chosenAlternative: string;
}



export interface AttackRecord {
  q: number;
  r: number;
  isRanged: boolean;
  turnNumber: number;
}


/**
 * Rules engine terrain definition
 */
export interface TerrainDefinition {
  id: number;
  name: string;
  /** double base_move_cost = 3;     // Base movement cost
 double defense_bonus = 4;      // Defense bonus multiplier (0.0 to 1.0) */
  type: number;
  description: string;
  /** How this terrain impacts */
  unitProperties: Record<number, TerrainUnitProperties>;
  /** List of units that can be built on this terrain */
  buildableUnitIds: number[];
}


/**
 * Rules engine unit definition
 */
export interface UnitDefinition {
  id: number;
  name: string;
  description: string;
  health: number;
  coins: number;
  movementPoints: number;
  retreatPoints: number;
  defense: number;
  attackRange: number;
  minAttackRange: number;
  splashDamage: number;
  terrainProperties: Record<number, TerrainUnitProperties>;
  properties: string[];
  /** Unit classification for attack calculations */
  unitClass: string;
  unitTerrain: string;
  /** Attack table: base attack values against different unit classes
 Key format: "Light:Air", "Heavy:Land", "Stealth:Water", etc.
 Value 0 or missing key means "n/a" (cannot attack) */
  attackVsClass: Record<string, number>;
  /** Ordered list of allowed actions this turn
 Examples:
   ["move", "attack"] - can move then attack
   ["move", "attack|capture"] - can move then either attack or capture
   ["attack"] - can only attack (no movement)
 Default if empty: ["move", "attack|capture"] */
  actionOrder: string[];
  /** How many times each action type can be performed per turn
 Key: action name, Value: max count
 Example: {"attack": 2} means can attack twice
 Default if not specified: 1 per action type */
  actionLimits: Record<string, number>;
}


/**
 * Properties that are specific to unit on a particular terrain
 */
export interface TerrainUnitProperties {
  terrainId: number;
  unitId: number;
  movementCost: number;
  healingBonus: number;
  canBuild: boolean;
  canCapture: boolean;
  attackBonus: number;
  defenseBonus: number;
  attackRange: number;
  minAttackRange: number;
}


/**
 * Properties for unit-vs-unit combat interactions
 */
export interface UnitUnitProperties {
  attackerId: number;
  defenderId: number;
  attackOverride?: number | undefined;
  defenseOverride?: number | undefined;
  damage?: DamageDistribution;
}


/**
 * Damage distribution for combat calculations
 */
export interface DamageDistribution {
  minDamage: number;
  maxDamage: number;
  expectedDamage: number;
  ranges?: DamageRange[];
}



export interface DamageRange {
  minValue: number;
  maxValue: number;
  probability: number;
}


/**
 * Main rules engine definition - centralized source of truth
 */
export interface RulesEngine {
  /** Core entity definitions */
  units: Record<number, UnitDefinition>;
  terrains: Record<number, TerrainDefinition>;
  /** Centralized property definitions (source of truth)
 Key format: "terrain_id:unit_id" (e.g., "1:3" for terrain 1, unit 3) */
  terrainUnitProperties: Record<string, TerrainUnitProperties>;
  /** Key format: "attacker_id:defender_id" (e.g., "1:2" for unit 1 attacking unit 2) */
  unitUnitProperties: Record<string, UnitUnitProperties>;
}


/**
 * Describes a game and its metadata
 */
export interface Game {
  createdAt?: Timestamp;
  updatedAt?: Timestamp;
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
  /** URL to screenshot/preview image (defaults to /games/{id}/screenshot)
 Can be overridden to point to CDN or external hosting */
  screenshotUrl: string;
}



export interface GameConfiguration {
  /** Player configuration */
  players?: GamePlayer[];
  /** Team configuration */
  teams?: GameTeam[];
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
  /** Nickname for the player in this game */
  name: string;
  /** Whether play is still in the game - can this just be inferred? */
  isActive: boolean;
}



export interface GameTeam {
  /** ID of the team within the game (unique to the game) */
  teamId: number;
  /** Name of the team - in a game */
  name: string;
  /** Just a color for this team */
  color: string;
  /** Whether team has active players - can also be inferred */
  isActive: boolean;
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
  updatedAt?: Timestamp;
  /** ID of the game whos state is being tracked */
  gameId: string;
  turnCounter: number;
  currentPlayer: number;
  /** Current world state */
  worldData?: WorldData;
  /** Current state hash for validation */
  stateHash: string;
  /** Version number for optimistic locking */
  version: number;
  status: GameStatus;
  /** Only set after a win has been possible */
  finished: boolean;
  winningPlayer: number;
  winningTeam: number;
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
  startedAt?: Timestamp;
  endedAt?: Timestamp;
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
  timestamp?: Timestamp;
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
  /** Complete unit state before the move */
  previousUnit?: Unit;
  /** Complete unit state after the move (includes updated position, distanceLeft, etc.) */
  updatedUnit?: Unit;
}


/**
 * *
 A unit took damage
 */
export interface UnitDamagedChange {
  /** Complete unit state before taking damage */
  previousUnit?: Unit;
  /** Complete unit state after taking damage */
  updatedUnit?: Unit;
}


/**
 * *
 A unit was killed
 */
export interface UnitKilledChange {
  /** Complete unit state before being killed */
  previousUnit?: Unit;
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
  /** Units that had their movement/health reset for the new turn */
  resetUnits?: Unit[];
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
  updateMask?: FieldMask;
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
  games: Record<string, Game>;
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
  fieldErrors: Record<string, string>;
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
  /** A Path from source to dest along with cost on each tile for tracking */
  allPaths?: AllPaths;
}


/**
 * Compact representation of all reachable paths from a source
 */
export interface AllPaths {
  /** Starting coordinate for all paths */
  sourceQ: number;
  sourceR: number;
  /** Map of edges: key is "toQ,toR" for quick parent lookup
 Each edge represents the optimal way to reach 'to' from its parent */
  edges: Record<string, PathEdge>;
}


/**
 * A single edge in a path with movement details
 */
export interface PathEdge {
  fromQ: number;
  fromR: number;
  toQ: number;
  toR: number;
  movementCost: number;
  totalCost: number;
  terrainType: string;
  explanation: string;
}


/**
 * Full path from source to destination (constructed on-demand from AllPaths)
 */
export interface Path {
  /** Edges in order from source to destination */
  edges?: PathEdge[];
  /** len(directions) = len(edges) - 1
 and directions[i] = direction from edge[i - 1] -> edge[i] */
  directions: PathDirection[];
  /** Sum of all edge costs */
  totalCost: number;
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
  movementCost: number;
  /** Ready-to-use action object for ProcessMoves */
  action?: MoveUnitAction;
  /** Debug fields */
  reconstructedPath?: Path;
}


/**
 * *
 A possible attack target
 */
export interface AttackOption {
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
  unitType: number;
  cost: number;
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
 * *
 Request for simulating combat between two units
 */
export interface SimulateAttackRequest {
  attackerUnitType: number;
  attackerTerrain: number;
  attackerHealth: number;
  defenderUnitType: number;
  defenderTerrain: number;
  defenderHealth: number;
  woundBonus: number;
  numSimulations: number;
}


/**
 * *
 Response containing damage distribution statistics
 */
export interface SimulateAttackResponse {
  /** Damage distributions: damage_value -> number_of_occurrences */
  attackerDamageDistribution: Record<number, number>;
  defenderDamageDistribution: Record<number, number>;
  /** Statistical summary */
  attackerMeanDamage: number;
  defenderMeanDamage: number;
  attackerKillProbability: number;
  defenderKillProbability: number;
}



export interface EmptyRequest {
}



export interface EmptyResponse {
}


/**
 * Request to fetch data from a URL
 */
export interface SetContentRequest {
  innerHtml: string;
}


/**
 * Response from fetch
 */
export interface SetContentResponse {
}


/**
 * Request to fetch data from a URL
 */
export interface LogMessageRequest {
  message: string;
}


/**
 * Response from fetch
 */
export interface LogMessageResponse {
}


/**
 * Request to fetch data from a URL
 */
export interface SetGameStateRequest {
  game?: Game;
  state?: GameState;
}


/**
 * Response from fetch
 */
export interface SetGameStateResponse {
}


/**
 * Request to update game UI status (current player, turn counter)
 */
export interface UpdateGameStatusRequest {
  currentPlayer: number;
  turnCounter: number;
}



export interface UpdateGameStatusResponse {
}


/**
 * Request to set a tile at a specific coordinate
 */
export interface SetTileAtRequest {
  q: number;
  r: number;
  tile?: Tile;
}



export interface SetTileAtResponse {
}


/**
 * Request to set a unit at a specific coordinate
 */
export interface SetUnitAtRequest {
  q: number;
  r: number;
  unit?: Unit;
}



export interface SetUnitAtResponse {
}


/**
 * Request to remove a tile at a specific coordinate
 */
export interface RemoveTileAtRequest {
  q: number;
  r: number;
}



export interface RemoveTileAtResponse {
}


/**
 * Request to remove a unit at a specific coordinate
 */
export interface RemoveUnitAtRequest {
  q: number;
  r: number;
}



export interface RemoveUnitAtResponse {
}


/**
 * Request to show highlights on the game board
 */
export interface ShowHighlightsRequest {
  highlights?: HighlightSpec[];
}



export interface ShowHighlightsResponse {
}


/**
 * Specification for a single highlight
 */
export interface HighlightSpec {
  q: number;
  r: number;
  type: string;
}


/**
 * Request to clear highlights
 */
export interface ClearHighlightsRequest {
  types: string[];
}



export interface ClearHighlightsResponse {
}


/**
 * Request to show a path on the game board
 */
export interface ShowPathRequest {
  coords: number[];
  color: number;
  thickness: number;
}



export interface ShowPathResponse {
}


/**
 * Request to clear paths
 */
export interface ClearPathsRequest {
}



export interface ClearPathsResponse {
}


/**
 * Request to animate unit movement along a path
 */
export interface MoveUnitAnimationRequest {
  unit?: Unit;
  path?: HexCoord[];
}



export interface MoveUnitAnimationResponse {
}


/**
 * Hex coordinate for paths
 */
export interface HexCoord {
  q: number;
  r: number;
}


/**
 * Request to show attack effect animation
 */
export interface ShowAttackEffectRequest {
  fromQ: number;
  fromR: number;
  toQ: number;
  toR: number;
  damage: number;
  splashTargets?: SplashTarget[];
}



export interface SplashTarget {
  q: number;
  r: number;
  damage: number;
}



export interface ShowAttackEffectResponse {
}


/**
 * Request to show heal effect animation
 */
export interface ShowHealEffectRequest {
  q: number;
  r: number;
  amount: number;
}



export interface ShowHealEffectResponse {
}


/**
 * Request to show capture effect animation
 */
export interface ShowCaptureEffectRequest {
  q: number;
  r: number;
}



export interface ShowCaptureEffectResponse {
}


/**
 * Request to set unit with optional animation
 */
export interface SetUnitAtAnimationRequest {
  q: number;
  r: number;
  unit?: Unit;
  flash: boolean;
  appear: boolean;
}



export interface SetUnitAtAnimationResponse {
}


/**
 * Request to remove unit with optional animation
 */
export interface RemoveUnitAtAnimationRequest {
  q: number;
  r: number;
  animate: boolean;
}



export interface RemoveUnitAtAnimationResponse {
}


/**
 * Called when a turn option is clicked in TurnOptionsPanel
 */
export interface TurnOptionClickedRequest {
  gameId: string;
  optionIndex: number;
  optionType: string;
  q: number;
  r: number;
}


/**
 * Response of a turn option click
 */
export interface TurnOptionClickedResponse {
  gameId: string;
}


/**
 * Called when the scene was clicked
 */
export interface SceneClickedRequest {
  gameId: string;
  q: number;
  r: number;
  layer: string;
}


/**
 * Response of a turn option click
 */
export interface SceneClickedResponse {
  gameId: string;
}


/**
 * Called when the end turn button was clicked
 */
export interface EndTurnButtonClickedRequest {
  gameId: string;
}


/**
 * Response of a turn option click
 */
export interface EndTurnButtonClickedResponse {
  gameId: string;
}


/**
 * Called when the end turn button was clicked
 */
export interface InitializeGameRequest {
  gameData: string;
  gameState: string;
  moveHistory: string;
}


/**
 * Response of a turn option click
 */
export interface InitializeGameResponse {
  success: boolean;
  error: string;
  /** Initial UI state information */
  currentPlayer: number;
  turnCounter: number;
  gameName: string;
}


/**
 * ThemeInfo contains metadata about a theme
 */
export interface ThemeInfo {
  name: string;
  version: string;
  basePath: string;
  assetType: string;
  needsPostProcessing: boolean;
}


/**
 * UnitMapping maps a unit ID to its theme-specific representation
 */
export interface UnitMapping {
  old: string;
  name: string;
  image: string;
  description: string;
}


/**
 * TerrainMapping maps a terrain ID to its theme-specific representation
 */
export interface TerrainMapping {
  old: string;
  name: string;
  image: string;
  description: string;
}


/**
 * ThemeManifest represents the full theme configuration
 This matches the structure of mapping.json files
 */
export interface ThemeManifest {
  themeInfo?: ThemeInfo;
  units: Record<number, UnitMapping>;
  terrains: Record<number, TerrainMapping>;
}


/**
 * PlayerColor defines the color scheme for a player
 */
export interface PlayerColor {
  primary: string;
  secondary: string;
}


/**
 * AssetResult represents a rendered asset
 */
export interface AssetResult {
  type: Type;
  data: string;
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
  updateMask?: FieldMask;
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
  users: Record<string, User>;
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
  fieldErrors: Record<string, string>;
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
  updateMask?: FieldMask;
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
  worlds: Record<string, World>;
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
  fieldErrors: Record<string, string>;
}

