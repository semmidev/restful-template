package httpapi

import (
	"context"
	"errors"
	"net/http"
	"testing"

	apperrors "github.com/semmidev/restful-template/internal/shared/errors"
	. "github.com/smartystreets/goconvey/convey"
)

func TestToHumaErr(t *testing.T) {
	Convey("Given ToHumaErr function", t, func() {
		ctx := context.Background()

		Convey("When error is SafeError with code NOT_FOUND and nil internal", func() {
			err := apperrors.NewNotFound("User not found", nil)
			apiErr := ToHumaErr(ctx, err)
			var hErr *APIError
			So(errors.As(apiErr, &hErr), ShouldBeTrue)
			So(hErr.Status, ShouldEqual, http.StatusNotFound)
			So(hErr.Code, ShouldEqual, "NOT_FOUND")
			So(hErr.Detail, ShouldEqual, "User not found")
		})

		Convey("When error is SafeError with code FORBIDDEN and nil internal", func() {
			err := apperrors.NewForbidden("No access", nil)
			apiErr := ToHumaErr(ctx, err)
			var hErr *APIError
			So(errors.As(apiErr, &hErr), ShouldBeTrue)
			So(hErr.Status, ShouldEqual, http.StatusForbidden)
			So(hErr.Code, ShouldEqual, "FORBIDDEN")
			So(hErr.Detail, ShouldEqual, "No access")
		})

		Convey("When error is SafeError wrapping another sentinel error", func() {
			err := apperrors.NewNotFound("Todo not found", apperrors.ErrNotFound)
			apiErr := ToHumaErr(ctx, err)
			var hErr *APIError
			So(errors.As(apiErr, &hErr), ShouldBeTrue)
			So(hErr.Status, ShouldEqual, http.StatusNotFound)
			So(hErr.Code, ShouldEqual, "NOT_FOUND")
		})

		Convey("When error is raw apperrors.ErrConflict sentinel", func() {
			err := apperrors.ErrConflict
			apiErr := ToHumaErr(ctx, err)
			var hErr *APIError
			So(errors.As(apiErr, &hErr), ShouldBeTrue)
			So(hErr.Status, ShouldEqual, http.StatusConflict)
		})
	})
}
