package app

import (
	"arturgudiev/memoryguard/ent"
	"arturgudiev/memoryguard/ent/migrate"
	"arturgudiev/memoryguard/ent/schema"
	"arturgudiev/memoryguard/repositories"
	"arturgudiev/memoryguard/services"
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/url"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func databaseURL() string {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL != "" {
		return dbURL
	}

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
	return dbURL
}

func provideEntClient() (*ent.Client, error) {
	dbURL := databaseURL()

	ctx := context.Background()
	if err := migratePlainPasswordsToHash(ctx, dbURL); err != nil {
		return nil, fmt.Errorf("password hash migration: %w", err)
	}

	client, err := ent.Open("postgres", dbURL)
	if err != nil {
		return nil, err
	}

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

// migratePlainPasswordsToHash moves legacy plaintext users.password into bcrypt
// users.password_hash before Ent applies NOT NULL on the new column.
func migratePlainPasswordsToHash(ctx context.Context, dbURL string) error {
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		return err
	}
	defer db.Close()

	var usersExists bool
	if err := db.QueryRowContext(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'users'
		)`).Scan(&usersExists); err != nil {
		return err
	}
	if !usersExists {
		return nil
	}

	hasColumn := func(name string) (bool, error) {
		var exists bool
		err := db.QueryRowContext(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'users' AND column_name = $1
			)`, name).Scan(&exists)
		return exists, err
	}

	hasPassword, err := hasColumn("password")
	if err != nil {
		return err
	}
	hasPasswordHash, err := hasColumn("password_hash")
	if err != nil {
		return err
	}

	if !hasPassword && !hasPasswordHash {
		return nil
	}

	if hasPassword && !hasPasswordHash {
		if _, err := db.ExecContext(ctx, `ALTER TABLE users ADD COLUMN password_hash VARCHAR`); err != nil {
			return err
		}
		hasPasswordHash = true
	}

	if hasPassword && hasPasswordHash {
		rows, err := db.QueryContext(ctx, `
			SELECT id, password FROM users
			WHERE password_hash IS NULL OR password_hash = ''`)
		if err != nil {
			return err
		}
		defer rows.Close()

		type row struct {
			id       int
			password string
		}
		var toHash []row
		for rows.Next() {
			var r row
			if err := rows.Scan(&r.id, &r.password); err != nil {
				return err
			}
			toHash = append(toHash, r)
		}
		if err := rows.Err(); err != nil {
			return err
		}
		_ = rows.Close()

		for _, r := range toHash {
			hash, err := bcrypt.GenerateFromPassword([]byte(r.password), bcrypt.DefaultCost)
			if err != nil {
				return err
			}
			if _, err := db.ExecContext(ctx,
				`UPDATE users SET password_hash = $1 WHERE id = $2`,
				string(hash), r.id,
			); err != nil {
				return err
			}
		}

		if _, err := db.ExecContext(ctx, `ALTER TABLE users DROP COLUMN password`); err != nil {
			return err
		}
		log.Printf("Migrated %d user password(s) to password_hash", len(toHash))
	}

	if hasPasswordHash {
		if _, err := db.ExecContext(ctx, `DELETE FROM users WHERE password_hash IS NULL OR password_hash = ''`); err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, `ALTER TABLE users ALTER COLUMN password_hash SET NOT NULL`); err != nil {
			// Column may already be NOT NULL.
			if !strings.Contains(strings.ToLower(err.Error()), "already") {
				return err
			}
		}
	}

	return nil
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
