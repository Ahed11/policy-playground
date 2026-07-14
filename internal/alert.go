package policy

type Alert struct {
	PolicyID string `json:"policy_id"`
	PolicyName string `json:"policy_name"`
	Severity string `json:"severity"`
	EventID string `json:"event_id"`
	UserID string `json:"user_id"`
	Matched bool `json:"matched"`
	Reasons []string `json:"reasons"`
}