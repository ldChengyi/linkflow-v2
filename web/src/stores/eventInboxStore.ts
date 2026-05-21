import type { DeviceEventReceivedPayload } from '@/hooks/useDevicesRealtime';
import { create } from 'zustand';

export interface EventInboxItem extends DeviceEventReceivedPayload {
  event_id: string;
  occurred_at: string;
  received_at: string;
  read: boolean;
}

interface EventInboxState {
  items: EventInboxItem[];
  unreadCount: number;
  pushEvent: (
    payload: DeviceEventReceivedPayload,
    meta: { eventId: string; occurredAt: string },
  ) => void;
  markAllRead: () => void;
  markRead: (eventID: string) => void;
}

const maxInboxItems = 20;

const countUnread = (items: EventInboxItem[]) => {
  return items.reduce((total, item) => total + (item.read ? 0 : 1), 0);
};

export const useEventInboxStore = create<EventInboxState>()((set) => ({
  items: [],
  unreadCount: 0,

  pushEvent: (payload, meta) => {
    set((state) => {
      if (state.items.some((item) => item.event_id === meta.eventId)) {
        return state;
      }

      const item: EventInboxItem = {
        ...payload,
        event_id: meta.eventId,
        occurred_at: meta.occurredAt,
        received_at: new Date().toISOString(),
        read: false,
      };
      const items = [item, ...state.items].slice(0, maxInboxItems);

      return {
        items,
        unreadCount: countUnread(items),
      };
    });
  },

  markAllRead: () => {
    set((state) => {
      const items = state.items.map((item) =>
        item.read ? item : { ...item, read: true },
      );
      return { items, unreadCount: 0 };
    });
  },

  markRead: (eventID) => {
    set((state) => {
      const items = state.items.map((item) =>
        item.event_id === eventID ? { ...item, read: true } : item,
      );
      return { items, unreadCount: countUnread(items) };
    });
  },
}));
