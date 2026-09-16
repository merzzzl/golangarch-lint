package dto

type Violation struct {
	Severity string `json:"severity"`
	Check    string `json:"check"`
	Rule     string `json:"rule"`
	Path     string `json:"path"`
	Pos      string `json:"pos"`
	Message  string `json:"message"`
}
