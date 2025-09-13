package services

type SingletonGameViewPresenterServiceImpl struct {
	BaseGameViewPresenterServiceImpl
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
