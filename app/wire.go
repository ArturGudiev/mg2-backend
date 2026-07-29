//go:build wireinject
// +build wireinject

package app

import (
	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/ent/migrate"
	"arturgudiev/memoryguard/ent/schema"
	"arturgudiev/memoryguard/repositories"
	"arturgudiev/memoryguard/services"
	"context"
	"log"
	"net/url"
	"os"
	"strings"

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

func provideEntClient() (*ent.Client, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbName := os.Getenv("DB_NAME")
		if dbName == "" {
			dbName = "memoryguard"
		}
		dbPassword := os.Getenv("DB_PASSWORD")
		if dbPassword == "" {
			dbPassword = "postgres"
		}
		dbHost := os.Getenv("DB_HOST")
		if dbHost == "" {
			dbHost = "localhost"
		}
		dbUser := os.Getenv("DB_USER")
		if dbUser == "" {
			dbUser = "postgres"
		}
		u := &url.URL{
			Scheme:   "postgres",
			User:     url.UserPassword(dbUser, dbPassword),
			Host:     dbHost,
			Path:     "/" + url.PathEscape(dbName),
			RawQuery: "sslmode=disable",
		}
		dbURL = u.String()
		log.Printf("DATABASE_URL was not set; using constructed DB URL: %s", dbURL)
	}

	client, err := ent.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	if err := client.Schema.Create(ctx, migrate.WithDropColumn(true), migrate.WithDropIndex(true)); err != nil {
		errMsg := strings.ToLower(err.Error())
		if strings.Contains(errMsg, "already exists") ||
			strings.Contains(errMsg, "permission denied") ||
			strings.Contains(errMsg, "unexpected attribute change") ||
			strings.Contains(errMsg, "expect identity") {
			log.Printf("Warning: Schema migration had issues: %v", err)
			log.Println("Continuing with existing schema...")
		} else {
			client.Close()
			return nil, err
		}
	} else {
		log.Println("Schema migration completed successfully")
	}

	// One-time / opt-in: set role=admin for all existing users.
	if os.Getenv("MIGRATE_USERS_ROLE_ADMIN") == "true" {
		n, err := client.User.Update().SetRole(schema.UserRoleAdmin).Save(ctx)
		if err != nil {
			log.Printf("Warning: MIGRATE_USERS_ROLE_ADMIN failed: %v", err)
		} else {
			log.Printf("MIGRATE_USERS_ROLE_ADMIN: set role=admin on %d user(s)", n)
		}
	}

	return client, nil
}

func provideApp(
	client *ent.Client,
	usersRepository *repositories.UsersRepository,
	refreshTokensRepository *repositories.RefreshTokensRepository,
	memoryNodesRepository *repositories.MemoryNodesRepository,
	cardsRepository *repositories.CardsRepository,
	cardItemsRepository *repositories.CardItemsRepository,
	cardUserCountsRepository *repositories.CardUserCountsRepository,
	cardUsersRepository *repositories.CardUsersRepository,
	memoryNodeUsersRepository *repositories.MemoryNodeUsersRepository,
	memoryNodesService *services.MemoryNodesService,
	cardsService *services.CardsService,
	cardItemsService *services.CardItemsService,
) *App {
	return &App{
		Client:                    client,
		UsersRepository:           usersRepository,
		RefreshTokensRepository:   refreshTokensRepository,
		MemoryNodesRepository:     memoryNodesRepository,
		CardsRepository:           cardsRepository,
		CardItemsRepository:       cardItemsRepository,
		CardUserCountsRepository:  cardUserCountsRepository,
		CardUsersRepository:       cardUsersRepository,
		MemoryNodeUsersRepository: memoryNodeUsersRepository,
		MemoryNodesService:        memoryNodesService,
		CardsService:              cardsService,
		CardItemsService:          cardItemsService,
		ctx:                       context.Background(),
	}
}
