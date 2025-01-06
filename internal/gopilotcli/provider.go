package gopilotcli

// Provider is a singleton that holds the current GopilotCLI implementation
var provider GopilotCLI = NewRealGopilot()

// SetProvider allows tests to inject a mock implementation
func SetProvider(p GopilotCLI) {
	provider = p
}

// GetProvider returns the current GopilotCLI implementation
func GetProvider() GopilotCLI {
	return provider
}

// Reset restores the default real implementation
func Reset() {
	provider = NewRealGopilot()
}
