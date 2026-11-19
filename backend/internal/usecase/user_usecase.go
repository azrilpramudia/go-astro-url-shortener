package usecase

import (
	"context"
	"errors"
	"strings"

	"go-astro-url-shortener/internal/entity"
	"go-astro-url-shortener/internal/model"
	"go-astro-url-shortener/internal/model/converter"
	"go-astro-url-shortener/internal/repository"
	"go-astro-url-shortener/internal/util"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthResult struct {
	User  *model.UserResponse
	Token string
}

type UserUseCase struct {
	DB             *gorm.DB
	Log            *logrus.Logger
	Validate       *validator.Validate
	UserRepository *repository.UserRepository
	Tokens         *util.TokenManager
	dummyHash      []byte
}

func NewUserUseCase(db *gorm.DB, log *logrus.Logger, validate *validator.Validate,
	userRepository *repository.UserRepository, tokens *util.TokenManager) *UserUseCase {
	dummy, _ := bcrypt.GenerateFromPassword([]byte("dummy-password"), bcrypt.DefaultCost)
	return &UserUseCase{
		DB: db, Log: log, Validate: validate,
		UserRepository: userRepository, Tokens: tokens, dummyHash: dummy,
	}
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func (c *UserUseCase) Register(ctx context.Context, request *model.RegisterRequest) (*AuthResult, error) {
	if err := c.Validate.Struct(request); err != nil {
		c.Log.Warnf("invalid register request: %+v", err)
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid request: name, a valid email, and a password of 8-72 characters are required")
	}
	// bcrypt hanya memproses 72 byte pertama; validator menghitung karakter, bukan byte
	if len(request.Password) > 72 {
		return nil, fiber.NewError(fiber.StatusBadRequest, "password must be at most 72 bytes")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(request.Password), bcrypt.DefaultCost)
	if err != nil {
		c.Log.Warnf("failed to hash password: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	user := &entity.User{
		Name:         strings.TrimSpace(request.Name),
		Email:        normalizeEmail(request.Email),
		PasswordHash: string(hash),
	}

	if err := c.UserRepository.Create(c.DB.WithContext(ctx), user); err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, fiber.NewError(fiber.StatusConflict, "email already registered")
		}
		c.Log.Warnf("failed to create user: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	token, err := c.Tokens.Generate(user.ID)
	if err != nil {
		c.Log.Warnf("failed to generate token: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	return &AuthResult{User: converter.UserToResponse(user), Token: token}, nil
}

func (c *UserUseCase) Login(ctx context.Context, request *model.LoginRequest) (*AuthResult, error) {
	if err := c.Validate.Struct(request); err != nil {
		return nil, fiber.NewError(fiber.StatusBadRequest, "invalid request: email and password are required")
	}

	invalid := fiber.NewError(fiber.StatusUnauthorized, "invalid email or password")

	user := new(entity.User)
	err := c.UserRepository.FindByEmail(c.DB.WithContext(ctx), normalizeEmail(request.Email), user)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		_ = bcrypt.CompareHashAndPassword(c.dummyHash, []byte(request.Password))
		return nil, invalid
	}
	if err != nil {
		c.Log.Warnf("failed to find user: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(request.Password)) != nil {
		return nil, invalid
	}

	token, err := c.Tokens.Generate(user.ID)
	if err != nil {
		c.Log.Warnf("failed to generate token: %+v", err)
		return nil, fiber.ErrInternalServerError
	}

	return &AuthResult{User: converter.UserToResponse(user), Token: token}, nil
}

func (c *UserUseCase) Current(ctx context.Context, userID uint) (*model.UserResponse, error) {
	user := new(entity.User)
	if err := c.UserRepository.FindByID(c.DB.WithContext(ctx), userID, user); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		c.Log.Warnf("failed to find user: %+v", err)
		return nil, fiber.ErrInternalServerError
	}
	return converter.UserToResponse(user), nil
}
