# helper

## gin

1. 提供了更多的 handler 选择: `func(*gin.Context) error`, `func(*gin.Context, *reqType) error`, `func(*gin.Context) (*respType, error)`, `func(*gin.Context, *reqType) (*respType, error)`。
2. 默认使用中文的 validator。 
3. 通过 tag 为 `reqType` 提供默认值，默认支持: `string`, `[]byte`, `int`, `float64`, `bool`, `time.Duration`, `time.Time`。使用 `mapstructure` 支持自定义类型。
   
   -  `time.Time`: 使用 `time.RFC3339` 或者 `time.RFC3339Nano`，支持 `now+{time.Duration}`, `now-{time.Duration}`
5. 支持使用 `zerolog` 覆盖默认的 `gin.DefaultWriter`, `gin.DefaultErrorWriter`, `DebugPrintRouteFunc`。 

### usage

```go
package main

import (
	"github.com/fioepq9/helper"
	"github.com/gin-gonic/gin"
)

type EchoRequest struct {
	Message string `form:"message"`
}

type EchoResponse struct {
	Message string `json:"message"`
}

// curl -X GET "http://localhost:8080/echo?message=hello"
// {"message": "hello"}
func main() {
	e := gin.New()
	
	r := helper.Gin().Router(e)
	
	r.GET("/echo", func(c *gin.Context, req *EchoRequest) (resp *EchoResponse, err error) {
		return &EchoResponse{Message: req.Message}, nil
    })
	
	if err := e.Run(":8080"); err != nil {
		panic(err)
    }
}
```

### examples

1. [默认值的使用](./examples/gin/default_binding)