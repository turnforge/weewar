//go:build !wasm
// +build !wasm

package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/panyam/gocurrent"
	v1 "github.com/turnforge/lilbattle/gen/go/lilbattle/v1/models"
	lib "github.com/turnforge/lilbattle/lib"
	"github.com/turnforge/lilbattle/web/assets/themes"
)

type ScreenShotItem struct {
	Kind      string
	Id        string
	Version   int64
	WorldData *v1.WorldData

	ThemeErrors map[string]error
	ThemeFiles  map[string]*v1.File
}

// ScreenshotCompletionCallback is called after all screenshots for an item are complete
type ScreenshotCompletionCallback func([]ScreenShotItem) error

type ScreenShotIndexer struct {
	// The chan where we write world event updates to
	reducer *gocurrent.Reducer2[ScreenShotItem, map[string]ScreenShotItem]

	ClientMgr *ClientMgr

	// Callback invoked after all screenshots for an item are complete
	OnComplete ScreenshotCompletionCallback
}

func NewScreenShotIndexer(clientMgr *ClientMgr) *ScreenShotIndexer {
	s := ScreenShotIndexer{ClientMgr: clientMgr}
	s.reducer = gocurrent.NewReducer2(
		gocurrent.WithFlushPeriod2[ScreenShotItem, map[string]ScreenShotItem](5 * time.Second),
	)
	s.reducer.CollectFunc = func(collection map[string]ScreenShotItem, items ...ScreenShotItem) (map[string]ScreenShotItem, bool) {
		if collection == nil {
			collection = map[string]ScreenShotItem{}
		}
		for _, item := range items {
			curr, ok := collection[item.Id]
			if curr.WorldData == nil || !ok {
				collection[item.Id] = item
			} else if curr.WorldData.ScreenshotIndexInfo != nil && item.WorldData.ScreenshotIndexInfo != nil {
				currLastUpdated := curr.WorldData.ScreenshotIndexInfo.LastUpdatedAt.AsTime()
				itemLastUpdated := item.WorldData.ScreenshotIndexInfo.LastUpdatedAt.AsTime()
				if currLastUpdated.Before(itemLastUpdated) {
					collection[item.Id] = item
				}
			}
		}
		return collection, false
	}
	go s.start()
	return &s
}

func (s *ScreenShotIndexer) start() {
	checkerTicker := time.NewTicker(60 * time.Second) // we also see which items have "changed" but missed the indexing through some loss

	for {
		select {
		case newItems := <-s.reducer.OutputChan():
			s.startBatchProcessing(newItems)
		case <-checkerTicker.C:
			// log.Println("Time to proactively find items that need to be re-indexed")
		}
	}
}

// TODO - Put this into a worker pool later and/or with rate limiting
func (s *ScreenShotIndexer) startBatchProcessing(batch map[string]ScreenShotItem) {
	results := []ScreenShotItem{}
	for _, item := range batch {
		log.Printf("Creating screenshots for %s: %s", item.Kind, item.Id)

		// Render all themes for this item
		for _, themeName := range []string{"default", "modern", "fantasy"} {
			err := s.renderScreenshot(themeName, &item)
			if err != nil {
				log.Printf("Failed to render %s screenshot for %s/%s: %v", themeName, item.Kind, item.Id, err)
				item.ThemeErrors[themeName] = err
				break // Stop processing other themes if one fails
			}
		}

		results = append(results, item)
	}

	// Notify completion with all results
	if s.OnComplete != nil {
		if err := s.OnComplete(results); err != nil {
			log.Printf("Failed to process completion callback: %v", err)
		}
	}
}

func (s *ScreenShotIndexer) Send(kind string, id string, version int64, worldData *v1.WorldData) {
	s.reducer.InputChan() <- ScreenShotItem{kind, id, version, worldData, make(map[string]error), make(map[string]*v1.File)}
}

func (s *ScreenShotIndexer) renderScreenshot(themeName string, item *ScreenShotItem) error {
	// Create theme
	re := lib.DefaultRulesEngine()
	theme, err := themes.CreateTheme(themeName, re.GetCityTerrains())
	if err != nil {
		log.Printf("Failed to create theme %s: %v", themeName, err)
		return err
	}

	// Create renderer for this theme
	renderer, err := themes.CreateWorldRenderer(theme)
	if err != nil {
		log.Printf("Failed to create renderer for theme %s: %v", themeName, err)
		return err
	}

	// Render the image
	imageBytes, contentType, err := renderer.Render(item.WorldData.TilesMap, item.WorldData.UnitsMap, nil)
	if err != nil {
		log.Printf("Failed to render screenshot: %v", err)
		return err
	}

	// Determine file extension from content type
	extension := "png"
	if contentType == "image/svg+xml" {
		extension = "svg"
	}

	// Create path: screenshots/{kind}/{id}/{theme}.{ext}
	filePath := fmt.Sprintf("screenshots/%s/%s/%s.%s", item.Kind, item.Id, themeName, extension)

	// Upload to filestore
	filestoreSvcClient := s.ClientMgr.GetFileStoreSvcClient()
	resp, err := filestoreSvcClient.PutFile(context.Background(), &v1.PutFileRequest{
		File: &v1.File{
			Path:        filePath,
			ContentType: contentType,
		},
		Content: imageBytes,
	})
	if err != nil {
		log.Printf("Failed to upload screenshot to filestore: %v", err)
		return err
	}

	// Store the file in ThemeFiles map
	item.ThemeFiles[themeName] = resp.File

	log.Printf("Successfully uploaded screenshot: %s", filePath)
	return nil
}
