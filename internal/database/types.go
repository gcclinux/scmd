package database

// CommandRecord represents a stored command in the database.
type CommandRecord struct {
	Id   int    `json:"id"`
	Key  string `json:"key"`
	Data string `json:"data"`
}
