import { IndexStateGORM as IndexStateGORMInterface, IndexRecordsLROGORM as IndexRecordsLROGORMInterface } from "./interfaces";




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


