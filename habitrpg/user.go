package habitrpg

import "time"

type TaskChallenge struct {
	ID     string `json:"id,omitempty"`
	Broken string `json:"broken,omitempty"`
	Winner string `json:"winner,omitempty"`
}

type TaskHistory struct {
	Date  int     `json:"date,omitempty"` // WTF: Though docs says this is a Date, there is an timestamp in it
	Value float64 `json:"value,omitempty"`
}

type TaskRepeat struct {
	Monday    bool `json:"m,omitempty"`
	Tuesday   bool `json:"t,omitempty"`
	Wednesday bool `json:"w,omitempty"`
	Thursday  bool `json:"th,omitempty"`
	Friday    bool `json:"f,omitempty"`
	Saturday  bool `json:"s,omitempty"`
	Sunday    bool `json:"su,omitempty"`
}

type TaskChecklistEntry struct {
	Completed bool   `json:"completed,omitempty"`
	Text      string `json:"text,omitempty"`
	ID        string `json:"id,omitempty"`
}

type Task struct {
	// General
	ID          string          `json:"id,omitempty"`
	Type        string          `json:"type,omitempty"`
	DateCreated time.Time       `json:"dateCreated,omitempty"`
	Text        string          `json:"text,omitempty"`
	Notes       string          `json:"notes,omitempty"`
	Tags        map[string]bool `json:"tags,omitempty"`
	Value       float64         `json:"value,omitempty"`
	Priority    int             `json:"priority,omitempty"`
	Attribute   string          `json:"attribute,omitempty"`
	Challenge   TaskChallenge   `json:"challenge,omitempty"`

	// Multiple types
	History           []TaskHistory        `json:"history,omitempty"`
	Completed         bool                 `json:"completed,omitempty"`
	CollapseChecklist bool                 `json:"collapseChecklist,omitempty"`
	Checklist         []TaskChecklistEntry `json:"checklist,omitempty"`

	// Habit
	Up   bool `json:"up,omitempty"`
	Down bool `json:"down,omitempty"`

	// Daily
	Repeat TaskRepeat `json:"repeat,omitempty"`
	Streak int        `json:"streak,omitempty"`

	// Todo
	DateCompleted time.Time `json:"dateCompleted,omitempty"`
	Date          string    `json:"date,omitempty"`
}
