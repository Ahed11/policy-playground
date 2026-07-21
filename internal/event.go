package policy

type Event struct {
	EventID string `yaml:"event_id"`
	Time string `yaml:"time"`
	UserID string `yaml:"user_id"`
	Action string `yaml:"action"`
	ObjectType string `yaml:"object_type"`
	FileName *string `yaml:"file_name"`
	FileExt *string `yaml:"file_ext"`
	ContentClasses *[]string `yaml:"content_classes"`
	Channel string `yaml:"channel"`
	DestinationType string `yaml:"destination_type"`
	SizeBytes *int `yaml:"size_bytes"`
}