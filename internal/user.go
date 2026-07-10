package internal

import (
	"gopkg.in/yaml.v2"
)

type User struct {
	User_id string `yaml:"user_id"`
	Department string `yaml:"department"`
	Role string `yaml:"role"`
}
