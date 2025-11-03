package singleton

import (
	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
	"github.com/panyam/turnengine/games/weewar/services"
	pj "google.golang.org/protobuf/encoding/protojson"
)

type SingletonWorldsService struct {
	services.BaseWorldsService
	SingletonWorld     *v1.World
	SingletonWorldData *v1.WorldData

	RuntimeWorld *services.World
}

// NOTE - ONly API really needed here are "getters" and "move processors" so no Creations, Deletions, Listing or even
// GetWorld needed - GetWorld data is set when we create this
func NewSingletonWorldsService() *SingletonWorldsService {
	w := &SingletonWorldsService{
		BaseWorldsService: services.BaseWorldsService{
			// WorldsService: SingletonWorldsService
		},
		SingletonWorld:     &v1.World{},
		SingletonWorldData: &v1.WorldData{},
	}
	w.Self = w
	return w
}

func (w *SingletonWorldsService) GetRuntimeWorld(gameId string) (*services.World, error) {
	return w.RuntimeWorld, nil
}

func (w *SingletonWorldsService) SaveWorld(game *v1.World, state *v1.WorldData) error {
	return nil
}

func (w *SingletonWorldsService) Load(
	worldBytes []byte,
	worldDataBytes []byte,
) {
	// Now load data from the bytes
	if err := pj.Unmarshal(worldBytes, w.SingletonWorld); err != nil {
		panic(err)
	}
	if err := pj.Unmarshal(worldDataBytes, w.SingletonWorldData); err != nil {
		panic(err)
	}
}
