import { useI18n } from '@/contexts/I18nContext';
import { register } from '@/services/auth';
import { useAuthStore } from '@/stores/authStore';
import { Link, history, useSearchParams } from '@umijs/max';
import { notification } from 'antd';
import { useState } from 'react';
import AuthCard from './components/AuthCard';
import EmailInput from './components/EmailInput';
import PasswordInput from './components/PasswordInput';

const RegisterPage = () => {
  const { t } = useI18n();
  const setSession = useAuthStore((state) => state.setSession);
  const [searchParams] = useSearchParams();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const redirect = searchParams.get('redirect');
  const loginPath = redirect
    ? `/login?redirect=${encodeURIComponent(redirect)}`
    : '/login';
  const passwordMatches = password === confirmPassword;
  const showPasswordMismatch = confirmPassword !== '' && !passwordMatches;
  const canSubmit =
    email.trim() !== '' &&
    password !== '' &&
    confirmPassword !== '' &&
    passwordMatches &&
    !submitting;

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    if (!canSubmit) {
      return;
    }

    setSubmitting(true);

    try {
      const result = await register({
        email: email.trim(),
        password,
      });
      setSession(result);
      notification.success({
        message: t('registerSuccess'),
        placement: 'topRight',
      });
      history.push(searchParams.get('redirect') || '/admin');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <AuthCard title={t('registerTitle')}>
      <form className="grid gap-4" onSubmit={handleSubmit}>
        <EmailInput value={email} onChange={setEmail} />
        <PasswordInput
          id="register-password"
          autoComplete="new-password"
          value={password}
          onChange={setPassword}
        />
        <div className="grid gap-2">
          <PasswordInput
            id="confirm-password"
            autoComplete="new-password"
            label={t('confirmPassword')}
            placeholder={t('confirmPasswordPlaceholder')}
            value={confirmPassword}
            onChange={setConfirmPassword}
          />
          {showPasswordMismatch ? (
            <p className="text-sm text-red-600 dark:text-red-300">
              {t('passwordMismatch')}
            </p>
          ) : null}
        </div>

        <button
          className="min-h-11 rounded-md bg-linkflow-primary px-4 font-bold text-white transition duration-200 hover:-translate-y-0.5 hover:bg-linkflow-primary-hover hover:shadow-lg hover:shadow-linkflow-primary-soft active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-55 disabled:hover:translate-y-0 disabled:hover:shadow-none dark:bg-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-hover"
          type="submit"
          disabled={!canSubmit}
        >
          {submitting ? t('registering') : t('registerButton')}
        </button>

        <p className="text-center text-sm text-linkflow-muted dark:text-linkflow-dark-muted">
          {t('hasAccount')}{' '}
          <Link
            to={loginPath}
            className="font-semibold text-linkflow-primary transition hover:text-linkflow-primary-hover dark:text-linkflow-dark-primary dark:hover:text-linkflow-dark-primary-hover"
          >
            {t('goLogin')}
          </Link>
        </p>
      </form>
    </AuthCard>
  );
};

export default RegisterPage;
