import { useCallback, useEffect, useRef, useState } from "react";
import type { SignalMessage } from "./types";

type ConnectionState = "disconnected" | "connecting" | "connected";

interface UseWebSocketOptions {
  room: string;
  onMessage?: (msg: SignalMessage) => void;
  autoReconnect?: boolean;
}

export function useWebSocket({
  room,
  onMessage,
  autoReconnect = true,
}: UseWebSocketOptions) {
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttempts = useRef(0);
  const onMessageRef = useRef(onMessage);
  const [state, setState] = useState<ConnectionState>("disconnected");

  useEffect(() => {
    onMessageRef.current = onMessage;
  });

  const disconnect = useCallback(() => {
    reconnectAttempts.current = 0;
    wsRef.current?.close();
    wsRef.current = null;
  }, []);

  useEffect(() => {
    if (!room) return;

    let cancelled = false;

    function connect() {
      if (cancelled) return;
      if (wsRef.current?.readyState === WebSocket.OPEN) return;

      const proto = window.location.protocol === "https:" ? "wss" : "ws";
      const url = `${proto}://${window.location.host}/v1/signal?room=${encodeURIComponent(room)}`;

      setState("connecting");
      const ws = new WebSocket(url);
      wsRef.current = ws;

      ws.onopen = () => {
        if (cancelled) return;
        setState("connected");
        reconnectAttempts.current = 0;
      };

      ws.onmessage = (event) => {
        if (cancelled) return;
        try {
          const msg: SignalMessage = JSON.parse(event.data);
          onMessageRef.current?.(msg);
        } catch {
          console.error("[ws] Failed to parse message");
        }
      };

      ws.onclose = () => {
        wsRef.current = null;
        if (cancelled) return;
        setState("disconnected");

        if (autoReconnect) {
          const delay = Math.min(
            1000 * 2 ** reconnectAttempts.current,
            30000,
          );
          reconnectAttempts.current += 1;
          setTimeout(connect, delay);
        }
      };

      ws.onerror = () => {
        ws.close();
      };
    }

    connect();

    return () => {
      cancelled = true;
      reconnectAttempts.current = 0;
      wsRef.current?.close();
      wsRef.current = null;
    };
  }, [room, autoReconnect]);

  const send = useCallback((msg: SignalMessage) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(msg));
    }
  }, []);

  return { state, send, disconnect };
}
