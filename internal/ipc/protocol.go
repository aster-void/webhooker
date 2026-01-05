package ipc

// Client → Server
type Request struct {
	Type string `json:"type"` // "register"
}

// Server → Client
type Response struct {
	Type string `json:"type"` // "registered", "webhook", "error"
	Path string `json:"path,omitempty"`
	Data string `json:"data,omitempty"`
}
