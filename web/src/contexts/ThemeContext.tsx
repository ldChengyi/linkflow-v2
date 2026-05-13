import { useAdminUiStore } from '@/stores/adminUiStore';
import { createContext, useContext, useMemo } from 'react';

interface ThemeContextValue {
  darkMode: boolean;
  setDarkMode: React.Dispatch<React.SetStateAction<boolean>>;
  toggleDarkMode: () => void;
}

const ThemeContext = createContext<ThemeContextValue | null>(null);

interface ThemeProviderProps {
  children: React.ReactNode;
}

export const ThemeProvider = ({ children }: ThemeProviderProps) => {
  const darkMode = useAdminUiStore((state) => state.darkMode);
  const setDarkMode = useAdminUiStore((state) => state.setDarkMode);
  const toggleDarkMode = useAdminUiStore((state) => state.toggleDarkMode);

  const value = useMemo(
    () => ({
      darkMode,
      setDarkMode,
      toggleDarkMode,
    }),
    [darkMode, setDarkMode, toggleDarkMode],
  );

  return (
    <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
  );
};

export const useTheme = () => {
  const value = useContext(ThemeContext);

  if (value === null) {
    throw new Error('useTheme must be used inside ThemeProvider');
  }

  return value;
};
