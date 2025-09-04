package common

import (
	"bufio"
	"fmt"
	"os"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
)

func (c *Client) SendBetBatch(filePath string, agencyID uint8) error {
	c.createClientSocket()

	file, err := os.Open(filePath)
	if err != nil {
		log.Criticalf("action: open_file | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	betBatch := protocol.NewBatch(agencyID, c.config.MaxBatchAmount)

	for scanner.Scan() && !c.hasSignal() {
		// sleep de 0.1 segundo para testear cierre gracefull
		// time.Sleep(5 * time.Millisecond)

		bet_line := scanner.Text()
		// If batch is full, send it
		if !betBatch.AddBetLine(bet_line) {
			log.Debugf("action: sending_batch | result: in_progress | batch: %d",
				betBatch.GetBetCount(),
			)
			if err := c.sendBatch(betBatch); err != nil {
				log.Errorf("action: send_batch | result: fail | client_id: %v | error: %v",
					c.config.ID,
					err,
				)
				c.conn.Close()
				return err
			}
			betBatch = protocol.NewBatch(agencyID, c.config.MaxBatchAmount)
			betBatch.AddBetLine(bet_line)
		}
	}
	if err := scanner.Err(); err != nil {
		log.Fatal(err)
		return err
	}
	if c.hasSignal() {
		c.sendTerminate()
		return nil
	}

	// send last batch
	if err := c.sendBatch(betBatch); err != nil {
		log.Errorf("action: send_batch | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		c.conn.Close()
		return err
	}
	c.sendAgencyReady(agencyID)
	c.sendTerminate()
	return nil
}

func (c *Client) sendBatch(batch *protocol.BetBatchRegister) error {
	data := batch.ToBytes()
	// Send the data to the server
	if err := c.FullWrite(data); err != nil {
		return fmt.Errorf("error while sending batch: %v", err)
	}

	// Await response
	responseData, err := c.ReadWithPayloadLength()
	if err != nil {
		return fmt.Errorf("error while reading response: %v", err)
	}

	// Check confirmation
	confirmation, err := protocol.DeserializeConfirmation(responseData)
	if err != nil {
		return fmt.Errorf("error while deserializing confirmation: %v", err)
	}

	if confirmation.Success {
		log.Debugf("action: send_batch | result: success | client_id: %v | confirmation: %v",
			c.config.ID,
			confirmation.Message,
		)
	}
	return nil
}

func (c *Client) sendAgencyReady(agencyID uint8) {
	msg := []byte{byte(protocol.READY), 0x00, 0x00, 0x00, 0x00, agencyID}
	c.FullWrite(msg)
	log.Infof("action: send_agency_ready | result: success")
	c.conn.Close()
}

func (c *Client) sendTerminate() {
	msg := []byte{byte(protocol.TERMINATE), 0x00, 0x00, 0x00, 0x00}
	c.FullWrite(msg)
	log.Infof("action: send_terminate | result: success")
	c.conn.Close()
}
