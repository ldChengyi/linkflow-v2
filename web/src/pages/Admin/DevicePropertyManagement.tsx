import { useI18n } from '@/contexts/I18nContext';
import {
  getDeviceLatestProperties,
  listDevices,
  type Device,
  type DeviceLatestProperties,
} from '@/services/devices';
import { listProducts, type Product } from '@/services/products';
import { listTenants, type Tenant } from '@/services/tenants';
import { formatDateTime } from '@/utils/date';
import {
  AppstoreOutlined,
  FieldTimeOutlined,
  LineChartOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { notification } from 'antd';
import { useEffect, useState } from 'react';

type DevicePropertyTab = 'latest' | 'history';

const tabButtonClassName = (active: boolean) => {
  return [
    'inline-flex min-h-11 items-center gap-2 border-b-2 px-3 text-sm font-bold transition',
    active
      ? 'border-linkflow-primary text-linkflow-primary'
      : 'border-transparent text-linkflow-muted hover:text-linkflow-primary dark:text-linkflow-dark-muted dark:hover:text-linkflow-dark-primary',
  ].join(' ');
};

const statusBadgeClassName = (reported: boolean) => {
  return [
    'inline-flex rounded-full px-2.5 py-1 text-xs font-bold',
    reported
      ? 'bg-linkflow-primary-soft text-linkflow-primary'
      : 'bg-slate-100 text-linkflow-muted dark:bg-slate-800 dark:text-linkflow-dark-muted',
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

interface DevicePropertyCardProps {
  name: string;
  value: unknown;
}

const DevicePropertyCard = ({ name, value }: DevicePropertyCardProps) => {
  const { t } = useI18n();

  return (
    <article className="min-h-36 rounded-lg border border-linkflow-border bg-white p-4 shadow-[0_8px_24px_rgba(23,32,51,0.06)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/20">
      <div className="mb-4 flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="m-0 text-xs font-bold uppercase text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDevicePropertyKey')}
          </p>
          <h4 className="m-0 mt-1 break-all text-base font-bold">{name}</h4>
        </div>
        <span className="shrink-0 rounded-full bg-slate-100 px-2.5 py-1 text-xs font-bold text-linkflow-muted dark:bg-slate-800 dark:text-linkflow-dark-muted">
          {valueTypeOf(value)}
        </span>
      </div>
      <pre className="m-0 max-h-40 overflow-auto whitespace-pre-wrap break-words rounded-md bg-slate-50 p-3 text-sm font-semibold leading-6 text-linkflow-text dark:bg-slate-900/60 dark:text-linkflow-dark-text">
        {formatPropertyValue(value)}
      </pre>
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

      <div className="grid gap-4 [grid-template-columns:repeat(auto-fit,minmax(240px,1fr))]">
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

    const run = async () => {
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
        if (cancelled) {
          return;
        }
        setDevices(result.items);
        setDeviceId((current) =>
          current && result.items.some((device) => device.id === current)
            ? current
            : result.items[0]?.id || '',
        );
      } catch (error) {
        if (cancelled) {
          return;
        }
        notification.error({
          message: t('adminDeviceLoadFailed'),
          description: error instanceof Error ? error.message : undefined,
          placement: 'topRight',
        });
      }
    };

    run();

    return () => {
      cancelled = true;
    };
  }, [productFilter, tenantId, t]);

  const loadLatestProperties = async (nextDeviceId = deviceId) => {
    if (!nextDeviceId) {
      setLatest(null);
      return;
    }

    setLoadingLatest(true);
    try {
      const result = await getDeviceLatestProperties(nextDeviceId);
      setLatest(result);
    } catch (error) {
      notification.error({
        message: t('adminDevicePropertyLatestFailed'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    } finally {
      setLoadingLatest(false);
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

  const handleTenantChange = (nextTenantId: string) => {
    setTenantId(nextTenantId);
    setProductFilter('');
    setDeviceId('');
    setLatest(null);
  };

  const handleProductFilterChange = (nextProductId: string) => {
    setProductFilter(nextProductId);
    setDeviceId('');
    setLatest(null);
  };

  const selectedDevice = devices.find((device) => device.id === deviceId);

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
              onChange={(event) => setDeviceId(event.target.value)}
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

      <section className="rounded-lg border border-linkflow-border bg-white p-4 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div className="min-w-0">
            <p className="m-0 text-sm font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
              {selectedDevice?.device_slug ?? '--'}
            </p>
            <h3 className="m-0 mt-1 truncate text-xl font-bold">
              {selectedDevice?.device_name ?? t('adminDevicePropertyNoDevice')}
            </h3>
          </div>
          <span className={statusBadgeClassName(latest?.reported ?? false)}>
            {latest?.reported
              ? t('adminDevicePropertyReported')
              : t('adminDevicePropertyNotReported')}
          </span>
        </div>
      </section>

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
      </div>

      {activeTab === 'latest' ? (
        <LatestPropertiesPanel latest={latest} loading={loadingLatest} />
      ) : (
        <HistoryPropertiesPanel />
      )}
    </section>
  );
};

export default DevicePropertyManagement;
