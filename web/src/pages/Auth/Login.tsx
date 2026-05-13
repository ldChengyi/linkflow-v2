import { useI18n } from '@/contexts/I18nContext';
import { login } from '@/services/auth';
import { useAuthStore } from '@/stores/authStore';
import { Link, history, useSearchParams } from '@umijs/max';
import { notification } from 'antd';
import { useState } from 'react';
import AuthCard from './components/AuthCard';
import EmailInput from './components/EmailInput';
import PasswordInput from './components/PasswordInput';

const LoginPage = () => {
  const { t } = useI18n();
  const setSession = useAuthStore((state) => state.setSession);
  const [searchParams] = useSearchParams();
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [submitting, setSubmitting] = useState(false);

  const redirect = searchParams.get('redirect');
  const registerPath = redirect
    ? `/register?redirect=${encodeURIComponent(redirect)}`
    : '/register';
  const canSubmit = email.trim() !== '' && password !== '' && !submitting;

  const handleSubmit = async (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();

    if (!canSubmit) {
      return;
    }

    setSubmitting(true);

    try {
      const result = await login({
        email: email.trim(),
        password,
      });
      setSession(result);
      notification.success({
        message: t('loginSuccess'),
        placement: 'topRight',
      });
      history.push(searchParams.get('redirect') || '/admin');
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <AuthCard title={t('loginTitle')}>
      <form className="grid gap-4" onSubmit={handleSubmit}>
        <EmailInput value={email} onChange={setEmail} />
        <PasswordInput id="password" value={password} onChange={setPassword} />

        <button
          className="min-h-11 rounded-md bg-linkflow-primary px-4 font-bold text-white transition duration-200 hover:-translate-y-0.5 hover:bg-linkflow-primary-hover hover:shadow-lg hover:shadow-linkflow-primary-soft active:scale-[0.98] disabled:cursor-not-allowed disabled:opacity-55 disabled:hover:translate-y-0 disabled:hover:shadow-none dark:bg-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-hover"
          type="submit"
          disabled={!canSubmit}
        >
          {submitting ? t('loggingIn') : t('loginButton')}
        </button>

        <p className="text-center text-sm text-linkflow-muted dark:text-linkflow-dark-muted">
          {t('noAccount')}{' '}
          <Link
            to={registerPath}
            className="font-semibold text-linkflow-primary transition hover:text-linkflow-primary-hover dark:text-linkflow-dark-primary dark:hover:text-linkflow-dark-primary-hover"
          >
            {t('goRegister')}
          </Link>
        </p>
      </form>
    </AuthCard>
  );
};

export default LoginPage;
