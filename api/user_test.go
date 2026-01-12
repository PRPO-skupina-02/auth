package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PRPO-skupina-02/auth/auth"
	"github.com/PRPO-skupina-02/auth/db"
	"github.com/PRPO-skupina-02/auth/models"
	"github.com/PRPO-skupina-02/common/database"
	"github.com/PRPO-skupina-02/common/xtesting"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestUsersList(t *testing.T) {
	db, fixtures := database.PrepareTestDatabase(t, db.FixtureFS, db.MigrationsFS)
	r := TestingRouter(t, db)

	adminID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	adminToken, _ := auth.GenerateToken(adminID, "admin@example.com")

	customerID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	customerToken, _ := auth.GenerateToken(customerID, "customer@example.com")

	tests := []struct {
		name   string
		token  string
		params string
		status int
	}{
		{
			name:   "ok",
			token:  adminToken,
			status: http.StatusOK,
		},
		{
			name:   "ok-paginated",
			token:  adminToken,
			params: "?limit=2&offset=1",
			status: http.StatusOK,
		},
		{
			name:   "ok-sorted",
			token:  adminToken,
			params: "?sort=email",
			status: http.StatusOK,
		},
		{
			name:   "forbidden-customer",
			token:  customerToken,
			status: http.StatusForbidden,
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

			targetURL := fmt.Sprintf("/api/v1/auth/users%s", testCase.params)

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

func TestUsersShow(t *testing.T) {
	db, fixtures := database.PrepareTestDatabase(t, db.FixtureFS, db.MigrationsFS)
	r := TestingRouter(t, db)

	adminID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	adminToken, _ := auth.GenerateToken(adminID, "admin@example.com")

	customerID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	customerToken, _ := auth.GenerateToken(customerID, "customer@example.com")

	tests := []struct {
		name   string
		token  string
		userID string
		status int
	}{
		{
			name:   "ok",
			token:  adminToken,
			userID: "00000000-0000-0000-0000-000000000003",
			status: http.StatusOK,
		},
		{
			name:   "not-found",
			token:  adminToken,
			userID: "00000000-0000-0000-0000-999999999999",
			status: http.StatusNotFound,
		},
		{
			name:   "invalid-uuid",
			token:  adminToken,
			userID: "invalid-uuid",
			status: http.StatusBadRequest,
		},
		{
			name:   "forbidden-customer",
			token:  customerToken,
			userID: "00000000-0000-0000-0000-000000000002",
			status: http.StatusForbidden,
		},
		{
			name:   "no-token",
			userID: "00000000-0000-0000-0000-000000000003",
			status: http.StatusUnauthorized,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := fixtures.Load()
			assert.NoError(t, err)

			targetURL := fmt.Sprintf("/api/v1/auth/users/%s", testCase.userID)

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

func TestAdminCreateUser(t *testing.T) {
	db, fixtures := database.PrepareTestDatabase(t, db.FixtureFS, db.MigrationsFS)
	r := TestingRouter(t, db)

	adminID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	adminToken, _ := auth.GenerateToken(adminID, "admin@example.com")

	customerID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	customerToken, _ := auth.GenerateToken(customerID, "customer@example.com")

	tests := []struct {
		name   string
		token  string
		body   AdminCreateUserRequest
		status int
	}{
		{
			name:  "ok-employee",
			token: adminToken,
			body: AdminCreateUserRequest{
				Email:     "newemployee@example.com",
				Password:  "password123",
				FirstName: "New",
				LastName:  "Employee",
				Role:      models.RoleEmployee,
				Active:    true,
			},
			status: http.StatusCreated,
		},
		{
			name:  "ok-customer",
			token: adminToken,
			body: AdminCreateUserRequest{
				Email:     "newcustomer@example.com",
				Password:  "password123",
				FirstName: "New",
				LastName:  "Customer",
				Role:      models.RoleCustomer,
				Active:    true,
			},
			status: http.StatusCreated,
		},
		{
			name:  "ok-admin",
			token: adminToken,
			body: AdminCreateUserRequest{
				Email:     "newadmin@example.com",
				Password:  "password123",
				FirstName: "New",
				LastName:  "Admin",
				Role:      models.RoleAdmin,
				Active:    true,
			},
			status: http.StatusCreated,
		},
		{
			name:  "duplicate-email",
			token: adminToken,
			body: AdminCreateUserRequest{
				Email:     "customer@example.com",
				Password:  "password123",
				FirstName: "Duplicate",
				LastName:  "User",
				Role:      models.RoleEmployee,
				Active:    true,
			},
			status: http.StatusConflict,
		},
		{
			name:  "validation-error-email",
			token: adminToken,
			body: AdminCreateUserRequest{
				Email:     "invalid-email",
				Password:  "password123",
				FirstName: "Test",
				LastName:  "User",
				Role:      models.RoleEmployee,
				Active:    true,
			},
			status: http.StatusBadRequest,
		},
		{
			name:  "validation-error-password-too-short",
			token: adminToken,
			body: AdminCreateUserRequest{
				Email:     "test@example.com",
				Password:  "short",
				FirstName: "Test",
				LastName:  "User",
				Role:      models.RoleEmployee,
				Active:    true,
			},
			status: http.StatusBadRequest,
		},
		{
			name:  "forbidden-customer",
			token: customerToken,
			body: AdminCreateUserRequest{
				Email:     "test@example.com",
				Password:  "password123",
				FirstName: "Test",
				LastName:  "User",
				Role:      models.RoleEmployee,
				Active:    true,
			},
			status: http.StatusForbidden,
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

			targetURL := "/api/v1/auth/users"

			req := xtesting.NewTestingRequest(t, targetURL, http.MethodPost, testCase.body)
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

func TestUsersUpdate(t *testing.T) {
	db, fixtures := database.PrepareTestDatabase(t, db.FixtureFS, db.MigrationsFS)
	r := TestingRouter(t, db)

	adminID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	adminToken, _ := auth.GenerateToken(adminID, "admin@example.com")

	customerID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	customerToken, _ := auth.GenerateToken(customerID, "customer@example.com")

	firstName := "UpdatedName"
	active := false

	tests := []struct {
		name   string
		token  string
		userID string
		body   AdminUpdateUserRequest
		status int
	}{
		{
			name:   "ok",
			token:  adminToken,
			userID: "00000000-0000-0000-0000-000000000003",
			body: AdminUpdateUserRequest{
				FirstName: &firstName,
			},
			status: http.StatusOK,
		},
		{
			name:   "ok-deactivate",
			token:  adminToken,
			userID: "00000000-0000-0000-0000-000000000003",
			body: AdminUpdateUserRequest{
				Active: &active,
			},
			status: http.StatusOK,
		},
		{
			name:   "not-found",
			token:  adminToken,
			userID: "00000000-0000-0000-0000-999999999999",
			body: AdminUpdateUserRequest{
				FirstName: &firstName,
			},
			status: http.StatusNotFound,
		},
		{
			name:   "forbidden-customer",
			token:  customerToken,
			userID: "00000000-0000-0000-0000-000000000002",
			body: AdminUpdateUserRequest{
				FirstName: &firstName,
			},
			status: http.StatusForbidden,
		},
		{
			name:   "no-token",
			userID: "00000000-0000-0000-0000-000000000003",
			status: http.StatusUnauthorized,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := fixtures.Load()
			assert.NoError(t, err)

			targetURL := fmt.Sprintf("/api/v1/auth/users/%s", testCase.userID)

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

func TestUsersDelete(t *testing.T) {
	db, fixtures := database.PrepareTestDatabase(t, db.FixtureFS, db.MigrationsFS)
	r := TestingRouter(t, db)

	adminID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	adminToken, _ := auth.GenerateToken(adminID, "admin@example.com")

	customerID := uuid.MustParse("00000000-0000-0000-0000-000000000003")
	customerToken, _ := auth.GenerateToken(customerID, "customer@example.com")

	tests := []struct {
		name   string
		token  string
		userID string
		status int
	}{
		{
			name:   "ok",
			token:  adminToken,
			userID: "00000000-0000-0000-0000-000000000003",
			status: http.StatusNoContent,
		},
		{
			name:   "not-found",
			token:  adminToken,
			userID: "00000000-0000-0000-0000-999999999999",
			status: http.StatusNotFound,
		},
		{
			name:   "forbidden-customer",
			token:  customerToken,
			userID: "00000000-0000-0000-0000-000000000002",
			status: http.StatusForbidden,
		},
		{
			name:   "no-token",
			userID: "00000000-0000-0000-0000-000000000003",
			status: http.StatusUnauthorized,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			err := fixtures.Load()
			assert.NoError(t, err)

			targetURL := fmt.Sprintf("/api/v1/auth/users/%s", testCase.userID)

			req := xtesting.NewTestingRequest(t, targetURL, http.MethodDelete, nil)
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
