package NoQ_RoomQ_Exception

type InvalidTokenException struct{}

func (ex *InvalidTokenException) Error() string {
	return "Invalid Token"
}
