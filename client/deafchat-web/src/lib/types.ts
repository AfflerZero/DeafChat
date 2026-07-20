export type MessageType =
  | "join"
  | "leave"
  | "ready"
  | "offer"
  | "answer"
  | "candidate"
  | "error"
  | "peer_count";

export interface SignalMessage {
  type: MessageType;
  room: string;
  payload?: string;
  sender_id?: string;
}

export interface SDPPayload {
  sdp: RTCSessionDescriptionInit;
}

export interface ICECandidatePayload {
  candidate: RTCIceCandidateInit;
}

export interface ErrorPayload {
  message: string;
}

export interface PeerCountPayload {
  count: number;
}

export interface LeavePayload {
  name: string;
}

export interface LoginRequest {
  email: string;
  password: string;
}

export interface TokenResponse {
  authentication_token: {
    token: string;
    expiry: string;
  };
}

export interface User {
  id: number;
  created_at: string;
  name: string;
  email: string;
  activated: boolean;
}
