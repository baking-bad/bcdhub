package channels

// StatsChannel -
type StatsChannel struct {
	messages chan Message
}

// NewStatsChannel -
func NewStatsChannel() *StatsChannel {
	return &StatsChannel{
		messages: make(chan Message),
	}
}

// GetName -
func (c *StatsChannel) GetName() string {
	return "stats"
}

// Run -
func (c *StatsChannel) Run() {

}

// Listen -
func (c *StatsChannel) Listen() <-chan Message {
	return c.messages
}

// Stop -
func (c *StatsChannel) Stop() {
	close(c.messages)
}
