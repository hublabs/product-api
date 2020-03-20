package controllers

import (
	"errors"
	"net/http"

	"github.com/hublabs/common/api"
	"github.com/hublabs/product-api/factory"
	"github.com/pangpanglabs/goutils/behaviorlog"

	"github.com/go-xorm/xorm"
	"github.com/labstack/echo"
)

func renderFail(c echo.Context, err error) error {
	if err == nil {
		err = api.ErrorUnknown.New(nil)
	}
	behaviorlog.FromCtx(c.Request().Context()).WithError(err)
	var apiError *api.Error
	if ok := errors.As(err, &apiError); ok {
		return c.JSON(apiError.Status(), api.Result{
			Success: false,
			Error:   *apiError,
		})
	}
	return err
}

func renderSuccArray(c echo.Context, withHasMore, hasMore bool, totalCount int64, result interface{}) error {
	if withHasMore {
		return renderSucc(c, http.StatusOK, api.ArrayResultMore{
			HasMore: hasMore,
			Items:   result,
		})
	} else {
		return renderSucc(c, http.StatusOK, api.ArrayResult{
			TotalCount: totalCount,
			Items:      result,
		})
	}
}

func renderSucc(c echo.Context, status int, result interface{}) error {
	req := c.Request()
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "DELETE" {
		if session, ok := factory.DB(req.Context()).(*xorm.Session); ok {
			if err := session.Commit(); err != nil {
				return api.ErrorDB.New(err)
			}
		}
	}

	return c.JSON(status, api.Result{
		Success: true,
		Result:  result,
	})
}
