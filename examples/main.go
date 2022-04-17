package main

import (
	"embed"
	"fmt"

	_ "embed"

	"github.com/alextanhongpin/go-fixture"
)

//go:embed fixture.yaml
var raw []byte

//go:embed *.yaml
var yamls embed.FS

func main() {
	fmt.Println(fixture.Parse(raw))
	fmt.Println(fixture.ParseFS(yamls, "."))
}
