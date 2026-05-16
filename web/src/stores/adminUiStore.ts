import type { AppLocale } from '@/i18n/auth';
import { create } from 'zustand';
import { persist } from 'zustand/middleware';

export type AdminSectionId =
  | 'overview'
  | 'tenants'
  | 'products'
  | 'thingsModels'
  | 'devices'
  | 'auditLogs'
  | 'settings';

const validSectionIds: AdminSectionId[] = [
  'overview',
  'tenants',
  'products',
  'thingsModels',
  'devices',
  'auditLogs',
  'settings',
];

interface AdminUiState {
  darkMode: boolean;
  locale: AppLocale;
  sidebarExpanded: boolean;
  visitedTabs: AdminSectionId[];
  addVisitedTab: (id: AdminSectionId) => void;
  removeVisitedTab: (id: AdminSectionId) => void;
  setDarkMode: (value: boolean | ((current: boolean) => boolean)) => void;
  setLocale: (value: AppLocale | ((current: AppLocale) => AppLocale)) => void;
  setSidebarExpanded: (
    value: boolean | ((current: boolean) => boolean),
  ) => void;
  toggleDarkMode: () => void;
}

const resolveValue = <T>(value: T | ((current: T) => T), current: T) => {
  if (typeof value === 'function') {
    return (value as (current: T) => T)(current);
  }

  return value;
};

const isAdminSectionId = (value: unknown): value is AdminSectionId => {
  return (
    typeof value === 'string' &&
    validSectionIds.includes(value as AdminSectionId)
  );
};

const normalizeVisitedTabs = (value: unknown): AdminSectionId[] => {
  if (!Array.isArray(value)) {
    return ['overview'];
  }

  const tabs = value.filter(isAdminSectionId);
  return tabs.length > 0 ? tabs : ['overview'];
};

export const useAdminUiStore = create<AdminUiState>()(
  persist(
    (set) => ({
      darkMode: false,
      locale: 'zh-CN',
      sidebarExpanded: false,
      visitedTabs: ['overview'],

      addVisitedTab: (id) => {
        set((state) => {
          if (state.visitedTabs.includes(id)) {
            return state;
          }

          return {
            visitedTabs: [...state.visitedTabs, id],
          };
        });
      },

      removeVisitedTab: (id) => {
        set((state) => {
          const visitedTabs = state.visitedTabs.filter((tabId) => tabId !== id);

          return {
            visitedTabs: visitedTabs.length > 0 ? visitedTabs : ['overview'],
          };
        });
      },

      setDarkMode: (value) => {
        set((state) => ({
          darkMode: resolveValue(value, state.darkMode),
        }));
      },

      setLocale: (value) => {
        set((state) => ({
          locale: resolveValue(value, state.locale),
        }));
      },

      setSidebarExpanded: (value) => {
        set((state) => ({
          sidebarExpanded: resolveValue(value, state.sidebarExpanded),
        }));
      },

      toggleDarkMode: () => {
        set((state) => ({
          darkMode: !state.darkMode,
        }));
      },
    }),
    {
      name: 'linkflow.admin-ui',
      merge: (persistedState, currentState) => {
        if (typeof persistedState !== 'object' || persistedState === null) {
          return currentState;
        }

        const state = persistedState as Partial<AdminUiState>;

        return {
          ...currentState,
          ...state,
          visitedTabs: normalizeVisitedTabs(state.visitedTabs),
        };
      },
    },
  ),
);
