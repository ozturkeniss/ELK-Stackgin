package repository

import (
	"context"
	"gorm.io/gorm"
	"elk-stack-user/internal/model"
	"time"
)

type UserRepository interface {
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id uint) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetByUsername(ctx context.Context, username string) (*model.User, error)
	GetAll(ctx context.Context, limit, offset int) ([]*model.User, error)
	Update(ctx context.Context, user *model.User) error
	Delete(ctx context.Context, id uint) error
	Count(ctx context.Context) (int64, error)
	// Login methods
	GetUserForLogin(ctx context.Context, usernameOrEmail string) (*model.User, error)
	RecordLoginAttempt(ctx context.Context, attempt *model.LoginAttempt) error
	GetFailedAttempts(ctx context.Context, username, ipAddress string, since time.Time) ([]*model.LoginAttempt, error)
	RecordBan(ctx context.Context, ban *model.BanRecord) error
	IsBanned(ctx context.Context, username, ipAddress string) (*model.BanRecord, error)
	RemoveExpiredBans(ctx context.Context) error
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) Create(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) GetByID(ctx context.Context, id uint) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetAll(ctx context.Context, limit, offset int) ([]*model.User, error) {
	var users []*model.User
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&model.User{}, id).Error
}

func (r *userRepository) Count(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.User{}).Count(&count).Error
	return count, err
}

func (r *userRepository) GetUserForLogin(ctx context.Context, usernameOrEmail string) (*model.User, error) {
	var user model.User
	err := r.db.WithContext(ctx).Where("username = ? OR email = ?", usernameOrEmail, usernameOrEmail).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) RecordLoginAttempt(ctx context.Context, attempt *model.LoginAttempt) error {
	return r.db.WithContext(ctx).Create(attempt).Error
}

func (r *userRepository) GetFailedAttempts(ctx context.Context, username, ipAddress string, since time.Time) ([]*model.LoginAttempt, error) {
	var attempts []*model.LoginAttempt
	err := r.db.WithContext(ctx).
		Where("(username = ? OR ip_address = ?) AND success = ? AND timestamp > ?", 
			username, ipAddress, false, since).
		Find(&attempts).Error
	return attempts, err
}

func (r *userRepository) RecordBan(ctx context.Context, ban *model.BanRecord) error {
	return r.db.WithContext(ctx).Create(ban).Error
}

func (r *userRepository) IsBanned(ctx context.Context, username, ipAddress string) (*model.BanRecord, error) {
	var ban model.BanRecord
	err := r.db.WithContext(ctx).
		Where("(username = ? OR ip_address = ?) AND expires_at > ?", 
			username, ipAddress, time.Now()).
		First(&ban).Error
	if err != nil {
		return nil, err
	}
	return &ban, nil
}

func (r *userRepository) RemoveExpiredBans(ctx context.Context) error {
	return r.db.WithContext(ctx).Where("expires_at <= ?", time.Now()).Delete(&model.BanRecord{}).Error
}
