package channels

// Channel -
type Channel interface {
	GetName() string
	Run()
	Listen() <-chan Message
	Stop()
}

// Message -
type Message struct {
	ChannelName string      `json:"channel_name"`
	Body        interface{} `json:"body"`
}
