package common

import (
	"bufio"
	"os"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
)

func (c *Client) SendBetBatch(filePath string, agencyID uint8) {
	c.createClientSocket()

	file, err := os.Open(filePath)
	if err != nil {
		log.Criticalf("action: open_file | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	scanner := bufio.NewScanner(file)

	sigterm := false
	hasSignal := func() bool {
		select {
		case <-c.signal_chan:
			sigterm = true
			return true
		default:
			return false
		}
	}

	betBatch := protocol.NewBatch(agencyID, c.config.MaxBatchAmount)
	for scanner.Scan() && !hasSignal() {
		bet_line := scanner.Text()
		log.Debugf("action: read_bet_line | result: success | bet_line: %s", bet_line)
		// If batch is full, send it
		if !betBatch.AddBetLine(bet_line) {
			log.Infof("action: sending_batch | result: in_progress | batch: %d",
				betBatch.GetBetCount(),
			)
			c.sendBatch(betBatch)
			betBatch = protocol.NewBatch(agencyID, c.config.MaxBatchAmount)
			betBatch.AddBetLine(bet_line)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
	}
	if sigterm {
		file.Close()
		return
	}
	// Last batch
	c.sendBatch(betBatch)
	c.sendTerminate()

	file.Close()
}

func (c *Client) sendBatch(batch *protocol.BetBatchRegister) {
	data := batch.ToBytes()
	// Send the data to the server
	if err := FullWrite(c, data); err != nil {
		log.Errorf("action: send_batch | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		c.conn.Close()
		return
	}

	// Await response
	responseData, err := readWithPayloadLength(c.conn)

	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	// Check confirmation
	confirmation, err := protocol.DeserializeConfirmation(responseData)
	if err != nil {
		log.Errorf("action: deserialize_confirmation | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	if confirmation.Success {
		log.Infof("action: apuesta_enviada | result: success | apuestas_guardadas: %s",
			confirmation.Message,
		)
	}
}

func (c *Client) sendTerminate() {
	msg := []byte{byte(protocol.TERMINATE), 0x00, 0x00, 0x00, 0x00}
	FullWrite(c, msg)
	log.Infof("action: send_terminate | result: success")
	c.conn.Close()
}
