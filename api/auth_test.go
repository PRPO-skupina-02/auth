package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PRPO-skupina-02/auth/auth"
	"github.com/PRPO-skupina-02/auth/db"
	"github.com/PRPO-skupina-02/common/database"
	"github.com/PRPO-skupina-02/common/xtesting"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRegister(t *testing.T) {
	db, fixtures := database.PrepareTestDatabase(t, db.FixtureFS, db.MigrationsFS)
	r := TestingRouter(t, db)

	tests := []struct {
		name   string
		body   RegisterRequest
		status int
	}{
		{
			name: "ok",
			body: RegisterRequest{
				Email:     "newcustomer@example.com",
				Password:  "password123",
				FirstName: "New",
				LastName:  "Customer",
			},
			status: http.StatusCreated,
		},
		{
			name: "duplicate-email",
			body: RegisterRequest{
				Email:     "customer@example.com",
				Password:  "password123",
				FirstName: "Duplicate",
				LastName:  "User",
			},
			status: http.StatusConflict,
		},
		{
			name: "validation-error-email",
			body: RegisterRequest{
				Email:     "invalid-email",
				Password:  "password123",
				FirstName: "Test",
				LastName:  "User",
			},
			status: http.StatusBadRequest,
		},
		{
			name: "validation-error-password-too-short",
			body: RegisterRequest{
				Email:     "test@example.com",
				Password:  "short",
				FirstName: "Test",
				LastName:  "User",
			},
			status: http.StatusBadRequest,
		},
		{
			name:   "no-body",
			status: http.StatusBadRequest,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := fixtures.Load()
			assert.NoError(t, err)

			targetURL := "/api/v1/auth/register"

			req := xtesting.NewTestingRequest(t, targetURL, http.MethodPost, testCase.body)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			ignoreResp := xtesting.ValuesCheckers{
				"id":         xtesting.ValueUUID(),
				"created_at": xtesting.ValueTimeInPastDuration(time.Second),
				"updated_at": xtesting.ValueTimeInPastDuration(time.Second),
			}

			assert.Equal(t, testCase.status, w.Code)
			if testCase.status == http.StatusCreated {
				xtesting.AssertGoldenJSON(t, w, ignoreResp)
			} else {
				xtesting.AssertGoldenJSON(t, w)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	db, fixtures := database.PrepareTestDatabase(t, db.FixtureFS, db.MigrationsFS)
	r := TestingRouter(t, db)

	tests := []struct {
		name   string
		body   LoginRequest
		status int
	}{
		{
			name: "ok-customer",
			body: LoginRequest{
				Email:    "customer@example.com",
				Password: "customer123",
			},
			status: http.StatusOK,
		},
		{
			name: "ok-employee",
			body: LoginRequest{
				Email:    "employee@example.com",
				Password: "employee123",
			},
			status: http.StatusOK,
		},
		{
			name: "ok-admin",
			body: LoginRequest{
				Email:    "admin@example.com",
				Password: "admin123",
			},
			status: http.StatusOK,
		},
		{
			name: "wrong-password",
			body: LoginRequest{
				Email:    "customer@example.com",
				Password: "wrongpassword",
			},
			status: http.StatusUnauthorized,
		},
		{
			name: "user-not-found",
			body: LoginRequest{
				Email:    "nonexistent@example.com",
				Password: "password123",
			},
			status: http.StatusUnauthorized,
		},
		{
			name:   "no-body",
			status: http.StatusBadRequest,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := fixtures.Load()
			assert.NoError(t, err)

			targetURL := "/api/v1/auth/login"

			req := xtesting.NewTestingRequest(t, targetURL, http.MethodPost, testCase.body)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			ignoreResp := xtesting.ValuesCheckers{
				"access_token":  xtesting.ValueNotEqual(""),
				"refresh_token": xtesting.ValueNotEqual(""),
			}

			assert.Equal(t, testCase.status, w.Code)
			if testCase.status == http.StatusOK {
				xtesting.AssertGoldenJSON(t, w, ignoreResp)
			} else {
				xtesting.AssertGoldenJSON(t, w)
			}
		})
	}
}

func TestVerifyToken(t *testing.T) {
	db, fixtures := database.PrepareTestDatabase(t, db.FixtureFS, db.MigrationsFS)
	r := TestingRouter(t, db)

	// Generate valid tokens for testing
	customerID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	validToken, _ := auth.GenerateToken(customerID, "customer@example.com")

	tests := []struct {
		name   string
		body   map[string]string
		status int
	}{
		{
			name: "ok",
			body: map[string]string{
				"token": validToken,
			},
			status: http.StatusOK,
		},
		{
			name: "invalid-token",
			body: map[string]string{
				"token": "invalid.jwt.token",
			},
			status: http.StatusUnauthorized,
		},
		{
			name:   "no-body",
			status: http.StatusBadRequest,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := fixtures.Load()
			assert.NoError(t, err)

			targetURL := "/api/v1/auth/verify"

			req := xtesting.NewTestingRequest(t, targetURL, http.MethodPost, testCase.body)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.status, w.Code)
			xtesting.AssertGoldenJSON(t, w)
		})
	}
}

func TestRefreshToken(t *testing.T) {
	db, fixtures := database.PrepareTestDatabase(t, db.FixtureFS, db.MigrationsFS)
	r := TestingRouter(t, db)

	// Generate valid refresh token for testing
	customerID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	validRefreshToken, _ := auth.GenerateRefreshToken(customerID, "customer@example.com")

	tests := []struct {
		name   string
		body   map[string]string
		status int
	}{
		{
			name: "ok",
			body: map[string]string{
				"refresh_token": validRefreshToken,
			},
			status: http.StatusOK,
		},
		{
			name: "invalid-token",
			body: map[string]string{
				"refresh_token": "invalid.jwt.token",
			},
			status: http.StatusUnauthorized,
		},
		{
			name:   "no-body",
			status: http.StatusBadRequest,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := fixtures.Load()
			assert.NoError(t, err)

			targetURL := "/api/v1/auth/refresh"

			req := xtesting.NewTestingRequest(t, targetURL, http.MethodPost, testCase.body)
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			ignoreResp := xtesting.ValuesCheckers{
				"access_token":  xtesting.ValueNotEqual(""),
				"refresh_token": xtesting.ValueNotEqual(""),
			}

			assert.Equal(t, testCase.status, w.Code)
			if testCase.status == http.StatusOK {
				xtesting.AssertGoldenJSON(t, w, ignoreResp)
			} else {
				xtesting.AssertGoldenJSON(t, w)
			}
		})
	}
}

func TestGetCurrentUser(t *testing.T) {
	db, fixtures := database.PrepareTestDatabase(t, db.FixtureFS, db.MigrationsFS)
	r := TestingRouter(t, db)

	customerID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	validToken, _ := auth.GenerateToken(customerID, "customer@example.com")

	tests := []struct {
		name   string
		token  string
		status int
	}{
		{
			name:   "ok",
			token:  validToken,
			status: http.StatusOK,
		},
		{
			name:   "no-token",
			status: http.StatusUnauthorized,
		},
		{
			name:   "invalid-token",
			token:  "invalid.jwt.token",
			status: http.StatusUnauthorized,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := fixtures.Load()
			assert.NoError(t, err)

			targetURL := "/api/v1/auth/me"

			req := xtesting.NewTestingRequest(t, targetURL, http.MethodGet, nil)
			if testCase.token != "" {
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testCase.token))
			}
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.status, w.Code)
			xtesting.AssertGoldenJSON(t, w)
		})
	}
}

func TestUpdateCurrentUser(t *testing.T) {
	db, fixtures := database.PrepareTestDatabase(t, db.FixtureFS, db.MigrationsFS)
	r := TestingRouter(t, db)

	customerID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	validToken, _ := auth.GenerateToken(customerID, "customer@example.com")

	tests := []struct {
		name   string
		token  string
		body   UpdateUserRequest
		status int
	}{
		{
			name:  "ok",
			token: validToken,
			body: UpdateUserRequest{
				FirstName: "UpdatedFirst",
				LastName:  "UpdatedLast",
			},
			status: http.StatusOK,
		},
		{
			name:  "ok-partial",
			token: validToken,
			body: UpdateUserRequest{
				FirstName: "OnlyFirst",
			},
			status: http.StatusOK,
		},
		{
			name:   "no-token",
			status: http.StatusUnauthorized,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := fixtures.Load()
			assert.NoError(t, err)

			targetURL := "/api/v1/auth/me"

			req := xtesting.NewTestingRequest(t, targetURL, http.MethodPut, testCase.body)
			if testCase.token != "" {
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testCase.token))
			}
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			ignoreResp := xtesting.ValuesCheckers{
				"updated_at": xtesting.ValueTimeInPastDuration(time.Second),
			}

			assert.Equal(t, testCase.status, w.Code)
			if testCase.status == http.StatusOK {
				xtesting.AssertGoldenJSON(t, w, ignoreResp)
			} else {
				xtesting.AssertGoldenJSON(t, w)
			}
		})
	}
}

func TestChangePassword(t *testing.T) {
	db, fixtures := database.PrepareTestDatabase(t, db.FixtureFS, db.MigrationsFS)
	r := TestingRouter(t, db)

	customerID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	validToken, _ := auth.GenerateToken(customerID, "customer@example.com")

	tests := []struct {
		name   string
		token  string
		body   ChangePasswordRequest
		status int
	}{
		{
			name:  "ok",
			token: validToken,
			body: ChangePasswordRequest{
				OldPassword: "customer123",
				NewPassword: "newpassword123",
			},
			status: http.StatusOK,
		},
		{
			name:  "wrong-old-password",
			token: validToken,
			body: ChangePasswordRequest{
				OldPassword: "wrongpassword",
				NewPassword: "newpassword123",
			},
			status: http.StatusUnauthorized,
		},
		{
			name:  "new-password-too-short",
			token: validToken,
			body: ChangePasswordRequest{
				OldPassword: "customer123",
				NewPassword: "short",
			},
			status: http.StatusBadRequest,
		},
		{
			name:   "no-token",
			status: http.StatusUnauthorized,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := fixtures.Load()
			assert.NoError(t, err)

			targetURL := "/api/v1/auth/me/password"

			req := xtesting.NewTestingRequest(t, targetURL, http.MethodPut, testCase.body)
			if testCase.token != "" {
				req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", testCase.token))
			}
			w := httptest.NewRecorder()

			r.ServeHTTP(w, req)

			assert.Equal(t, testCase.status, w.Code)
			xtesting.AssertGoldenJSON(t, w)
		})
	}
}
