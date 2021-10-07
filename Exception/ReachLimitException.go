package NoQ_RoomQ_Exception

type ReachLimitException struct{}

func (ex *ReachLimitException) Error() string {
	return "Reach limit"
}
