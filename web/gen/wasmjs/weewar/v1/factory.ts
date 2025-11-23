// Generated TypeScript factory + deserializer (annotation-based)
// DO NOT EDIT - This file is auto-generated


import { MessageTypeConstructor, BaseDeserializer, FactoryInterface } from "@protoc-gen-go-wasmjs/runtime";


import { weewar_v1SchemaRegistry } from "./schemas";




import {AllPaths as AllPathsInterface,AssetResult as AssetResultInterface,AttackRecord as AttackRecordInterface,AttackUnitAction as AttackUnitActionInterface,BuildOptionClickedRequest as BuildOptionClickedRequestInterface,BuildOptionClickedResponse as BuildOptionClickedResponseInterface,BuildUnitAction as BuildUnitActionInterface,CaptureBuildingAction as CaptureBuildingActionInterface,ClearHighlightsRequest as ClearHighlightsRequestInterface,ClearHighlightsResponse as ClearHighlightsResponseInterface,ClearPathsRequest as ClearPathsRequestInterface,ClearPathsResponse as ClearPathsResponseInterface,CoinsChangedChange as CoinsChangedChangeInterface,CreateGameRequest as CreateGameRequestInterface,CreateGameResponse as CreateGameResponseInterface,CreateIndexRecordsLRORequest as CreateIndexRecordsLRORequestInterface,CreateIndexRecordsLROResponse as CreateIndexRecordsLROResponseInterface,CreateUserRequest as CreateUserRequestInterface,CreateUserResponse as CreateUserResponseInterface,CreateWorldRequest as CreateWorldRequestInterface,CreateWorldResponse as CreateWorldResponseInterface,DamageDistribution as DamageDistributionInterface,DamageRange as DamageRangeInterface,DeleteGameRequest as DeleteGameRequestInterface,DeleteGameResponse as DeleteGameResponseInterface,DeleteIndexStatesRequest as DeleteIndexStatesRequestInterface,DeleteIndexStatesResponse as DeleteIndexStatesResponseInterface,DeleteUserRequest as DeleteUserRequestInterface,DeleteUserResponse as DeleteUserResponseInterface,DeleteWorldRequest as DeleteWorldRequestInterface,DeleteWorldResponse as DeleteWorldResponseInterface,EmptyRequest as EmptyRequestInterface,EmptyResponse as EmptyResponseInterface,EndTurnAction as EndTurnActionInterface,EndTurnButtonClickedRequest as EndTurnButtonClickedRequestInterface,EndTurnButtonClickedResponse as EndTurnButtonClickedResponseInterface,EnsureIndexStateRequest as EnsureIndexStateRequestInterface,EnsureIndexStateResponse as EnsureIndexStateResponseInterface,Game as GameInterface,GameConfiguration as GameConfigurationInterface,GameMove as GameMoveInterface,GameMoveGroup as GameMoveGroupInterface,GameMoveHistory as GameMoveHistoryInterface,GameOption as GameOptionInterface,GamePlayer as GamePlayerInterface,GameSettings as GameSettingsInterface,GameState as GameStateInterface,GameTeam as GameTeamInterface,GetGameContentRequest as GetGameContentRequestInterface,GetGameContentResponse as GetGameContentResponseInterface,GetGameRequest as GetGameRequestInterface,GetGameResponse as GetGameResponseInterface,GetGameStateRequest as GetGameStateRequestInterface,GetGameStateResponse as GetGameStateResponseInterface,GetGamesRequest as GetGamesRequestInterface,GetGamesResponse as GetGamesResponseInterface,GetIndexRecordsLRORequest as GetIndexRecordsLRORequestInterface,GetIndexRecordsLROResponse as GetIndexRecordsLROResponseInterface,GetIndexStatesRequest as GetIndexStatesRequestInterface,GetIndexStatesResponse as GetIndexStatesResponseInterface,GetOptionsAtRequest as GetOptionsAtRequestInterface,GetOptionsAtResponse as GetOptionsAtResponseInterface,GetUserContentRequest as GetUserContentRequestInterface,GetUserContentResponse as GetUserContentResponseInterface,GetUserRequest as GetUserRequestInterface,GetUserResponse as GetUserResponseInterface,GetUsersRequest as GetUsersRequestInterface,GetUsersResponse as GetUsersResponseInterface,GetWorldRequest as GetWorldRequestInterface,GetWorldResponse as GetWorldResponseInterface,GetWorldsRequest as GetWorldsRequestInterface,GetWorldsResponse as GetWorldsResponseInterface,HexCoord as HexCoordInterface,HighlightSpec as HighlightSpecInterface,IncomeConfig as IncomeConfigInterface,IndexInfo as IndexInfoInterface,IndexRecord as IndexRecordInterface,IndexRecordsLRO as IndexRecordsLROInterface,IndexState as IndexStateInterface,IndexStateList as IndexStateListInterface,InitializeGameRequest as InitializeGameRequestInterface,InitializeGameResponse as InitializeGameResponseInterface,InitializeSingletonRequest as InitializeSingletonRequestInterface,InitializeSingletonResponse as InitializeSingletonResponseInterface,Job as JobInterface,ListGamesRequest as ListGamesRequestInterface,ListGamesResponse as ListGamesResponseInterface,ListIndexStatesRequest as ListIndexStatesRequestInterface,ListIndexStatesResponse as ListIndexStatesResponseInterface,ListMovesRequest as ListMovesRequestInterface,ListMovesResponse as ListMovesResponseInterface,ListUsersRequest as ListUsersRequestInterface,ListUsersResponse as ListUsersResponseInterface,ListWorldsRequest as ListWorldsRequestInterface,ListWorldsResponse as ListWorldsResponseInterface,LogMessageRequest as LogMessageRequestInterface,LogMessageResponse as LogMessageResponseInterface,MoveUnitAction as MoveUnitActionInterface,MoveUnitRequest as MoveUnitRequestInterface,MoveUnitResponse as MoveUnitResponseInterface,Pagination as PaginationInterface,PaginationResponse as PaginationResponseInterface,Path as PathInterface,PathEdge as PathEdgeInterface,PlayerChangedChange as PlayerChangedChangeInterface,PlayerColor as PlayerColorInterface,ProcessMovesRequest as ProcessMovesRequestInterface,ProcessMovesResponse as ProcessMovesResponseInterface,RemoveTileAtRequest as RemoveTileAtRequestInterface,RemoveTileAtResponse as RemoveTileAtResponseInterface,RemoveUnitAtRequest as RemoveUnitAtRequestInterface,RemoveUnitAtResponse as RemoveUnitAtResponseInterface,RepeatInfo as RepeatInfoInterface,RulesEngine as RulesEngineInterface,Run as RunInterface,SceneClickedRequest as SceneClickedRequestInterface,SceneClickedResponse as SceneClickedResponseInterface,SetAllowedPanelsRequest as SetAllowedPanelsRequestInterface,SetAllowedPanelsResponse as SetAllowedPanelsResponseInterface,SetContentRequest as SetContentRequestInterface,SetContentResponse as SetContentResponseInterface,SetGameStateRequest as SetGameStateRequestInterface,SetGameStateResponse as SetGameStateResponseInterface,SetTileAtRequest as SetTileAtRequestInterface,SetTileAtResponse as SetTileAtResponseInterface,SetUnitAtRequest as SetUnitAtRequestInterface,SetUnitAtResponse as SetUnitAtResponseInterface,ShowAttackEffectRequest as ShowAttackEffectRequestInterface,ShowAttackEffectResponse as ShowAttackEffectResponseInterface,ShowBuildOptionsRequest as ShowBuildOptionsRequestInterface,ShowBuildOptionsResponse as ShowBuildOptionsResponseInterface,ShowCaptureEffectRequest as ShowCaptureEffectRequestInterface,ShowCaptureEffectResponse as ShowCaptureEffectResponseInterface,ShowHealEffectRequest as ShowHealEffectRequestInterface,ShowHealEffectResponse as ShowHealEffectResponseInterface,ShowHighlightsRequest as ShowHighlightsRequestInterface,ShowHighlightsResponse as ShowHighlightsResponseInterface,ShowPathRequest as ShowPathRequestInterface,ShowPathResponse as ShowPathResponseInterface,SimulateAttackRequest as SimulateAttackRequestInterface,SimulateAttackResponse as SimulateAttackResponseInterface,SplashTarget as SplashTargetInterface,TerrainDefinition as TerrainDefinitionInterface,TerrainMapping as TerrainMappingInterface,TerrainUnitProperties as TerrainUnitPropertiesInterface,ThemeInfo as ThemeInfoInterface,ThemeManifest as ThemeManifestInterface,Tile as TileInterface,TurnOptionClickedRequest as TurnOptionClickedRequestInterface,TurnOptionClickedResponse as TurnOptionClickedResponseInterface,Unit as UnitInterface,UnitBuiltChange as UnitBuiltChangeInterface,UnitDamagedChange as UnitDamagedChangeInterface,UnitDefinition as UnitDefinitionInterface,UnitKilledChange as UnitKilledChangeInterface,UnitMapping as UnitMappingInterface,UnitMovedChange as UnitMovedChangeInterface,UnitUnitProperties as UnitUnitPropertiesInterface,UpdateGameRequest as UpdateGameRequestInterface,UpdateGameResponse as UpdateGameResponseInterface,UpdateGameStatusRequest as UpdateGameStatusRequestInterface,UpdateGameStatusResponse as UpdateGameStatusResponseInterface,UpdateIndexRecordsLRORequest as UpdateIndexRecordsLRORequestInterface,UpdateIndexRecordsLROResponse as UpdateIndexRecordsLROResponseInterface,UpdateUserRequest as UpdateUserRequestInterface,UpdateUserResponse as UpdateUserResponseInterface,UpdateWorldRequest as UpdateWorldRequestInterface,UpdateWorldResponse as UpdateWorldResponseInterface,User as UserInterface,UserInfo as UserInfoInterface,World as WorldInterface,WorldChange as WorldChangeInterface,WorldData as WorldDataInterface,WorldInfo as WorldInfoInterface} from "./models/interfaces";

import {AllPaths as ConcreteAllPaths,AssetResult as ConcreteAssetResult,AttackRecord as ConcreteAttackRecord,AttackUnitAction as ConcreteAttackUnitAction,BuildOptionClickedRequest as ConcreteBuildOptionClickedRequest,BuildOptionClickedResponse as ConcreteBuildOptionClickedResponse,BuildUnitAction as ConcreteBuildUnitAction,CaptureBuildingAction as ConcreteCaptureBuildingAction,ClearHighlightsRequest as ConcreteClearHighlightsRequest,ClearHighlightsResponse as ConcreteClearHighlightsResponse,ClearPathsRequest as ConcreteClearPathsRequest,ClearPathsResponse as ConcreteClearPathsResponse,CoinsChangedChange as ConcreteCoinsChangedChange,CreateGameRequest as ConcreteCreateGameRequest,CreateGameResponse as ConcreteCreateGameResponse,CreateIndexRecordsLRORequest as ConcreteCreateIndexRecordsLRORequest,CreateIndexRecordsLROResponse as ConcreteCreateIndexRecordsLROResponse,CreateUserRequest as ConcreteCreateUserRequest,CreateUserResponse as ConcreteCreateUserResponse,CreateWorldRequest as ConcreteCreateWorldRequest,CreateWorldResponse as ConcreteCreateWorldResponse,DamageDistribution as ConcreteDamageDistribution,DamageRange as ConcreteDamageRange,DeleteGameRequest as ConcreteDeleteGameRequest,DeleteGameResponse as ConcreteDeleteGameResponse,DeleteIndexStatesRequest as ConcreteDeleteIndexStatesRequest,DeleteIndexStatesResponse as ConcreteDeleteIndexStatesResponse,DeleteUserRequest as ConcreteDeleteUserRequest,DeleteUserResponse as ConcreteDeleteUserResponse,DeleteWorldRequest as ConcreteDeleteWorldRequest,DeleteWorldResponse as ConcreteDeleteWorldResponse,EmptyRequest as ConcreteEmptyRequest,EmptyResponse as ConcreteEmptyResponse,EndTurnAction as ConcreteEndTurnAction,EndTurnButtonClickedRequest as ConcreteEndTurnButtonClickedRequest,EndTurnButtonClickedResponse as ConcreteEndTurnButtonClickedResponse,EnsureIndexStateRequest as ConcreteEnsureIndexStateRequest,EnsureIndexStateResponse as ConcreteEnsureIndexStateResponse,Game as ConcreteGame,GameConfiguration as ConcreteGameConfiguration,GameMove as ConcreteGameMove,GameMoveGroup as ConcreteGameMoveGroup,GameMoveHistory as ConcreteGameMoveHistory,GameOption as ConcreteGameOption,GamePlayer as ConcreteGamePlayer,GameSettings as ConcreteGameSettings,GameState as ConcreteGameState,GameTeam as ConcreteGameTeam,GetGameContentRequest as ConcreteGetGameContentRequest,GetGameContentResponse as ConcreteGetGameContentResponse,GetGameRequest as ConcreteGetGameRequest,GetGameResponse as ConcreteGetGameResponse,GetGameStateRequest as ConcreteGetGameStateRequest,GetGameStateResponse as ConcreteGetGameStateResponse,GetGamesRequest as ConcreteGetGamesRequest,GetGamesResponse as ConcreteGetGamesResponse,GetIndexRecordsLRORequest as ConcreteGetIndexRecordsLRORequest,GetIndexRecordsLROResponse as ConcreteGetIndexRecordsLROResponse,GetIndexStatesRequest as ConcreteGetIndexStatesRequest,GetIndexStatesResponse as ConcreteGetIndexStatesResponse,GetOptionsAtRequest as ConcreteGetOptionsAtRequest,GetOptionsAtResponse as ConcreteGetOptionsAtResponse,GetUserContentRequest as ConcreteGetUserContentRequest,GetUserContentResponse as ConcreteGetUserContentResponse,GetUserRequest as ConcreteGetUserRequest,GetUserResponse as ConcreteGetUserResponse,GetUsersRequest as ConcreteGetUsersRequest,GetUsersResponse as ConcreteGetUsersResponse,GetWorldRequest as ConcreteGetWorldRequest,GetWorldResponse as ConcreteGetWorldResponse,GetWorldsRequest as ConcreteGetWorldsRequest,GetWorldsResponse as ConcreteGetWorldsResponse,HexCoord as ConcreteHexCoord,HighlightSpec as ConcreteHighlightSpec,IncomeConfig as ConcreteIncomeConfig,IndexInfo as ConcreteIndexInfo,IndexRecord as ConcreteIndexRecord,IndexRecordsLRO as ConcreteIndexRecordsLRO,IndexState as ConcreteIndexState,IndexStateList as ConcreteIndexStateList,InitializeGameRequest as ConcreteInitializeGameRequest,InitializeGameResponse as ConcreteInitializeGameResponse,InitializeSingletonRequest as ConcreteInitializeSingletonRequest,InitializeSingletonResponse as ConcreteInitializeSingletonResponse,Job as ConcreteJob,ListGamesRequest as ConcreteListGamesRequest,ListGamesResponse as ConcreteListGamesResponse,ListIndexStatesRequest as ConcreteListIndexStatesRequest,ListIndexStatesResponse as ConcreteListIndexStatesResponse,ListMovesRequest as ConcreteListMovesRequest,ListMovesResponse as ConcreteListMovesResponse,ListUsersRequest as ConcreteListUsersRequest,ListUsersResponse as ConcreteListUsersResponse,ListWorldsRequest as ConcreteListWorldsRequest,ListWorldsResponse as ConcreteListWorldsResponse,LogMessageRequest as ConcreteLogMessageRequest,LogMessageResponse as ConcreteLogMessageResponse,MoveUnitAction as ConcreteMoveUnitAction,MoveUnitRequest as ConcreteMoveUnitRequest,MoveUnitResponse as ConcreteMoveUnitResponse,Pagination as ConcretePagination,PaginationResponse as ConcretePaginationResponse,Path as ConcretePath,PathEdge as ConcretePathEdge,PlayerChangedChange as ConcretePlayerChangedChange,PlayerColor as ConcretePlayerColor,ProcessMovesRequest as ConcreteProcessMovesRequest,ProcessMovesResponse as ConcreteProcessMovesResponse,RemoveTileAtRequest as ConcreteRemoveTileAtRequest,RemoveTileAtResponse as ConcreteRemoveTileAtResponse,RemoveUnitAtRequest as ConcreteRemoveUnitAtRequest,RemoveUnitAtResponse as ConcreteRemoveUnitAtResponse,RepeatInfo as ConcreteRepeatInfo,RulesEngine as ConcreteRulesEngine,Run as ConcreteRun,SceneClickedRequest as ConcreteSceneClickedRequest,SceneClickedResponse as ConcreteSceneClickedResponse,SetAllowedPanelsRequest as ConcreteSetAllowedPanelsRequest,SetAllowedPanelsResponse as ConcreteSetAllowedPanelsResponse,SetContentRequest as ConcreteSetContentRequest,SetContentResponse as ConcreteSetContentResponse,SetGameStateRequest as ConcreteSetGameStateRequest,SetGameStateResponse as ConcreteSetGameStateResponse,SetTileAtRequest as ConcreteSetTileAtRequest,SetTileAtResponse as ConcreteSetTileAtResponse,SetUnitAtRequest as ConcreteSetUnitAtRequest,SetUnitAtResponse as ConcreteSetUnitAtResponse,ShowAttackEffectRequest as ConcreteShowAttackEffectRequest,ShowAttackEffectResponse as ConcreteShowAttackEffectResponse,ShowBuildOptionsRequest as ConcreteShowBuildOptionsRequest,ShowBuildOptionsResponse as ConcreteShowBuildOptionsResponse,ShowCaptureEffectRequest as ConcreteShowCaptureEffectRequest,ShowCaptureEffectResponse as ConcreteShowCaptureEffectResponse,ShowHealEffectRequest as ConcreteShowHealEffectRequest,ShowHealEffectResponse as ConcreteShowHealEffectResponse,ShowHighlightsRequest as ConcreteShowHighlightsRequest,ShowHighlightsResponse as ConcreteShowHighlightsResponse,ShowPathRequest as ConcreteShowPathRequest,ShowPathResponse as ConcreteShowPathResponse,SimulateAttackRequest as ConcreteSimulateAttackRequest,SimulateAttackResponse as ConcreteSimulateAttackResponse,SplashTarget as ConcreteSplashTarget,TerrainDefinition as ConcreteTerrainDefinition,TerrainMapping as ConcreteTerrainMapping,TerrainUnitProperties as ConcreteTerrainUnitProperties,ThemeInfo as ConcreteThemeInfo,ThemeManifest as ConcreteThemeManifest,Tile as ConcreteTile,TurnOptionClickedRequest as ConcreteTurnOptionClickedRequest,TurnOptionClickedResponse as ConcreteTurnOptionClickedResponse,Unit as ConcreteUnit,UnitBuiltChange as ConcreteUnitBuiltChange,UnitDamagedChange as ConcreteUnitDamagedChange,UnitDefinition as ConcreteUnitDefinition,UnitKilledChange as ConcreteUnitKilledChange,UnitMapping as ConcreteUnitMapping,UnitMovedChange as ConcreteUnitMovedChange,UnitUnitProperties as ConcreteUnitUnitProperties,UpdateGameRequest as ConcreteUpdateGameRequest,UpdateGameResponse as ConcreteUpdateGameResponse,UpdateGameStatusRequest as ConcreteUpdateGameStatusRequest,UpdateGameStatusResponse as ConcreteUpdateGameStatusResponse,UpdateIndexRecordsLRORequest as ConcreteUpdateIndexRecordsLRORequest,UpdateIndexRecordsLROResponse as ConcreteUpdateIndexRecordsLROResponse,UpdateUserRequest as ConcreteUpdateUserRequest,UpdateUserResponse as ConcreteUpdateUserResponse,UpdateWorldRequest as ConcreteUpdateWorldRequest,UpdateWorldResponse as ConcreteUpdateWorldResponse,User as ConcreteUser,UserInfo as ConcreteUserInfo,World as ConcreteWorld,WorldChange as ConcreteWorldChange,WorldData as ConcreteWorldData,WorldInfo as ConcreteWorldInfo} from "./models/models";




/**
 * Factory result interface for enhanced factory methods
 */
export interface FactoryResult<T> {
  instance: T;
  fullyLoaded: boolean;
}

/**
 * Enhanced factory with context-aware object construction
 */
export class Weewar_v1Factory {


  /**
   * Enhanced factory method for IndexInfo
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newIndexInfo = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<IndexInfoInterface> => {
    const out = new ConcreteIndexInfo();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for User
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUser = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UserInterface> => {
    const out = new ConcreteUser();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for Pagination
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newPagination = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<PaginationInterface> => {
    const out = new ConcretePagination();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for PaginationResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newPaginationResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<PaginationResponseInterface> => {
    const out = new ConcretePaginationResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for World
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newWorld = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<WorldInterface> => {
    const out = new ConcreteWorld();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for WorldData
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newWorldData = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<WorldDataInterface> => {
    const out = new ConcreteWorldData();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for Tile
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newTile = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<TileInterface> => {
    const out = new ConcreteTile();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for Unit
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUnit = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UnitInterface> => {
    const out = new ConcreteUnit();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for AttackRecord
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newAttackRecord = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<AttackRecordInterface> => {
    const out = new ConcreteAttackRecord();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for TerrainDefinition
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newTerrainDefinition = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<TerrainDefinitionInterface> => {
    const out = new ConcreteTerrainDefinition();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for UnitDefinition
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUnitDefinition = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UnitDefinitionInterface> => {
    const out = new ConcreteUnitDefinition();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for TerrainUnitProperties
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newTerrainUnitProperties = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<TerrainUnitPropertiesInterface> => {
    const out = new ConcreteTerrainUnitProperties();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for UnitUnitProperties
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUnitUnitProperties = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UnitUnitPropertiesInterface> => {
    const out = new ConcreteUnitUnitProperties();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for DamageDistribution
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newDamageDistribution = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<DamageDistributionInterface> => {
    const out = new ConcreteDamageDistribution();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for DamageRange
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newDamageRange = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<DamageRangeInterface> => {
    const out = new ConcreteDamageRange();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for RulesEngine
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newRulesEngine = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<RulesEngineInterface> => {
    const out = new ConcreteRulesEngine();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for Game
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGame = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GameInterface> => {
    const out = new ConcreteGame();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GameConfiguration
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGameConfiguration = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GameConfigurationInterface> => {
    const out = new ConcreteGameConfiguration();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for IncomeConfig
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newIncomeConfig = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<IncomeConfigInterface> => {
    const out = new ConcreteIncomeConfig();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GamePlayer
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGamePlayer = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GamePlayerInterface> => {
    const out = new ConcreteGamePlayer();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GameTeam
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGameTeam = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GameTeamInterface> => {
    const out = new ConcreteGameTeam();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GameSettings
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGameSettings = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GameSettingsInterface> => {
    const out = new ConcreteGameSettings();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GameState
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGameState = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GameStateInterface> => {
    const out = new ConcreteGameState();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GameMoveHistory
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGameMoveHistory = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GameMoveHistoryInterface> => {
    const out = new ConcreteGameMoveHistory();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GameMoveGroup
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGameMoveGroup = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GameMoveGroupInterface> => {
    const out = new ConcreteGameMoveGroup();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GameMove
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGameMove = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GameMoveInterface> => {
    const out = new ConcreteGameMove();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for MoveUnitAction
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newMoveUnitAction = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<MoveUnitActionInterface> => {
    const out = new ConcreteMoveUnitAction();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for AttackUnitAction
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newAttackUnitAction = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<AttackUnitActionInterface> => {
    const out = new ConcreteAttackUnitAction();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for BuildUnitAction
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newBuildUnitAction = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<BuildUnitActionInterface> => {
    const out = new ConcreteBuildUnitAction();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for CaptureBuildingAction
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newCaptureBuildingAction = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<CaptureBuildingActionInterface> => {
    const out = new ConcreteCaptureBuildingAction();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for EndTurnAction
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newEndTurnAction = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<EndTurnActionInterface> => {
    const out = new ConcreteEndTurnAction();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for WorldChange
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newWorldChange = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<WorldChangeInterface> => {
    const out = new ConcreteWorldChange();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for UnitMovedChange
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUnitMovedChange = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UnitMovedChangeInterface> => {
    const out = new ConcreteUnitMovedChange();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for UnitDamagedChange
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUnitDamagedChange = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UnitDamagedChangeInterface> => {
    const out = new ConcreteUnitDamagedChange();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for UnitKilledChange
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUnitKilledChange = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UnitKilledChangeInterface> => {
    const out = new ConcreteUnitKilledChange();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for PlayerChangedChange
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newPlayerChangedChange = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<PlayerChangedChangeInterface> => {
    const out = new ConcretePlayerChangedChange();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for UnitBuiltChange
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUnitBuiltChange = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UnitBuiltChangeInterface> => {
    const out = new ConcreteUnitBuiltChange();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for CoinsChangedChange
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newCoinsChangedChange = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<CoinsChangedChangeInterface> => {
    const out = new ConcreteCoinsChangedChange();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for AllPaths
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newAllPaths = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<AllPathsInterface> => {
    const out = new ConcreteAllPaths();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for PathEdge
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newPathEdge = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<PathEdgeInterface> => {
    const out = new ConcretePathEdge();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for Path
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newPath = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<PathInterface> => {
    const out = new ConcretePath();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for IndexState
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newIndexState = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<IndexStateInterface> => {
    const out = new ConcreteIndexState();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for EnsureIndexStateRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newEnsureIndexStateRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<EnsureIndexStateRequestInterface> => {
    const out = new ConcreteEnsureIndexStateRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for EnsureIndexStateResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newEnsureIndexStateResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<EnsureIndexStateResponseInterface> => {
    const out = new ConcreteEnsureIndexStateResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetIndexStatesRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetIndexStatesRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetIndexStatesRequestInterface> => {
    const out = new ConcreteGetIndexStatesRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for IndexStateList
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newIndexStateList = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<IndexStateListInterface> => {
    const out = new ConcreteIndexStateList();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetIndexStatesResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetIndexStatesResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetIndexStatesResponseInterface> => {
    const out = new ConcreteGetIndexStatesResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ListIndexStatesRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newListIndexStatesRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ListIndexStatesRequestInterface> => {
    const out = new ConcreteListIndexStatesRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ListIndexStatesResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newListIndexStatesResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ListIndexStatesResponseInterface> => {
    const out = new ConcreteListIndexStatesResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for DeleteIndexStatesRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newDeleteIndexStatesRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<DeleteIndexStatesRequestInterface> => {
    const out = new ConcreteDeleteIndexStatesRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for DeleteIndexStatesResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newDeleteIndexStatesResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<DeleteIndexStatesResponseInterface> => {
    const out = new ConcreteDeleteIndexStatesResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for IndexRecord
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newIndexRecord = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<IndexRecordInterface> => {
    const out = new ConcreteIndexRecord();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for IndexRecordsLRO
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newIndexRecordsLRO = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<IndexRecordsLROInterface> => {
    const out = new ConcreteIndexRecordsLRO();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for CreateIndexRecordsLRORequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newCreateIndexRecordsLRORequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<CreateIndexRecordsLRORequestInterface> => {
    const out = new ConcreteCreateIndexRecordsLRORequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for CreateIndexRecordsLROResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newCreateIndexRecordsLROResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<CreateIndexRecordsLROResponseInterface> => {
    const out = new ConcreteCreateIndexRecordsLROResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for UpdateIndexRecordsLRORequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUpdateIndexRecordsLRORequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UpdateIndexRecordsLRORequestInterface> => {
    const out = new ConcreteUpdateIndexRecordsLRORequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for UpdateIndexRecordsLROResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUpdateIndexRecordsLROResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UpdateIndexRecordsLROResponseInterface> => {
    const out = new ConcreteUpdateIndexRecordsLROResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetIndexRecordsLRORequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetIndexRecordsLRORequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetIndexRecordsLRORequestInterface> => {
    const out = new ConcreteGetIndexRecordsLRORequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetIndexRecordsLROResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetIndexRecordsLROResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetIndexRecordsLROResponseInterface> => {
    const out = new ConcreteGetIndexRecordsLROResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for WorldInfo
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newWorldInfo = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<WorldInfoInterface> => {
    const out = new ConcreteWorldInfo();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ListWorldsRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newListWorldsRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ListWorldsRequestInterface> => {
    const out = new ConcreteListWorldsRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ListWorldsResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newListWorldsResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ListWorldsResponseInterface> => {
    const out = new ConcreteListWorldsResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetWorldRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetWorldRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetWorldRequestInterface> => {
    const out = new ConcreteGetWorldRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetWorldResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetWorldResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetWorldResponseInterface> => {
    const out = new ConcreteGetWorldResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for UpdateWorldRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUpdateWorldRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UpdateWorldRequestInterface> => {
    const out = new ConcreteUpdateWorldRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for UpdateWorldResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUpdateWorldResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UpdateWorldResponseInterface> => {
    const out = new ConcreteUpdateWorldResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for DeleteWorldRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newDeleteWorldRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<DeleteWorldRequestInterface> => {
    const out = new ConcreteDeleteWorldRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for DeleteWorldResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newDeleteWorldResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<DeleteWorldResponseInterface> => {
    const out = new ConcreteDeleteWorldResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetWorldsRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetWorldsRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetWorldsRequestInterface> => {
    const out = new ConcreteGetWorldsRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetWorldsResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetWorldsResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetWorldsResponseInterface> => {
    const out = new ConcreteGetWorldsResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for CreateWorldRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newCreateWorldRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<CreateWorldRequestInterface> => {
    const out = new ConcreteCreateWorldRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for CreateWorldResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newCreateWorldResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<CreateWorldResponseInterface> => {
    const out = new ConcreteCreateWorldResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ListGamesRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newListGamesRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ListGamesRequestInterface> => {
    const out = new ConcreteListGamesRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ListGamesResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newListGamesResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ListGamesResponseInterface> => {
    const out = new ConcreteListGamesResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetGameRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetGameRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetGameRequestInterface> => {
    const out = new ConcreteGetGameRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetGameResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetGameResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetGameResponseInterface> => {
    const out = new ConcreteGetGameResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetGameContentRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetGameContentRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetGameContentRequestInterface> => {
    const out = new ConcreteGetGameContentRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetGameContentResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetGameContentResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetGameContentResponseInterface> => {
    const out = new ConcreteGetGameContentResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for UpdateGameRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUpdateGameRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UpdateGameRequestInterface> => {
    const out = new ConcreteUpdateGameRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for UpdateGameResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUpdateGameResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UpdateGameResponseInterface> => {
    const out = new ConcreteUpdateGameResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for DeleteGameRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newDeleteGameRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<DeleteGameRequestInterface> => {
    const out = new ConcreteDeleteGameRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for DeleteGameResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newDeleteGameResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<DeleteGameResponseInterface> => {
    const out = new ConcreteDeleteGameResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetGamesRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetGamesRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetGamesRequestInterface> => {
    const out = new ConcreteGetGamesRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetGamesResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetGamesResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetGamesResponseInterface> => {
    const out = new ConcreteGetGamesResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for CreateGameRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newCreateGameRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<CreateGameRequestInterface> => {
    const out = new ConcreteCreateGameRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for CreateGameResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newCreateGameResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<CreateGameResponseInterface> => {
    const out = new ConcreteCreateGameResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ProcessMovesRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newProcessMovesRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ProcessMovesRequestInterface> => {
    const out = new ConcreteProcessMovesRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ProcessMovesResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newProcessMovesResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ProcessMovesResponseInterface> => {
    const out = new ConcreteProcessMovesResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetGameStateRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetGameStateRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetGameStateRequestInterface> => {
    const out = new ConcreteGetGameStateRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetGameStateResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetGameStateResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetGameStateResponseInterface> => {
    const out = new ConcreteGetGameStateResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ListMovesRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newListMovesRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ListMovesRequestInterface> => {
    const out = new ConcreteListMovesRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ListMovesResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newListMovesResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ListMovesResponseInterface> => {
    const out = new ConcreteListMovesResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetOptionsAtRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetOptionsAtRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetOptionsAtRequestInterface> => {
    const out = new ConcreteGetOptionsAtRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetOptionsAtResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetOptionsAtResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetOptionsAtResponseInterface> => {
    const out = new ConcreteGetOptionsAtResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GameOption
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGameOption = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GameOptionInterface> => {
    const out = new ConcreteGameOption();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for SimulateAttackRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newSimulateAttackRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<SimulateAttackRequestInterface> => {
    const out = new ConcreteSimulateAttackRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for SimulateAttackResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newSimulateAttackResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<SimulateAttackResponseInterface> => {
    const out = new ConcreteSimulateAttackResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for InitializeSingletonRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newInitializeSingletonRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<InitializeSingletonRequestInterface> => {
    const out = new ConcreteInitializeSingletonRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for InitializeSingletonResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newInitializeSingletonResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<InitializeSingletonResponseInterface> => {
    const out = new ConcreteInitializeSingletonResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for TurnOptionClickedRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newTurnOptionClickedRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<TurnOptionClickedRequestInterface> => {
    const out = new ConcreteTurnOptionClickedRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for TurnOptionClickedResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newTurnOptionClickedResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<TurnOptionClickedResponseInterface> => {
    const out = new ConcreteTurnOptionClickedResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for SceneClickedRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newSceneClickedRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<SceneClickedRequestInterface> => {
    const out = new ConcreteSceneClickedRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for SceneClickedResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newSceneClickedResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<SceneClickedResponseInterface> => {
    const out = new ConcreteSceneClickedResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for EndTurnButtonClickedRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newEndTurnButtonClickedRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<EndTurnButtonClickedRequestInterface> => {
    const out = new ConcreteEndTurnButtonClickedRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for EndTurnButtonClickedResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newEndTurnButtonClickedResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<EndTurnButtonClickedResponseInterface> => {
    const out = new ConcreteEndTurnButtonClickedResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for BuildOptionClickedRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newBuildOptionClickedRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<BuildOptionClickedRequestInterface> => {
    const out = new ConcreteBuildOptionClickedRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for BuildOptionClickedResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newBuildOptionClickedResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<BuildOptionClickedResponseInterface> => {
    const out = new ConcreteBuildOptionClickedResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for InitializeGameRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newInitializeGameRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<InitializeGameRequestInterface> => {
    const out = new ConcreteInitializeGameRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for InitializeGameResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newInitializeGameResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<InitializeGameResponseInterface> => {
    const out = new ConcreteInitializeGameResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for EmptyRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newEmptyRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<EmptyRequestInterface> => {
    const out = new ConcreteEmptyRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for EmptyResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newEmptyResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<EmptyResponseInterface> => {
    const out = new ConcreteEmptyResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for SetContentRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newSetContentRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<SetContentRequestInterface> => {
    const out = new ConcreteSetContentRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for SetContentResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newSetContentResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<SetContentResponseInterface> => {
    const out = new ConcreteSetContentResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ShowBuildOptionsRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newShowBuildOptionsRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ShowBuildOptionsRequestInterface> => {
    const out = new ConcreteShowBuildOptionsRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ShowBuildOptionsResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newShowBuildOptionsResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ShowBuildOptionsResponseInterface> => {
    const out = new ConcreteShowBuildOptionsResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for LogMessageRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newLogMessageRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<LogMessageRequestInterface> => {
    const out = new ConcreteLogMessageRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for LogMessageResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newLogMessageResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<LogMessageResponseInterface> => {
    const out = new ConcreteLogMessageResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for SetGameStateRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newSetGameStateRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<SetGameStateRequestInterface> => {
    const out = new ConcreteSetGameStateRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for SetGameStateResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newSetGameStateResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<SetGameStateResponseInterface> => {
    const out = new ConcreteSetGameStateResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for UpdateGameStatusRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUpdateGameStatusRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UpdateGameStatusRequestInterface> => {
    const out = new ConcreteUpdateGameStatusRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for UpdateGameStatusResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUpdateGameStatusResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UpdateGameStatusResponseInterface> => {
    const out = new ConcreteUpdateGameStatusResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for SetTileAtRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newSetTileAtRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<SetTileAtRequestInterface> => {
    const out = new ConcreteSetTileAtRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for SetTileAtResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newSetTileAtResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<SetTileAtResponseInterface> => {
    const out = new ConcreteSetTileAtResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for SetUnitAtRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newSetUnitAtRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<SetUnitAtRequestInterface> => {
    const out = new ConcreteSetUnitAtRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for SetUnitAtResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newSetUnitAtResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<SetUnitAtResponseInterface> => {
    const out = new ConcreteSetUnitAtResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for RemoveTileAtRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newRemoveTileAtRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<RemoveTileAtRequestInterface> => {
    const out = new ConcreteRemoveTileAtRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for RemoveTileAtResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newRemoveTileAtResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<RemoveTileAtResponseInterface> => {
    const out = new ConcreteRemoveTileAtResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for RemoveUnitAtRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newRemoveUnitAtRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<RemoveUnitAtRequestInterface> => {
    const out = new ConcreteRemoveUnitAtRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for RemoveUnitAtResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newRemoveUnitAtResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<RemoveUnitAtResponseInterface> => {
    const out = new ConcreteRemoveUnitAtResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ShowHighlightsRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newShowHighlightsRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ShowHighlightsRequestInterface> => {
    const out = new ConcreteShowHighlightsRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ShowHighlightsResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newShowHighlightsResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ShowHighlightsResponseInterface> => {
    const out = new ConcreteShowHighlightsResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for HighlightSpec
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newHighlightSpec = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<HighlightSpecInterface> => {
    const out = new ConcreteHighlightSpec();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ClearHighlightsRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newClearHighlightsRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ClearHighlightsRequestInterface> => {
    const out = new ConcreteClearHighlightsRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ClearHighlightsResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newClearHighlightsResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ClearHighlightsResponseInterface> => {
    const out = new ConcreteClearHighlightsResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ShowPathRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newShowPathRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ShowPathRequestInterface> => {
    const out = new ConcreteShowPathRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ShowPathResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newShowPathResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ShowPathResponseInterface> => {
    const out = new ConcreteShowPathResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ClearPathsRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newClearPathsRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ClearPathsRequestInterface> => {
    const out = new ConcreteClearPathsRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ClearPathsResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newClearPathsResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ClearPathsResponseInterface> => {
    const out = new ConcreteClearPathsResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for MoveUnitRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newMoveUnitRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<MoveUnitRequestInterface> => {
    const out = new ConcreteMoveUnitRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for MoveUnitResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newMoveUnitResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<MoveUnitResponseInterface> => {
    const out = new ConcreteMoveUnitResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for HexCoord
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newHexCoord = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<HexCoordInterface> => {
    const out = new ConcreteHexCoord();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ShowAttackEffectRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newShowAttackEffectRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ShowAttackEffectRequestInterface> => {
    const out = new ConcreteShowAttackEffectRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for SplashTarget
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newSplashTarget = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<SplashTargetInterface> => {
    const out = new ConcreteSplashTarget();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ShowAttackEffectResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newShowAttackEffectResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ShowAttackEffectResponseInterface> => {
    const out = new ConcreteShowAttackEffectResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ShowHealEffectRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newShowHealEffectRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ShowHealEffectRequestInterface> => {
    const out = new ConcreteShowHealEffectRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ShowHealEffectResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newShowHealEffectResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ShowHealEffectResponseInterface> => {
    const out = new ConcreteShowHealEffectResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ShowCaptureEffectRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newShowCaptureEffectRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ShowCaptureEffectRequestInterface> => {
    const out = new ConcreteShowCaptureEffectRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ShowCaptureEffectResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newShowCaptureEffectResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ShowCaptureEffectResponseInterface> => {
    const out = new ConcreteShowCaptureEffectResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for SetAllowedPanelsRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newSetAllowedPanelsRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<SetAllowedPanelsRequestInterface> => {
    const out = new ConcreteSetAllowedPanelsRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for SetAllowedPanelsResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newSetAllowedPanelsResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<SetAllowedPanelsResponseInterface> => {
    const out = new ConcreteSetAllowedPanelsResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for UserInfo
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUserInfo = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UserInfoInterface> => {
    const out = new ConcreteUserInfo();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ListUsersRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newListUsersRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ListUsersRequestInterface> => {
    const out = new ConcreteListUsersRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ListUsersResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newListUsersResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ListUsersResponseInterface> => {
    const out = new ConcreteListUsersResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetUserRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetUserRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetUserRequestInterface> => {
    const out = new ConcreteGetUserRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetUserResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetUserResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetUserResponseInterface> => {
    const out = new ConcreteGetUserResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetUserContentRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetUserContentRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetUserContentRequestInterface> => {
    const out = new ConcreteGetUserContentRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetUserContentResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetUserContentResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetUserContentResponseInterface> => {
    const out = new ConcreteGetUserContentResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for UpdateUserRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUpdateUserRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UpdateUserRequestInterface> => {
    const out = new ConcreteUpdateUserRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for UpdateUserResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUpdateUserResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UpdateUserResponseInterface> => {
    const out = new ConcreteUpdateUserResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for DeleteUserRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newDeleteUserRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<DeleteUserRequestInterface> => {
    const out = new ConcreteDeleteUserRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for DeleteUserResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newDeleteUserResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<DeleteUserResponseInterface> => {
    const out = new ConcreteDeleteUserResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetUsersRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetUsersRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetUsersRequestInterface> => {
    const out = new ConcreteGetUsersRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for GetUsersResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newGetUsersResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<GetUsersResponseInterface> => {
    const out = new ConcreteGetUsersResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for CreateUserRequest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newCreateUserRequest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<CreateUserRequestInterface> => {
    const out = new ConcreteCreateUserRequest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for CreateUserResponse
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newCreateUserResponse = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<CreateUserResponseInterface> => {
    const out = new ConcreteCreateUserResponse();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for Job
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newJob = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<JobInterface> => {
    const out = new ConcreteJob();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for RepeatInfo
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newRepeatInfo = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<RepeatInfoInterface> => {
    const out = new ConcreteRepeatInfo();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for Run
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newRun = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<RunInterface> => {
    const out = new ConcreteRun();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ThemeInfo
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newThemeInfo = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ThemeInfoInterface> => {
    const out = new ConcreteThemeInfo();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for UnitMapping
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newUnitMapping = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<UnitMappingInterface> => {
    const out = new ConcreteUnitMapping();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for TerrainMapping
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newTerrainMapping = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<TerrainMappingInterface> => {
    const out = new ConcreteTerrainMapping();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for ThemeManifest
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newThemeManifest = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<ThemeManifestInterface> => {
    const out = new ConcreteThemeManifest();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for PlayerColor
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newPlayerColor = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<PlayerColorInterface> => {
    const out = new ConcretePlayerColor();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }

  /**
   * Enhanced factory method for AssetResult
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw data to potentially populate from
   * @returns Factory result with instance and population status
   */
  newAssetResult = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<AssetResultInterface> => {
    const out = new ConcreteAssetResult();
    
    // Factory does not populate by default - let deserializer handle it
    return { instance: out, fullyLoaded: false };
  }



  /**
   * Get factory method for a fully qualified message type
   * Enables cross-package factory delegation
   */
  getFactoryMethod(messageType: string): ((parent?: any, attributeName?: string, attributeKey?: string | number, data?: any) => FactoryResult<any>) | undefined {
    // Extract package from message type (e.g., "library.common.BaseMessage" -> "library.common")
    const parts = messageType.split('.');
    if (parts.length < 2) {
      return undefined;
    }
    
    const packageName = parts.slice(0, -1).join('.');
    const typeName = parts[parts.length - 1];
    const methodName = 'new' + typeName;
    
    // Check if this is our own package first
    const currentPackage = "weewar.v1";
    if (packageName === currentPackage) {
      return (this as any)[methodName];
    }
    
    // Check external type factory mappings
    const externalFactory = this.externalTypeFactories()[messageType];
    if (externalFactory) {
      return externalFactory;
    }
    
    // Delegate to appropriate dependency factory

    
    return undefined;
  }

  /**
   * Generic object deserializer that respects factory decisions
   */
  protected deserializeObject(instance: any, data: any): any {
    if (!data || typeof data !== 'object') return instance;
    
    for (const [key, value] of Object.entries(data)) {
      if (value !== null && value !== undefined) {
        instance[key] = value;
      }
    }
    return instance;
  }

  // External type conversion methods

  /**
   * Mapping of external types to their factory methods
   */
  private externalTypeFactories(): Record<string, (parent?: any, attributeName?: string, attributeKey?: string | number, data?: any) => FactoryResult<any>> { 
      return {
          "google.protobuf.Timestamp": this.newTimestamp,
          "google.protobuf.FieldMask": this.newFieldMask,
      }
  };

  /**
   * Convert native Date to protobuf Timestamp format for serialization
   */
  serializeTimestamp(date: Date): any {
    if (!date) return null;
    return {
      seconds: Math.floor(date.getTime() / 1000).toString(),
      nanos: (date.getTime() % 1000) * 1000000
    };
  }

  /**
   * Factory method for converting protobuf Timestamp data to native Date
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object  
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw protobuf timestamp data
   * @returns Factory result with Date instance
   */
  newTimestamp = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<Date> => {
    if (!data) {
      return { instance: new Date(), fullyLoaded: true };
    }
    
    let date: Date;
    if (typeof data === 'string') {
      // Handle ISO string format
      date = new Date(data);
    } else if (data.seconds !== undefined) {
      // Handle protobuf format with seconds/nanos
      const seconds = typeof data.seconds === 'string' 
        ? parseInt(data.seconds, 10) 
        : data.seconds;
      const nanos = data.nanos || 0;
      date = new Date(seconds * 1000 + Math.floor(nanos / 1000000));
    } else {
      date = new Date();
    }
    
    return { instance: date, fullyLoaded: true };
  }

  /**
   * Convert native string array to protobuf FieldMask format for serialization
   */
  serializeFieldMask(paths: string[]): any {
    if (!paths || !Array.isArray(paths)) return null;
    return { paths };
  }

  /**
   * Factory method for converting protobuf FieldMask data to native string array
   * @param parent Parent object containing this field
   * @param attributeName Field name in parent object
   * @param attributeKey Array index, map key, or union tag (for containers)
   * @param data Raw protobuf field mask data
   * @returns Factory result with string array instance
   */
  newFieldMask = (
    parent?: any,
    attributeName?: string,
    attributeKey?: string | number,
    data?: any
  ): FactoryResult<string[]> => {
    if (!data) {
      return { instance: [], fullyLoaded: true };
    }

    let paths: string[];
    if (Array.isArray(data)) {
      paths = data;
    } else if (data.paths && Array.isArray(data.paths)) {
      paths = data.paths;
    } else {
      paths = [];
    }

    return { instance: paths, fullyLoaded: true };
  }
}

// Shared factory instance to avoid creating new instances on every deserializer construction
const DEFAULT_FACTORY = new Weewar_v1Factory();

/**
 * Schema-aware deserializer for weewar.v1 package
 * Extends BaseDeserializer with package-specific configuration
 */
export class Weewar_v1Deserializer extends BaseDeserializer {
  constructor(
    schemaRegistry = weewar_v1SchemaRegistry,
    factory: FactoryInterface = DEFAULT_FACTORY
  ) {
    super(schemaRegistry, factory);
  }

  /**
   * Static utility method to create and deserialize a message without needing a deserializer instance
   * @param messageType Fully qualified message type (use Class.MESSAGE_TYPE)
   * @param data Raw data to deserialize
   * @returns Deserialized instance or null if creation failed
   */
  static fromMsgType<T>(messageType: string, data: any): T {
    const deserializer = new Weewar_v1Deserializer(); // Uses default factory and schema registry
    return deserializer.createAndDeserialize<T>(messageType, data);
  }


  /**
   * Static utility method - infers messageType from type parameter
   * Type-safe convenience method
   */
  static from<T>(typeConstructor: MessageTypeConstructor<T>, data: any): T {
    const deserializer = new Weewar_v1Deserializer();
    return deserializer.createAndDeserialize<T>(typeConstructor.MESSAGE_TYPE, data);
  }

  // Deserialize if data is already a partial instance
  static fromPartial<T extends { __MESSAGE_TYPE: string }>(data: T): T {
    const messageType = data.__MESSAGE_TYPE;
    return this.fromMsgType<T>(messageType, data);
  }
}
