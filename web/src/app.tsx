import { history } from '@umijs/max';
import { notification } from 'antd';
import { I18nProvider } from './contexts/I18nContext';
import { ThemeProvider } from './contexts/ThemeContext';
import { useAuthStore } from './stores/authStore';

interface ApiErrorBody {
  code?: number;
  msg?: string;
}

interface RequestErrorLike {
  message?: string;
  response?: {
    status?: number;
    data?: unknown;
  };
}

interface RequestInterceptorConfig {
  headers?: Record<string, unknown>;
  [key: string]: unknown;
}

interface RequestRuntimeConfig {
  errorConfig: {
    errorHandler: (error: unknown) => void;
  };
  requestInterceptors: Array<
    (config: RequestInterceptorConfig) => RequestInterceptorConfig
  >;
}

const isApiErrorBody = (value: unknown): value is ApiErrorBody => {
  if (typeof value !== 'object' || value === null) {
    return false;
  }

  return 'msg' in value || 'code' in value;
};

const getRequestErrorMessage = (error: unknown) => {
  const requestError = error as RequestErrorLike;
  const body = requestError.response?.data;

  if (isApiErrorBody(body) && typeof body.msg === 'string') {
    return body.msg;
  }

  if (typeof requestError.message === 'string' && requestError.message !== '') {
    return requestError.message;
  }

  return 'Request failed';
};

const getRequestErrorStatus = (error: unknown) => {
  const requestError = error as RequestErrorLike;
  const body = requestError.response?.data;

  if (typeof requestError.response?.status === 'number') {
    return requestError.response.status;
  }

  if (isApiErrorBody(body) && typeof body.code === 'number') {
    return body.code;
  }

  return undefined;
};

const redirectToLogin = () => {
  const pathname = history.location.pathname;

  if (pathname === '/login' || pathname === '/register') {
    return;
  }

  const redirect = `${pathname}${history.location.search}${history.location.hash}`;
  useAuthStore.getState().clearSession();
  history.push(`/login?redirect=${encodeURIComponent(redirect)}`);
};

export function rootContainer(container: React.ReactNode) {
  return (
    <I18nProvider>
      <ThemeProvider>{container}</ThemeProvider>
    </I18nProvider>
  );
}

// 全局初始化数据配置，用于 Layout 用户信息和权限初始化
// 更多信息见文档：https://umijs.org/docs/api/runtime-config#getinitialstate
export async function getInitialState(): Promise<{ name: string }> {
  return { name: '@umijs/max' };
}

export const layout = () => {
  return {
    logo: 'https://img.alicdn.com/tfs/TB1YHEpwUT1gK0jSZFhXXaAtVXa-28-27.svg',
    menu: {
      locale: false,
    },
  };
};

export const request: RequestRuntimeConfig = {
  errorConfig: {
    errorHandler: (error) => {
      if (getRequestErrorStatus(error) === 401) {
        redirectToLogin();
        return;
      }

      notification.error({
        message: 'Request failed',
        description: getRequestErrorMessage(error),
        placement: 'topRight',
      });
    },
  },
  requestInterceptors: [
    (config) => {
      const token = useAuthStore.getState().accessToken;

      if (!token) {
        return config;
      }

      return {
        ...config,
        headers: {
          ...config.headers,
          Authorization: `Bearer ${token}`,
        },
      };
    },
  ],
};
