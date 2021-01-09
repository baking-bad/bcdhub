package channels

import (
	"sync"
	"time"

	"github.com/baking-bad/bcdhub/cmd/api/ws/datasources"
	"github.com/baking-bad/bcdhub/internal/logger"
	"github.com/baking-bad/bcdhub/internal/models/block"
	"github.com/baking-bad/bcdhub/internal/mq"
	jsoniter "github.com/json-iterator/go"
)

var json = jsoniter.ConfigCompatibleWithStandardLibrary

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
	return c.initMessage()
}

func (c *StatsChannel) listen(source datasources.DataSource) {
	defer c.wg.Done()

	ch := source.Subscribe()
	for {
		select {
		case <-c.stop:
			source.Unsubscribe(ch)
			return
		case data := <-ch:
			if data.Type != datasources.RabbitType || data.Kind != mq.QueueBlocks {
				continue
			}
			if err := c.createMessage(data.Body.([]byte)); err != nil {
				logger.Error(err)
			}
		}
	}
}

func (c *StatsChannel) initMessage() error {
	blocks, err := c.ctx.Blocks.LastByNetworks()
	if err != nil {
		return err
	}
	stats, err := c.getStats(blocks)
	if err != nil {
		return err
	}
	c.messages <- Message{
		ChannelName: c.GetName(),
		Body:        stats,
	}
	return nil
}

func (c *StatsChannel) createMessage(data []byte) error {
	var b block.Block
	if err := json.Unmarshal(data, &b); err != nil {
		return err
	}
	stats, err := c.getStats([]block.Block{b})
	if err != nil {
		return err
	}
	c.messages <- Message{
		ChannelName: c.GetName(),
		Body:        stats,
	}
	return nil
}

func (c *StatsChannel) getStats(blocks []block.Block) ([]StatsBody, error) {
	var network string
	if len(blocks) == 1 {
		network = blocks[0].Network
	}
	callCounts, err := c.ctx.Storage.GetCallsCountByNetwork(network)
	if err != nil {
		return nil, err
	}
	contractStats, err := c.ctx.Storage.GetContractStatsByNetwork(network)
	if err != nil {
		return nil, err
	}
	faCount, err := c.ctx.Storage.GetFACountByNetwork(network)
	if err != nil {
		return nil, err
	}
	body := make([]StatsBody, len(blocks))
	for i := range blocks {
		body[i] = StatsBody{
			Network:   blocks[i].Network,
			Level:     blocks[i].Level,
			Timestamp: blocks[i].Timestamp,
			Protocol:  blocks[i].Protocol,
		}
		calls, ok := callCounts[blocks[i].Network]
		if ok {
			body[i].ContractCalls = calls
		}
		fa, ok := faCount[blocks[i].Network]
		if ok {
			body[i].FACount = fa
		}
		stats, ok := contractStats[blocks[i].Network]
		if ok {
			body[i].Total = stats.Total
			body[i].TotalBalance = stats.Balance
			body[i].UniqueContracts = stats.SameCount
		}
	}

	return body, nil
}
