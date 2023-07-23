package camera

const TypeStreamingStatus = "120"

type StreamingStatus struct {
	Status byte `tlv8:"1"`
}

const (
	StreamingStatusAvailable   = 0
	StreamingStatusBusy        = 1
	StreamingStatusUnavailable = 2
)
