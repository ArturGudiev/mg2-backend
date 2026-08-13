package app

import (
	"arturgudiev/memoryguard/repositories"
	"arturgudiev/memoryguard/services"
	"context"

	"arturgudiev/memoryguard/ent"
)

// App holds all application dependencies.
type App struct {
	Client                    *ent.Client
	UsersRepository           *repositories.UsersRepository
	RefreshTokensRepository   *repositories.RefreshTokensRepository
	MemoryNodesRepository     *repositories.MemoryNodesRepository
	CardsRepository           *repositories.CardsRepository
	CardItemsRepository       *repositories.CardItemsRepository
	CardUserCountsRepository  *repositories.CardUserCountsRepository
	CardUsersRepository       *repositories.CardUsersRepository
	MemoryNodeUsersRepository *repositories.MemoryNodeUsersRepository
	MemoryNodesService        *services.MemoryNodesService
	CardsService              *services.CardsService
	CardItemsService          *services.CardItemsService
	EmailService              *services.EmailService
	ctx                       context.Context
}

// NewApp creates a new App instance with all dependencies initialized using Wire.
func NewApp() (*App, error) {
	return InitializeApp()
}

// Close closes all resources.
func (a *App) Close() error {
	return a.Client.Close()
}
