package jobs

type Job struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Attempts int    `json:"attempts"`
}

type JobStatus struct {
	ID        string `json:"id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	Attempts  int    `json:"attempts"`
	UpdatedAt string `json:"updated_at"`
}

func StatusKey(id string) string {
	return "job_status:" + id
}
