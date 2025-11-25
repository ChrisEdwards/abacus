package beads

// LiteIssue represents a minimal issue record returned by `bd list`.
type LiteIssue struct {
	ID string `json:"id"`
}

// Comment models a Beads issue comment entry.
type Comment struct {
	ID        int    `json:"id"`
	IssueID   string `json:"issue_id"`
	Author    string `json:"author"`
	Text      string `json:"text"`
	CreatedAt string `json:"created_at"`
}

// Dependency captures dependency metadata from the Beads API.
type Dependency struct {
	TargetID string `json:"id"`
	Type     string `json:"dependency_type"`
}

// Dependent represents a reverse dependency entry.
type Dependent struct {
	ID   string `json:"id"`
	Type string `json:"dependency_type"`
}

// FullIssue models the expanded issue data used by the UI.
type FullIssue struct {
	ID                 string       `json:"id"`
	Title              string       `json:"title"`
	Status             string       `json:"status"`
	IssueType          string       `json:"issue_type"`
	Priority           int          `json:"priority"`
	Description        string       `json:"description"`
	Design             string       `json:"design"`
	AcceptanceCriteria string       `json:"acceptance_criteria"`
	Notes              string       `json:"notes"`
	CreatedAt          string       `json:"created_at"`
	UpdatedAt          string       `json:"updated_at"`
	ClosedAt           string       `json:"closed_at"`
	ExternalRef        string       `json:"external_ref"`
	Labels             []string     `json:"labels"`
	Comments           []Comment    `json:"comments"`
	Dependencies       []Dependency `json:"dependencies"`
	Dependents         []Dependent  `json:"dependents"`
}
