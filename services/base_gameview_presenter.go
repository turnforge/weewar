package services

import (
	v1 "github.com/panyam/turnengine/games/weewar/gen/go/weewar/v1"
)

type GameViewPresenterImpl interface {
	v1.GameViewPresenterServer
}

type BaseGameViewPresenterImpl struct {
	v1.UnimplementedGameViewPresenterServer
}
