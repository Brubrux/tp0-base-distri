package common

import (
	"time"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
)

func (c *Client) SendBetsAwaitWinners(filePath string, agencyID uint8) {
	if err := c.SendBetBatch(filePath, agencyID); err != nil {
		log.Errorf("action: send_bets | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	} else if c.hasSignal() {
		return
	}

	hasWinners := false
	var winners *protocol.Winners
	for !c.hasSignal() && !hasWinners {
		// sleep de 1 segundo para testear cierre gracefull
		time.Sleep(1 * time.Second)
		c.createClientSocket()
		if err := c.sendGetWinners(agencyID); err != nil {
			log.Errorf("action: send_bets | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			c.conn.Close()
			return
		}

		response, err := c.ReadWithPayloadLength()
		c.conn.Close()

		if err != nil {
			log.Errorf("action: send_bets | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}
		hasWinners, winners, err = protocol.DeserializeWinnersResponse(response)
		if err != nil {
			log.Errorf("action: send_bets | result: fail | client_id: %v | error: %v",
				c.config.ID,
				err,
			)
			return
		}
	}

	if hasWinners {
		log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d",
			winners.WinnersCount,
		)
	} else {
		log.Infof("action: consulta_ganadores | result: fail")
	}
}

func (c *Client) sendGetWinners(agencyID uint8) error {
	msg := protocol.NewGetWinnersMessage(agencyID)
	return c.FullWrite(msg)
}
