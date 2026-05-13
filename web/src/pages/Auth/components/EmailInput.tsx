import { useI18n } from '@/contexts/I18nContext';
import { CloseOutlined } from '@ant-design/icons';

interface EmailInputProps {
  value: string;
  onChange: (value: string) => void;
}

const EmailInput = ({ value, onChange }: EmailInputProps) => {
  const { t } = useI18n();

  return (
    <label className="grid gap-2 font-semibold" htmlFor="email">
      <span>{t('email')}</span>
      <div className="group flex min-h-11 items-center rounded-md border border-linkflow-border bg-white px-3 transition hover:-translate-y-0.5 hover:border-linkflow-primary focus-within:border-linkflow-primary focus-within:ring-4 focus-within:ring-linkflow-primary-soft dark:border-linkflow-dark-border dark:bg-slate-900 dark:focus-within:ring-linkflow-dark-primary-soft">
        <input
          id="email"
          name="email"
          type="email"
          autoComplete="email"
          value={value}
          onChange={(event) => onChange(event.target.value)}
          placeholder={t('emailPlaceholder')}
          className="min-w-0 flex-1 bg-transparent font-normal outline-none placeholder:text-linkflow-subtle dark:placeholder:text-linkflow-dark-subtle"
        />

        {value !== '' && (
          <button
            type="button"
            onClick={() => onChange('')}
            className="ml-2 grid h-7 w-7 place-items-center rounded-full text-linkflow-subtle opacity-70 transition hover:bg-slate-100 hover:text-linkflow-text group-hover:opacity-100 dark:text-linkflow-dark-subtle dark:hover:bg-slate-800 dark:hover:text-linkflow-dark-text"
            aria-label={t('clearEmail')}
          >
            <CloseOutlined aria-hidden="true" />
          </button>
        )}
      </div>
    </label>
  );
};

export default EmailInput;
