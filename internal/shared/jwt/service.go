package jwt

import (
	"context"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
)

type TokenClaims struct {
	UserID     uuid.UUID
	Email      string
	ActiveRole string
	Roles      []string
}

type JWTService struct {
	secret     []byte
	accessTTL  time.Duration
	refreshTTL time.Duration
	issuer     string
	audience   string
}

// NewJWTService constructs a JWTService.
// issuer and audience are embedded as `iss` / `aud` claims and validated on
// every parse to prevent token substitution attacks across environments.
func NewJWTService(secret string, accessTTL, refreshTTL time.Duration, issuer, audience string) *JWTService {
	return &JWTService{
		secret:     []byte(secret),
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
		issuer:     issuer,
		audience:   audience,
	}
}

func (s *JWTService) GeneratePair(ctx context.Context, userID uuid.UUID, email string, activeRole string, roles []string) (string, string, int64, int64, error) {
	now := time.Now()
	accessExp := now.Add(s.accessTTL).Unix()

	access := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"jti":         uuid.New().String(),
		"sub":         userID.String(),
		"email":       email,
		"active_role": activeRole,
		"roles":       roles,
		"iss":         s.issuer,
		"aud":         s.audience,
		"exp":         accessExp,
		"type":        "access",
	})
	aStr, err := access.SignedString(s.secret)
	if err != nil {
		return "", "", 0, 0, err
	}

	refreshExp := now.Add(s.refreshTTL).Unix()
	refresh := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"jti":         uuid.New().String(),
		"sub":         userID.String(),
		"email":       email,
		"active_role": activeRole,
		"roles":       roles,
		"iss":         s.issuer,
		"aud":         s.audience,
		"exp":         refreshExp,
		"type":        "refresh",
	})
	rStr, err := refresh.SignedString(s.secret)
	return aStr, rStr, accessExp, refreshExp, err
}

func (s *JWTService) ParseAccess(ctx context.Context, token string) (*TokenClaims, error) {
	return s.parse(token, "access")
}

func (s *JWTService) ParseRefresh(ctx context.Context, token string) (*TokenClaims, error) {
	return s.parse(token, "refresh")
}

func (s *JWTService) parse(token, typ string) (*TokenClaims, error) {
	parsed, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return s.secret, nil
	})
	if err != nil || !parsed.Valid {
		return nil, apperrors.ErrUnauthorized
	}

	claims, ok := parsed.Claims.(jwt.MapClaims)
	if !ok {
		return nil, apperrors.ErrUnauthorized
	}

	// validate token type, issuer, and audience to prevent token substitution
	// across environments and between access/refresh endpoints.
	if claims["type"] != typ {
		return nil, apperrors.ErrUnauthorized
	}

	if iss, _ := claims["iss"].(string); iss != s.issuer {
		return nil, apperrors.ErrUnauthorized
	}

	if aud, _ := claims["aud"].(string); aud != s.audience {
		return nil, apperrors.ErrUnauthorized
	}

	sub, ok1 := claims["sub"].(string)
	email, ok2 := claims["email"].(string)
	if !ok1 || !ok2 {
		return nil, apperrors.ErrUnauthorized
	}

	uid, err := uuid.Parse(sub)
	if err != nil {
		return nil, apperrors.ErrUnauthorized
	}

	activeRole, _ := claims["active_role"].(string)
	var roles []string
	if rawRoles, ok := claims["roles"].([]interface{}); ok {
		for _, r := range rawRoles {
			if rStr, ok := r.(string); ok {
				roles = append(roles, rStr)
			}
		}
	}

	return &TokenClaims{
		UserID:     uid,
		Email:      email,
		ActiveRole: activeRole,
		Roles:      roles,
	}, nil
}
