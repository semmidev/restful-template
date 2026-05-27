package jwt

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/semmidev/restful-template/internal/modules/auth"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
)

type JWTService struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewJWTService(secret string, accessTTL, refreshTTL time.Duration) *JWTService {
	return &JWTService{secret: []byte(secret), accessTTL: accessTTL, refreshTTL: refreshTTL}
}

func (s *JWTService) GeneratePair(ctx context.Context, userID uuid.UUID, email string) (string, string, int64, int64, error) {
	now := time.Now()
	accessExp := now.Add(s.accessTTL).Unix()
	access := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"jti":   uuid.New().String(),
		"sub":   userID.String(),
		"email": email,
		"exp":   accessExp,
		"type":  "access",
	})
	aStr, err := access.SignedString(s.secret)
	if err != nil {
		return "", "", 0, 0, err
	}
	refreshExp := now.Add(s.refreshTTL).Unix()
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"jti":   uuid.New().String(),
		"sub":   userID.String(),
		"email": email,
		"exp":   refreshExp,
		"type":  "refresh",
	})
	rStr, err := refresh.SignedString(s.secret)
	return aStr, rStr, accessExp, refreshExp, err
}

func (s *JWTService) ParseAccess(ctx context.Context, token string) (*auth.TokenClaims, error) {
	return s.parse(token, "access")
}

func (s *JWTService) ParseRefresh(ctx context.Context, token string) (*auth.TokenClaims, error) {
	return s.parse(token, "refresh")
}

func (s *JWTService) parse(token, typ string) (*auth.TokenClaims, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) { return s.secret, nil })
	if err != nil || !parsed.Valid {
		return nil, err
	}
	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != typ {
		return nil, jwt.ErrTokenInvalidClaims
	}
	sub, ok1 := claims["sub"].(string)
	email, ok2 := claims["email"].(string)
	if !ok1 || !ok2 {
		return nil, apperrors.ErrUnauthorized
	}
	uid, _ := uuid.Parse(sub)
	return &auth.TokenClaims{UserID: uid, Email: email}, nil
}
