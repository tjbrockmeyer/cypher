package neohttp

type request struct {
	Statements []query `json:"statements"`
}

type query struct {
	Statement  string                 `json:"statement"`
	Parameters map[string]interface{} `json:"parameters,omitempty"`
	IncludeStats bool `json:"includeStats,omitempty"`
}

