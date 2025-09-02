package common

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("log")

// ClientConfig Configuration used by the client
type ClientConfig struct {
	ID            string
	ServerAddress string
	LoopAmount    int
	LoopPeriod    time.Duration
}

// Client Entity that encapsulates how
type Client struct {
	config ClientConfig
	conn   net.Conn
}

// NewClient Initializes a new client receiving the configuration
// as a parameter
func NewClient(config ClientConfig) *Client {
	client := &Client{
		config: config,
	}
	return client
}

// CreateClientSocket Initializes client socket. In case of
// failure, error is printed in stdout/stderr and exit 1
// is returned
func (c *Client) createClientSocket() error {
	conn, err := net.Dial("tcp", c.config.ServerAddress)
	if err != nil {
		log.Criticalf(
			"action: connect | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
	}
	c.conn = conn
	return nil
}

// StartClientLoop Send messages to the client until some time threshold is met
func (c *Client) StartClientLoop(sigterm_channel chan os.Signal) {
	// There is an autoincremental msgID to identify every message sent
	// Messages if the message amount threshold has not been surpassed
	for msgID := 1; msgID <= c.config.LoopAmount; msgID++ {
		// Create the connection the server in every loop iteration. Send an
		select {
		case <-sigterm_channel:
			log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
			return
		default:
			exit := sendMessage(c, msgID)
			if exit {
				return
			}
		}
	}
	log.Infof("action: loop_finished | result: success | client_id: %v", c.config.ID)
}

func sendMessage(c *Client, msgID int) bool {
	c.createClientSocket()

	// TODO: Modify the send to avoid short-write
	msg_to_send := []byte(fmt.Sprintf("[CLIENT %v] Message N°%v\n", c.config.ID, msgID))
	fullWrite(c, msg_to_send)

	msg, err := bufio.NewReader(c.conn).ReadString('\n')
	c.conn.Close()

	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return true
	}

	log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
		c.config.ID,
		msg,
	)

	// Wait a time between sending one message and the next one
	time.Sleep(c.config.LoopPeriod)
	return false
}

// ------- Ej5 --------

func (c *Client) SendBetInfo(betInfo string) error {
	msg := []byte(fmt.Sprintf("[CLIENT %v] Bet Info: %v\n", c.config.ID, betInfo))

	c.createClientSocket()
	if err := fullWrite(c, msg); err != nil {
		return err
	}

	msgRcv, err := bufio.NewReader(c.conn).ReadString('\n')
	c.conn.Close()

	if err != nil {
		log.Errorf("action: receive_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return nil
	}

	log.Infof("action: receive_message | result: success | client_id: %v | msg: %v",
		c.config.ID,
		msgRcv,
	)
	return nil
}

// Sends a message to the server making sure to write the full message
func fullWrite(c *Client, msg []byte) error {
	for written := 0; written < len(msg); {
		n, err := c.conn.Write(msg[written:])
		if err != nil {
			return err
		}
		written += n
	}
	return nil
}
