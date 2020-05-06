package ws

import "github.com/baking-bad/bcdhub/cmd/api/ws/channels"

// Hub -
type Hub struct {
	public  []channels.Channel
	private []channels.Channel
	clients []Client

	stop chan struct{}
}

// NewHub -
func NewHub() *Hub {
	return &Hub{
		public:  make([]channels.Channel, 0),
		private: make([]channels.Channel, 0),
		clients: make([]Client, 0),

		stop: make(chan struct{}),
	}
}

// AddPublic -
func (h *Hub) AddPublic(channel channels.Channel) {
	h.public = append(h.public, channel)
}

// AddPrivate -
func (h *Hub) AddPrivate(channel channels.Channel) {
	h.private = append(h.private, channel)
}

// AddClient -
func (h *Hub) AddClient(client Client) {
	h.clients = append(h.clients, client)
}

// Run -
func (h *Hub) Run() {
	for i := range h.public {
		go func() {
			for {
				select {
				case <- h.stop:
					return
				case msg := <-h.public[i].Listen():
				}
			}
		}
	}


}

// Stop -
func (h *Hub) Stop() {
	h.stop <- struct{}{}
	for i := range h.public {
		h.public[i].Stop()
	}
}
