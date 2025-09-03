package common

import (
	"net"

	"github.com/7574-sistemas-distribuidos/docker-compose-init/client/protocol"
)

// ------- Ej5 --------

// Sends a bet registration message to the server and awaits confirmation
func (c *Client) SendBetRegister(b protocol.BetRegister) {

	c.createClientSocket()

	// Serialize and send
	msg := b.ToBytes()
	if err := FullWrite(c, msg); err != nil {
		log.Errorf("action: send_message | result: fail | client_id: %v | error: %v",
			c.config.ID,
			err,
		)
		return
	}

	// Await response
	responseData, err := readWithPayloadLength(c.conn)
	c.conn.Close()

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
		log.Infof("action: apuesta_enviada | result: success | dni: %v | numero: %v",
			b.ID,
			b.Number,
		)
	}
}

// Sends a message to the server making sure to write the full message
func FullWrite(c *Client, msg []byte) error {
	for written := 0; written < len(msg); {
		n, err := c.conn.Write(msg[written:])
		if err != nil {
			return err
		}
		written += n
	}
	return nil
}

// Reads exactly size bytes from the connection, handling short-reads
func fullRead(conn net.Conn, size int) ([]byte, error) {
	buffer := make([]byte, size)
	totalRead := 0

	for totalRead < size {
		n, err := conn.Read(buffer[totalRead:])
		if err != nil {
			return nil, err
		}
		totalRead += n
	}
	return buffer, nil
}

// Reads a complete message using the payload length field
func readWithPayloadLength(conn net.Conn) ([]byte, error) {

	// 1 byte
	opCodeData, err := fullRead(conn, 1)
	if err != nil {
		return nil, err
	}

	// 4 bytes, big-endian
	payloadLengthData, err := fullRead(conn, 4)
	if err != nil {
		return nil, err
	}
	payloadLength := parseUint32BigEndian(payloadLengthData)

	payloadData, err := fullRead(conn, int(payloadLength))
	if err != nil {
		return nil, err
	}

	fullMessage := make([]byte, 0, 1+4+len(payloadData))
	fullMessage = append(fullMessage, opCodeData...)
	fullMessage = append(fullMessage, payloadLengthData...)
	fullMessage = append(fullMessage, payloadData...)

	return fullMessage, nil
}

func parseUint32BigEndian(data []byte) uint32 {
	if len(data) < 4 {
		return 0
	}
	return uint32(data[0])<<24 |
		uint32(data[1])<<16 |
		uint32(data[2])<<8 |
		uint32(data[3])
}
