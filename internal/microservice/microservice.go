package microservice

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/aopoltorzhicky/bcdhub/internal/db"
	"github.com/aopoltorzhicky/bcdhub/internal/jsonload"
	"github.com/aopoltorzhicky/bcdhub/internal/mq"
	"github.com/aopoltorzhicky/bcdhub/internal/noderpc"
	"github.com/jinzhu/gorm"
	"github.com/streadway/amqp"
)

// Handler -
type Handler func(amqp.Delivery) error

// Microservice -
type Microservice struct {
	RPCs map[string]*noderpc.NodeRPC
	DB   *gorm.DB
	MQ   *mq.MQ

	queue   string
	handler Handler
	close   chan os.Signal
}

// New - create microservice
func New(configFile string, handler Handler) (*Microservice, error) {
	var cfg config
	if err := jsonload.StructFromFile(configFile, &cfg); err != nil {
		return nil, err
	}
	cfg.print()

	database, err := db.Database(cfg.Db.URI, cfg.Db.Log)
	if err != nil {
		return nil, err
	}

	messageQueue, err := mq.New(cfg.Mq.URI, []string{mq.QueueTags})
	if err != nil {
		return nil, err
	}

	rpcs := cfg.createRPCs()

	return &Microservice{
		DB:      database,
		MQ:      messageQueue,
		RPCs:    rpcs,
		queue:   cfg.Mq.Queue,
		handler: handler,
		close:   make(chan os.Signal, 2),
	}, err
}

// Close - close microservice
func (m *Microservice) Close() {
	if m.DB != nil {
		m.DB.Close()
	}
	if m.MQ != nil {
		m.MQ.Close()
	}
	close(m.close)
}

// Start - run microservice
func (m *Microservice) Start() {
	signal.Notify(m.close, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	msgs, err := m.MQ.Consume(m.queue)
	if err != nil {
		panic(err)
	}

	log.Print("Started")
	for {
		select {
		case <-m.close:
			log.Print("Stopped")
			return
		case msg := <-msgs:
			if err := m.handler(msg); err != nil {
				log.Printf("Microservice error: %s", err.Error())
			}
		}
	}
}
