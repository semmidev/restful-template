package tracing

import (
	"context"

	"github.com/semmidev/restful-template/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type AuthDecorator struct {
	base   domain.AuthUsecase
	tracer trace.Tracer
}

func NewAuthDecorator(base domain.AuthUsecase) *AuthDecorator {
	return &AuthDecorator{
		base:   base,
		tracer: otel.Tracer("auth_service"),
	}
}

func (d *AuthDecorator) Register(ctx context.Context, in domain.RegisterInput) (domain.TokenPair, error) {
	ctx, span := d.tracer.Start(ctx, "Auth.Register")
	defer span.End()

	res, err := d.base.Register(ctx, in)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return res, err
}

func (d *AuthDecorator) Login(ctx context.Context, in domain.LoginInput) (domain.TokenPair, error) {
	ctx, span := d.tracer.Start(ctx, "Auth.Login")
	defer span.End()

	res, err := d.base.Login(ctx, in)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return res, err
}

func (d *AuthDecorator) Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error) {
	ctx, span := d.tracer.Start(ctx, "Auth.Refresh")
	defer span.End()

	res, err := d.base.Refresh(ctx, refreshToken)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return res, err
}
