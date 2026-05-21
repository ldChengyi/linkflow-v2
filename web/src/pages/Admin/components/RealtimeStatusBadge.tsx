import { useI18n } from '@/contexts/I18nContext';
import type { DeviceRealtimeStatus } from '@/hooks/useDevicesRealtime';
import type { AdminMessageKey } from '@/i18n/admin';

export const realtimeFallbackRefreshMs = 10000;

export const isRealtimeFallbackStatus = (status: DeviceRealtimeStatus) => {
  return status === 'reconnecting' || status === 'closed';
};

const realtimeStatusLabelKey: Record<DeviceRealtimeStatus, AdminMessageKey> = {
  idle: 'adminRealtimeIdle',
  connecting: 'adminRealtimeConnecting',
  open: 'adminRealtimeOpen',
  reconnecting: 'adminRealtimeReconnecting',
  closed: 'adminRealtimeClosed',
};

const dotClassName = (status: DeviceRealtimeStatus) => {
  if (status === 'open') {
    return 'bg-emerald-500';
  }

  if (status === 'connecting' || status === 'reconnecting') {
    return 'bg-amber-500';
  }

  return 'bg-slate-400';
};

const badgeClassName = (status: DeviceRealtimeStatus) => {
  return [
    'inline-flex h-11 items-center gap-2 rounded-md border px-3 text-xs font-bold',
    status === 'open'
      ? 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-900/70 dark:bg-emerald-950/40 dark:text-emerald-300'
      : status === 'connecting' || status === 'reconnecting'
      ? 'border-amber-200 bg-amber-50 text-amber-700 dark:border-amber-900/70 dark:bg-amber-950/40 dark:text-amber-300'
      : 'border-linkflow-border bg-white text-linkflow-muted dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted',
  ].join(' ');
};

interface RealtimeStatusBadgeProps {
  status: DeviceRealtimeStatus;
}

const RealtimeStatusBadge = ({ status }: RealtimeStatusBadgeProps) => {
  const { t } = useI18n();

  return (
    <span
      className={badgeClassName(status)}
      title={t(realtimeStatusLabelKey[status])}
    >
      <span
        className={['h-2 w-2 rounded-full', dotClassName(status)].join(' ')}
        aria-hidden="true"
      />
      {t(realtimeStatusLabelKey[status])}
    </span>
  );
};

export default RealtimeStatusBadge;
