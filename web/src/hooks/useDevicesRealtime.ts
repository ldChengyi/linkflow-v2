import { useEffect, useRef, useState } from 'react';

export const DEVICE_REALTIME_PATH = '/api/v1/ws/devices';

export type DeviceConnectionStatus = 'online' | 'offline';

export interface DeviceConnectionChangedPayload {
  tenant_id: string;
  product_id: string;
  device_id: string;
  tenant_slug: string;
  product_key: string;
  device_slug: string;
  status: DeviceConnectionStatus;
  reason?: string;
}

export interface DevicePropertyChangedPayload {
  tenant_id: string;
  product_id: string;
  device_id: string;
  tenant_slug: string;
  product_key: string;
  device_slug: string;
  properties: Record<string, unknown>;
}

interface DeviceEnvelope<T> {
  event_id: string;
  event_type: string;
  event_version: number;
  occurred_at: string;
  producer: string;
  tenant_id: string;
  payload: T;
}

export type DeviceRealtimeStatus =
  | 'idle'
  | 'connecting'
  | 'open'
  | 'reconnecting'
  | 'closed';

export interface UseDevicesRealtimeOptions {
  enabled?: boolean;
  onConnectionChanged?: (
    payload: DeviceConnectionChangedPayload,
    envelope: { occurredAt: string; tenantId: string },
  ) => void;
  onPropertyChanged?: (
    payload: DevicePropertyChangedPayload,
    envelope: { occurredAt: string; tenantId: string },
  ) => void;
}

export interface UseDevicesRealtimeResult {
  status: DeviceRealtimeStatus;
}

const minBackoffMs = 1000;
const maxBackoffMs = 30000;
const unauthorizedCloseCode = 1008;

const buildEndpoint = () => {
  if (typeof window === 'undefined') {
    return '';
  }
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const { hostname, host, port } = window.location;
  const wsHost = port === '8000' ? `${hostname}:18080` : host;
  return `${protocol}//${wsHost}${DEVICE_REALTIME_PATH}`;
};

export const useDevicesRealtime = (
  options: UseDevicesRealtimeOptions = {},
): UseDevicesRealtimeResult => {
  const { enabled = true } = options;
  const optionsRef = useRef(options);
  optionsRef.current = options;

  const [status, setStatus] = useState<DeviceRealtimeStatus>('idle');

  useEffect(() => {
    if (!enabled || typeof window === 'undefined') {
      setStatus('idle');
      return;
    }

    let cancelled = false;
    let socket: WebSocket | null = null;
    let reconnectTimer: number | null = null;
    let backoffMs = minBackoffMs;

    const cleanup = () => {
      if (reconnectTimer !== null) {
        window.clearTimeout(reconnectTimer);
        reconnectTimer = null;
      }
      if (socket) {
        socket.onopen = null;
        socket.onmessage = null;
        socket.onerror = null;
        socket.onclose = null;
        if (
          socket.readyState === WebSocket.OPEN ||
          socket.readyState === WebSocket.CONNECTING
        ) {
          socket.close(1000, 'unmount');
        }
        socket = null;
      }
    };

    const scheduleReconnect = () => {
      if (cancelled) {
        return;
      }
      setStatus('reconnecting');
      reconnectTimer = window.setTimeout(() => {
        backoffMs = Math.min(backoffMs * 2, maxBackoffMs);
        connect();
      }, backoffMs);
    };

    const handleMessage = (raw: string) => {
      let envelope: DeviceEnvelope<unknown>;
      try {
        envelope = JSON.parse(raw) as DeviceEnvelope<unknown>;
      } catch (error) {
        return;
      }
      if (!envelope || typeof envelope.event_type !== 'string') {
        return;
      }
      const meta = {
        occurredAt: envelope.occurred_at,
        tenantId: envelope.tenant_id,
      };
      const handlers = optionsRef.current;
      switch (envelope.event_type) {
        case 'device.connection.changed':
          handlers.onConnectionChanged?.(
            envelope.payload as DeviceConnectionChangedPayload,
            meta,
          );
          return;
        case 'device.property.changed':
          handlers.onPropertyChanged?.(
            envelope.payload as DevicePropertyChangedPayload,
            meta,
          );
          return;
        default:
          return;
      }
    };

    const connect = () => {
      if (cancelled) {
        return;
      }
      const endpoint = buildEndpoint();
      if (!endpoint) {
        setStatus('idle');
        return;
      }

      setStatus('connecting');
      try {
        socket = new WebSocket(endpoint);
      } catch (error) {
        scheduleReconnect();
        return;
      }

      socket.onopen = () => {
        if (cancelled) {
          return;
        }
        backoffMs = minBackoffMs;
        setStatus('open');
      };

      socket.onmessage = (event) => {
        if (typeof event.data === 'string') {
          handleMessage(event.data);
        }
      };

      socket.onerror = () => {
        // onclose will follow; rely on it to decide whether to reconnect.
      };

      socket.onclose = (event) => {
        socket = null;
        if (cancelled) {
          return;
        }
        if (event.code === unauthorizedCloseCode) {
          setStatus('closed');
          return;
        }
        scheduleReconnect();
      };
    };

    connect();

    return () => {
      cancelled = true;
      cleanup();
      setStatus('idle');
    };
  }, [enabled]);

  return { status };
};
