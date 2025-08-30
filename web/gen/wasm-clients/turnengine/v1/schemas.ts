
// Generated TypeScript schemas from proto file
// DO NOT EDIT - This file is auto-generated

import { FieldType, FieldSchema, MessageSchema } from "./deserializer_schemas";


/**
 * Schema for ProposalInfo message
 */
export const ProposalInfoSchema: MessageSchema = {
  name: "ProposalInfo",
  fields: [
    {
      name: "proposalId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "sessionId",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "proposerId",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "fromStateHash",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "toStateHash",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "movesBlob",
      type: FieldType.STRING,
      id: 6,
    },
    {
      name: "changesBlob",
      type: FieldType.STRING,
      id: 7,
    },
    {
      name: "newStateBlob",
      type: FieldType.STRING,
      id: 8,
    },
    {
      name: "assignedValidators",
      type: FieldType.REPEATED,
      id: 9,
      repeated: true,
    },
    {
      name: "votes",
      type: FieldType.MESSAGE,
      id: 10,
      messageType: "turnengine.v1.VotesEntry",
    },
    {
      name: "phase",
      type: FieldType.STRING,
      id: 11,
    },
    {
      name: "createdAt",
      type: FieldType.MESSAGE,
      id: 12,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "deadline",
      type: FieldType.MESSAGE,
      id: 13,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "nonce",
      type: FieldType.STRING,
      id: 14,
    },
  ],
};


/**
 * Schema for ValidationVote message
 */
export const ValidationVoteSchema: MessageSchema = {
  name: "ValidationVote",
  fields: [
    {
      name: "validatorId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "approved",
      type: FieldType.BOOLEAN,
      id: 2,
    },
    {
      name: "computedHash",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "errorReason",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "submittedAt",
      type: FieldType.MESSAGE,
      id: 5,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "signature",
      type: FieldType.STRING,
      id: 6,
    },
  ],
};


/**
 * Schema for ProposalTrackingInfo message
 */
export const ProposalTrackingInfoSchema: MessageSchema = {
  name: "ProposalTrackingInfo",
  fields: [
    {
      name: "proposalId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "proposerId",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "phase",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "createdAt",
      type: FieldType.MESSAGE,
      id: 4,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "validatorCount",
      type: FieldType.NUMBER,
      id: 5,
    },
    {
      name: "votesReceived",
      type: FieldType.NUMBER,
      id: 6,
    },
  ],
};


/**
 * Schema for SubmitProposalRequest message
 */
export const SubmitProposalRequestSchema: MessageSchema = {
  name: "SubmitProposalRequest",
  fields: [
    {
      name: "sessionId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "proposerId",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "fromStateHash",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "toStateHash",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "movesBlob",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "changesBlob",
      type: FieldType.STRING,
      id: 6,
    },
    {
      name: "newStateBlob",
      type: FieldType.STRING,
      id: 7,
    },
    {
      name: "nonce",
      type: FieldType.STRING,
      id: 8,
    },
  ],
};


/**
 * Schema for SubmitProposalResponse message
 */
export const SubmitProposalResponseSchema: MessageSchema = {
  name: "SubmitProposalResponse",
  fields: [
    {
      name: "status",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "proposalId",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "reason",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "assignedValidators",
      type: FieldType.REPEATED,
      id: 4,
      repeated: true,
    },
  ],
};


/**
 * Schema for GetPendingValidationsRequest message
 */
export const GetPendingValidationsRequestSchema: MessageSchema = {
  name: "GetPendingValidationsRequest",
  fields: [
    {
      name: "validatorId",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for GetPendingValidationsResponse message
 */
export const GetPendingValidationsResponseSchema: MessageSchema = {
  name: "GetPendingValidationsResponse",
  fields: [
    {
      name: "validations",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "turnengine.v1.PendingValidation",
      repeated: true,
    },
  ],
};


/**
 * Schema for PendingValidation message
 */
export const PendingValidationSchema: MessageSchema = {
  name: "PendingValidation",
  fields: [
    {
      name: "sessionId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "proposalId",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "proposerId",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "fromStateHash",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "movesBlob",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "changesBlob",
      type: FieldType.STRING,
      id: 6,
    },
    {
      name: "deadline",
      type: FieldType.MESSAGE,
      id: 7,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "nonce",
      type: FieldType.STRING,
      id: 8,
    },
  ],
};


/**
 * Schema for SubmitValidationRequest message
 */
export const SubmitValidationRequestSchema: MessageSchema = {
  name: "SubmitValidationRequest",
  fields: [
    {
      name: "sessionId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "proposalId",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "validatorId",
      type: FieldType.STRING,
      id: 3,
    },
    {
      name: "approved",
      type: FieldType.BOOLEAN,
      id: 4,
    },
    {
      name: "computedHash",
      type: FieldType.STRING,
      id: 5,
    },
    {
      name: "errorReason",
      type: FieldType.STRING,
      id: 6,
    },
    {
      name: "signature",
      type: FieldType.STRING,
      id: 7,
    },
  ],
};


/**
 * Schema for SubmitValidationResponse message
 */
export const SubmitValidationResponseSchema: MessageSchema = {
  name: "SubmitValidationResponse",
  fields: [
    {
      name: "recorded",
      type: FieldType.BOOLEAN,
      id: 1,
    },
    {
      name: "consensusReached",
      type: FieldType.BOOLEAN,
      id: 2,
    },
    {
      name: "consensusApproved",
      type: FieldType.BOOLEAN,
      id: 3,
    },
  ],
};


/**
 * Schema for GetProposalStatusRequest message
 */
export const GetProposalStatusRequestSchema: MessageSchema = {
  name: "GetProposalStatusRequest",
  fields: [
    {
      name: "proposalId",
      type: FieldType.STRING,
      id: 1,
    },
  ],
};


/**
 * Schema for GetProposalStatusResponse message
 */
export const GetProposalStatusResponseSchema: MessageSchema = {
  name: "GetProposalStatusResponse",
  fields: [
    {
      name: "proposal",
      type: FieldType.MESSAGE,
      id: 1,
      messageType: "turnengine.v1.ProposalInfo",
    },
    {
      name: "phase",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "votesReceived",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "votesRequired",
      type: FieldType.NUMBER,
      id: 4,
    },
  ],
};


/**
 * Schema for GameSession message
 */
export const GameSessionSchema: MessageSchema = {
  name: "GameSession",
  fields: [
    {
      name: "sessionId",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "gameType",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "playerIds",
      type: FieldType.REPEATED,
      id: 3,
      repeated: true,
    },
    {
      name: "currentPlayerId",
      type: FieldType.STRING,
      id: 4,
    },
    {
      name: "requiredValidators",
      type: FieldType.NUMBER,
      id: 5,
    },
    {
      name: "currentTick",
      type: FieldType.NUMBER,
      id: 6,
    },
    {
      name: "currentStateHash",
      type: FieldType.STRING,
      id: 7,
    },
    {
      name: "currentStateBlob",
      type: FieldType.STRING,
      id: 8,
    },
    {
      name: "activeProposal",
      type: FieldType.MESSAGE,
      id: 9,
      messageType: "turnengine.v1.ProposalInfo",
    },
    {
      name: "createdAt",
      type: FieldType.MESSAGE,
      id: 10,
      messageType: "google.protobuf.Timestamp",
    },
    {
      name: "updatedAt",
      type: FieldType.MESSAGE,
      id: 11,
      messageType: "google.protobuf.Timestamp",
    },
  ],
};


/**
 * Schema for Pagination message
 */
export const PaginationSchema: MessageSchema = {
  name: "Pagination",
  fields: [
    {
      name: "pageKey",
      type: FieldType.STRING,
      id: 1,
    },
    {
      name: "pageOffset",
      type: FieldType.NUMBER,
      id: 2,
    },
    {
      name: "pageSize",
      type: FieldType.NUMBER,
      id: 3,
    },
  ],
};


/**
 * Schema for PaginationResponse message
 */
export const PaginationResponseSchema: MessageSchema = {
  name: "PaginationResponse",
  fields: [
    {
      name: "nextPageKey",
      type: FieldType.STRING,
      id: 2,
    },
    {
      name: "nextPageOffset",
      type: FieldType.NUMBER,
      id: 3,
    },
    {
      name: "hasMore",
      type: FieldType.BOOLEAN,
      id: 4,
    },
    {
      name: "totalResults",
      type: FieldType.NUMBER,
      id: 5,
    },
  ],
};



/**
 * Package-scoped schema registry for turnengine.v1
 */
export const TurnengineV1SchemaRegistry: Record<string, MessageSchema> = {
  "turnengine.v1.ProposalInfo": ProposalInfoSchema,
  "turnengine.v1.ValidationVote": ValidationVoteSchema,
  "turnengine.v1.ProposalTrackingInfo": ProposalTrackingInfoSchema,
  "turnengine.v1.SubmitProposalRequest": SubmitProposalRequestSchema,
  "turnengine.v1.SubmitProposalResponse": SubmitProposalResponseSchema,
  "turnengine.v1.GetPendingValidationsRequest": GetPendingValidationsRequestSchema,
  "turnengine.v1.GetPendingValidationsResponse": GetPendingValidationsResponseSchema,
  "turnengine.v1.PendingValidation": PendingValidationSchema,
  "turnengine.v1.SubmitValidationRequest": SubmitValidationRequestSchema,
  "turnengine.v1.SubmitValidationResponse": SubmitValidationResponseSchema,
  "turnengine.v1.GetProposalStatusRequest": GetProposalStatusRequestSchema,
  "turnengine.v1.GetProposalStatusResponse": GetProposalStatusResponseSchema,
  "turnengine.v1.GameSession": GameSessionSchema,
  "turnengine.v1.Pagination": PaginationSchema,
  "turnengine.v1.PaginationResponse": PaginationResponseSchema,
};

/**
 * Get schema for a message type from turnengine.v1 package
 */
export function getSchema(messageType: string): MessageSchema | undefined {
  return TurnengineV1SchemaRegistry[messageType];
}

/**
 * Get field schema by name from turnengine.v1 package
 */
export function getFieldSchema(messageType: string, fieldName: string): FieldSchema | undefined {
  const schema = getSchema(messageType);
  return schema?.fields.find(field => field.name === fieldName);
}

/**
 * Get field schema by proto field ID from turnengine.v1 package
 */
export function getFieldSchemaById(messageType: string, fieldId: number): FieldSchema | undefined {
  const schema = getSchema(messageType);
  return schema?.fields.find(field => field.id === fieldId);
}

/**
 * Check if field is part of a oneof group in turnengine.v1 package
 */
export function isOneofField(messageType: string, fieldName: string): boolean {
  const fieldSchema = getFieldSchema(messageType, fieldName);
  return fieldSchema?.oneofGroup !== undefined;
}

/**
 * Get all fields in a oneof group from turnengine.v1 package
 */
export function getOneofFields(messageType: string, oneofGroup: string): FieldSchema[] {
  const schema = getSchema(messageType);
  return schema?.fields.filter(field => field.oneofGroup === oneofGroup) || [];
}