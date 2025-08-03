
// Generated TypeScript schema-aware deserializer
// DO NOT EDIT - This file is auto-generated

import { FieldType, FieldSchema, MessageSchema } from "./deserializer_schemas";
import { WeewarV1Factory } from "./factory";
import { WeewarV1SchemaRegistry } from "./schemas";

// Shared factory instance to avoid creating new instances on every deserializer construction
const DEFAULT_FACTORY = new WeewarV1Factory();

/**
 * Factory interface that deserializer expects
 */
export interface FactoryInterface {
  /**
   * Get factory method for a fully qualified message type
   * This enables cross-package factory delegation
   */
  getFactoryMethod?(messageType: string): ((parent?: any, attributeName?: string, attributeKey?: string | number, data?: any) => FactoryResult<any>) | undefined;
}

/**
 * Factory result interface
 */
export interface FactoryResult<T> {
  instance: T;
  fullyLoaded: boolean;
}

/**
 * Schema-aware deserializer for weewar.v1 package
 */
export class WeewarV1Deserializer {
  constructor(
    private schemaRegistry: Record<string, MessageSchema> = WeewarV1SchemaRegistry,
    private factory: FactoryInterface = DEFAULT_FACTORY
  ) {}

  /**
   * Deserialize an object using schema information
   * @param instance The target instance to populate
   * @param data The source data to deserialize from
   * @param messageType The fully qualified message type (e.g., "library.v1.Book")
   * @returns The populated instance
   */
  deserialize<T>(instance: T, data: any, messageType: string): T {
    if (!data || typeof data !== 'object') {
      return instance;
    }

    const schema = this.schemaRegistry[messageType];
    if (!schema) {
      // Fallback to simple property copying if no schema found
      return this.fallbackDeserialize(instance, data);
    }

    // Process each field according to its schema
    for (const fieldSchema of schema.fields) {
      const fieldValue = data[fieldSchema.name];
      if (fieldValue === null || fieldValue === undefined) {
        continue;
      }

      this.deserializeField(instance, fieldSchema, fieldValue);
    }

    return instance;
  }

  /**
   * Deserialize a single field based on its schema
   */
  private deserializeField(instance: any, fieldSchema: FieldSchema, fieldValue: any): void {
    const fieldName = fieldSchema.name;

    switch (fieldSchema.type) {
      case FieldType.STRING:
      case FieldType.NUMBER:
      case FieldType.BOOLEAN:
        // Simple primitive types - direct assignment
        instance[fieldName] = fieldValue;
        break;

      case FieldType.MESSAGE:
        if (fieldSchema.repeated) {
          // Handle repeated message fields (arrays)
          instance[fieldName] = this.deserializeMessageArray(
            fieldValue,
            fieldSchema.messageType!,
            instance,
            fieldName
          );
        } else {
          // Handle single message field
          instance[fieldName] = this.deserializeMessageField(
            fieldValue,
            fieldSchema.messageType!,
            instance,
            fieldName
          );
        }
        break;

      case FieldType.REPEATED:
        // Handle repeated primitive fields
        if (Array.isArray(fieldValue)) {
          instance[fieldName] = [...fieldValue]; // Simple copy for primitives
        }
        break;

      case FieldType.ONEOF:
        // Handle oneof fields (would need additional logic for union types)
        instance[fieldName] = fieldValue;
        break;

      case FieldType.MAP:
        // Handle map fields (would need additional schema info for key/value types)
        instance[fieldName] = { ...fieldValue };
        break;

      default:
        // Fallback to direct assignment
        instance[fieldName] = fieldValue;
        break;
    }
  }

  /**
   * Deserialize a single message field
   */
  private deserializeMessageField(
    fieldValue: any,
    messageType: string,
    parent: any,
    attributeName: string
  ): any {
    // Try to get factory method using cross-package delegation
    let factoryMethod;
    
    if (this.factory.getFactoryMethod) {
      factoryMethod = this.factory.getFactoryMethod(messageType);
    } else {
      // Fallback to simple method name lookup
      const factoryMethodName = this.getFactoryMethodName(messageType);
      factoryMethod = (this.factory as any)[factoryMethodName];
    }

    if (factoryMethod) {
      const result = factoryMethod(parent, attributeName, undefined, fieldValue);
      if (result.fullyLoaded) {
        return result.instance;
      } else {
        // Factory created instance but didn't populate - use deserializer
        return this.deserialize(result.instance, fieldValue, messageType);
      }
    }

    // No factory method found - fallback
    return this.fallbackDeserialize({}, fieldValue);
  }

  /**
   * Deserialize an array of message objects
   */
  private deserializeMessageArray(
    fieldValue: any[],
    messageType: string,
    parent: any,
    attributeName: string
  ): any[] {
    if (!Array.isArray(fieldValue)) {
      return [];
    }

    // Try to get factory method using cross-package delegation
    let factoryMethod;
    
    if (this.factory.getFactoryMethod) {
      factoryMethod = this.factory.getFactoryMethod(messageType);
    } else {
      // Fallback to simple method name lookup
      const factoryMethodName = this.getFactoryMethodName(messageType);
      factoryMethod = (this.factory as any)[factoryMethodName];
    }

    return fieldValue.map((item, index) => {
      if (factoryMethod) {
        const result = factoryMethod(parent, attributeName, index, item);
        if (result.fullyLoaded) {
          return result.instance;
        } else {
          // Factory created instance but didn't populate - use deserializer
          return this.deserialize(result.instance, item, messageType);
        }
      }

      // No factory method found - fallback
      return this.fallbackDeserialize({}, item);
    });
  }

  /**
   * Convert message type to factory method name
   * "library.v1.Book" -> "newBook"
   */
  private getFactoryMethodName(messageType: string): string {
    const parts = messageType.split('.');
    const typeName = parts[parts.length - 1]; // Get last part (e.g., "Book")
    return 'new' + typeName;
  }

  /**
   * Fallback deserializer for when no schema is available
   */
  private fallbackDeserialize<T>(instance: T, data: any): T {
    if (!data || typeof data !== 'object') {
      return instance;
    }

    for (const [key, value] of Object.entries(data)) {
      if (value !== null && value !== undefined) {
        (instance as any)[key] = value;
      }
    }

    return instance;
  }

  /**
   * Create and deserialize a new instance of a message type
   */
  createAndDeserialize<T>(messageType: string, data: any): T {
    // Try to get factory method using cross-package delegation
    let factoryMethod;
    
    if (this.factory.getFactoryMethod) {
      factoryMethod = this.factory.getFactoryMethod(messageType);
    } else {
      // Fallback to simple method name lookup
      const factoryMethodName = this.getFactoryMethodName(messageType);
      factoryMethod = (this.factory as any)[factoryMethodName];
    }

    if (!factoryMethod) {
      throw new Error(`Could not find factory method to deserialize: ${messageType}`)
    }

    const result = factoryMethod(undefined, undefined, undefined, data);
    if (result.fullyLoaded) {
      return result.instance;
    } else {
      return this.deserialize(result.instance, data, messageType);
    }
  }

  /**
   * Static utility method to create and deserialize a message without needing a deserializer instance
   * @param messageType Fully qualified message type (use Class.MESSAGE_TYPE)
   * @param data Raw data to deserialize
   * @returns Deserialized instance or null if creation failed
   */
  static from<T>(messageType: string, data: any) {
    const deserializer = new WeewarV1Deserializer(); // Uses default factory and schema registry
    return deserializer.createAndDeserialize<T>(messageType, data);
  }
}
