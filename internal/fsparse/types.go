package fsparse

type Workflow struct {
	Name         string
	Tasks        []Task
	Dependencies map[string][]string
}

type Task struct {
	ID           string
	MarkdownPath string
	Command      string
	Dependencies []string
	Priority     string
	Retries      int
	Timeout      string
	Status       string
	Output       string
	Duration     string
}
