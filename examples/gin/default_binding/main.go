package main

import (
	"time"

	"github.com/fioepq9/helper"
	"github.com/gin-gonic/gin"
)

type EchoRequest struct {
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

type EchoResponse struct {
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

// curl -X GET "http://localhost:8080/echo"
// {"limit":10,"offset":1,"order":"asc","names":["alice","bob","charlie"],"percent":0.5,"all":true,"data":"d2hhdCBpcyB0aGUgcHJvYmxlbSA/","timeout":5000000000,"start":"2020-01-01T00:00:00Z","end":"2023-07-01T00:00:00Z","today":"2023-07-01T22:41:43.2329482+08:00","tomorrow":"2023-07-02T22:41:43.2329511+08:00"}
func main() {
	e := gin.New()

	r := helper.Gin().Router(e)

	r.GET("/echo", func(c *gin.Context, req *EchoRequest) (resp *EchoResponse, err error) {
		return &EchoResponse{
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

	if err := e.Run(":8080"); err != nil {
		panic(err)
	}
}
