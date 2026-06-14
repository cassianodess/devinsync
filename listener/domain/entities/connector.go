package entities

import (
	"log"

	"github.com/gorilla/websocket"
)

type Connector struct {
	Connection *websocket.Conn
	BaseURL    string
}

func NewConnector(baseURL string) (*Connector, error) {
	connection, _, err := websocket.DefaultDialer.Dial(
		baseURL,
		nil,
	)

	if err != nil {
		return nil, err
	}

	return &Connector{
		Connection: connection,
		BaseURL:    baseURL,
	}, nil
}

func (this *Connector) SendMessage(event *Event) error {
	err := this.Connection.WriteJSON(event)
	if err != nil {
		log.Println("erro while send json: ", err)
		return err
	}

	return nil

}
