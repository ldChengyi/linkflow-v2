import { useI18n } from '@/contexts/I18nContext';
import {
  CloseOutlined,
  EyeInvisibleOutlined,
  EyeOutlined,
} from '@ant-design/icons';
import { useState } from 'react';

interface PasswordInputProps {
  autoComplete?: string;
  id: string;
  label?: string;
  placeholder?: string;
  value: string;
  onChange: (value: string) => void;
}

const PasswordInput = ({
  autoComplete = 'current-password',
  id,
  label,
  placeholder,
  value,
  onChange,
}: PasswordInputProps) => {
  const { t } = useI18n();
  const [showPassword, setShowPassword] = useState(false);

  return (
    <label className="grid gap-2 font-semibold" htmlFor={id}>
      <span>{label ?? t('password')}</span>
      <div className="group flex min-h-11 items-center rounded-md border border-linkflow-border bg-white px-3 transition hover:-translate-y-0.5 hover:border-linkflow-primary focus-within:border-linkflow-primary focus-within:ring-4 focus-within:ring-linkflow-primary-soft dark:border-linkflow-dark-border dark:bg-slate-900 dark:focus-within:ring-linkflow-dark-primary-soft">
        <input
          id={id}
          name="password"
          type={showPassword ? 'text' : 'password'}
          autoComplete={autoComplete}
          value={value}
          onChange={(event) => onChange(event.target.value)}
          placeholder={placeholder ?? t('passwordPlaceholder')}
          className="min-w-0 flex-1 bg-transparent font-normal outline-none placeholder:text-linkflow-subtle dark:placeholder:text-linkflow-dark-subtle"
        />

        {value !== '' && (
          <button
            type="button"
            onClick={() => onChange('')}
            className="ml-2 grid h-7 w-7 place-items-center rounded-full text-linkflow-subtle opacity-70 transition hover:bg-slate-100 hover:text-linkflow-text group-hover:opacity-100 dark:text-linkflow-dark-subtle dark:hover:bg-slate-800 dark:hover:text-linkflow-dark-text"
            aria-label={t('clearPassword')}
          >
            <CloseOutlined aria-hidden="true" />
          </button>
        )}

        <button
          type="button"
          onClick={() => setShowPassword((current) => !current)}
          className="ml-2 grid h-8 w-8 place-items-center rounded-md text-linkflow-primary transition hover:bg-blue-50 active:scale-95 dark:text-linkflow-dark-primary dark:hover:bg-blue-950"
          aria-label={showPassword ? t('hidePassword') : t('showPassword')}
        >
          {showPassword ? (
            <EyeInvisibleOutlined aria-hidden="true" />
          ) : (
            <EyeOutlined aria-hidden="true" />
          )}
        </button>
      </div>
    </label>
  );
};

export default PasswordInput;
