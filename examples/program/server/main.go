package main

import (
	"context"

	"github.com/gonitro/nitro/app"
	"github.com/gonitro/nitro/examples/program/types"
)

type Handler struct{}

func (h *Handler) Call(ctx context.Context, req *types.Request, rsp *types.Response) error {
	rsp.Message = "Hello " + req.Name
	return nil
}

func main() {
	// create a new program
	prog := app.New()
	// name it
	prog.Name("helloworld")
	// register a function
	prog.Register(new(Handler))
	// run the program
	prog.Run()
}
