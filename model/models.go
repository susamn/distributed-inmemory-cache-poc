package model

type DataPayload struct {
	DataVersion  int64             `json:"data_version"`
	Data         map[string]string `json:"data"`
	PID          int               `json:"pid"`
	RunningSince int64             `json:"running_since"`
}
