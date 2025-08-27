package handler

import (
	"net/http"
	"strconv"
	"time"
	"elk-stack-user/internal/model"
	"elk-stack-user/internal/service"
	"elk-stack-user/internal/logger"
	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService service.UserService
}

func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// CreateUser godoc
// @Summary Create a new user
// @Description Create a new user with the provided information
// @Tags users
// @Accept json
// @Produce json
// @Param user body model.CreateUserRequest true "User information"
// @Success 201 {object} model.UserResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /users [post]
func (h *UserHandler) CreateUser(c *gin.Context) {
	start := time.Now()
	requestID := c.GetString("request_id")
	
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		logger.Logger.Error("Invalid request body",
			logger.RequestID(requestID),
			logger.Error(err),
			logger.String("path", c.Request.URL.Path),
			logger.String("method", c.Request.Method),
		)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logger.Logger.Info("Creating user",
		logger.RequestID(requestID),
		logger.String("username", req.Username),
		logger.String("email", req.Email),
		logger.String("ip", c.ClientIP()),
	)

	user, err := h.userService.CreateUser(c.Request.Context(), &req)
	if err != nil {
		latency := time.Since(start)
		if err.Error() == "email already exists" || err.Error() == "username already exists" {
			logger.Logger.Warn("User creation failed - duplicate data",
				logger.RequestID(requestID),
				logger.String("username", req.Username),
				logger.String("email", req.Email),
				logger.String("error", err.Error()),
				logger.ResponseTime(latency),
			)
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		
		logger.Logger.Error("User creation failed",
			logger.RequestID(requestID),
			logger.String("username", req.Username),
			logger.String("email", req.Email),
			logger.Error(err),
			logger.ResponseTime(latency),
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}

	latency := time.Since(start)
	logger.Logger.Info("User created successfully",
		logger.RequestID(requestID),
		logger.UserID(user.ID),
		logger.String("username", user.Username),
		logger.String("email", user.Email),
		logger.ResponseTime(latency),
		logger.StatusCode(http.StatusCreated),
	)

	c.JSON(http.StatusCreated, user)
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Get user information by user ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 200 {object} model.UserResponse
// @Failure 404 {object} map[string]interface{}
// @Router /users/{id} [get]
func (h *UserHandler) GetUserByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	user, err := h.userService.GetUserByID(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetUserByEmail godoc
// @Summary Get user by email
// @Description Get user information by email address
// @Tags users
// @Accept json
// @Produce json
// @Param email query string true "User email"
// @Success 200 {object} model.UserResponse
// @Failure 404 {object} map[string]interface{}
// @Router /users/email [get]
func (h *UserHandler) GetUserByEmail(c *gin.Context) {
	email := c.Query("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email parameter is required"})
		return
	}

	user, err := h.userService.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetAllUsers godoc
// @Summary Get all users
// @Description Get paginated list of all users
// @Tags users
// @Accept json
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10)"
// @Success 200 {object} map[string]interface{}
// @Router /users [get]
func (h *UserHandler) GetAllUsers(c *gin.Context) {
	pageStr := c.DefaultQuery("page", "1")
	pageSizeStr := c.DefaultQuery("page_size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil {
		page = 1
	}

	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		pageSize = 10
	}

	users, total, err := h.userService.GetAllUsers(c.Request.Context(), page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch users"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"users": users,
		"pagination": gin.H{
			"page":       page,
			"page_size":  pageSize,
			"total":      total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

// UpdateUser godoc
// @Summary Update user
// @Description Update user information by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Param user body model.UpdateUserRequest true "User update information"
// @Success 200 {object} model.UserResponse
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 409 {object} map[string]interface{}
// @Router /users/{id} [put]
func (h *UserHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.userService.UpdateUser(c.Request.Context(), uint(id), &req)
	if err != nil {
		if err.Error() == "User not found" {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		if err.Error() == "email already exists" || err.Error() == "username already exists" {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUser godoc
// @Summary Delete user
// @Description Delete user by ID
// @Tags users
// @Accept json
// @Produce json
// @Param id path int true "User ID"
// @Success 204 "No Content"
// @Failure 400 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /users/{id} [delete]
func (h *UserHandler) DeleteUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	err = h.userService.DeleteUser(c.Request.Context(), uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}

	c.Status(http.StatusNoContent)
}
