package compilation

// compilation kinds
const (
	KindVerification = "verification"
	KindCompilation  = "compilation"
	KindDeployment   = "deployment"
)

// compilation statuses
const (
	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusFailed     = "failed"
	StatusSuccess    = "success"
	StatusError      = "error"
	StatusMismatch   = "mismatch"
)
