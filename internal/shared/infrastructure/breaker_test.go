package infrastructure

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/redis/go-redis/v9"
	. "github.com/smartystreets/goconvey/convey"
)

func TestBreakerErrorSuccess(t *testing.T) {
	Convey("Given isDBErrorSuccess function", t, func() {
		Convey("When error is nil", func() {
			So(isDBErrorSuccess(nil), ShouldBeTrue)
		})

		Convey("When error is pgx.ErrNoRows", func() {
			So(isDBErrorSuccess(pgx.ErrNoRows), ShouldBeTrue)
		})

		Convey("When error is context.Canceled", func() {
			So(isDBErrorSuccess(context.Canceled), ShouldBeTrue)
		})

		Convey("When error is context.DeadlineExceeded", func() {
			So(isDBErrorSuccess(context.DeadlineExceeded), ShouldBeFalse)
		})

		Convey("When error is a general pgconn.PgError (e.g. Unique Violation)", func() {
			pgErr := &pgconn.PgError{
				Code: "23505", // unique_violation
			}
			So(isDBErrorSuccess(pgErr), ShouldBeTrue)
		})

		Convey("When error is a system pgconn.PgError (e.g. Class 57 admin_shutdown)", func() {
			pgErr := &pgconn.PgError{
				Code: "57P01",
			}
			So(isDBErrorSuccess(pgErr), ShouldBeFalse)
		})

		Convey("When error is a generic network/connection error", func() {
			err := errors.New("connection reset by peer")
			So(isDBErrorSuccess(err), ShouldBeFalse)
		})
	})

	Convey("Given isRedisErrorSuccess function", t, func() {
		Convey("When error is nil", func() {
			So(isRedisErrorSuccess(nil), ShouldBeTrue)
		})

		Convey("When error is redis.Nil", func() {
			So(isRedisErrorSuccess(redis.Nil), ShouldBeTrue)
		})

		Convey("When error is context.Canceled", func() {
			So(isRedisErrorSuccess(context.Canceled), ShouldBeTrue)
		})

		Convey("When error is generic redis error", func() {
			err := errors.New("redis connection timeout")
			So(isRedisErrorSuccess(err), ShouldBeFalse)
		})
	})
}
