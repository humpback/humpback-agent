package main

import (
	"context"
	"humpback-agent/app"
)

func main() {
	app.Bootstrap(context.Background())
}
