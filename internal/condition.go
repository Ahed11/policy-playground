package policy

type Condition struct {
	Field string `yaml:"field"`
	Equals string `yaml:"equals"`
	In []string `yaml:"in"`
	Exists *bool `yaml:"exists"`
	Contains string `yaml:"contains"`
	All []Condition `yaml:"all"`
	Any []Condition `yaml:"any"`
}