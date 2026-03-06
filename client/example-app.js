let localStream, pc;
const room = prompt("Enter Room ID:") || "default";
const protocol = window.location.protocol === "https:" ? "wss" : "ws";
const socket = new WebSocket(`${protocol}://${window.location.host}/signal`);

const iceConfig = {
  iceServers: [
    // Replace with self-hosted STUN
    { urls: "stun:stun.l.google.com:19302" }
    //{ urls: "136.36.253.240 " }
    //{ urls: "stun:192.168.1.133:3478" } // on LAN
  ]
};

let isOfferer = false;

socket.onopen = () => {
  console.log("[Socket] Connected");
  socket.send(JSON.stringify({ type: "join", room }));
  console.log("[Socket] Sent join request for room:", room);
};

socket.onmessage = async (event) => {
  const msg = JSON.parse(event.data);
  console.log("[Socket] Message received:", msg);

  if (msg.room !== room) return;

  switch (msg.type) {
    case "ready":
      console.log("[Room] A second peer joined — acting as offerer");
      isOfferer = true;
      await startLocal();
      await createPeerConnection();
      await createOffer();
      break;

    case "offer":
      console.log("[Signal] Received offer");
      await startLocal();
      await createPeerConnection();
      await pc.setRemoteDescription(new RTCSessionDescription(msg.sdp));
      const answer = await pc.createAnswer();
      await pc.setLocalDescription(answer);
      socket.send(JSON.stringify({ type: "answer", sdp: pc.localDescription, room }));
      break;

    case "answer":
      console.log("[Signal] Received answer");
      await pc.setRemoteDescription(new RTCSessionDescription(msg.sdp));
      break;

    case "candidate":
      console.log("[Signal] Received ICE candidate");
      if (pc) {
        await pc.addIceCandidate(new RTCIceCandidate(msg.candidate));
      }
      break;
  }
};

async function startLocal() {
  if (localStream) return;

  console.log("[Media] Requesting camera/mic...");
  localStream = await navigator.mediaDevices.getUserMedia({ video: true, audio: true });
  document.getElementById("localVideo").srcObject = localStream;
  console.log("[Media] Local stream ready");
}

async function createPeerConnection() {
  if (pc) return;

  console.log("[Peer] Creating RTCPeerConnection...");
  pc = new RTCPeerConnection(iceConfig);

  localStream.getTracks().forEach((track) => pc.addTrack(track, localStream));

  pc.onicecandidate = (event) => {
    if (event.candidate) {
      console.log("[Peer] Sending ICE candidate");
      socket.send(JSON.stringify({ type: "candidate", candidate: event.candidate, room }));
    }
  };
  pc.oniceconnectionstatechange = () => {
    console.log("[ICE] Connection State:", pc.iceConnectionState);
  };
  pc.ontrack = (event) => {
    console.log("[Peer] Received remote track");
    document.getElementById("remoteVideo").srcObject = event.streams[0];
  };
}

async function createOffer() {
  const offer = await pc.createOffer();
  await pc.setLocalDescription(offer);
  console.log("[Peer] Sending offer");
  socket.send(JSON.stringify({ type: "offer", sdp: pc.localDescription, room }));
}

