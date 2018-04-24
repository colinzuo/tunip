package generator

// FilebeatFields simulate for filebeat
type FilebeatFields struct {
	ContainerType string `json:"container_type"`
}

// FilebeatWrapper simulate for filebeat
type FilebeatWrapper struct {
	Timestamp   string         `json:"@timestamp"`
	Hostname    string         `json:"hostname"`
	Fields      FilebeatFields `json:"fields"`
	JSONExtract interface{}    `json:"json_extract"`
}
