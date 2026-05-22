package tracing

import (
	"context"

	"github.com/semmidev/restful-template/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type AuthServiceDecorator struct {
	base   domain.AuthUsecase
	tracer trace.Tracer
}

func NewAuthServiceDecorator(base domain.AuthUsecase) *AuthServiceDecorator {
	return &AuthServiceDecorator{
		base:   base,
		tracer: otel.Tracer("auth_service"),
	}
}

func (d *AuthServiceDecorator) Register(ctx context.Context, in domain.RegisterInput) (domain.TokenPair, error) {
	ctx, span := d.tracer.Start(ctx, "AuthService.Register")
	defer span.End()

	res, err := d.base.Register(ctx, in)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return res, err
}

func (d *AuthServiceDecorator) Login(ctx context.Context, in domain.LoginInput) (domain.TokenPair, error) {
	ctx, span := d.tracer.Start(ctx, "AuthService.Login")
	defer span.End()

	res, err := d.base.Login(ctx, in)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return res, err
}

func (d *AuthServiceDecorator) Refresh(ctx context.Context, refreshToken string) (domain.TokenPair, error) {
	ctx, span := d.tracer.Start(ctx, "AuthService.Refresh")
	defer span.End()

	res, err := d.base.Refresh(ctx, refreshToken)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	}
	return res, err
}
