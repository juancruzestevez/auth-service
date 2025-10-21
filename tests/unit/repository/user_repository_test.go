package repository_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/juancruzestevez/auth-service/models"
	"github.com/juancruzestevez/auth-service/repository"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock, func()) {
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to create mock database: %v", err)
	}

	dialector := postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "postgres",
	})

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to open gorm connection: %v", err)
	}

	cleanup := func() {
		sqlDB.Close()
	}

	return db, mock, cleanup
}

func TestUserRepository_Create(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewUserRepository(db)

	user := &models.User{
		FirstName: "Juan",
		LastName:  "Pérez",
		Nickname:  "juanp",
		Email:     "juan@example.com",
		Password:  "hashedpassword",
	}

	mock.ExpectBegin()
	mock.ExpectQuery(regexp.QuoteMeta(`INSERT INTO "users"`)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

	err := repo.Create(user)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestUserRepository_FindByEmail(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewUserRepository(db)

	tests := []struct {
		name      string
		email     string
		mockSetup func()
		wantUser  bool
		wantErr   bool
	}{
		{
			name:  "Usuario encontrado",
			email: "test@example.com",
			mockSetup: func() {
				rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "nickname", "email", "password"}).
					AddRow(1, time.Now(), time.Now(), nil, "Test", "User", "testuser", "test@example.com", "hashedpass")

				// GORM agrega deleted_at IS NULL y LIMIT
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
					WithArgs("test@example.com", 1).
					WillReturnRows(rows)
			},
			wantUser: true,
			wantErr:  false,
		},
		{
			name:  "Usuario no encontrado",
			email: "notfound@example.com",
			mockSetup: func() {
				mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE email = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
					WithArgs("notfound@example.com", 1).
					WillReturnError(gorm.ErrRecordNotFound)
			},
			wantUser: false,
			wantErr:  false, // El repositorio retorna nil, no error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockSetup()

			user, err := repo.FindByEmail(tt.email)

			if (err != nil) != tt.wantErr {
				t.Errorf("FindByEmail() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.wantUser && user == nil {
				t.Error("Expected user to be found")
			}

			if !tt.wantUser && user != nil {
				t.Error("Expected user to be nil")
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Errorf("Unfulfilled expectations: %v", err)
			}
		})
	}
}

func TestUserRepository_FindByNickname(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "nickname", "email", "password"}).
		AddRow(1, time.Now(), time.Now(), nil, "Test", "User", "testuser", "test@example.com", "hashedpass")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE nickname = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs("testuser", 1).
		WillReturnRows(rows)

	user, err := repo.FindByNickname("testuser")

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if user == nil {
		t.Error("Expected user to be found")
	}

	if user != nil && user.Nickname != "testuser" {
		t.Errorf("Expected nickname testuser, got %s", user.Nickname)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}

func TestUserRepository_FindByID(t *testing.T) {
	db, mock, cleanup := setupMockDB(t)
	defer cleanup()

	repo := repository.NewUserRepository(db)

	rows := sqlmock.NewRows([]string{"id", "created_at", "updated_at", "deleted_at", "first_name", "last_name", "nickname", "email", "password"}).
		AddRow(1, time.Now(), time.Now(), nil, "Test", "User", "testuser", "test@example.com", "hashedpass")

	mock.ExpectQuery(regexp.QuoteMeta(`SELECT * FROM "users" WHERE "users"."id" = $1 AND "users"."deleted_at" IS NULL ORDER BY "users"."id" LIMIT $2`)).
		WithArgs(1, 1).
		WillReturnRows(rows)

	user, err := repo.FindByID(1)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if user == nil {
		t.Error("Expected user to be found")
	}

	if user != nil && user.ID != 1 {
		t.Errorf("Expected ID 1, got %d", user.ID)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unfulfilled expectations: %v", err)
	}
}
