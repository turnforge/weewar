package weewar

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall/js"
)

// =============================================================================
// Memory SaveHandler (for testing)
// =============================================================================

// MemorySaveHandler stores sessions in memory - useful for testing
type MemorySaveHandler struct {
	sessions map[string][]byte
	mutex    sync.RWMutex
}

// NewMemorySaveHandler creates a new in-memory save handler
func NewMemorySaveHandler() *MemorySaveHandler {
	return &MemorySaveHandler{
		sessions: make(map[string][]byte),
	}
}

func (h *MemorySaveHandler) Save(sessionData []byte) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	// Extract session ID from the data to use as key
	// For now, we'll use a simple approach - in real implementation,
	// we'd parse the JSON to get the sessionId
	sessionID := fmt.Sprintf("session_%d", len(h.sessions))
	h.sessions[sessionID] = sessionData
	
	return nil
}

func (h *MemorySaveHandler) Load(sessionID string) ([]byte, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	data, exists := h.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	
	return data, nil
}

func (h *MemorySaveHandler) List() ([]string, error) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	
	sessionIDs := make([]string, 0, len(h.sessions))
	for id := range h.sessions {
		sessionIDs = append(sessionIDs, id)
	}
	
	return sessionIDs, nil
}

func (h *MemorySaveHandler) Delete(sessionID string) error {
	h.mutex.Lock()
	defer h.mutex.Unlock()
	
	if _, exists := h.sessions[sessionID]; !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	
	delete(h.sessions, sessionID)
	return nil
}

// =============================================================================
// File SaveHandler (for CLI mode)
// =============================================================================

// FileSaveHandler stores sessions as JSON files on disk
type FileSaveHandler struct {
	saveDirectory string
}

// NewFileSaveHandler creates a new file-based save handler
func NewFileSaveHandler(saveDirectory string) (*FileSaveHandler, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(saveDirectory, 0755); err != nil {
		return nil, fmt.Errorf("failed to create save directory: %w", err)
	}
	
	return &FileSaveHandler{
		saveDirectory: saveDirectory,
	}, nil
}

func (h *FileSaveHandler) Save(sessionData []byte) error {
	// For now, generate a filename based on timestamp
	// In real implementation, we'd extract sessionId from the JSON
	filename := fmt.Sprintf("session_%d.json", len(sessionData))
	filepath := filepath.Join(h.saveDirectory, filename)
	
	if err := os.WriteFile(filepath, sessionData, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}
	
	return nil
}

func (h *FileSaveHandler) Load(sessionID string) ([]byte, error) {
	filename := sessionID
	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}
	
	filepath := filepath.Join(h.saveDirectory, filename)
	
	data, err := os.ReadFile(filepath)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}
	
	return data, nil
}

func (h *FileSaveHandler) List() ([]string, error) {
	entries, err := os.ReadDir(h.saveDirectory)
	if err != nil {
		return nil, fmt.Errorf("failed to read save directory: %w", err)
	}
	
	sessionIDs := make([]string, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".json") {
			// Remove .json extension for session ID
			sessionID := strings.TrimSuffix(entry.Name(), ".json")
			sessionIDs = append(sessionIDs, sessionID)
		}
	}
	
	return sessionIDs, nil
}

func (h *FileSaveHandler) Delete(sessionID string) error {
	filename := sessionID
	if !strings.HasSuffix(filename, ".json") {
		filename += ".json"
	}
	
	filepath := filepath.Join(h.saveDirectory, filename)
	
	if err := os.Remove(filepath); err != nil {
		return fmt.Errorf("failed to delete session file: %w", err)
	}
	
	return nil
}

// =============================================================================
// Browser SaveHandler (for WASM in browser)
// =============================================================================

// BrowserSaveHandler calls JavaScript functions to save/load via browser APIs
type BrowserSaveHandler struct {
	apiEndpoint string
}

// NewBrowserSaveHandler creates a new browser-based save handler
func NewBrowserSaveHandler(apiEndpoint string) *BrowserSaveHandler {
	return &BrowserSaveHandler{
		apiEndpoint: apiEndpoint,
	}
}

func (h *BrowserSaveHandler) Save(sessionData []byte) error {
	// Check if we're running in a browser environment
	if js.Global().IsUndefined() {
		return fmt.Errorf("browser save handler requires browser environment")
	}
	
	// Call JavaScript function to handle the save
	// The JS function should make an HTTP POST to the API endpoint
	saveHandler := js.Global().Get("gameSaveHandler")
	if saveHandler.IsUndefined() {
		return fmt.Errorf("gameSaveHandler JavaScript function not found")
	}
	
	// Convert to string for JS
	sessionDataStr := string(sessionData)
	
	// Call the JS function and wait for result
	result := saveHandler.Invoke(sessionDataStr, h.apiEndpoint)
	
	// Check if the save was successful
	success := result.Get("success").Bool()
	if !success {
		errorMsg := result.Get("error").String()
		return fmt.Errorf("browser save failed: %s", errorMsg)
	}
	
	return nil
}

func (h *BrowserSaveHandler) Load(sessionID string) ([]byte, error) {
	if js.Global().IsUndefined() {
		return nil, fmt.Errorf("browser save handler requires browser environment")
	}
	
	loadHandler := js.Global().Get("gameLoadHandler")
	if loadHandler.IsUndefined() {
		return nil, fmt.Errorf("gameLoadHandler JavaScript function not found")
	}
	
	result := loadHandler.Invoke(sessionID, h.apiEndpoint)
	
	success := result.Get("success").Bool()
	if !success {
		errorMsg := result.Get("error").String()
		return nil, fmt.Errorf("browser load failed: %s", errorMsg)
	}
	
	data := result.Get("data").String()
	return []byte(data), nil
}

func (h *BrowserSaveHandler) List() ([]string, error) {
	if js.Global().IsUndefined() {
		return nil, fmt.Errorf("browser save handler requires browser environment")
	}
	
	listHandler := js.Global().Get("gameListHandler")
	if listHandler.IsUndefined() {
		return nil, fmt.Errorf("gameListHandler JavaScript function not found")
	}
	
	result := listHandler.Invoke(h.apiEndpoint)
	
	success := result.Get("success").Bool()
	if !success {
		errorMsg := result.Get("error").String()
		return nil, fmt.Errorf("browser list failed: %s", errorMsg)
	}
	
	// Convert JS array to Go slice
	jsArray := result.Get("data")
	length := jsArray.Get("length").Int()
	sessionIDs := make([]string, length)
	
	for i := 0; i < length; i++ {
		sessionIDs[i] = jsArray.Index(i).String()
	}
	
	return sessionIDs, nil
}

func (h *BrowserSaveHandler) Delete(sessionID string) error {
	if js.Global().IsUndefined() {
		return fmt.Errorf("browser save handler requires browser environment")
	}
	
	deleteHandler := js.Global().Get("gameDeleteHandler")
	if deleteHandler.IsUndefined() {
		return fmt.Errorf("gameDeleteHandler JavaScript function not found")
	}
	
	result := deleteHandler.Invoke(sessionID, h.apiEndpoint)
	
	success := result.Get("success").Bool()
	if !success {
		errorMsg := result.Get("error").String()
		return fmt.Errorf("browser delete failed: %s", errorMsg)
	}
	
	return nil
}

// =============================================================================
// SaveHandler Factory
// =============================================================================

// SaveHandlerConfig configures how to create a SaveHandler
type SaveHandlerConfig struct {
	Type          string `json:"type"`          // "memory", "file", "browser"
	SaveDirectory string `json:"saveDirectory"` // For file handler
	APIEndpoint   string `json:"apiEndpoint"`   // For browser handler
}

// CreateSaveHandler creates a SaveHandler based on configuration
func CreateSaveHandler(config SaveHandlerConfig) (SaveHandler, error) {
	switch config.Type {
	case "memory":
		return NewMemorySaveHandler(), nil
		
	case "file":
		if config.SaveDirectory == "" {
			config.SaveDirectory = "./saves"
		}
		return NewFileSaveHandler(config.SaveDirectory)
		
	case "browser":
		if config.APIEndpoint == "" {
			config.APIEndpoint = "/api/v1/games/sessions"
		}
		return NewBrowserSaveHandler(config.APIEndpoint), nil
		
	default:
		return nil, fmt.Errorf("unknown save handler type: %s", config.Type)
	}
}