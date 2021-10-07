package NoQ_RoomQ_Exception

type QueueStoppedException struct{}

func (ex *QueueStoppedException) Error() string {
	return "Queue is stopped"
}
