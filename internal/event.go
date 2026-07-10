package internal

import (
	"gopkg.in/yaml.v2"
)

type Event struct {
	Event_id string `yaml:"event_id"`
	Time string `yaml:"time"`
	User_id string `yaml:"user_id"`
	Action string `yaml:"action"`
	Object_type string `yaml:"object_type"`
	File_name string `yaml:"file_name"`
	File_ext string `yaml:"file_ext"`
	Content_classes []string `yaml:"content_classes"`
	Channel string `yaml:"channel"`
	Destination_type string `yaml:"destination_type"`
	Size_bytes int `yaml:"size_bytes"`
}