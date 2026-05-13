import { adminMessages, type AdminMessageKey } from '@/i18n/admin';
import { AppLocale, AuthMessageKey, authMessages } from '@/i18n/auth';
import { useAdminUiStore } from '@/stores/adminUiStore';
import { createContext, useContext, useMemo } from 'react';

type MessageKey = AuthMessageKey | AdminMessageKey;

const messages = {
  'zh-CN': {
    ...authMessages['zh-CN'],
    ...adminMessages['zh-CN'],
  },
  'en-US': {
    ...authMessages['en-US'],
    ...adminMessages['en-US'],
  },
} as const;

interface I18nContextValue {
  locale: AppLocale;
  setLocale: React.Dispatch<React.SetStateAction<AppLocale>>;
  t: (key: MessageKey) => string;
}

const I18nContext = createContext<I18nContextValue | null>(null);

interface I18nProviderProps {
  children: React.ReactNode;
}

export const I18nProvider = ({ children }: I18nProviderProps) => {
  const locale = useAdminUiStore((state) => state.locale);
  const setLocale = useAdminUiStore((state) => state.setLocale);

  const value = useMemo(
    () => ({
      locale,
      setLocale,
      t: (key: MessageKey) => messages[locale][key],
    }),
    [locale],
  );

  return <I18nContext.Provider value={value}>{children}</I18nContext.Provider>;
};

export const useI18n = () => {
  const value = useContext(I18nContext);

  if (value === null) {
    throw new Error('useI18n must be used inside I18nProvider');
  }

  return value;
};
