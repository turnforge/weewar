# Services Next Steps

## Completed
- ✅ FileStore service with local filesystem backend
- ✅ R2 FileStore service with presigned URLs
- ✅ Screenshot indexing pipeline with batch processing
- ✅ Optimistic locking for WorldData updates
- ✅ BackendWorldsService abstraction for GORM and FS implementations
- ✅ WorldDataUpdater interface for storage-agnostic operations
- ✅ Google Cloud Datastore backend (gaebe package) for App Engine deployment
  - Uses protoc-gen-dal generated entities with proper datastore_tags
  - Supports composite indexes for needs_indexing queries
  - Cross-entity transactions for SaveMoveGroup atomicity
  - Backend selectable at runtime via WORLDS_SERVICE_BE/GAMES_SERVICE_BE=gae

## TODO

### FileStore
- [ ] Implement file size limits and content-type validation
- [ ] Add file metadata caching to avoid repeated HeadObject calls
- [ ] Consider adding file versioning support
- [ ] Add cleanup/garbage collection for orphaned screenshots

### Screenshot Indexing
- [ ] Implement worker pool for parallel screenshot generation
- [ ] Add rate limiting to prevent overwhelming filestore
- [ ] Implement retry logic for failed theme renders
- [ ] Add proactive re-indexing for items with NeedsIndexing=true (periodic checker)
- [ ] Support screenshot generation for games (currently only worlds)

### Testing
- [ ] Add unit tests for path security (directory traversal attempts)
- [ ] Add integration tests for screenshot pipeline
- [ ] Test optimistic locking conflicts with concurrent updates
- [ ] Test R2 presigned URL generation and expiry

### Documentation
- [ ] Document screenshot URL structure and theme naming
- [ ] Add examples for FileStore API usage
- [ ] Document WorldDataUpdater interface contract
