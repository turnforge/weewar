package services

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	pj "google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// Generic file storage we can use for various kinds of entities
// These all share a few things in common:
// 1. A file storage dir
// 2. Each dir in this storage dir represents a unique entity by id
// 3. Each directory will have a metadata.json - that represents the main metadata for this entity
// 4. Can have other xyz.json for xyz specific attributes
type FileStorage struct {
	storageDir string
}

func NewFileStorage(storageDir string) *FileStorage {
	if storageDir == "" {
		storageDir = GAMES_STORAGE_DIR
	}

	// Ensure storage directory exists
	if err := os.MkdirAll(storageDir, 0755); err != nil {
		log.Printf("Failed to create games storage directory: %v", err)
		panic(err)
	}
	return &FileStorage{storageDir: storageDir}
}

func (f *FileStorage) CreateEntity(customId string) (newId string, err error) {
	if customId != "" {
		// Game ID provided - check if it's available
		if exists, err := f.EntityExists(customId); exists || err != nil {
			return "", fmt.Errorf("ID check failed or ID already exists: %w", err)
		} else {
			return customId, nil
		}
	}

	// No game ID provided, generate a new one
	const MaxRetries = 5
	for range MaxRetries {
		customId, err := newRandomId()
		if err != nil {
			return "", fmt.Errorf("failed to generate game ID: %w", err)
		}

		// Check if this ID is already taken
		exists, err := f.EntityExists(customId)
		if !exists {
			return customId, nil
		}
	}
	// ID collision, try again
	return "", fmt.Errorf("ID Generation failed")
}

func (f *FileStorage) EntityExists(id string) (exists bool, err error) {
	ed := f.getEntityDir(id)
	if _, err := os.Stat(ed); err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, nil
	}
}

func (f *FileStorage) DeleteEntity(id string) error {
	entityPath := f.getEntityDir(id)
	err := os.RemoveAll(entityPath)
	if err != nil {
		if os.IsNotExist(err) {
			err = nil
		}
	}
	return err
}

func ListFSEntities[T proto.Message](f *FileStorage, validate func(entry T) bool) (entities []T, err error) {
	// Read all game directories
	entries, err := os.ReadDir(f.storageDir)
	if err != nil {
		if os.IsNotExist(err) {
			// Storage directory doesn't exist yet, return empty list
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read games storage directory: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		entityId := entry.Name()
		newInstance, err := LoadFSArtifact[T](f, entityId, "metadata")
		if err != nil {
			log.Printf("Failed to artifact for entity %s: %v", entityId, err)
			continue
		}
		// if message != nil { // it may have been filtered out
		// if !reflect.ValueOf(message).IsZero() {

		if validate == nil || validate(newInstance) {
			// Only return metadata for listing (not full game data)
			entities = append(entities, newInstance)
		}
	}
	return
}

func LoadFSArtifact[T proto.Message](f *FileStorage, id string, name string) (out T, err error) {
	data, err := f.ReadArtifactFile(id, name)
	if err != nil {
		log.Printf("Failed to load artifact (%s) for entity %s: %v", name, id, err)
		return
	}

	out = newProtoInstance[T]()
	err = pj.Unmarshal(data, out)
	return
}

func (f *FileStorage) LoadArtifact(id string, name string, m proto.Message) error {
	data, err := f.ReadArtifactFile(id, name)
	if err != nil {
		return err
	}
	return pj.Unmarshal(data, m)
}

func (f *FileStorage) SaveArtifact(id string, name string, m proto.Message) error {
	entityDir := f.getEntityDir(id)
	if err := os.MkdirAll(entityDir, 0755); err != nil {
		return fmt.Errorf("failed to create entity directory %s: %w", entityDir, err)
	}

	mo := pj.MarshalOptions{
		Indent:            "  ",
		UseProtoNames:     true,
		EmitDefaultValues: true,
	}
	data, err := mo.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata for entity %s: %w", id, err)
	}

	artifactPath := f.getArtifactPath(id, name)
	if err := os.WriteFile(artifactPath, data, 0644); err != nil {
		return fmt.Errorf("failed to write metadata for entity %s: %w", id, err)
	}

	return nil
}

func (f *FileStorage) getEntityDir(entityId string) string {
	return filepath.Join(f.storageDir, entityId)
}

func (f *FileStorage) getArtifactPath(entityId string, name string) string {
	return filepath.Join(f.getEntityDir(entityId), fmt.Sprintf("%s.json", name))
}

func (f *FileStorage) ReadArtifactFile(id string, name string) ([]byte, error) {
	path := f.getArtifactPath(id, name)
	return os.ReadFile(path)
}
