package main

import (
	"embed"
	"fmt"

	_ "embed"

	"github.com/alextanhongpin/go-fixture"
	"gopkg.in/yaml.v3"
)

//go:embed fixture.yaml
var raw []byte

//go:embed *.yaml
var yamls embed.FS

func main() {
	fmt.Println(fixture.Parse(raw, yaml.Unmarshal))
	fmt.Println(fixture.ParseFS(yamls, ".", yaml.Unmarshal))
}
