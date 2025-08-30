// Generated TypeScript interfaces from proto file
// DO NOT EDIT - This file is auto-generated




export enum ProposalPhase {
  PROPOSAL_PHASE_UNSPECIFIED = 0,
  PROPOSAL_PHASE_OPEN = 1,
  PROPOSAL_PHASE_COLLECTING = 2,
  PROPOSAL_PHASE_FINALIZING = 3,
  PROPOSAL_PHASE_COMMITTED = 4,
  PROPOSAL_PHASE_REJECTED = 5,
  PROPOSAL_PHASE_TIMEOUT = 6,
}


export enum Status {
  STATUS_UNSPECIFIED = 0,
  STATUS_ACCEPTED = 1,
  STATUS_REJECTED = 2,
}



export interface ProposalInfo {
  proposalId: string;
  sessionId: string;
  proposerId: string;
  /** State transition (opaque to coordinator) */
  fromStateHash: string;
  toStateHash: string;
  movesBlob: Uint8Array;
  changesBlob: Uint8Array;
  newStateBlob: Uint8Array;
  /** Validation tracking */
  assignedValidators: string[];
  votes?: Map<string, ValidationVote>;
  phase: ProposalPhase;
  /** Timing */
  createdAt?: Date;
  deadline?: Date;
  /** Anti-replay */
  nonce: string;
}



export interface ValidationVote {
  validatorId: string;
  approved: boolean;
  computedHash: string;
  errorReason: string;
  submittedAt?: Date;
  signature: string;
}


/**
 * Lightweight proposal tracking for game state (game-agnostic)
 */
export interface ProposalTrackingInfo {
  /** ID of the active proposal */
  proposalId: string;
  /** Player who made the proposal */
  proposerId: string;
  /** Current phase of the proposal */
  phase: ProposalPhase;
  /** Creation time */
  createdAt?: Date;
  /** Number of validators assigned */
  validatorCount: number;
  /** Number of votes received */
  votesReceived: number;
}



export interface SubmitProposalRequest {
  sessionId: string;
  proposerId: string;
  fromStateHash: string;
  toStateHash: string;
  movesBlob: Uint8Array;
  changesBlob: Uint8Array;
  newStateBlob: Uint8Array;
  nonce: string;
}



export interface SubmitProposalResponse {
  status: Status;
  proposalId: string;
  reason: string;
  assignedValidators: string[];
}



export interface GetPendingValidationsRequest {
  validatorId: string;
}



export interface GetPendingValidationsResponse {
  validations?: PendingValidation[];
}



export interface PendingValidation {
  sessionId: string;
  proposalId: string;
  proposerId: string;
  fromStateHash: string;
  movesBlob: Uint8Array;
  changesBlob: Uint8Array;
  deadline?: Date;
  nonce: string;
}



export interface SubmitValidationRequest {
  sessionId: string;
  proposalId: string;
  validatorId: string;
  approved: boolean;
  computedHash: string;
  errorReason: string;
  signature: string;
}



export interface SubmitValidationResponse {
  recorded: boolean;
  consensusReached: boolean;
  consensusApproved: boolean;
}



export interface GetProposalStatusRequest {
  proposalId: string;
}



export interface GetProposalStatusResponse {
  proposal?: ProposalInfo;
  phase: ProposalPhase;
  votesReceived: number;
  votesRequired: number;
}



export interface GameSession {
  sessionId: string;
  gameType: string;
  playerIds: string[];
  currentPlayerId: string;
  requiredValidators: number;
  /** Current state (opaque to coordinator) */
  currentTick: number;
  currentStateHash: string;
  currentStateBlob: Uint8Array;
  /** Active proposal if any */
  activeProposal?: ProposalInfo;
  createdAt?: Date;
  updatedAt?: Date;
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

