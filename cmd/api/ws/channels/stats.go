package channels

import (
	"sort"
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/cmd/api/ws/datasources"
	"github.com/baking-bad/bcdhub/internal/logger"
)

// StatsChannel -
type StatsChannel struct {
	*DefaultChannel

	messages chan Message
	stop     chan struct{}
	wg       sync.WaitGroup
}

// StatsBody -
type StatsBody struct {
	Network         string    `json:"network"`
	Level           int64     `json:"level"`
	Timestamp       time.Time `json:"time"`
	Protocol        string    `json:"protocol"`
	Total           int64     `json:"total"`
	ContractCalls   int64     `json:"contract_calls"`
	UniqueContracts int64     `json:"unique_contracts"`
	TotalBalance    int64     `json:"total_balance"`
	TotalWithdrawn  int64     `json:"total_withdrawn"`
	FACount         int64     `json:"fa_count"`
}

type byNetwork []StatsBody

func (a byNetwork) Len() int           { return len(a) }
func (a byNetwork) Less(i, j int) bool { return a[i].Network < a[j].Network }
func (a byNetwork) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }

// NewStatsChannel -
func NewStatsChannel(opts ...ChannelOption) *StatsChannel {
	return &StatsChannel{
		DefaultChannel: NewDefaultChannel(opts...),
		messages:       make(chan Message, 10),
		stop:           make(chan struct{}),
	}
}

// GetName -
func (c *StatsChannel) GetName() string {
	return "stats"
}

// Run -
func (c *StatsChannel) Run() {
	if len(c.sources) == 0 {
		logger.Errorf("[%s] Empty source list", c.GetName())
		return
	}

	for i := range c.sources {
		c.wg.Add(1)
		go c.listen(c.sources[i])
	}
}

// Listen -
func (c *StatsChannel) Listen() <-chan Message {
	return c.messages
}

// Stop -
func (c *StatsChannel) Stop() {
	close(c.stop)
	c.wg.Wait()
	close(c.messages)
}

// Init -
func (c *StatsChannel) Init() error {
	return c.createMessage()
}

func (c *StatsChannel) listen(source datasources.DataSource) {
	defer c.wg.Done()

	ch := source.Subscribe()
	for {
		select {
		case <-c.stop:
			source.Unsubscribe(ch)
		case data := <-ch:
			if data.Type != datasources.RabbitType {
				continue
			}
			if err := c.createMessage(); err != nil {
				logger.Error(err)
			}
		}
	}
}

func (c *StatsChannel) createMessage() error {
	states, err := c.es.GetAllStates()
	if err != nil {
		return err
	}
	callCounts, err := c.es.GetCallsCountByNetwork()
	if err != nil {
		return err
	}
	contractStats, err := c.es.GetContractStatsByNetwork()
	if err != nil {
		return err
	}

	faCount, err := c.es.GetFACountByNetwork()
	if err != nil {
		return err
	}

	body := make([]StatsBody, len(states))
	for i := range states {
		body[i] = StatsBody{
			Network:   states[i].Network,
			Level:     states[i].Level,
			Timestamp: states[i].Timestamp,
			Protocol:  states[i].Protocol,
		}
		calls, ok := callCounts[states[i].Network]
		if ok {
			body[i].ContractCalls = calls
		}
		fa, ok := faCount[states[i].Network]
		if ok {
			body[i].FACount = fa
		}
		stats, ok := contractStats[states[i].Network]
		if ok {
			body[i].Total = stats.Total
			body[i].TotalBalance = stats.Balance
			body[i].TotalWithdrawn = stats.TotalWithdrawn
			body[i].UniqueContracts = stats.SameCount
		}
	}

	sort.Sort(byNetwork(body))
	c.messages <- Message{
		ChannelName: c.GetName(),
		Body:        body,
	}
	return nil
}
