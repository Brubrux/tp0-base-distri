package common

func (c *Client) SendBetsAwaitWinners(filePath string, agencyID uint8) error {
	c.SendBetBatch(filePath, agencyID)

	return nil
}
