package services

type Service interface {
	// Start starts the service
	Start() error

	// Stop stops the service
	Stop() error

	// Reload reloads the service
	Reload() error

	// IsRunning returns true if the service is running
	IsRunning() bool

	// IsEnabled returns true if the service is enabled
	IsEnabled() bool
}
