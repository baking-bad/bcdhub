package providers

// references types
const (
	RefTypeBranch = "branch"
	RefTypeTag    = "tag"
)

// Project -
type Project struct {
	User    string `json:"user"`
	Project string `json:"project"`
	URL     string `json:"url"`
}

// Account -
type Account struct {
	Login     string `json:"login"`
	AvatarURL string `json:"avatarURL"`
}

// Ref -
type Ref struct {
	Name string `json:"name"`
	URL  string `json:"url"`
	Type string `json:"type"`
}
