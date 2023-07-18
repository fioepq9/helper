package helper_test

import (
	"net/http/httptest"
	"time"

	"github.com/gin-gonic/gin"
	req "github.com/imroc/req/v3"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/rs/zerolog"

	"github.com/fioepq9/helper"
)

var _ = Describe("Checking Binding", Label("gin", "binding"), func() {
	type EchoRequest struct {
		Token   string `header:"token" binding:"required"`
		Name    string `uri:"name" binding:"required"`
		Message string `form:"message" default:"hello"`
	}

	type EchoResponse struct {
		Token   string `json:"token"`
		Name    string `json:"name"`
		Message string `json:"message"`
	}

	type LoginRequest struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	type LoginResponse struct {
		Username string `json:"username"`
		Password string `json:"token"`
	}

	type ListRequest struct {
		Limit    int           `form:"limit" default:"10"`
		Offset   int           `form:"offset" default:"1"`
		Order    string        `form:"order" default:"asc"`
		Names    []string      `form:"names" default:"alice,bob,charlie"`
		Percent  float64       `form:"percent" default:"0.5"`
		All      bool          `form:"all" default:"true"`
		Data     []byte        `form:"data" default:"what is the problem ?"`
		Timeout  time.Duration `form:"timeout" default:"5s"`
		Start    time.Time     `form:"start" default:"2020-01-01T00:00:00Z"`
		End      time.Time     `form:"end" default:"2023-07-01T00:00:00Z"`
		Today    time.Time     `form:"today" default:"now"`
		Tomorrow time.Time     `form:"tomorrow" default:"now+24h"`
	}

	type ListResponse struct {
		Limit    int           `json:"limit"`
		Offset   int           `json:"offset"`
		Order    string        `json:"order"`
		Names    []string      `json:"names"`
		Percent  float64       `json:"percent"`
		All      bool          `json:"all"`
		Data     []byte        `json:"data"`
		Timeout  time.Duration `json:"timeout"`
		Start    time.Time     `json:"start"`
		End      time.Time     `json:"end"`
		Today    time.Time     `json:"today"`
		Tomorrow time.Time     `json:"tomorrow"`
	}

	type NestedRequest struct {
		ListRequest `mapstructure:",squash"`
	}

	type NestedResponse struct {
		ListResponse
	}

	var (
		svc *httptest.Server
		c   *req.Client
	)

	BeforeEach(func() {
		gin.SetMode(gin.TestMode)
		e := gin.New()
		e.Use(func(c *gin.Context) {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
			log := helper.Zerolog().NewLogger(GinkgoWriter)
			ctx := log.WithContext(c.Request.Context())
			c.Request = c.Request.WithContext(ctx)
			c.Next()
		})
		r := helper.Gin().Router(e)
		r.GET("/echo/:name", func(c *gin.Context, req *EchoRequest) (resp *EchoResponse, err error) {
			return &EchoResponse{
				Token:   req.Token,
				Name:    req.Name,
				Message: req.Message,
			}, nil
		})
		r.POST("/login", func(c *gin.Context, req *LoginRequest) (resp *LoginResponse, err error) {
			return &LoginResponse{
				Username: req.Username,
				Password: req.Password,
			}, nil
		})
		r.GET("/list", func(c *gin.Context, req *ListRequest) (resp *ListResponse, err error) {
			return &ListResponse{
				Limit:    req.Limit,
				Offset:   req.Offset,
				Order:    req.Order,
				Names:    req.Names,
				Percent:  req.Percent,
				All:      req.All,
				Data:     req.Data,
				Timeout:  req.Timeout,
				Start:    req.Start,
				End:      req.End,
				Today:    req.Today,
				Tomorrow: req.Tomorrow,
			}, nil
		})
		r.GET("/list/nested", func(c *gin.Context, req *NestedRequest) (resp *NestedResponse, err error) {
			return &NestedResponse{
				ListResponse: ListResponse{
					Limit:    req.Limit,
					Offset:   req.Offset,
					Order:    req.Order,
					Names:    req.Names,
					Percent:  req.Percent,
					All:      req.All,
					Data:     req.Data,
					Timeout:  req.Timeout,
					Start:    req.Start,
					End:      req.End,
					Today:    req.Today,
					Tomorrow: req.Tomorrow,
				},
			}, nil
		})
		svc = httptest.NewServer(e)
		c = req.C().SetBaseURL(svc.URL)
	})
	When("method is GET and request has tag [ header, uri, form, default ]", func() {
		Context("and all fields are provided", func() {
			It("should return success", func(ctx SpecContext) {
				var resp EchoResponse
				httpResp, err := c.R().
					SetHeader("token", "foo").
					SetQueryParam("message", "bar").
					SetSuccessResult(&resp).
					Get("/echo/alice")
				Expect(err).To(BeNil())
				Expect(httpResp.IsSuccessState()).To(BeTrue())
				Expect(resp.Token).To(Equal("foo"))
				Expect(resp.Message).To(Equal("bar"))
				Expect(resp.Name).To(Equal("alice"))
			})
		})

		Context("and one field which has default tag is missing", func() {
			It("should return success", func(ctx SpecContext) {
				var resp EchoResponse
				httpResp, err := c.R().
					SetHeader("token", "foo").
					SetSuccessResult(&resp).
					Get("/echo/alice")
				Expect(err).To(BeNil())
				Expect(httpResp.IsSuccessState()).To(BeTrue())
				Expect(resp.Token).To(Equal("foo"))
				Expect(resp.Message).To(Equal("hello"))
				Expect(resp.Name).To(Equal("alice"))
			})
		})
	})
	When("method is POST and request has tag [ json ]", func() {
		Context("and all fields are provided", func() {
			It("should return success", func(ctx SpecContext) {
				var resp LoginResponse
				httpResp, err := c.R().
					SetBodyJsonMarshal(LoginRequest{Username: "foo", Password: "bar"}).
					SetSuccessResult(&resp).
					Post("/login")
				Expect(err).To(BeNil())
				Expect(httpResp.IsSuccessState()).To(BeTrue())
				Expect(resp.Username).To(Equal("foo"))
				Expect(resp.Password).To(Equal("bar"))
			})
		})
	})
	When("method is GET and request has tag [ form, default ]", func() {
		Context("and all fields are not provided", func() {
			It("should return success with default value", func(ctx SpecContext) {
				var resp ListResponse
				httpResp, err := c.R().SetSuccessResult(&resp).Get("/list")
				Expect(err).To(BeNil())
				Expect(httpResp.IsSuccessState()).To(BeTrue())
				Expect(resp.Limit).To(Equal(10))
				Expect(resp.Offset).To(Equal(1))
				Expect(resp.Order).To(Equal("asc"))
				Expect(resp.Names).To(Equal([]string{"alice", "bob", "charlie"}))
				Expect(resp.Percent).To(Equal(0.5))
				Expect(resp.All).To(BeTrue())
				Expect(resp.Data).To(Equal([]byte("what is the problem ?")))
				Expect(resp.Timeout).To(Equal(5 * time.Second))
				Expect(resp.Start).To(Equal(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)))
				Expect(resp.End).To(Equal(time.Date(2023, 7, 1, 0, 0, 0, 0, time.UTC)))
				Expect(resp.Today.Truncate(24 * time.Hour)).To(Equal(time.Now().Truncate(24 * time.Hour)))
				Expect(resp.Tomorrow.Truncate(24 * time.Hour)).To(Equal(time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)))
			})
		})

		Context("and all fields are not provided and the request has nested struct field", func() {
			It("should return success with default value", func(ctx SpecContext) {
				var resp NestedResponse
				httpResp, err := c.R().SetSuccessResult(&resp).Get("/list/nested")
				Expect(err).To(BeNil())
				Expect(httpResp.IsSuccessState()).To(BeTrue())
				Expect(resp.Limit).To(Equal(10))
				Expect(resp.Offset).To(Equal(1))
				Expect(resp.Order).To(Equal("asc"))
				Expect(resp.Names).To(Equal([]string{"alice", "bob", "charlie"}))
				Expect(resp.Percent).To(Equal(0.5))
				Expect(resp.All).To(BeTrue())
				Expect(resp.Data).To(Equal([]byte("what is the problem ?")))
				Expect(resp.Timeout).To(Equal(5 * time.Second))
				Expect(resp.Start).To(Equal(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)))
				Expect(resp.End).To(Equal(time.Date(2023, 7, 1, 0, 0, 0, 0, time.UTC)))
				Expect(resp.Today.Truncate(24 * time.Hour)).To(Equal(time.Now().Truncate(24 * time.Hour)))
				Expect(resp.Tomorrow.Truncate(24 * time.Hour)).To(Equal(time.Now().Truncate(24 * time.Hour).Add(24 * time.Hour)))
			})
		})
	})

	AfterEach(func() {
		svc.Close()
	})
})
