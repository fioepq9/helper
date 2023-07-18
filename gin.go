package helper

import (
	"net/http"
	"reflect"
	"strings"
	"sync"

	"github.com/cockroachdb/errors"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	validator "github.com/go-playground/validator/v10"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"
	"github.com/rs/zerolog"
)

var (
	ginHelper     *GinHelper
	ginHelperOnce sync.Once
)

type GinHelper struct {
	Bindings            []GinBinding
	BindingValidator    binding.StructValidator
	BindingErrorHandler func(*gin.Context, error)
	SuccessHandler      func(*gin.Context, any)
	ErrorHandler        func(*gin.Context, error)
}

// Gin
//   - Notes: This function first call will disable the gin default validator
func Gin(options ...func(*GinHelper)) *GinHelper {
	ginHelperOnce.Do(func() {
		ginHelper = &GinHelper{
			Bindings: []GinBinding{
				NewGinDefaultBinding(),
				NewGinBinding(binding.Header),
				NewGinURIBinding(),
				NewGinBinding(binding.Form),
				NewGinBinding(binding.JSON),
			},
			BindingValidator: NewGinValidator(),
			BindingErrorHandler: func(c *gin.Context, err error) {
				c.AbortWithStatusJSON(http.StatusBadRequest, err.Error())
			},
			ErrorHandler: func(c *gin.Context, err error) {
				c.AbortWithStatusJSON(http.StatusInternalServerError, err.Error())
			},
			SuccessHandler: func(c *gin.Context, resp any) {
				c.JSON(http.StatusOK, resp)
			},
		}
		gin.DisableBindValidation()
	})
	for _, opt := range options {
		opt(ginHelper)
	}
	return ginHelper
}

// SetZerologWriter set zerolog writer
//   - gin.DefaultWriter
//   - gin.DefaultErrorWriter
//   - gin.DebugPrintRouteFunc
func (h *GinHelper) SetZerologWriter(log zerolog.Logger, lvl zerolog.Level) *GinHelper {
	zw := ginZerologWriter{log: log, lvl: lvl}
	zw.SetAll()
	return h
}

func (h *GinHelper) Router(routes gin.IRoutes) *GinRouter {
	return &GinRouter{
		helper: h,
		routes: routes,
	}
}

type GinRouter struct {
	routes gin.IRoutes
	helper *GinHelper
}

func (r *GinRouter) GET(path string, handler any) *GinRouter {
	return r.Handle(http.MethodGet, path, handler)
}

func (r *GinRouter) POST(path string, handler any) *GinRouter {
	return r.Handle(http.MethodPost, path, handler)
}

func (r *GinRouter) Handle(method string, path string, handler any) *GinRouter {
	assertHandler(handler)
	v := reflect.ValueOf(handler)
	t := v.Type()

	request := func(c *gin.Context) ([]reflect.Value, error) {
		in := make([]reflect.Value, 0, t.NumIn())
		in = append(in, reflect.ValueOf(c))
		if t.NumIn() == 2 {
			reqV := reflect.New(t.In(1).Elem())
			reqT := reqV.Elem().Type()
			// check if request struct has tags
			hasTags := make(map[string]bool)
			hasTags["default"] = true
			for i := 0; i < reqT.NumField(); i++ {
				for _, b := range r.helper.Bindings {
					tag := b.Name()
					if hasTags[tag] {
						continue
					}
					if _, ok := reqT.Field(i).Tag.Lookup(tag); ok {
						hasTags[tag] = true
					}
				}
			}
			// call BeforeBind hook
			if beforeBinding, ok := reqV.Interface().(BeforeBinding); ok {
				if err := beforeBinding.BeforeBind(c); err != nil {
					return nil, errors.Wrap(err, "hook BeforeBind failed")
				}
			}
			// bind
			for _, b := range r.helper.Bindings {
				if !hasTags[b.Name()] {
					continue
				}
				err := b.Bind(c, reqV.Interface())
				if err != nil {
					return nil, errors.Wrapf(err, "bind %s failed", b.Name())
				}
			}
			// call AfterBind hook
			if afterBinding, ok := reqV.Interface().(AfterBinding); ok {
				if err := afterBinding.AfterBind(c); err != nil {
					return nil, errors.Wrap(err, "hook AfterBind failed")
				}
			}
			// call BeforeValidate hook
			if beforeValidation, ok := reqV.Interface().(BeforeValidation); ok {
				if err := beforeValidation.BeforeValidate(c); err != nil {
					return nil, errors.Wrap(err, "hook BeforeValidate failed")
				}
			}
			// validate
			err := r.helper.BindingValidator.ValidateStruct(reqV.Elem().Interface())
			if err != nil {
				return nil, errors.Wrap(err, "validate failed")
			}
			// call AfterValidate hook
			if afterValidation, ok := reqV.Interface().(AfterValidation); ok {
				if err := afterValidation.AfterValidate(c); err != nil {
					return nil, errors.Wrap(err, "hook AfterValidate failed")
				}
			}
			in = append(in, reqV)
		}
		return in, nil
	}

	r.routes.Handle(method, path, func(c *gin.Context) {
		in, err := request(c)
		if err != nil {
			r.helper.BindingErrorHandler(c, err)
			return
		}
		out := v.Call(in)
		var resp any
		switch len(out) {
		case 0:
			return
		case 1:
			if errVal := out[0].Interface(); errVal != nil {
				err = errVal.(error)
			}
		case 2:
			resp = out[0].Interface()
			if errVal := out[1].Interface(); errVal != nil {
				err = errVal.(error)
			}
		default:
			panic("invalid count for handler return values")
		}
		if err != nil {
			r.helper.ErrorHandler(c, err)
			return
		}
		if resp != nil {
			r.helper.SuccessHandler(c, resp)
			return
		}
	})

	return r
}

type BeforeBinding interface {
	BeforeBind(c *gin.Context) error
}

type AfterBinding interface {
	AfterBind(c *gin.Context) error
}

type BeforeValidation interface {
	BeforeValidate(c *gin.Context) error
}

type AfterValidation interface {
	AfterValidate(c *gin.Context) error
}

type GinValidator struct {
	Validate           *validator.Validate
	Translator         locales.Translator
	TranslatorRegister func(v *validator.Validate, trans ut.Translator) error
	utTranslator       *ut.UniversalTranslator
}

var _ binding.StructValidator = (*GinValidator)(nil)

func NewGinValidator(options ...func(*GinValidator)) *GinValidator {
	v := validator.New()
	v.SetTagName("binding")

	gv := &GinValidator{
		Validate:           v,
		Translator:         zh.New(),
		TranslatorRegister: zhTranslations.RegisterDefaultTranslations,
	}

	for _, opt := range options {
		opt(gv)
	}

	gv.utTranslator = ut.New(gv.Translator)

	err := gv.TranslatorRegister(v, gv.utTranslator.GetFallback())
	if err != nil {
		panic(err)
	}

	return gv
}

func (v *GinValidator) ValidateStruct(obj any) error {
	val := reflect.ValueOf(obj)
	typ := val.Type()

	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	if typ.Kind() != reflect.Struct {
		return nil
	}

	err := v.Validate.Struct(obj)
	if err != nil {
		errs := err.(validator.ValidationErrors)
		kvTuple := make([]string, 0)
		for k, v := range errs.Translate(v.utTranslator.GetFallback()) {
			kvTuple = append(kvTuple, k+"="+v)
		}
		return errors.Newf("[%s]", strings.Join(kvTuple, ","))
	}

	return nil
}

func (v *GinValidator) Engine() any {
	return v.Validate
}

// assertHandler checks if handler is valid
// handler must be a function
// handler's first argument must be *gin.Context
// handler's second argument must be a struct
// handler's last return value must be error
// handler's first return value must be a pointer
// example:
//   - func(c *gin.Context)
//   - func(c *gin.Context) error
//   - func(c *gin.Context, *req) error
//   - func(c *gin.Context) (*resp, error)
//   - func(c *gin.Context, *req) (*resp, error)
func assertHandler(handler any) {
	v := reflect.ValueOf(handler)
	t := v.Type()

	if t.Kind() != reflect.Func {
		panic("handler must be a function")
	}

	if t.NumIn() == 0 || t.NumIn() > 2 {
		panic("handler must have 1 or 2 arguments")
	}
	if t.In(0) != reflect.TypeOf(&gin.Context{}) {
		panic("handler's first argument must be *gin.Context")
	}
	if t.NumIn() == 2 &&
		(t.In(1).Kind() != reflect.Ptr || t.In(1).Elem().Kind() != reflect.Struct) {
		panic("handler's second argument must be a struct pointer")
	}

	if t.NumOut() > 2 {
		panic("handler return values count must be 2 or less")
	}
	if t.NumOut() != 0 && !t.Out(t.NumOut()-1).Implements(reflect.TypeOf((*error)(nil)).Elem()) {
		panic("handler's last return value must be error")
	}
	if t.NumOut() == 2 && t.Out(0).Kind() != reflect.Ptr {
		panic("handler's first return value must be a pointer")
	}
}
