package NoQ_RoomQ_Exception

type NotServingException struct{}

func (ex *NotServingException) Error() string {
	return "Not serving"
}
