/**
 * IndexedDB storage for reference images
 *
 * Stores reference images per world ID to prevent cross-contamination.
 * Gracefully degrades if IndexedDB is not available.
 */

export interface ReferenceImageRecord {
    worldId: string;
    imageBlob: Blob;
    filename?: string;
    timestamp: number;
}

export class ReferenceImageDB {
    private dbName = 'WeeWarReferenceImages';
    private storeName = 'images';
    private version = 1;
    private db: IDBDatabase | null = null;
    private isAvailable = false;

    /**
     * Initialize the IndexedDB connection
     * Gracefully returns false if IndexedDB is not available
     */
    async init(): Promise<boolean> {
        // Check if IndexedDB is available
        if (typeof indexedDB === 'undefined') {
            console.warn('[ReferenceImageDB] IndexedDB not available in this browser');
            this.isAvailable = false;
            return false;
        }

        try {
            this.db = await this.openDatabase();
            this.isAvailable = true;
            return true;
        } catch (error) {
            console.error('[ReferenceImageDB] Failed to initialize IndexedDB:', error);
            this.isAvailable = false;
            return false;
        }
    }

    /**
     * Open or create the IndexedDB database
     */
    private openDatabase(): Promise<IDBDatabase> {
        return new Promise((resolve, reject) => {
            const request = indexedDB.open(this.dbName, this.version);

            request.onerror = () => {
                reject(new Error('Failed to open IndexedDB'));
            };

            request.onsuccess = () => {
                resolve(request.result);
            };

            request.onupgradeneeded = (event) => {
                const db = (event.target as IDBOpenDBRequest).result;

                // Create object store if it doesn't exist
                if (!db.objectStoreNames.contains(this.storeName)) {
                    db.createObjectStore(this.storeName, { keyPath: 'worldId' });
                }
            };
        });
    }

    /**
     * Save a reference image for a world
     * @param worldId The world ID
     * @param imageBlob The image blob to store
     * @param filename Optional filename for the image
     */
    async saveImage(worldId: string, imageBlob: Blob, filename?: string): Promise<void> {
        if (!this.isAvailable || !this.db) {
            console.warn('[ReferenceImageDB] IndexedDB not available, skipping save');
            return;
        }

        try {
            const record: ReferenceImageRecord = {
                worldId,
                imageBlob,
                filename,
                timestamp: Date.now()
            };

            await this.putRecord(record);
            console.log(`[ReferenceImageDB] Saved reference image for world ${worldId}`);
        } catch (error) {
            console.error('[ReferenceImageDB] Failed to save image:', error);
        }
    }

    /**
     * Get the reference image record for a world
     * @param worldId The world ID
     * @returns The full record with blob and metadata, or null if not found
     */
    async getImageRecord(worldId: string): Promise<ReferenceImageRecord | null> {
        if (!this.isAvailable || !this.db) {
            console.warn('[ReferenceImageDB] IndexedDB not available, cannot get image');
            return null;
        }

        try {
            const record = await this.getRecord(worldId);
            if (record) {
                console.log(`[ReferenceImageDB] Retrieved reference image for world ${worldId}`);
                return record;
            }
            return null;
        } catch (error) {
            console.error('[ReferenceImageDB] Failed to get image:', error);
            return null;
        }
    }

    /**
     * Delete the reference image for a world
     * @param worldId The world ID
     */
    async deleteImage(worldId: string): Promise<void> {
        if (!this.isAvailable || !this.db) {
            console.warn('[ReferenceImageDB] IndexedDB not available, skipping delete');
            return;
        }

        try {
            await this.deleteRecord(worldId);
            console.log(`[ReferenceImageDB] Deleted reference image for world ${worldId}`);
        } catch (error) {
            console.error('[ReferenceImageDB] Failed to delete image:', error);
        }
    }

    /**
     * Put a record into the object store
     */
    private putRecord(record: ReferenceImageRecord): Promise<void> {
        return new Promise((resolve, reject) => {
            if (!this.db) {
                reject(new Error('Database not initialized'));
                return;
            }

            const transaction = this.db.transaction([this.storeName], 'readwrite');
            const store = transaction.objectStore(this.storeName);
            const request = store.put(record);

            request.onerror = () => {
                reject(new Error('Failed to put record'));
            };

            request.onsuccess = () => {
                resolve();
            };
        });
    }

    /**
     * Get a record from the object store
     */
    private getRecord(worldId: string): Promise<ReferenceImageRecord | null> {
        return new Promise((resolve, reject) => {
            if (!this.db) {
                reject(new Error('Database not initialized'));
                return;
            }

            const transaction = this.db.transaction([this.storeName], 'readonly');
            const store = transaction.objectStore(this.storeName);
            const request = store.get(worldId);

            request.onerror = () => {
                reject(new Error('Failed to get record'));
            };

            request.onsuccess = () => {
                resolve(request.result || null);
            };
        });
    }

    /**
     * Delete a record from the object store
     */
    private deleteRecord(worldId: string): Promise<void> {
        return new Promise((resolve, reject) => {
            if (!this.db) {
                reject(new Error('Database not initialized'));
                return;
            }

            const transaction = this.db.transaction([this.storeName], 'readwrite');
            const store = transaction.objectStore(this.storeName);
            const request = store.delete(worldId);

            request.onerror = () => {
                reject(new Error('Failed to delete record'));
            };

            request.onsuccess = () => {
                resolve();
            };
        });
    }

    /**
     * Check if IndexedDB is available and initialized
     */
    public get available(): boolean {
        return this.isAvailable;
    }
}
