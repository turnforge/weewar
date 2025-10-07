package services

import (
	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

type SingletonGameViewPresenterServiceImpl struct {
	BaseGameViewPresenterServiceImpl
	GameViewerPage v1.GameViewerPageClient
}

// NOTE - ONly API really needed here are "getters" and "move processors" so no Creations, Deletions, Listing or even
// GetGame needed - GetGame data is set when we create this
func NewSingletonGameViewPresenterServiceImpl() *SingletonGameViewPresenterServiceImpl {
	w := &SingletonGameViewPresenterServiceImpl{
		BaseGameViewPresenterServiceImpl: BaseGameViewPresenterServiceImpl{
			// WorldsService: SingletonWorldsService
		},
	}
	return w
}
