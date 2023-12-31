# helper

## gin

1. 提供了更多的 handler 选择。
   - `func(*gin.Context)` 
   - `func(*gin.Context) error`
   - `func(*gin.Context, *reqType) error`
   - `func(*gin.Context) (*respType, error)`
   - `func(*gin.Context, *reqType) (*respType, error)`

2. 自动参数绑定。[如何添加更多支持的 tag ?](./examples/gin/add_new_binding/main.go)
   - `header`
   - `uri`
   - `form`
   - `json`

3. 默认使用中文的 validator。 
 
4. 通过 tag `default` 为 `reqType` 提供默认值，默认支持: 
   - `string`: `default:"foo"`
   - `[]byte`: `default:"bar"`
   - `int`: `default:"10"`
   - `float64`: `default:"10.0"`
   - `bool`: `default:"true"`
   - `time.Duration`: `default:"10s"`, 使用 `time.ParseDuration` 进行解析。
   - `time.Time`: 使用 `time.RFC3339` 或者 `time.RFC3339Nano`，支持 `now+{time.Duration}`, `now-{time.Duration}`
   - 使用 `mapstructure` 支持自定义类型。
   
5. 支持使用 `zerolog` 覆盖以下 `gin` 的配置。 
   - `gin.DefaultWriter`
   - `gin.DefaultErrorWriter`
   - `gin.DebugPrintRouteFunc` 

6. `reqType` 支持下列钩子。
   - `BeforeBind(*gin.Context)`
   - `AfterBind(*gin.Context)`
   - `BeforeValidate(*gin.Context)`
   - `AfterValidate(*gin.Context)`

### Usage

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

### Examples

1. [默认值的使用](./examples/gin/default_binding/main.go)

## Contributing
![Alt](https://repobeats.axiom.co/api/embed/fc33fc4f571db13b097859952614b06b48f46bbe.svg "Repobeats analytics image")