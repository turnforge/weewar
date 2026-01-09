//go:build !wasm
// +build !wasm

package gaebe

import (
	"context"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/datastore"
)

// GenId tracks allocated IDs to prevent duplicates
type GenId struct {
	Key       *datastore.Key `datastore:"-"`
	Class     string         `datastore:"class"`
	Id        string         `datastore:"id"`
	CreatedAt time.Time      `datastore:"created_at"`
}

func randid() string {
	max_id := int64(math.Pow(36, 8))
	randval := rand.Int63() % max_id
	return strconv.FormatInt(randval, 36)
}

// GenerateShortID generates a random 8-character ID
func GenerateShortID() string {
	return randid()
}

// NewID generates or validates an ID for a given class
// If existingId is provided and available, returns it
// Otherwise generates a new random ID
func NewID(ctx context.Context, client *datastore.Client, namespace string, cls string, existingId string) string {
	existingId = strings.ToLower(existingId)

	if existingId != "" {
		// Check if ID already exists
		keyName := cls + ":" + existingId
		key := NamespacedKey("GenId", keyName, namespace)
		var existing GenId
		err := client.Get(ctx, key, &existing)
		if err == nil {
			// ID already taken
			return ""
		}
		if err != datastore.ErrNoSuchEntity {
			// Some other error
			return ""
		}

		// ID is available, register it
		gid := &GenId{
			Class:     cls,
			Id:        existingId,
			CreatedAt: time.Now(),
		}
		_, err = client.Put(ctx, key, gid)
		if err == nil {
			return existingId
		}
		return ""
	}

	// Generate random ID
	for i := 0; i < 5; i++ {
		newId := randid()
		keyName := cls + ":" + newId
		key := NamespacedKey("GenId", keyName, namespace)

		gid := &GenId{
			Class:     cls,
			Id:        newId,
			CreatedAt: time.Now(),
		}
		_, err := client.Put(ctx, key, gid)
		if err == nil {
			return newId
		}
	}

	return ""
}
