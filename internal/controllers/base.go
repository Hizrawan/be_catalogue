package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"

	"be20250107/internal/app"
	"be20250107/internal/errors"
	"be20250107/internal/middlewares"
	"be20250107/internal/reqdata"
	"be20250107/utils/reftype"
)

// Controller is the base struct for other controllers. This struct implements
// several helper methods to assist in controllers implementation as well as
// provide central access to the app Registry.
type Controller struct {
	App *app.Registry
}

// Parse decodes request body to the provided Form instance according to the
// content type. It will return nil if the Form can be properly decoded or that
// it is empty.
func (c *Controller) Parse(reqPtr reqdata.Form, r *http.Request) error {
	//	This ensures empty body doesn't break the parsing process
	if r.ContentLength == 0 {
		return nil
	}

	if strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data") {
		err := ParseMultipartFormData(reqPtr, r)
		if err != nil {
			return errors.ErrMalformedRequest
		}
	} else {
		err := json.NewDecoder(r.Body).Decode(reqPtr)
		if err != nil {
			return errors.ErrMalformedRequest
		}
	}
	return nil
}

// Validate decodes request body to the provided Form instance and performs
// authorization and validation check. It will return nil if the Form can be
// decoded and pass both checks, and will error if at least one of the checks
// fails.
func (c *Controller) Validate(reqPtr reqdata.Form, r *http.Request) error {
	err := c.Parse(reqPtr, r)
	if err != nil {
		return err
	}

	ctx := c.RequestContext(r)
	if ok := reqPtr.Authorized(ctx); !ok {
		panic(errors.ErrForbidden)
	}
	if err = reqPtr.Validate(ctx); err != nil {
		return err
	}
	return nil
}

// AssertAuthenticated checks if the request has valid user authentication info.
// If the user is logged out, the function will panic with
// errors.ErrUnauthenticated error.
func (c *Controller) AssertAuthenticated(r *http.Request) reqdata.AuthInformation {
	ctx := c.RequestContext(r)
	if !ctx.Auth.IsLoggedIn() {
		panic(errors.ErrUnauthenticated)
	}
	return ctx.Auth
}

// AssertAuthorized checks if the request has user authentication information and
// whether the user is authorized to access the resource. If the user is not
// logged in, the function will panic with HTTP ErrUnauthenticated error. If the
// user doesn't pass the checkFn provided for authorization, the function will
// panic with errors.ErrForbidden error.
func (c *Controller) AssertAuthorized(checkFn func(ctx *reqdata.Context) bool, r *http.Request) reqdata.AuthInformation {
	ctx := c.RequestContext(r)
	if !ctx.Auth.IsLoggedIn() {
		panic(errors.ErrUnauthenticated)
	}
	if !checkFn(ctx) {
		panic(errors.ErrForbidden)
	}
	return ctx.Auth
}

// parseMultipartFormData parses a request body encoded as multipart/form-data
// into a provided struct
func ParseMultipartFormData(reqPtr reqdata.Form, r *http.Request) error {
	t := reflect.TypeOf(reqPtr)
	if t.Kind() != reflect.Ptr {
		return fmt.Errorf("request must be a pointer, received: %s", t.Kind())
	}

	err := r.ParseMultipartForm(8 << 20)
	if err != nil {
		return err
	}

	t = t.Elem()
	req := reflect.ValueOf(reqPtr).Elem()
	tagName := "json"
	upFile := reqdata.UploadedFile{}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		elemType := field.Type
		if elemType.Kind() == reflect.Slice {
			elemType = elemType.Elem()
		}

		name := field.Tag.Get(tagName)

		isSlice := field.Type.Kind() == reflect.Slice
		isUploadedFile := reftype.IsTypeOf(elemType, upFile)
		embedsUploadedFile := reftype.IsStructEmbeds(elemType, upFile)

		if isUploadedFile || embedsUploadedFile {
			files := r.MultipartForm.File[name]
			cv := reflect.MakeSlice(reflect.SliceOf(elemType), len(files), len(files))

			for fIdx, f := range files {
				fHndl, err := f.Open()
				if err != nil {
					return err
				}
				upf := reqdata.UploadedFile{
					Filename:    f.Filename,
					Size:        f.Size,
					ContentType: f.Header.Get("Content-Type"),
					File:        fHndl,
				}
				var val reflect.Value

				if embedsUploadedFile {
					upfWrapperPtr := reflect.New(elemType)
					upfWrapper := upfWrapperPtr.Elem()
					upfWrapperType := upfWrapper.Type()
					for fldIdx := 0; fldIdx < upfWrapper.NumField(); fldIdx++ {
						fld := upfWrapper.Field(fldIdx)
						if upfWrapperType.Field(fldIdx).Anonymous && reftype.IsTypeOf(fld.Type(), upf) {
							fld.Set(reflect.ValueOf(upf))
						}
					}
					val = upfWrapper
				} else {
					val = reflect.ValueOf(upf)
				}

				cv.Index(fIdx).Set(val)
				if !isSlice {
					break
				}
			}

			if cv.Len() > 0 {
				if isSlice {
					req.Field(i).Set(cv)
				} else {
					req.Field(i).Set(cv.Index(0))
				}
			}
		} else {
			values := r.MultipartForm.Value[name]
			cv := reflect.MakeSlice(reflect.SliceOf(elemType), len(values), len(values))

			for i, v := range values {
				switch elemType.Kind() {
				case reflect.String:
					cv.Index(i).Set(reflect.ValueOf(v))
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					iVal, err := strconv.ParseInt(v, 10, 64)
					if err != nil {
						return err
					}
					if reflect.TypeOf(iVal).ConvertibleTo(elemType) {
						iType := reflect.ValueOf(iVal).Convert(elemType)
						cv.Index(i).Set(iType)
					} else {
						return fmt.Errorf("value cannot be converted")
					}
				case reflect.Float32, reflect.Float64:
					fVal, err := strconv.ParseFloat(v, 64)
					if err != nil {
						return err
					}
					if reflect.TypeOf(fVal).ConvertibleTo(elemType) {
						fType := reflect.ValueOf(fVal).Convert(elemType)
						cv.Index(i).Set(fType)
					} else {
						return fmt.Errorf("value cannot be converted")
					}
				case reflect.Bool:
					bVal, err := strconv.ParseBool(v)
					if err != nil {
						return err
					}
					cv.Index(i).Set(reflect.ValueOf(bVal))
				}

				if !isSlice {
					break
				}
			}

			if cv.Len() > 0 {
				if isSlice {
					req.Field(i).Set(cv)
				} else {
					req.Field(i).Set(cv.Index(0))
				}
			}
		}
	}
	return nil
}

// RequestContext retrieves request-related data and information from the Context
// of the provided Request and returns a RequestContext object for easier access
func (c *Controller) RequestContext(r *http.Request) *reqdata.Context {
	var auth reqdata.AuthInformation
	if a, ok := r.Context().Value(middlewares.ContextAuth).(reqdata.AuthInformation); ok {
		auth = a
	}

	return &reqdata.Context{
		App:         c.App,
		Auth:        auth,
		HTTPContext: r.Context(),
	}
}

// PageMeta represents the pagination information of a dataset
type PageMeta struct {
	Page      int64 `json:"page"`
	TotalPage int64 `json:"total_page"`
	Total     int64 `json:"total"`
	Size      int64 `json:"size"`
	PerPage   int64 `json:"per_page"`
}

// getPaginationFrom retrieves the page and limit from the query parameter of the
// provided HTTP request to be used for pagination.
func (c *Controller) getPaginationFrom(r *http.Request) (page int, limit int) {
	pageQuery := r.URL.Query().Get("page")
	limitQuery := r.URL.Query().Get("per_page")

	page = 1
	limit = 50
	if pageQuery != "" {
		if p, err := strconv.Atoi(pageQuery); err == nil && p > 0 {
			page = p
		}
	}
	if limitQuery != "" {
		if l, err := strconv.Atoi(limitQuery); err == nil && l > 0 {
			limit = l
		}
	}
	return page, limit
}

//// Paginate will paginate a database query based on the pagination parameter
//// contained in the HTTP request object. It will then return the pagination
//// information as well as placing the paginated result in the provided pointer.
//func (c *Controller) Paginate(r *http.Request, query database.Queryer, outPtr any) (PageMeta, error) {
//	page, limit := c.getPaginationFrom(r)
//
//	return c.PaginateBy(page, limit, query, outPtr)
//}

//// PaginateBy receives a database query and paginates the result based on the
//// provided size and page number. The result will then be placed in the provided
//// pointer as well as returning the pagination information.
//func (c *Controller) PaginateBy(page int, limit int, query database.Queryer, outPtr any) (PageMeta, error) {
//	pMeta := PageMeta{
//		Page:    int64(page),
//		PerPage: int64(limit),
//	}
//
//	if pMeta.PerPage > 500 {
//		pMeta.PerPage = 500
//	}
//
//	res := query.Count(&pMeta.Total)
//	if res.Error != nil {
//		return pMeta, res.Error
//	}
//
//	pMeta.TotalPage = pMeta.Total / pMeta.PerPage
//	if pMeta.Total%pMeta.PerPage != 0 {
//		pMeta.TotalPage += 1
//	}
//
//	res = query.Offset((page - 1) * limit).Limit(limit).Find(outPtr)
//	if res.Error != nil {
//		return pMeta, res.Error
//	}
//	pMeta.Size = res.RowsAffected
//
//	return pMeta, nil
//}

//// FastPaginate will paginate a database query using the deferred join method for
//// quicker processing. It will return pagination information as well as placing
//// the paginated result in the provided pointer, similar in output to the
//// Paginate function.
//func (c Controller) FastPaginate(r *http.Request, fqpk string, query database.Queryer, outPtr any) (PageMeta, error) {
//	page, limit := c.getPaginationFrom(r)
//
//	return c.FastPaginateBy(fqpk, r.Context(), page, limit, query, outPtr)
//}

//// FastPaginateBy receives a database query and paginates the result based on the
//// provided size and page number using the deferred join method. The deferred
//// join requires a fully-qualified column name to be used as the primary key for
//// joining the filter with the actual data. The result will then be placed in the
//// provided pointer and the pagination information will be returned.
//func (c *Controller) FastPaginateBy(fqpk string, ctx context.Context, page int, limit int, query database.Queryer, outPtr any) (PageMeta, error) {
//	threshold := 500
//	if page*limit < threshold {
//		return c.PaginateBy(page, limit, query, outPtr)
//	}
//
//	pMeta := PageMeta{
//		Page:    int64(page),
//		PerPage: int64(limit),
//	}
//
//	if pMeta.PerPage > 500 {
//		pMeta.PerPage = 500
//	}
//
//	paginationQuery := query.WithContext(ctx).Select(fqpk)
//	res := paginationQuery.Count(&pMeta.Total)
//	if res.Error != nil {
//		return pMeta, res.Error
//	}
//
//	pMeta.TotalPage = pMeta.Total / pMeta.PerPage
//	if pMeta.Total%pMeta.PerPage != 0 {
//		pMeta.TotalPage += 1
//	}
//
//	paginationQuery = paginationQuery.Offset((page - 1) * limit).Limit(limit)
//	delete(query.Statement.Clauses, clause.Limit{}.Name())
//
//	res = query.Where("? IN (?)", fqpk, paginationQuery)
//	if res.Error != nil {
//		return pMeta, res.Error
//	}
//	pMeta.Size = res.RowsAffected
//
//	return pMeta, nil
//}

// Unauthenticated will send a panic of type errors.ErrUnauthenticated. Can be
// used to stop controllers execution in the case authentication is required.
// When the controllers is running in the default HTTP server, this error will be
// automatically converted to the proper HTTP 401 response.
func (c *Controller) Unauthenticated() {
	panic(errors.ErrUnauthenticated)
}

// Forbidden will send a panic of type errors.ErrForbidden. Can be used to stop
// controllers execution in case of lack of access privilege. When the controllers
// is running in the default HTTP server, this error will be automatically
// converted to the proper HTTP 403 response.
func (c *Controller) Forbidden() {
	panic(errors.ErrForbidden)
}

// NotFound will send a panic of type errors.ErrNotFound. Can be used to stop
// controllers execution in case a requested resource or data cannot be found.
// When the controllers is running in the default HTTP server, this error will be
// automatically converted to the proper HTTP 404 response.
func (c *Controller) NotFound() {
	panic(errors.ErrNotFound)
}

// TooManyRequests will send a panic of type errors.ErrTooManyRequests. Can be
// used to stop controllers execution in case a requested resource or data cannot
// be found. When the controllers is running in the default HTTP server, this
// error will be automatically converted to the proper HTTP 429 response.
func (c *Controller) TooManyRequests() {
	panic(errors.ErrTooManyRequests)
}
