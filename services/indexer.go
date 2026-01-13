package services

import (
	"fmt"
	"log"
	"time"

	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	v1s "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/services"
)

type IndexerService interface {
	v1s.IndexerServiceServer
}

type BaseIndexerService struct {
	Self IndexerService // The actual implementation
	v1s.UnimplementedIndexerServiceServer
	indexStateChan chan *v1.IndexState
}

/*
func (b *BaseIndexerService) GetIndexStates(contextcontext *v1.GetIndexStatesRequest) (resp *v1.GetIndexStatesResponse, err error) {
	return
}

func (b *BaseIndexerService) DeleteIndexStates(contextcontext *v1.DeleteIndexStatesRequest) (resp *v1.DeleteIndexStatesResponse, err error) {
	return
}

func (b *BaseIndexerService) CreateIndexRecordsLRO(context.Context, *v1.CreateIndexRecordsLRORequest) (resp *v1.CreateIndexRecordsLROResponse, err error) {
	return
}

func (b *BaseIndexerService) GetIndexRecordsLRO(context.Context, *v1.GetIndexRecordsLRORequest) (resp *v1.GetIndexRecordsLROResponse, err error) {
	return
}

func (b *BaseIndexerService) UpdateIndexRecordsLRO(context.Context, *v1.UpdateIndexRecordsLRORequest) (resp *v1.UpdateIndexRecordsLROResponse, err error) {
	return
}
*/

// Starts the bootstrapping loop where we identify which records need indexing and post it to the indexer for eventual
// indexing.  This will only run once.  After this is run on startup, new creations/updates will come to the indexer
// directly.  Ideally this should be limited in how many records it "finds".   So as to not DDOS our indexer on
// startup.  Only elements that are created or updated but not indexed since the last indexing should be considered
func (b *BaseIndexerService) Bootstrap() {
	// Find all games
	// for game := findGamesToIndex() { s.EnsureIndexState }
}

// The indexer loop is key - it is the liason between the bookkeeper and the workers.
//
// There are 2 main flows:
//
// Reactive Flow:
// ==============
//
// An entity is updated via an API (say Create or Update or even Delete)
// The API layer sends an "EnsureIndexState" call for this entity - ensure it has a updated_at field.
// If this is a new index state then the indexed_at would be null (or at 0) otherwise it would have updated_at >
// indexed_at.
// What this does is now indicate that we have these entries "ready" for indexing.
// The indexer loop runs periodically collecting items that are "ready" for indexing and proceeds to index them
//
// Bootstrap Flow:
// ===============
//
// The reactive flow is great when things are all running but we will have restarts (or the indexer will
// be restarted) or we may not even have a dedicated indexer and just run in batch mode manually
//
// Here we need something like:
// Give me N entities that need to be indexed but do not have an entry in our Index table (or updated_at
// is missing).   ie we could have a new entity that was created but it was never indexed so this would not exist
// in our table.  Alternatively due to a bug the update never may have sent an update to the indexer service
// (so updated_at would never be > indexed_at)
//
// So here the key is for us to search all entities at their source of truth, cross reference with
// index table.  The naive way of doing this is:
//
//	for each item in Items;
//		if item NOT IN indexerService:
//	    	indexerService.Ensure(item)		// we should do this for each index_type
//
// Problem is this unnecessarily going through the entire dataset just to find which "new items" have to be
// reindexed (and only useful for cases when outages caused af few issues to be not indexed).
//
// The other problem is this does not look at updated_at.  For example if an item exists in the indexer
// but hasnt been updated (say item udpate failed after saving into source of truth but somehow the indexer wasnt
// available).  here updated_at would < indexed_at but that is in correct.  So we could update our loop above
// to be:
//
//	for each item in Items;
//		if item NOT IN indexerService or item.updated_at > indexerService.getItem(item.id).updated_at:
//	    	indexerService.Ensure(item)		// we should do this for each index_type
//
// Here we are checking the updated_at recorded time.   This takes care of missing items but still need a
// full scan of our DB which is unnecesary (especially if items will change soon or for items that may
// never changes)
//
// What is needed is a quick way to get items that would have changed - or atleast do so incrementally.
// We have a few options:
//
//  1. Keep track of when the last indexing was (and on the indexer side we could only set this if no new
//     items were found in a time window) and only get items that were updated (or created) after that time.
//     Chances are those would be in need of an indexing (and ok to do so again).  The way we could get this
//     time is by simply querying the "latest" indexed item's indexed_at time
//
// The problem is this.  Consider the following:
//
// On the indexer we have items A, B, C, D updated at T0, T1, T1, T2(increasing order)
// on the Db we have items with A, B, C, D with their updated_at being T1, T1, T1.5, T2 (ie only A and C
// have updated).  If we pick T2 then since A and C have time "older" than T2 they wont be marked.
// But if we pick the "oldest" time - say T0 then what happens?  here B and D will have to be re-indexed
// even though they have not changed.  This in a way says every entry is behind by atmost X.
//
//  2. Keep track of when an entity was last indexed for each index_type on the source of truth itself (as
//     metadata) - then we could do a query where updated_at > indexed_at.  This however does not exploit
//     indexes due to comparing non static values.   So we will have a calculated "needs_indexing" field
//     that will be set each time we create/update an entity.  Given there are "few" index types we could
//     have an individual one for each field (and only update that when a particular type of work changes)
//
// Then by having a partial index on this field we can find items needing indexing (even for specific types)
func (b *BaseIndexerService) StartIndexer() {
	pickedInWindow := map[string]*v1.IndexState{}   // collects all updated items and stores entries that have updated in this window
	windowTimer := time.NewTimer(10 * time.Second)  // does the actual actioning the index update
	checkerTimer := time.NewTimer(10 * time.Second) // we also see which items have "changed" but missed the indexing through some loss
	for {
		select {
		case <-windowTimer.C:
			log.Println("Time to update index on batch")
		case newItem := <-b.indexStateChan:
			itemkey := fmt.Sprintf("%s:%s:%s", newItem.EntityType, newItem.EntityId, newItem.IndexType)
			pickedInWindow[itemkey] = newItem
		case <-checkerTimer.C:
			// log.Println("Time to proactively find items that need to be re-indexed")
		}
	}
}
