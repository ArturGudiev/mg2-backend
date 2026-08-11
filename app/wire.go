//go:build wireinject
// +build wireinject

package app

import (
	"arturgudiev/memoryguard/repositories"
	"arturgudiev/memoryguard/services"

	"github.com/google/wire"
	_ "github.com/lib/pq"
)

// InitializeApp creates App with all dependencies wired automatically.
func InitializeApp() (*App, error) {
	wire.Build(
		provideEntClient,
		repositories.NewUsersRepository,
		repositories.NewRefreshTokensRepository,
		repositories.NewMemoryNodesRepository,
		repositories.NewCardsRepository,
		repositories.NewCardItemsRepository,
		repositories.NewCardUserCountsRepository,
		repositories.NewCardUsersRepository,
		repositories.NewMemoryNodeUsersRepository,
		services.NewMemoryNodesService,
		services.NewCardItemsService,
		services.NewCardsService,
		provideApp,
	)
	return nil, nil
}
