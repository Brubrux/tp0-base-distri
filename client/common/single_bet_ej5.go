package common

import "fmt"

// ------- Ej5 --------

// Sends a message to the server making sure to write the full message
func (c *Client) FullWrite(msg []byte) error {
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
func (c *Client) FullRead(size int) ([]byte, error) {
	buffer := make([]byte, size)
	totalRead := 0

	for totalRead < size {
		if c.hasSignal() {
			return nil, fmt.Errorf("SIGTERM received")
		}

		resultChan := make(chan int)
		errorChan := make(chan error)

		go func() {
			n, err := c.conn.Read(buffer[totalRead:])
			if err != nil {
				errorChan <- err
			} else {
				resultChan <- n
			}
		}()

		select {
		case n := <-resultChan:
			totalRead += n
		case err := <-errorChan:
			return nil, err
		case <-c.signalChan:
			c.shuttingDown = true
			return nil, fmt.Errorf("SIGTERM received")
		}
	}

	return buffer, nil
}

// Reads a complete message using the payload length field
func (c *Client) ReadWithPayloadLength() ([]byte, error) {

	// 1 byte
	opCodeData, err := c.FullRead(1)
	if err != nil {
		return nil, err
	}

	// 4 bytes, big-endian
	payloadLengthData, err := c.FullRead(4)
	if err != nil {
		return nil, err
	}
	payloadLength := parseUint32BigEndian(payloadLengthData)

	payloadData, err := c.FullRead(int(payloadLength))
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
