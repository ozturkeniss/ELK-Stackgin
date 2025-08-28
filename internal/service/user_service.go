package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"elk-stack-user/internal/model"
	"elk-stack-user/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"time"
)

type UserService interface {
	CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.UserResponse, error)
	GetUserByID(ctx context.Context, id uint) (*model.UserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*model.UserResponse, error)
	GetAllUsers(ctx context.Context, page, pageSize int) ([]*model.UserResponse, int64, error)
	UpdateUser(ctx context.Context, id uint, req *model.UpdateUserRequest) (*model.UserResponse, error)
	DeleteUser(ctx context.Context, id uint) error
	// Login methods
	Login(ctx context.Context, req *model.LoginRequest, ipAddress, userAgent string) (*model.LoginResponse, error)
	IsUserBanned(ctx context.Context, username, ipAddress string) (*model.BanRecord, error)
}

type userService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) UserService {
	return &userService{userRepo: userRepo}
}

func (s *userService) CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.UserResponse, error) {
	// Email ve username kontrolü
	if _, err := s.userRepo.GetByEmail(ctx, req.Email); err == nil {
		return nil, errors.New("email already exists")
	}
	if _, err := s.userRepo.GetByUsername(ctx, req.Username); err == nil {
		return nil, errors.New("username already exists")
	}

	// Password hash'leme
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Age:       req.Age,
		IsActive:  true,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return s.toUserResponse(user), nil
}

func (s *userService) GetUserByID(ctx context.Context, id uint) (*model.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return s.toUserResponse(user), nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*model.UserResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	return s.toUserResponse(user), nil
}

func (s *userService) GetAllUsers(ctx context.Context, page, pageSize int) ([]*model.UserResponse, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize
	users, err := s.userRepo.GetAll(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}

	total, err := s.userRepo.Count(ctx)
	if err != nil {
		return nil, 0, err
	}

	userResponses := make([]*model.UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = s.toUserResponse(user)
	}

	return userResponses, total, nil
}

func (s *userService) UpdateUser(ctx context.Context, id uint, req *model.UpdateUserRequest) (*model.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Email değişikliği varsa kontrol et
	if req.Email != nil && *req.Email != user.Email {
		if _, err := s.userRepo.GetByEmail(ctx, *req.Email); err == nil {
			return nil, errors.New("email already exists")
		}
		user.Email = *req.Email
	}

	// Username değişikliği varsa kontrol et
	if req.Username != nil && *req.Username != user.Username {
		if _, err := s.userRepo.GetByUsername(ctx, *req.Username); err == nil {
			return nil, errors.New("username already exists")
		}
		user.Username = *req.Username
	}

	// Diğer alanları güncelle
	if req.FirstName != nil {
		user.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		user.LastName = *req.LastName
	}
	if req.Age != nil {
		user.Age = *req.Age
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.userRepo.Update(ctx, user); err != nil {
		return nil, err
	}

	return s.toUserResponse(user), nil
}

func (s *userService) DeleteUser(ctx context.Context, id uint) error {
	return s.userRepo.Delete(ctx, id)
}

func (s *userService) toUserResponse(user *model.User) *model.UserResponse {
	return &model.UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		FirstName: user.FirstName,
		LastName:  user.LastName,
		Age:       user.Age,
		IsActive:  user.IsActive,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
}

// Utility function for generating random strings
func generateRandomString(length int) string {
	bytes := make([]byte, length/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)
}

// Login method with ban system
func (s *userService) Login(ctx context.Context, req *model.LoginRequest, ipAddress, userAgent string) (*model.LoginResponse, error) {
	// Check if user is banned
	if ban, err := s.IsUserBanned(ctx, req.Username, ipAddress); err == nil && ban != nil {
		return nil, errors.New("account temporarily banned due to multiple failed login attempts")
	}

	// Get user for login
	user, err := s.userRepo.GetUserForLogin(ctx, req.Username)
	if err != nil {
		// Record failed attempt
		s.recordFailedAttempt(ctx, req.Username, ipAddress, userAgent)
		return nil, errors.New("invalid credentials")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		// Record failed attempt
		s.recordFailedAttempt(ctx, req.Username, ipAddress, userAgent)
		
		// Check if we should ban the user
		if s.shouldBanUser(ctx, req.Username, ipAddress) {
			s.banUser(ctx, req.Username, ipAddress, "Multiple failed login attempts")
		}
		
		return nil, errors.New("invalid credentials")
	}

	// Record successful attempt
	s.recordSuccessfulAttempt(ctx, req.Username, ipAddress, userAgent)

	// Generate simple token (in production, use JWT)
	token := generateRandomString(32)

	return &model.LoginResponse{
		User:  *s.toUserResponse(user),
		Token: token,
	}, nil
}

func (s *userService) IsUserBanned(ctx context.Context, username, ipAddress string) (*model.BanRecord, error) {
	return s.userRepo.IsBanned(ctx, username, ipAddress)
}

// Helper methods for ban system
func (s *userService) recordFailedAttempt(ctx context.Context, username, ipAddress, userAgent string) {
	attempt := &model.LoginAttempt{
		Username:  username,
		IPAddress: ipAddress,
		Success:   false,
		Timestamp: time.Now(),
		UserAgent: userAgent,
	}
	s.userRepo.RecordLoginAttempt(ctx, attempt)
}

func (s *userService) recordSuccessfulAttempt(ctx context.Context, username, ipAddress, userAgent string) {
	attempt := &model.LoginAttempt{
		Username:  username,
		IPAddress: ipAddress,
		Success:   true,
		Timestamp: time.Now(),
		UserAgent: userAgent,
	}
	s.userRepo.RecordLoginAttempt(ctx, attempt)
}

func (s *userService) shouldBanUser(ctx context.Context, username, ipAddress string) bool {
	// Check failed attempts in the last 2 minutes
	since := time.Now().Add(-2 * time.Minute)
	failedAttempts, err := s.userRepo.GetFailedAttempts(ctx, username, ipAddress, since)
	if err != nil {
		return false
	}
	
	// Ban if 3 or more failed attempts
	return len(failedAttempts) >= 3
}

func (s *userService) banUser(ctx context.Context, username, ipAddress, reason string) {
	ban := &model.BanRecord{
		Username:  username,
		IPAddress: ipAddress,
		BannedAt:  time.Now(),
		ExpiresAt: time.Now().Add(2 * time.Minute), // 2 minute ban
		Reason:    reason,
	}
	s.userRepo.RecordBan(ctx, ban)
}
