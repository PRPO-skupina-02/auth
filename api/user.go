package api

import (
	"net/http"

	"github.com/PRPO-skupina-02/auth/models"
	"github.com/PRPO-skupina-02/common/middleware"
	"github.com/PRPO-skupina-02/common/request"
	"github.com/gin-gonic/gin"
)

type AdminCreateUserRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"omitempty,min=1"`
	LastName  string `json:"last_name" binding:"omitempty,min=1"`
	Role      string `json:"role" binding:"required,oneof=customer employee admin"`
	Active    bool   `json:"active"`
}

// AdminCreateUser
//
//	@Id				AdminCreateUser
//	@Summary		Create user (admin)
//	@Description	Create a new user with any role (admin only)
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			request	body		AdminCreateUserRequest	true	"User creation details"
//	@Success		201		{object}	UserResponse
//	@Failure		400		{object}	middleware.HttpError
//	@Failure		401		{object}	middleware.HttpError
//	@Failure		403		{object}	middleware.HttpError
//	@Failure		409		{object}	middleware.HttpError
//	@Failure		500		{object}	middleware.HttpError
//	@Router			/users [post]
func AdminCreateUser(c *gin.Context) {
	tx := middleware.GetContextTransaction(c)

	var req AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	// Check if user already exists
	exists, err := models.UserExists(tx, req.Email)
	if err != nil {
		_ = c.Error(err)
		return
	}
	if exists {
		c.JSON(http.StatusConflict, gin.H{"error": "User with this email already exists"})
		return
	}

	// Create user
	user := models.User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Role:      models.UserRole(req.Role),
		Active:    req.Active,
	}

	if err := user.SetPassword(req.Password); err != nil {
		_ = c.Error(err)
		return
	}

	if err := user.Create(tx); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusCreated, newUserResponse(user))
}

// UsersList
//
//	@Id				UsersList
//	@Summary		List users
//	@Description	List all users (admin endpoint)
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			limit	query		int		false	"Limit the number of responses"	Default(10)
//	@Param			offset	query		int		false	"Offset the first response"		Default(0)
//	@Param			sort	query		string	false	"Sort results"
//	@Success		200		{object}	request.PaginatedResponse{data=[]UserResponse}
//	@Failure		400		{object}	middleware.HttpError
//	@Failure		401		{object}	middleware.HttpError
//	@Failure		500		{object}	middleware.HttpError
//	@Router			/users [get]
func UsersList(c *gin.Context) {
	tx := middleware.GetContextTransaction(c)
	pagination := request.GetNormalizedPaginationArgs(c)
	sort := request.GetSortOptions(c)

	users, total, err := models.GetUsers(tx, pagination, sort)
	if err != nil {
		_ = c.Error(err)
		return
	}

	response := []UserResponse{}
	for _, user := range users {
		response = append(response, newUserResponse(user))
	}

	request.RenderPaginatedResponse(c, response, int(total))
}

// UsersShow
//
//	@Id				UsersShow
//	@Summary		Get user by ID
//	@Description	Get a specific user by ID (admin endpoint)
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			userID	path		string	true	"User ID"
//	@Success		200		{object}	UserResponse
//	@Failure		400		{object}	middleware.HttpError
//	@Failure		401		{object}	middleware.HttpError
//	@Failure		404		{object}	middleware.HttpError
//	@Failure		500		{object}	middleware.HttpError
//	@Router			/users/{userID} [get]
func UsersShow(c *gin.Context) {
	tx := middleware.GetContextTransaction(c)

	userID, err := request.GetUUIDParam(c, "userID")
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := models.GetUser(tx, userID)
	if err != nil {
		_ = c.AbortWithError(http.StatusNotFound, err)
		return
	}

	c.JSON(http.StatusOK, newUserResponse(user))
}

type AdminUpdateUserRequest struct {
	FirstName *string `json:"first_name" binding:"omitempty,min=1"`
	LastName  *string `json:"last_name" binding:"omitempty,min=1"`
	Active    *bool   `json:"active" binding:"omitempty"`
}

// UsersUpdate
//
//	@Id				UsersUpdate
//	@Summary		Update user
//	@Description	Update a specific user (admin endpoint)
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			userID	path		string					true	"User ID"
//	@Param			request	body		AdminUpdateUserRequest	true	"User update details"
//	@Success		200		{object}	UserResponse
//	@Failure		400		{object}	middleware.HttpError
//	@Failure		401		{object}	middleware.HttpError
//	@Failure		404		{object}	middleware.HttpError
//	@Failure		500		{object}	middleware.HttpError
//	@Router			/users/{userID} [put]
func UsersUpdate(c *gin.Context) {
	tx := middleware.GetContextTransaction(c)

	userID, err := request.GetUUIDParam(c, "userID")
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	var req AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(err)
		return
	}

	user, err := models.GetUser(tx, userID)
	if err != nil {
		_ = c.AbortWithError(http.StatusNotFound, err)
		return
	}

	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Active != nil {
		user.Active = *req.Active
	}

	if err := user.Save(tx); err != nil {
		_ = c.Error(err)
		return
	}

	c.JSON(http.StatusOK, newUserResponse(user))
}

// UsersDelete
//
//	@Id				UsersDelete
//	@Summary		Delete user
//	@Description	Delete a specific user (admin endpoint)
//	@Tags			users
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			userID	path		string	true	"User ID"
//	@Success		204		{object}	nil
//	@Failure		400		{object}	middleware.HttpError
//	@Failure		401		{object}	middleware.HttpError
//	@Failure		404		{object}	middleware.HttpError
//	@Failure		500		{object}	middleware.HttpError
//	@Router			/users/{userID} [delete]
func UsersDelete(c *gin.Context) {
	tx := middleware.GetContextTransaction(c)

	userID, err := request.GetUUIDParam(c, "userID")
	if err != nil {
		_ = c.AbortWithError(http.StatusBadRequest, err)
		return
	}

	user, err := models.GetUser(tx, userID)
	if err != nil {
		_ = c.AbortWithError(http.StatusNotFound, err)
		return
	}

	if err := user.Delete(tx); err != nil {
		_ = c.Error(err)
		return
	}

	c.Status(http.StatusNoContent)
}
