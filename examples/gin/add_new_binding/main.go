package main

import (
	"github.com/fioepq9/helper"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

type EchoRequest struct {
	Message string `yaml:"message"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

// curl -X POST "http://localhost:8080/echo" -d 'message: hello'
// {"message":"hello"}
func main() {
	e := gin.New()

	// add yaml
	r := helper.
		Gin(func(ginHelper *helper.GinHelper) {
			ginHelper.Bindings = append(ginHelper.Bindings, helper.NewGinBinding(binding.YAML))
		}).
		Router(e)

	r.POST("/echo", func(c *gin.Context, req *EchoRequest) (resp *EchoResponse, err error) {
		return &EchoResponse{
			Message: req.Message,
		}, nil
	})

	if err := e.Run(":8080"); err != nil {
		panic(err)
	}
}
