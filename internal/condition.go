package policy

type Condition struct {
	Field string `yaml:"field"`
	Equals string `yaml:"equals"`
	In []string `yaml:"in"`
	Contains string `yaml:"contains"`
	All []Condition `yaml:"all"`
}