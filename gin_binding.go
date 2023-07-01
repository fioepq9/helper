package helper

import (
	"reflect"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/gin-gonic/gin/binding"

	"github.com/mitchellh/mapstructure"
)

type GinBinding interface {
	Name() string
	Bind(c *gin.Context, obj any) error
}

type GinDefaultBinding struct {
	TagName     string
	DecodeHooks []mapstructure.DecodeHookFunc
}

func NewGinDefaultBinding(options ...func(*GinDefaultBinding)) *GinDefaultBinding {
	b := &GinDefaultBinding{
		TagName: "json",
		DecodeHooks: []mapstructure.DecodeHookFunc{
			StringToSliceHookFunc(","),
			StringToBoolHookFunc(),
			StringToIntHookFunc(),
			StringToFloat64HookFunc(),
			StringToBytesHookFunc(),
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.OrComposeDecodeHookFunc(
				StringToTimeHookFunc(),
				mapstructure.StringToTimeHookFunc(time.RFC3339),
				mapstructure.StringToTimeHookFunc(time.RFC3339Nano),
			),
		},
	}

	for _, opt := range options {
		opt(b)
	}

	return b
}

func (b *GinDefaultBinding) Name() string {
	return "default"
}

func (b *GinDefaultBinding) Bind(_ *gin.Context, obj any) error {
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:    b.TagName,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(b.DecodeHooks...),
		Result:     obj,
	})
	if err != nil {
		return err
	}
	t := reflect.TypeOf(obj)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	if t.Kind() != reflect.Struct {
		return nil
	}

	dict := make(map[string]string)
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if defaultStr, ok := f.Tag.Lookup(b.Name()); ok {
			name := f.Name
			if tName, ok := f.Tag.Lookup(b.TagName); ok {
				name = tName
			}
			dict[name] = defaultStr
		}
	}

	if len(dict) == 0 {
		return nil
	}

	return decoder.Decode(dict)
}

type GinURIBinding struct {
	Params     func(*gin.Context) map[string][]string
	BindingURI binding.BindingUri
}

func NewGinURIBinding() *GinURIBinding {
	return &GinURIBinding{
		Params: func(c *gin.Context) map[string][]string {
			m := make(map[string][]string)
			for _, v := range c.Params {
				m[v.Key] = []string{v.Value}
			}
			return m
		},
		BindingURI: binding.Uri,
	}
}

func (b *GinURIBinding) Name() string {
	return "uri"
}

func (b *GinURIBinding) Bind(c *gin.Context, obj any) error {
	m := b.Params(c)
	return b.BindingURI.BindUri(m, obj)
}

type GinBindingWrapper struct {
	Binding binding.Binding
}

func NewGinBinding(b binding.Binding) *GinBindingWrapper {
	return &GinBindingWrapper{
		Binding: b,
	}
}

func (b *GinBindingWrapper) Name() string {
	return b.Binding.Name()
}

func (b *GinBindingWrapper) Bind(c *gin.Context, obj any) error {
	return b.Binding.Bind(c.Request, obj)
}
