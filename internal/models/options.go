package models

// CreateRepositoryOption -
type CreateRepositoryOption func(map[string]interface{})

// WithReadOnly -
func WithReadOnly() CreateRepositoryOption {
	return func(data map[string]interface{}) {
		data["readonly"] = true
	}
}

// WithCompress -
func WithCompress() CreateRepositoryOption {
	return func(data map[string]interface{}) {
		data["compress"] = "true"
	}
}

// WithMaxRetries -
func WithMaxRetries(maxRetries int64) CreateRepositoryOption {
	return func(data map[string]interface{}) {
		data["max_retries"] = maxRetries
	}
}
