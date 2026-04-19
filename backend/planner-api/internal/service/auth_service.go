package service

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"github.com/google/uuid"

	"planahead/planner-api/internal/repository"
)

type AuthPayload struct {
	Token       string
	StudentName string
	ExternalKey string
}

type AuthService struct {
	repo      *repository.StudentRepository
	jwtSecret []byte
}

func NewAuthService(repo *repository.StudentRepository, secret string) *AuthService {
	return &AuthService{
		repo:      repo,
		jwtSecret: []byte(secret),
	}
}

func (s *AuthService) Register(ctx context.Context, email, fullName, password string) (*AuthPayload, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	externalKey := uuid.New().String()
	student, err := s.repo.CreateStudent(ctx, externalKey, fullName, email, string(hashedPassword))
	if err != nil {
		return nil, err
	}

	token, err := s.generateToken(student)
	if err != nil {
		return nil, err
	}

	return &AuthPayload{
		Token:       token,
		StudentName: student.FullName,
		ExternalKey: student.ExternalKey,
	}, nil
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*AuthPayload, error) {
	student, err := s.repo.GetStudentByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(student.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	token, err := s.generateToken(student)
	if err != nil {
		return nil, err
	}

	return &AuthPayload{
		Token:       token,
		StudentName: student.FullName,
		ExternalKey: student.ExternalKey,
	}, nil
}

func (s *AuthService) generateToken(student *repository.PlannerStudent) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Subject:   student.ExternalKey,
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
	})
	return token.SignedString(s.jwtSecret)
}
