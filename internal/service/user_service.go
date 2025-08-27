package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"elk-stack-user/internal/model"
	"elk-stack-user/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	CreateUser(ctx context.Context, req *model.CreateUserRequest) (*model.UserResponse, error)
	GetUserByID(ctx context.Context, id uint) (*model.UserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*model.UserResponse, error)
	GetAllUsers(ctx context.Context, page, pageSize int) ([]*model.UserResponse, int64, error)
	UpdateUser(ctx context.Context, id uint, req *model.UpdateUserRequest) (*model.UserResponse, error)
	DeleteUser(ctx context.Context, id uint) error
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
