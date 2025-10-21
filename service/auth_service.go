package service

import (
	"errors"
	"strings"

	"github.com/juancruzestevez/auth-service/auth"
	"github.com/juancruzestevez/auth-service/dto"
	"github.com/juancruzestevez/auth-service/models"
	"github.com/juancruzestevez/auth-service/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidInput       = errors.New("invalid input")
)

// AuthService maneja la lógica de negocio de autenticación
type AuthService interface {
	Register(req dto.RegisterRequest) (*dto.AuthResponse, error)
	Login(req dto.LoginRequest) (*dto.AuthResponse, error)
}

type authService struct {
	userRepo repository.UserRepository
}

// NewAuthService crea una nueva instancia del servicio de autenticación
func NewAuthService(userRepo repository.UserRepository) AuthService {
	return &authService{
		userRepo: userRepo,
	}
}

func (s *authService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	// Validar campos requeridos
	if err := s.validateRegisterRequest(req); err != nil {
		return nil, err
	}

	// Normalizar el email
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	// Verificar si el email ya existe
	existingUser, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Verificar si el nickname ya existe
	existingUser, err = s.userRepo.FindByNickname(req.Nickname)
	if err != nil {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// Hashear la contraseña
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Crear el usuario
	user := &models.User{
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Nickname:  req.Nickname,
		Email:     req.Email,
		Password:  string(hashedPassword),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	// Generar token JWT
	token, expiresAt, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	// Preparar respuesta
	return &dto.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: dto.UserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Nickname:  user.Nickname,
			Email:     user.Email,
		},
	}, nil
}

func (s *authService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	// Validar campos requeridos
	if req.Email == "" || req.Password == "" {
		return nil, ErrInvalidInput
	}

	// Normalizar el email
	req.Email = strings.TrimSpace(strings.ToLower(req.Email))

	// Buscar usuario por email
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidCredentials
	}

	// Verificar la contraseña
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// Generar token JWT
	token, expiresAt, err := auth.GenerateToken(user.ID, user.Email)
	if err != nil {
		return nil, err
	}

	// Preparar respuesta
	return &dto.AuthResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User: dto.UserResponse{
			ID:        user.ID,
			FirstName: user.FirstName,
			LastName:  user.LastName,
			Nickname:  user.Nickname,
			Email:     user.Email,
		},
	}, nil
}

func (s *authService) validateRegisterRequest(req dto.RegisterRequest) error {
	if req.Email == "" || req.Password == "" || req.FirstName == "" || req.LastName == "" || req.Nickname == "" {
		return ErrInvalidInput
	}
	return nil
}
