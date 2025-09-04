package common

import (
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

	if err := c.sendGetWinners(agencyID); err != nil {
		log.Errorf("action: send_get_winners | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		c.sendTerminate()
		return
	}

	// Await winners response
	log.Debugf("action: espera_ganadores | result: in_progress")

	response, err := c.ReadWithPayloadLength()
	if err != nil {
		log.Errorf("action: read_response | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	hasWinners, winners, err := protocol.DeserializeWinnersResponse(response)
	if err != nil {
		log.Errorf("action: read_response | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	if hasWinners {
		log.Infof("action: consulta_ganadores | result: success | cant_ganadores: %d",
			winners.WinnersCount,
		)
	} else {
		log.Infof("action: consulta_ganadores | result: fail")
	}

	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *Client) sendGetWinners(agencyID uint8) error {
	msg := protocol.NewGetWinnersMessage(agencyID)
	return c.FullWrite(msg)
}
