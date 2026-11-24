package signal

import "encoding/json"

type MessageType string

const (
	MsgJoin      MessageType = "join"
	MsgLeave     MessageType = "leave"
	MsgReady     MessageType = "ready"
	MsgOffer     MessageType = "offer"
	MsgAnswer    MessageType = "answer"
	MsgCandidate MessageType = "candidate"
	MsgError     MessageType = "error"
	MsgPeerCount MessageType = "peer_count"
)

type Message struct {
	Type      MessageType     `json:"type"`
	Room      string          `json:"room"`
	Payload   json.RawMessage `json:"payload,omitempty"`
	SenderID  string          `json:"sender_id,omitempty"`
}

type JoinPayload struct {
	Name string `json:"name"`
}

type SDPPayload struct {
	SDP json.RawMessage `json:"sdp"`
}

type ICECandidatePayload struct {
	Candidate json.RawMessage `json:"candidate"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}

type PeerCountPayload struct {
	Count int `json:"count"`
}
