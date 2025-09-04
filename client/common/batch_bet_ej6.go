package common

import (
	"bufio"
	"fmt"
	"os"
	"time"

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
		// sleep de 0.1 segundo para testear cierre gracefull
		time.Sleep(5 * time.Millisecond)

		bet_line := scanner.Text()
		// If batch is full, send it
		if !betBatch.AddBetLine(bet_line) {
			log.Infof("action: sending_batch | result: in_progress | batch: %d",
				betBatch.GetBetCount(),
			)

			if err := c.sendBatch(betBatch); err != nil {
				log.Errorf("action: send_batch | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err,
				)

				file.Close()
				c.conn.Close()
				return
			} else {

			}

			betBatch = protocol.NewBatch(agencyID, c.config.MaxBatchAmount)
			betBatch.AddBetLine(bet_line)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
		file.Close()
		return
	}

	if sigterm {
		c.sendTerminate()
		file.Close()
		return
	}
	// Last batch
	if err := c.sendBatch(betBatch); err != nil {
		log.Errorf("action: send_batch | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		file.Close()
		c.conn.Close()
		return
	}
	c.sendTerminate()
	file.Close()
}

func (c *Client) sendBatch(batch *protocol.BetBatchRegister) error {
	data := batch.ToBytes()
	// Send the data to the server
	if err := FullWrite(c, data); err != nil {
		return fmt.Errorf("error while sending batch: %v", err)
	}

	// Await response
	responseData, err := readWithPayloadLength(c.conn)
	if err != nil {
		return fmt.Errorf("error while reading response: %v", err)
	}

	// Check confirmation
	confirmation, err := protocol.DeserializeConfirmation(responseData)
	if err != nil {
		return fmt.Errorf("error while deserializing confirmation: %v", err)
	}

	if confirmation.Success {
		log.Infof("action: send_batch | result: success | client_id: %v | confirmation: %v",
			c.config.ID,
			confirmation.Message,
		)
	}
	return nil
}

func (c *Client) sendTerminate() {
	msg := []byte{byte(protocol.TERMINATE), 0x00, 0x00, 0x00, 0x00}
	FullWrite(c, msg)
	log.Infof("action: send_terminate | result: success")
	c.conn.Close()
}
