

import { ProposalInfo as ProposalInfoInterface, ValidationVote as ValidationVoteInterface, ProposalTrackingInfo as ProposalTrackingInfoInterface, SubmitProposalRequest as SubmitProposalRequestInterface, SubmitProposalResponse as SubmitProposalResponseInterface, GetPendingValidationsRequest as GetPendingValidationsRequestInterface, GetPendingValidationsResponse as GetPendingValidationsResponseInterface, PendingValidation as PendingValidationInterface, SubmitValidationRequest as SubmitValidationRequestInterface, SubmitValidationResponse as SubmitValidationResponseInterface, GetProposalStatusRequest as GetProposalStatusRequestInterface, GetProposalStatusResponse as GetProposalStatusResponseInterface, GameSession as GameSessionInterface, Pagination as PaginationInterface, PaginationResponse as PaginationResponseInterface, ProposalPhase, Status } from "./interfaces";
import { TurnengineV1Deserializer } from "./deserializer";



export class ProposalInfo implements ProposalInfoInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "turnengine.v1.ProposalInfo";

  proposalId: string = "";
  sessionId: string = "";
  proposerId: string = "";
  /** State transition (opaque to coordinator) */
  fromStateHash: string = "";
  toStateHash: string = "";
  movesBlob: Uint8Array = new Uint8Array();
  changesBlob: Uint8Array = new Uint8Array();
  newStateBlob: Uint8Array = new Uint8Array();
  /** Validation tracking */
  assignedValidators: string[] = [];
  votes?: Map<string, ValidationVote>;
  phase: ProposalPhase = 0;
  /** Timing */
  createdAt?: Date;
  deadline?: Date;
  /** Anti-replay */
  nonce: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ProposalInfo instance or null if creation failed
   */
  static from(data: any) {
    return TurnengineV1Deserializer.from<ProposalInfo>(ProposalInfo.MESSAGE_TYPE, data);
  }
}



export class ValidationVote implements ValidationVoteInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "turnengine.v1.ValidationVote";

  validatorId: string = "";
  approved: boolean = false;
  computedHash: string = "";
  errorReason: string = "";
  submittedAt?: Date;
  signature: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ValidationVote instance or null if creation failed
   */
  static from(data: any) {
    return TurnengineV1Deserializer.from<ValidationVote>(ValidationVote.MESSAGE_TYPE, data);
  }
}


/**
 * Lightweight proposal tracking for game state (game-agnostic)
 */
export class ProposalTrackingInfo implements ProposalTrackingInfoInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "turnengine.v1.ProposalTrackingInfo";

  /** ID of the active proposal */
  proposalId: string = "";
  /** Player who made the proposal */
  proposerId: string = "";
  /** Current phase of the proposal */
  phase: ProposalPhase = 0;
  /** Creation time */
  createdAt?: Date;
  /** Number of validators assigned */
  validatorCount: number = 0;
  /** Number of votes received */
  votesReceived: number = 0;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized ProposalTrackingInfo instance or null if creation failed
   */
  static from(data: any) {
    return TurnengineV1Deserializer.from<ProposalTrackingInfo>(ProposalTrackingInfo.MESSAGE_TYPE, data);
  }
}



export class SubmitProposalRequest implements SubmitProposalRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "turnengine.v1.SubmitProposalRequest";

  sessionId: string = "";
  proposerId: string = "";
  fromStateHash: string = "";
  toStateHash: string = "";
  movesBlob: Uint8Array = new Uint8Array();
  changesBlob: Uint8Array = new Uint8Array();
  newStateBlob: Uint8Array = new Uint8Array();
  nonce: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized SubmitProposalRequest instance or null if creation failed
   */
  static from(data: any) {
    return TurnengineV1Deserializer.from<SubmitProposalRequest>(SubmitProposalRequest.MESSAGE_TYPE, data);
  }
}



export class SubmitProposalResponse implements SubmitProposalResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "turnengine.v1.SubmitProposalResponse";

  status: Status = 0;
  proposalId: string = "";
  reason: string = "";
  assignedValidators: string[] = [];

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized SubmitProposalResponse instance or null if creation failed
   */
  static from(data: any) {
    return TurnengineV1Deserializer.from<SubmitProposalResponse>(SubmitProposalResponse.MESSAGE_TYPE, data);
  }
}



export class GetPendingValidationsRequest implements GetPendingValidationsRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "turnengine.v1.GetPendingValidationsRequest";

  validatorId: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetPendingValidationsRequest instance or null if creation failed
   */
  static from(data: any) {
    return TurnengineV1Deserializer.from<GetPendingValidationsRequest>(GetPendingValidationsRequest.MESSAGE_TYPE, data);
  }
}



export class GetPendingValidationsResponse implements GetPendingValidationsResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "turnengine.v1.GetPendingValidationsResponse";

  validations: PendingValidation[] = [];

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetPendingValidationsResponse instance or null if creation failed
   */
  static from(data: any) {
    return TurnengineV1Deserializer.from<GetPendingValidationsResponse>(GetPendingValidationsResponse.MESSAGE_TYPE, data);
  }
}



export class PendingValidation implements PendingValidationInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "turnengine.v1.PendingValidation";

  sessionId: string = "";
  proposalId: string = "";
  proposerId: string = "";
  fromStateHash: string = "";
  movesBlob: Uint8Array = new Uint8Array();
  changesBlob: Uint8Array = new Uint8Array();
  deadline?: Date;
  nonce: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized PendingValidation instance or null if creation failed
   */
  static from(data: any) {
    return TurnengineV1Deserializer.from<PendingValidation>(PendingValidation.MESSAGE_TYPE, data);
  }
}



export class SubmitValidationRequest implements SubmitValidationRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "turnengine.v1.SubmitValidationRequest";

  sessionId: string = "";
  proposalId: string = "";
  validatorId: string = "";
  approved: boolean = false;
  computedHash: string = "";
  errorReason: string = "";
  signature: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized SubmitValidationRequest instance or null if creation failed
   */
  static from(data: any) {
    return TurnengineV1Deserializer.from<SubmitValidationRequest>(SubmitValidationRequest.MESSAGE_TYPE, data);
  }
}



export class SubmitValidationResponse implements SubmitValidationResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "turnengine.v1.SubmitValidationResponse";

  recorded: boolean = false;
  consensusReached: boolean = false;
  consensusApproved: boolean = false;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized SubmitValidationResponse instance or null if creation failed
   */
  static from(data: any) {
    return TurnengineV1Deserializer.from<SubmitValidationResponse>(SubmitValidationResponse.MESSAGE_TYPE, data);
  }
}



export class GetProposalStatusRequest implements GetProposalStatusRequestInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "turnengine.v1.GetProposalStatusRequest";

  proposalId: string = "";

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetProposalStatusRequest instance or null if creation failed
   */
  static from(data: any) {
    return TurnengineV1Deserializer.from<GetProposalStatusRequest>(GetProposalStatusRequest.MESSAGE_TYPE, data);
  }
}



export class GetProposalStatusResponse implements GetProposalStatusResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "turnengine.v1.GetProposalStatusResponse";

  proposal?: ProposalInfo;
  phase: ProposalPhase = 0;
  votesReceived: number = 0;
  votesRequired: number = 0;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GetProposalStatusResponse instance or null if creation failed
   */
  static from(data: any) {
    return TurnengineV1Deserializer.from<GetProposalStatusResponse>(GetProposalStatusResponse.MESSAGE_TYPE, data);
  }
}



export class GameSession implements GameSessionInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "turnengine.v1.GameSession";

  sessionId: string = "";
  gameType: string = "";
  playerIds: string[] = [];
  currentPlayerId: string = "";
  requiredValidators: number = 0;
  /** Current state (opaque to coordinator) */
  currentTick: number = 0;
  currentStateHash: string = "";
  currentStateBlob: Uint8Array = new Uint8Array();
  /** Active proposal if any */
  activeProposal?: ProposalInfo;
  createdAt?: Date;
  updatedAt?: Date;

  /**
   * Create and deserialize an instance from raw data
   * @param data Raw data to deserialize
   * @returns Deserialized GameSession instance or null if creation failed
   */
  static from(data: any) {
    return TurnengineV1Deserializer.from<GameSession>(GameSession.MESSAGE_TYPE, data);
  }
}



export class Pagination implements PaginationInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "turnengine.v1.Pagination";

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
    return TurnengineV1Deserializer.from<Pagination>(Pagination.MESSAGE_TYPE, data);
  }
}



export class PaginationResponse implements PaginationResponseInterface {
  /**
   * Fully qualified message type for schema resolution
   */
  static readonly MESSAGE_TYPE = "turnengine.v1.PaginationResponse";

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
    return TurnengineV1Deserializer.from<PaginationResponse>(PaginationResponse.MESSAGE_TYPE, data);
  }
}


