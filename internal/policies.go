package internal

import (
	"gopkg.in/yaml.v2"
)

type Policies struct {
	Policies []Policy `yaml:"policies"`
}