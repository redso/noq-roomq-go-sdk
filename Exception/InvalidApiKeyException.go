package NoQ_RoomQ_Exception

type InvalidApiKeyException struct{}

func (ex *InvalidApiKeyException) Error() string {
	return "Invalid api key"
}
