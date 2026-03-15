package linear

// LinearProject represents the subset of project fields used by control-api.
type LinearProject struct {
	ID   string
	Name string
	URL  string
}

// LinearCycle represents the subset of cycle fields used by control-api.
type LinearCycle struct {
	ID   string
	Name string
	URL  string
}

// LinearIssue represents the subset of issue fields used by control-api.
type LinearIssue struct {
	ID         string
	Identifier string
	Title      string
	URL        string
}
