import { useI18n } from '@/contexts/I18nContext';
import { useDevicesRealtime } from '@/hooks/useDevicesRealtime';
import {
  getDeviceLatestProperties,
  listDeviceEvents,
  listDevices,
  type Device,
  type DeviceEventEntry,
  type DeviceLatestProperties,
} from '@/services/devices';
import { listProducts, type Product } from '@/services/products';
import { listTenants, type Tenant } from '@/services/tenants';
import { useEventInboxStore } from '@/stores/eventInboxStore';
import { formatDateTime } from '@/utils/date';
import {
  AppstoreOutlined,
  FieldTimeOutlined,
  LineChartOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { notification } from 'antd';
import { useEffect, useState } from 'react';
import RealtimeStatusBadge, {
  isRealtimeFallbackStatus,
  realtimeFallbackRefreshMs,
} from './components/RealtimeStatusBadge';

type DevicePropertyTab = 'latest' | 'history' | 'events';

const tabButtonClassName = (active: boolean) => {
  return [
    'inline-flex min-h-11 items-center gap-2 border-b-2 px-3 text-sm font-bold transition',
    active
      ? 'border-linkflow-primary text-linkflow-primary'
      : 'border-transparent text-linkflow-muted hover:text-linkflow-primary dark:text-linkflow-dark-muted dark:hover:text-linkflow-dark-primary',
  ].join(' ');
};

const valueTypeOf = (value: unknown) => {
  if (Array.isArray(value)) {
    return 'array';
  }

  if (value === null) {
    return 'null';
  }

  return typeof value;
};

const formatPropertyValue = (value: unknown) => {
  if (typeof value === 'string') {
    return value;
  }

  if (
    typeof value === 'number' ||
    typeof value === 'boolean' ||
    value === null
  ) {
    return String(value);
  }

  return JSON.stringify(value, null, 2);
};

const isStructuredPropertyValue = (value: unknown) => {
  return typeof value === 'object' && value !== null;
};

const valueTypeBadgeClassName = (type: string) => {
  return [
    'shrink-0 rounded-full px-2.5 py-1 text-xs font-black uppercase leading-none',
    type === 'number'
      ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300'
      : type === 'boolean'
      ? 'bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-300'
      : type === 'string'
      ? 'bg-linkflow-primary-soft text-linkflow-primary dark:bg-linkflow-dark-primary-soft dark:text-linkflow-dark-primary'
      : 'bg-slate-100 text-linkflow-muted dark:bg-slate-800 dark:text-linkflow-dark-muted',
  ].join(' ');
};

interface DevicePropertyCardProps {
  name: string;
  value: unknown;
}

const DevicePropertyCard = ({ name, value }: DevicePropertyCardProps) => {
  const { t } = useI18n();
  const valueType = valueTypeOf(value);
  const formattedValue = formatPropertyValue(value);
  const structuredValue = isStructuredPropertyValue(value);

  return (
    <article className="group grid min-h-44 overflow-hidden rounded-lg border border-linkflow-border bg-white shadow-[0_8px_24px_rgba(23,32,51,0.06)] transition hover:-translate-y-0.5 hover:border-linkflow-primary hover:shadow-[0_12px_30px_rgba(18,137,153,0.12)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/20 dark:hover:border-linkflow-dark-primary">
      <div className="h-1 bg-linkflow-primary dark:bg-linkflow-dark-primary" />
      <div className="grid gap-4 p-4">
        <div className="flex items-start justify-between gap-3">
          <div className="min-w-0">
            <p className="m-0 text-xs font-bold uppercase text-linkflow-subtle dark:text-linkflow-dark-subtle">
              {t('adminDevicePropertyKey')}
            </p>
            <h4 className="m-0 mt-1 break-all text-base font-black text-linkflow-text dark:text-linkflow-dark-text">
              {name}
            </h4>
          </div>
          <span className={valueTypeBadgeClassName(valueType)}>
            {valueType}
          </span>
        </div>

        <div className="rounded-lg border border-linkflow-border bg-slate-50/80 p-3 dark:border-linkflow-dark-border dark:bg-slate-950/30">
          <p className="m-0 mb-2 text-xs font-bold uppercase text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDevicePropertyValue')}
          </p>
          {structuredValue ? (
            <pre className="m-0 max-h-44 overflow-auto whitespace-pre-wrap break-words rounded-md bg-linkflow-code p-3 font-mono text-xs font-semibold leading-5 text-linkflow-code-text">
              {formattedValue}
            </pre>
          ) : (
            <p className="m-0 break-words text-xl font-black leading-8 text-linkflow-text dark:text-linkflow-dark-text">
              {formattedValue}
            </p>
          )}
        </div>

        <dl className="m-0 flex items-center justify-between gap-3 text-xs font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
          <dt>{t('adminDevicePropertyValueType')}</dt>
          <dd className="m-0 font-mono text-linkflow-muted dark:text-linkflow-dark-muted">
            {valueType}
          </dd>
        </dl>
      </div>
    </article>
  );
};

interface LatestPropertiesPanelProps {
  latest: DeviceLatestProperties | null;
  loading: boolean;
}

const LatestPropertiesPanel = ({
  latest,
  loading,
}: LatestPropertiesPanelProps) => {
  const { t } = useI18n();
  const entries = Object.entries(latest?.properties ?? {});

  if (loading && latest === null) {
    return (
      <section className="grid min-h-72 place-items-center rounded-lg border border-linkflow-border bg-white p-8 text-center dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel">
        <p className="m-0 text-sm font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
          {t('adminDevicePropertyLoadingLatest')}
        </p>
      </section>
    );
  }

  if (latest === null || !latest.reported || entries.length === 0) {
    return (
      <section className="grid min-h-72 place-items-center rounded-lg border border-linkflow-border bg-white p-8 text-center dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel">
        <div>
          <p className="m-0 text-base font-bold">
            {t('adminDevicePropertyLatestEmpty')}
          </p>
          <p className="m-0 mt-2 max-w-xl leading-6 text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDevicePropertyLatestEmptyDescription')}
          </p>
        </div>
      </section>
    );
  }

  return (
    <section className="grid gap-5">
      <div className="grid gap-4 [grid-template-columns:repeat(auto-fit,minmax(220px,1fr))]">
        <article className="rounded-lg border border-linkflow-border bg-white p-4 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel">
          <p className="m-0 text-sm font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDevicePropertyProductKey')}
          </p>
          <strong className="mt-2 block break-all text-lg">
            {latest.product_key}
          </strong>
        </article>
        <article className="rounded-lg border border-linkflow-border bg-white p-4 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel">
          <p className="m-0 text-sm font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDevicePropertyDeviceSlug')}
          </p>
          <strong className="mt-2 block break-all text-lg">
            {latest.device_slug}
          </strong>
        </article>
        <article className="rounded-lg border border-linkflow-border bg-white p-4 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel">
          <p className="m-0 text-sm font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDevicePropertyOccurredAt')}
          </p>
          <strong className="mt-2 block text-lg">
            {formatDateTime(latest.occurred_at ?? '')}
          </strong>
        </article>
        <article className="rounded-lg border border-linkflow-border bg-white p-4 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel">
          <p className="m-0 text-sm font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDevicePropertyReceivedAt')}
          </p>
          <strong className="mt-2 block text-lg">
            {formatDateTime(latest.received_at ?? '')}
          </strong>
        </article>
      </div>

      <div className="grid gap-4 [grid-template-columns:repeat(auto-fit,minmax(260px,1fr))]">
        {entries.map(([key, value]) => (
          <DevicePropertyCard key={key} name={key} value={value} />
        ))}
      </div>
    </section>
  );
};

const HistoryPropertiesPanel = () => {
  const { t } = useI18n();

  return (
    <section className="grid gap-4 rounded-lg border border-linkflow-border bg-white p-5 shadow-[0_8px_24px_rgba(23,32,51,0.06)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/20">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h3 className="m-0 text-lg font-bold">
            {t('adminDevicePropertyHistoryTitle')}
          </h3>
          <p className="m-0 mt-2 leading-6 text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDevicePropertyHistoryHint')}
          </p>
        </div>
        <LineChartOutlined
          className="text-2xl text-linkflow-primary"
          aria-hidden="true"
        />
      </div>

      <div className="min-h-80 overflow-hidden rounded-lg border border-linkflow-border bg-slate-50 p-4 dark:border-linkflow-dark-border dark:bg-slate-900/30">
        <svg
          className="h-72 w-full text-linkflow-subtle dark:text-linkflow-dark-subtle"
          role="img"
          aria-label={t('adminDevicePropertyHistoryTitle')}
          viewBox="0 0 720 260"
          preserveAspectRatio="none"
        >
          <line
            x1="48"
            y1="24"
            x2="48"
            y2="220"
            stroke="currentColor"
            strokeWidth="1"
          />
          <line
            x1="48"
            y1="220"
            x2="690"
            y2="220"
            stroke="currentColor"
            strokeWidth="1"
          />
          {[60, 100, 140, 180].map((y) => (
            <line
              key={y}
              x1="48"
              y1={y}
              x2="690"
              y2={y}
              stroke="currentColor"
              strokeDasharray="6 8"
              strokeOpacity="0.35"
              strokeWidth="1"
            />
          ))}
        </svg>
        <p className="m-0 -mt-44 px-4 text-center text-sm font-semibold text-linkflow-subtle dark:text-linkflow-dark-subtle">
          {t('adminDevicePropertyHistoryEmpty')}
        </p>
      </div>
    </section>
  );
};

interface DeviceEventsPanelProps {
  events: DeviceEventEntry[];
  eventName: string;
  loading: boolean;
  page: number;
  pageSize: number;
  total: number;
  onEventNameChange: (eventName: string) => void;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  onReload: () => void;
}

const DeviceEventsPanel = ({
  events,
  eventName,
  loading,
  page,
  pageSize,
  total,
  onEventNameChange,
  onPageChange,
  onPageSizeChange,
  onReload,
}: DeviceEventsPanelProps) => {
  const { t } = useI18n();
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  return (
    <section className="grid gap-4">
      <div className="flex flex-wrap items-end justify-between gap-3 rounded-lg border border-linkflow-border bg-white p-4 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel">
        <label className="grid min-w-64 gap-2 text-sm font-bold">
          {t('adminDeviceEventName')}
          <input
            className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
            value={eventName}
            onChange={(event) => onEventNameChange(event.target.value)}
            placeholder={t('adminDeviceEventAllEvents')}
          />
        </label>
        <div className="flex flex-wrap items-center gap-2">
          <label className="grid gap-2 text-sm font-bold">
            {t('adminDevicePageSize')}
            <select
              className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
              value={pageSize}
              onChange={(event) => onPageSizeChange(Number(event.target.value))}
            >
              {[10, 20, 50, 100].map((size) => (
                <option key={size} value={size}>
                  {size}
                </option>
              ))}
            </select>
          </label>
          <button
            type="button"
            className="inline-flex h-11 items-center gap-2 self-end rounded-md border border-linkflow-primary bg-white/70 px-3 text-sm font-bold text-linkflow-primary shadow-sm transition hover:bg-linkflow-primary-soft disabled:cursor-not-allowed disabled:opacity-60 dark:border-linkflow-dark-primary dark:bg-linkflow-dark-panel/70 dark:text-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-soft"
            onClick={onReload}
            disabled={loading}
          >
            <ReloadOutlined aria-hidden="true" />
            {t('adminDeviceEventLoadHistory')}
          </button>
        </div>
      </div>

      {loading && events.length === 0 ? (
        <section className="grid min-h-72 place-items-center rounded-lg border border-linkflow-border bg-white p-8 text-center dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel">
          <p className="m-0 text-sm font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDeviceEventLoadingHistory')}
          </p>
        </section>
      ) : events.length === 0 ? (
        <section className="grid min-h-72 place-items-center rounded-lg border border-linkflow-border bg-white p-8 text-center dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel">
          <p className="m-0 text-base font-bold">
            {t('adminDeviceEventHistoryEmpty')}
          </p>
        </section>
      ) : (
        <div className="grid gap-3">
          {events.map((entry) => (
            <article
              key={entry.event_id}
              className="grid gap-4 rounded-lg border border-linkflow-border bg-white p-4 shadow-[0_8px_24px_rgba(23,32,51,0.06)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/20"
            >
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="m-0 text-xs font-bold uppercase text-linkflow-subtle dark:text-linkflow-dark-subtle">
                    {t('adminDeviceEventName')}
                  </p>
                  <h3 className="m-0 mt-1 break-all text-lg font-black text-linkflow-text dark:text-linkflow-dark-text">
                    {entry.event_name}
                  </h3>
                </div>
                <span className="rounded-full bg-linkflow-primary-soft px-3 py-1 text-xs font-black text-linkflow-primary dark:bg-linkflow-dark-primary-soft dark:text-linkflow-dark-primary">
                  {entry.product_key}/{entry.device_slug}
                </span>
              </div>
              <dl className="m-0 grid gap-3 text-sm [grid-template-columns:repeat(auto-fit,minmax(180px,1fr))]">
                <div>
                  <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
                    {t('adminDevicePropertyOccurredAt')}
                  </dt>
                  <dd className="m-0 mt-1 font-semibold">
                    {formatDateTime(entry.occurred_at)}
                  </dd>
                </div>
                <div>
                  <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
                    {t('adminDevicePropertyReceivedAt')}
                  </dt>
                  <dd className="m-0 mt-1 font-semibold">
                    {formatDateTime(entry.received_at)}
                  </dd>
                </div>
              </dl>
              <div className="rounded-lg border border-linkflow-border bg-slate-50/80 p-3 dark:border-linkflow-dark-border dark:bg-slate-950/30">
                <p className="m-0 mb-2 text-xs font-bold uppercase text-linkflow-subtle dark:text-linkflow-dark-subtle">
                  {t('adminDeviceEventParams')}
                </p>
                <pre className="m-0 max-h-52 overflow-auto whitespace-pre-wrap break-words rounded-md bg-linkflow-code p-3 font-mono text-xs font-semibold leading-5 text-linkflow-code-text">
                  {JSON.stringify(entry.params, null, 2)}
                </pre>
              </div>
            </article>
          ))}
        </div>
      )}

      <div className="flex flex-wrap items-center justify-between gap-3 rounded-lg border border-linkflow-border bg-white p-3 text-sm font-bold dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel">
        <span className="text-linkflow-subtle dark:text-linkflow-dark-subtle">
          {t('adminDeviceEventTotal').replace('{total}', String(total))}
        </span>
        <div className="flex items-center gap-2">
          <button
            type="button"
            className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-bold transition hover:border-linkflow-primary hover:text-linkflow-primary disabled:cursor-not-allowed disabled:opacity-60 dark:border-linkflow-dark-border dark:bg-linkflow-dark-page dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
            disabled={loading || page <= 1}
            onClick={() => onPageChange(page - 1)}
          >
            {t('adminDevicePrev')}
          </button>
          <span>
            {t('adminDeviceEventPage')
              .replace('{page}', String(page))
              .replace('{totalPages}', String(totalPages))}
          </span>
          <button
            type="button"
            className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-bold transition hover:border-linkflow-primary hover:text-linkflow-primary disabled:cursor-not-allowed disabled:opacity-60 dark:border-linkflow-dark-border dark:bg-linkflow-dark-page dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
            disabled={loading || page >= totalPages}
            onClick={() => onPageChange(page + 1)}
          >
            {t('adminDeviceNext')}
          </button>
        </div>
      </div>
    </section>
  );
};

const DevicePropertyManagement = () => {
  const { t } = useI18n();
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [tenantId, setTenantId] = useState('');
  const [products, setProducts] = useState<Product[]>([]);
  const [productFilter, setProductFilter] = useState('');
  const [devices, setDevices] = useState<Device[]>([]);
  const [deviceId, setDeviceId] = useState('');
  const [activeTab, setActiveTab] = useState<DevicePropertyTab>('latest');
  const [latest, setLatest] = useState<DeviceLatestProperties | null>(null);
  const [loadingLatest, setLoadingLatest] = useState(false);
  const [eventEntries, setEventEntries] = useState<DeviceEventEntry[]>([]);
  const [eventTotal, setEventTotal] = useState(0);
  const [eventPage, setEventPage] = useState(1);
  const [eventPageSize, setEventPageSize] = useState(20);
  const [eventNameFilter, setEventNameFilter] = useState('');
  const [loadingEvents, setLoadingEvents] = useState(false);
  const latestInboxEvent = useEventInboxStore((state) => state.items[0]);

  const loadSelectableDevices = async (
    options: { silent?: boolean; isCancelled?: () => boolean } = {},
  ) => {
    if (!tenantId) {
      setDevices([]);
      setDeviceId('');
      return;
    }

    try {
      const result = await listDevices({
        tenant_id: tenantId,
        product_id: productFilter || undefined,
        page: 1,
        page_size: 100,
      });
      if (options.isCancelled?.()) {
        return;
      }
      setDevices(result.items);
      setDeviceId((current) =>
        current && result.items.some((device) => device.id === current)
          ? current
          : result.items[0]?.id || '',
      );
    } catch (error) {
      if (options.isCancelled?.() || options.silent) {
        return;
      }
      notification.error({
        message: t('adminDeviceLoadFailed'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    }
  };

  useEffect(() => {
    let cancelled = false;

    const run = async () => {
      try {
        const result = await listTenants({ page: 1, page_size: 50 });
        if (cancelled) {
          return;
        }
        setTenants(result.items);
        setTenantId((current) => current || result.items[0]?.id || '');
      } catch (error) {
        if (cancelled) {
          return;
        }
        notification.error({
          message: t('adminTenantLoadFailed'),
          description: error instanceof Error ? error.message : undefined,
          placement: 'topRight',
        });
      }
    };

    run();

    return () => {
      cancelled = true;
    };
  }, [t]);

  useEffect(() => {
    let cancelled = false;

    const run = async () => {
      if (!tenantId) {
        setProducts([]);
        setProductFilter('');
        return;
      }

      try {
        const result = await listProducts({
          tenant_id: tenantId,
          page: 1,
          page_size: 100,
        });
        if (cancelled) {
          return;
        }
        setProducts(result.items);
        setProductFilter((current) =>
          current && result.items.some((product) => product.id === current)
            ? current
            : '',
        );
      } catch (error) {
        if (cancelled) {
          return;
        }
        notification.error({
          message: t('adminProductLoadFailed'),
          description: error instanceof Error ? error.message : undefined,
          placement: 'topRight',
        });
      }
    };

    run();

    return () => {
      cancelled = true;
    };
  }, [tenantId, t]);

  useEffect(() => {
    let cancelled = false;

    void loadSelectableDevices({ isCancelled: () => cancelled });

    return () => {
      cancelled = true;
    };
  }, [productFilter, tenantId, t]);

  const loadLatestProperties = async (
    nextDeviceId = deviceId,
    options: { silent?: boolean; isCancelled?: () => boolean } = {},
  ) => {
    if (!nextDeviceId) {
      setLatest(null);
      return;
    }

    if (!options.silent) {
      setLoadingLatest(true);
    }
    try {
      const result = await getDeviceLatestProperties(nextDeviceId);
      if (options.isCancelled?.()) {
        return;
      }
      setLatest(result);
    } catch (error) {
      if (options.isCancelled?.() || options.silent) {
        return;
      }
      notification.error({
        message: t('adminDevicePropertyLatestFailed'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    } finally {
      if (!options.silent) {
        setLoadingLatest(false);
      }
    }
  };

  const loadEventHistory = async (
    nextDeviceId = deviceId,
    options: {
      eventName?: string;
      page?: number;
      pageSize?: number;
      silent?: boolean;
      isCancelled?: () => boolean;
    } = {},
  ) => {
    if (!nextDeviceId) {
      setEventEntries([]);
      setEventTotal(0);
      return;
    }

    if (!options.silent) {
      setLoadingEvents(true);
    }
    try {
      const result = await listDeviceEvents(nextDeviceId, {
        event_name: (options.eventName ?? eventNameFilter).trim() || undefined,
        page: options.page ?? eventPage,
        page_size: options.pageSize ?? eventPageSize,
      });
      if (options.isCancelled?.()) {
        return;
      }
      setEventEntries(result.items);
      setEventTotal(result.total);
      setEventPage(result.page);
      setEventPageSize(result.page_size);
    } catch (error) {
      if (options.isCancelled?.() || options.silent) {
        return;
      }
      notification.error({
        message: t('adminDeviceEventHistoryFailed'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    } finally {
      if (!options.silent) {
        setLoadingEvents(false);
      }
    }
  };

  useEffect(() => {
    let cancelled = false;

    const run = async () => {
      if (!deviceId) {
        setLatest(null);
        return;
      }

      setLoadingLatest(true);
      try {
        const result = await getDeviceLatestProperties(deviceId);
        if (cancelled) {
          return;
        }
        setLatest(result);
      } catch (error) {
        if (cancelled) {
          return;
        }
        notification.error({
          message: t('adminDevicePropertyLatestFailed'),
          description: error instanceof Error ? error.message : undefined,
          placement: 'topRight',
        });
      } finally {
        if (!cancelled) {
          setLoadingLatest(false);
        }
      }
    };

    run();

    return () => {
      cancelled = true;
    };
  }, [deviceId, t]);

  useEffect(() => {
    let cancelled = false;

    const run = async () => {
      if (activeTab !== 'events') {
        return;
      }
      if (!deviceId) {
        setEventEntries([]);
        setEventTotal(0);
        return;
      }

      setLoadingEvents(true);
      try {
        const result = await listDeviceEvents(deviceId, {
          event_name: eventNameFilter.trim() || undefined,
          page: eventPage,
          page_size: eventPageSize,
        });
        if (cancelled) {
          return;
        }
        setEventEntries(result.items);
        setEventTotal(result.total);
        setEventPage(result.page);
        setEventPageSize(result.page_size);
      } catch (error) {
        if (cancelled) {
          return;
        }
        notification.error({
          message: t('adminDeviceEventHistoryFailed'),
          description: error instanceof Error ? error.message : undefined,
          placement: 'topRight',
        });
      } finally {
        if (!cancelled) {
          setLoadingEvents(false);
        }
      }
    };

    run();

    return () => {
      cancelled = true;
    };
  }, [activeTab, deviceId, eventNameFilter, eventPage, eventPageSize, t]);

  useEffect(() => {
    if (
      !latestInboxEvent ||
      activeTab !== 'events' ||
      eventPage !== 1 ||
      latestInboxEvent.device_id !== deviceId
    ) {
      return;
    }

    const eventName = eventNameFilter.trim();
    if (eventName && latestInboxEvent.event_name !== eventName) {
      return;
    }

    setEventEntries((current) => {
      if (
        current.some((entry) => entry.event_id === latestInboxEvent.event_id)
      ) {
        return current;
      }

      const next: DeviceEventEntry = {
        event_id: latestInboxEvent.event_id,
        tenant_id: latestInboxEvent.tenant_id,
        product_id: latestInboxEvent.product_id,
        product_key: latestInboxEvent.product_key,
        device_slug: latestInboxEvent.device_slug,
        event_name: latestInboxEvent.event_name,
        params: latestInboxEvent.params,
        occurred_at: latestInboxEvent.occurred_at,
        received_at: latestInboxEvent.received_at,
      };

      return [next, ...current].slice(0, eventPageSize);
    });
    setEventTotal((current) => current + 1);
  }, [
    activeTab,
    deviceId,
    eventNameFilter,
    eventPage,
    eventPageSize,
    latestInboxEvent,
  ]);

  const { status: realtimeStatus } = useDevicesRealtime({
    enabled: Boolean(tenantId),
    onConnectionChanged: (payload, meta) => {
      if (meta.tenantId !== tenantId) {
        return;
      }
      setDevices((current) =>
        current.map((device) =>
          device.id === payload.device_id
            ? {
                ...device,
                connection_status: payload.status,
                last_seen_at: meta.occurredAt,
                updated_at: meta.occurredAt,
              }
            : device,
        ),
      );
    },
    onPropertyChanged: (payload, meta) => {
      if (meta.tenantId !== tenantId || payload.device_id !== deviceId) {
        return;
      }

      setLatest({
        id: payload.device_id,
        tenant_id: payload.tenant_id,
        product_id: payload.product_id,
        product_key: payload.product_key,
        device_slug: payload.device_slug,
        reported: true,
        properties: payload.properties,
        occurred_at: meta.occurredAt,
        received_at: meta.occurredAt,
      });

      getDeviceLatestProperties(payload.device_id)
        .then((result) => {
          setLatest(result);
        })
        .catch((error) => {
          notification.error({
            message: t('adminDevicePropertyLatestFailed'),
            description: error instanceof Error ? error.message : undefined,
            placement: 'topRight',
          });
        });
    },
  });

  useEffect(() => {
    if (!tenantId || !isRealtimeFallbackStatus(realtimeStatus)) {
      return;
    }

    let active = true;
    const refresh = () => {
      const isCancelled = () => !active;
      void loadSelectableDevices({ silent: true, isCancelled });
      if (deviceId) {
        void loadLatestProperties(deviceId, { silent: true, isCancelled });
        if (activeTab === 'events') {
          void loadEventHistory(deviceId, { silent: true, isCancelled });
        }
      }
    };

    const timer = window.setInterval(refresh, realtimeFallbackRefreshMs);

    return () => {
      active = false;
      window.clearInterval(timer);
    };
  }, [
    activeTab,
    deviceId,
    eventNameFilter,
    eventPage,
    eventPageSize,
    productFilter,
    realtimeStatus,
    tenantId,
  ]);

  const handleTenantChange = (nextTenantId: string) => {
    setTenantId(nextTenantId);
    setProductFilter('');
    setDeviceId('');
    setLatest(null);
    setEventEntries([]);
    setEventTotal(0);
    setEventPage(1);
  };

  const handleProductFilterChange = (nextProductId: string) => {
    setProductFilter(nextProductId);
    setDeviceId('');
    setLatest(null);
    setEventEntries([]);
    setEventTotal(0);
    setEventPage(1);
  };

  const handleDeviceChange = (nextDeviceId: string) => {
    setDeviceId(nextDeviceId);
    setLatest(null);
    setEventEntries([]);
    setEventTotal(0);
    setEventPage(1);
  };

  const handleEventNameChange = (nextEventName: string) => {
    setEventNameFilter(nextEventName);
    setEventPage(1);
  };

  const handleEventPageSizeChange = (nextPageSize: number) => {
    setEventPageSize(nextPageSize);
    setEventPage(1);
  };

  return (
    <section className="grid gap-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div className="flex flex-wrap gap-3">
          <label className="grid min-w-64 gap-2 text-sm font-bold">
            {t('adminTenantName')}
            <select
              className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
              value={tenantId}
              onChange={(event) => handleTenantChange(event.target.value)}
            >
              {tenants.map((tenant) => (
                <option key={tenant.id} value={tenant.id}>
                  {tenant.tenant_name}
                </option>
              ))}
            </select>
          </label>
          <label className="grid min-w-64 gap-2 text-sm font-bold">
            {t('adminProductName')}
            <select
              className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
              value={productFilter}
              onChange={(event) =>
                handleProductFilterChange(event.target.value)
              }
            >
              <option value="">{t('adminDeviceAllProducts')}</option>
              {products.map((product) => (
                <option key={product.id} value={product.id}>
                  {product.product_name}
                </option>
              ))}
            </select>
          </label>
          <label className="grid min-w-64 gap-2 text-sm font-bold">
            {t('adminDevicePropertySelectDevice')}
            <select
              className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary disabled:opacity-70 dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
              value={deviceId}
              onChange={(event) => handleDeviceChange(event.target.value)}
              disabled={devices.length === 0}
            >
              {devices.length === 0 ? (
                <option value="">{t('adminDevicePropertyNoDevice')}</option>
              ) : null}
              {devices.map((device) => (
                <option key={device.id} value={device.id}>
                  {device.device_name}
                </option>
              ))}
            </select>
          </label>
        </div>
        <div className="flex flex-wrap items-center justify-end gap-2">
          <RealtimeStatusBadge status={realtimeStatus} />
          <button
            type="button"
            className="inline-flex h-11 items-center gap-2 rounded-md border border-linkflow-primary bg-white/70 px-3 text-sm font-bold text-linkflow-primary shadow-sm transition hover:bg-linkflow-primary-soft disabled:cursor-not-allowed disabled:opacity-60 dark:border-linkflow-dark-primary dark:bg-linkflow-dark-panel/70 dark:text-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-soft"
            onClick={() => loadLatestProperties()}
            disabled={!deviceId || loadingLatest}
          >
            <ReloadOutlined aria-hidden="true" />
            {t('adminDevicePropertyLoadLatest')}
          </button>
        </div>
      </div>

      <div
        className="flex gap-2 overflow-x-auto border-b border-linkflow-border dark:border-linkflow-dark-border"
        role="tablist"
        aria-label={t('adminDevicePropertiesNav')}
      >
        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'latest'}
          className={tabButtonClassName(activeTab === 'latest')}
          onClick={() => setActiveTab('latest')}
        >
          <AppstoreOutlined aria-hidden="true" />
          {t('adminDevicePropertyLatestTab')}
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'history'}
          className={tabButtonClassName(activeTab === 'history')}
          onClick={() => setActiveTab('history')}
        >
          <FieldTimeOutlined aria-hidden="true" />
          {t('adminDevicePropertyHistoryTab')}
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'events'}
          className={tabButtonClassName(activeTab === 'events')}
          onClick={() => setActiveTab('events')}
        >
          <LineChartOutlined aria-hidden="true" />
          {t('adminDeviceEventTab')}
        </button>
      </div>

      {activeTab === 'latest' ? (
        <LatestPropertiesPanel latest={latest} loading={loadingLatest} />
      ) : activeTab === 'history' ? (
        <HistoryPropertiesPanel />
      ) : (
        <DeviceEventsPanel
          events={eventEntries}
          eventName={eventNameFilter}
          loading={loadingEvents}
          onEventNameChange={handleEventNameChange}
          onPageChange={setEventPage}
          onPageSizeChange={handleEventPageSizeChange}
          onReload={() => loadEventHistory()}
          page={eventPage}
          pageSize={eventPageSize}
          total={eventTotal}
        />
      )}
    </section>
  );
};

export default DevicePropertyManagement;
