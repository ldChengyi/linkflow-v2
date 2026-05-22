import { useI18n } from '@/contexts/I18nContext';
import { useDevicesRealtime } from '@/hooks/useDevicesRealtime';
import {
  callDeviceService,
  getDeviceLatestProperties,
  listDeviceEvents,
  listDeviceServiceCalls,
  listDevices,
  type Device,
  type DeviceEventEntry,
  type DeviceLatestProperties,
  type DeviceServiceAckStatus,
  type DeviceServiceCallHistoryEntry,
  type DeviceServiceCallResult,
} from '@/services/devices';
import { listProducts, type Product } from '@/services/products';
import { listTenants, type Tenant } from '@/services/tenants';
import { listThingsModels, type ThingsModel } from '@/services/thingsmodels';
import { useEventInboxStore } from '@/stores/eventInboxStore';
import { formatDateTime } from '@/utils/date';
import {
  AppstoreOutlined,
  ApiOutlined,
  FieldTimeOutlined,
  LineChartOutlined,
  ReloadOutlined,
  SendOutlined,
} from '@ant-design/icons';
import { notification } from 'antd';
import { useEffect, useState } from 'react';
import AdminDataTable, {
  type AdminDataTableColumn,
} from './components/AdminDataTable';
import RealtimeStatusBadge, {
  isRealtimeFallbackStatus,
  realtimeFallbackRefreshMs,
} from './components/RealtimeStatusBadge';

type DevicePropertyTab =
  | 'latest'
  | 'history'
  | 'services'
  | 'serviceHistory'
  | 'eventHistory';

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

const isRecord = (value: unknown): value is Record<string, unknown> => {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
};

const formatFullDateTime = (value?: string) => {
  if (!value) {
    return '--';
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toISOString();
};

interface TimeCellProps {
  value?: string;
}

const TimeCell = ({ value }: TimeCellProps) => {
  const full = formatFullDateTime(value);

  return (
    <time
      className="block whitespace-nowrap text-sm font-semibold text-linkflow-text dark:text-linkflow-dark-text"
      dateTime={value}
      title={full}
    >
      {formatDateTime(value ?? '')}
    </time>
  );
};

interface JsonPreviewProps {
  value: unknown;
  maxHeightClassName?: string;
}

const JsonPreview = ({
  value,
  maxHeightClassName = 'max-h-36',
}: JsonPreviewProps) => {
  return (
    <pre
      className={[
        'm-0 overflow-auto whitespace-pre-wrap break-words rounded-md bg-linkflow-code p-3 font-mono text-xs font-semibold leading-5 text-linkflow-code-text',
        maxHeightClassName,
      ].join(' ')}
    >
      {JSON.stringify(value ?? {}, null, 2)}
    </pre>
  );
};

const ackStatusClassName = (status: DeviceServiceAckStatus) => {
  switch (status) {
    case 'success':
      return 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300';
    case 'failed':
      return 'bg-rose-50 text-rose-700 dark:bg-rose-950/40 dark:text-rose-300';
    case 'pending':
      return 'bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-300';
    default:
      return 'bg-slate-100 text-linkflow-muted dark:bg-slate-800 dark:text-linkflow-dark-muted';
  }
};

const AckStatusBadge = ({ status }: { status: DeviceServiceAckStatus }) => {
  const { t } = useI18n();
  const labelKey =
    status === 'success'
      ? 'adminDeviceServiceAckSuccess'
      : status === 'failed'
      ? 'adminDeviceServiceAckFailed'
      : 'adminDeviceServiceAckPending';

  return (
    <span
      className={[
        'inline-flex h-7 items-center rounded-full px-3 text-xs font-black leading-none',
        ackStatusClassName(status),
      ].join(' ')}
    >
      {t(labelKey)}
    </span>
  );
};

const serviceEntriesOf = (model: ThingsModel | null) => {
  return Object.entries(model?.services ?? {}).filter(([, value]) =>
    isRecord(value),
  );
};

const serviceDisplayName = (name: string, definition: unknown) => {
  if (!isRecord(definition) || typeof definition.name !== 'string') {
    return name;
  }
  return definition.name.trim() ? definition.name : name;
};

const serviceInputDefinition = (definition: unknown) => {
  if (!isRecord(definition) || !isRecord(definition.input)) {
    return {};
  }
  return definition.input;
};

const defaultServiceInputText = (definition: unknown) => {
  const input = serviceInputDefinition(definition);
  const values: Record<string, unknown> = {};

  Object.entries(input).forEach(([key, rawField]) => {
    if (!isRecord(rawField) || rawField.required !== true) {
      return;
    }

    switch (rawField.data_type) {
      case 'int':
      case 'float':
      case 'double':
        values[key] = 0;
        break;
      case 'bool':
        values[key] = false;
        break;
      case 'string':
        values[key] = '';
        break;
      default:
        values[key] = null;
        break;
    }
  });

  return JSON.stringify(values, null, 2);
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
  const columns: AdminDataTableColumn<DeviceEventEntry>[] = [
    {
      key: 'event_name',
      title: t('adminDeviceEventName'),
      render: (entry) => (
        <div className="min-w-0">
          <p className="m-0 break-all text-sm font-black text-linkflow-text dark:text-linkflow-dark-text">
            {entry.event_name}
          </p>
          <p className="m-0 mt-1 break-all text-xs font-semibold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {entry.product_key}/{entry.device_slug}
          </p>
        </div>
      ),
    },
    {
      key: 'occurred_at',
      title: t('adminDevicePropertyOccurredAt'),
      render: (entry) => <TimeCell value={entry.occurred_at} />,
    },
    {
      key: 'received_at',
      title: t('adminDevicePropertyReceivedAt'),
      render: (entry) => <TimeCell value={entry.received_at} />,
    },
    {
      key: 'params',
      title: t('adminDeviceEventParams'),
      className: 'min-w-80',
      render: (entry) => <JsonPreview value={entry.params} />,
    },
  ];

  const mobileItems = events.map((entry) => (
    <article
      key={entry.event_id}
      className="grid gap-3 rounded-lg border border-linkflow-border bg-white p-4 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel"
    >
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="m-0 break-all text-base font-black">
            {entry.event_name}
          </p>
          <p className="m-0 mt-1 break-all text-xs font-semibold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {entry.product_key}/{entry.device_slug}
          </p>
        </div>
      </div>
      <dl className="m-0 grid gap-3 text-sm">
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDevicePropertyOccurredAt')}
          </dt>
          <dd className="m-0 mt-1">
            <TimeCell value={entry.occurred_at} />
          </dd>
        </div>
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDevicePropertyReceivedAt')}
          </dt>
          <dd className="m-0 mt-1">
            <TimeCell value={entry.received_at} />
          </dd>
        </div>
      </dl>
      <JsonPreview value={entry.params} />
    </article>
  ));

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

      <AdminDataTable
        columns={columns}
        emptyDescription={t('adminDeviceEventHistoryEmpty')}
        emptyTitle={t('adminDeviceEventHistoryEmpty')}
        getRowKey={(entry) => entry.event_id}
        items={events}
        minWidthClassName="min-w-[980px]"
        mobileItems={mobileItems}
        pagination={{
          jumpLabel: t('adminDeviceJump'),
          jumpToLabel: t('adminDeviceJumpTo'),
          loading,
          loadingLabel: t('adminDeviceEventLoadingHistory'),
          nextLabel: t('adminDeviceNext'),
          onPageChange,
          onPageSizeChange,
          page,
          pageLabel: t('adminDeviceEventPage')
            .replace('{page}', String(page))
            .replace('{totalPages}', String(totalPages)),
          pageSize,
          pageSizeLabel: t('adminDevicePageSize'),
          pageSizeOptions: [10, 20, 50, 100],
          prevLabel: t('adminDevicePrev'),
          total,
          totalLabel: t('adminDeviceEventTotal').replace(
            '{total}',
            String(total),
          ),
        }}
      />
    </section>
  );
};

interface DeviceServiceCallPanelProps {
  calling: boolean;
  currentModel: ThingsModel | null;
  device: Device | undefined;
  inputText: string;
  loadingServices: boolean;
  result: DeviceServiceCallResult | null;
  serviceName: string;
  onCall: () => void;
  onInputTextChange: (text: string) => void;
  onServiceNameChange: (serviceName: string) => void;
}

const DeviceServiceCallPanel = ({
  calling,
  currentModel,
  device,
  inputText,
  loadingServices,
  result,
  serviceName,
  onCall,
  onInputTextChange,
  onServiceNameChange,
}: DeviceServiceCallPanelProps) => {
  const { t } = useI18n();
  const services = serviceEntriesOf(currentModel);
  const selectedDefinition = serviceName
    ? currentModel?.services[serviceName]
    : undefined;
  const canCall =
    Boolean(device) &&
    device?.status === 'active' &&
    device?.connection_status === 'online' &&
    Boolean(serviceName) &&
    !calling;

  return (
    <section className="grid gap-4">
      <div className="grid gap-4 rounded-lg border border-linkflow-border bg-white p-4 shadow-[0_8px_24px_rgba(23,32,51,0.06)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/20 lg:grid-cols-[minmax(0,1fr)_minmax(18rem,0.8fr)]">
        <div className="grid gap-4">
          <div className="flex flex-wrap items-end justify-between gap-3">
            <label className="grid min-w-64 flex-1 gap-2 text-sm font-bold">
              {t('adminDeviceServiceName')}
              <select
                className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary disabled:opacity-70 dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                value={serviceName}
                onChange={(event) => onServiceNameChange(event.target.value)}
                disabled={!device || loadingServices || services.length === 0}
              >
                {services.length === 0 ? (
                  <option value="">{t('adminDeviceServiceNoService')}</option>
                ) : null}
                {services.map(([name, definition]) => (
                  <option key={name} value={name}>
                    {serviceDisplayName(name, definition)} ({name})
                  </option>
                ))}
              </select>
            </label>
            <button
              type="button"
              className="inline-flex h-11 items-center gap-2 rounded-md border border-linkflow-primary bg-linkflow-primary px-4 text-sm font-black text-white shadow-sm transition hover:bg-linkflow-primary-hover disabled:cursor-not-allowed disabled:opacity-60 dark:border-linkflow-dark-primary dark:bg-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-hover"
              onClick={onCall}
              disabled={!canCall}
            >
              <SendOutlined aria-hidden="true" />
              {calling
                ? t('adminDeviceServiceCalling')
                : t('adminDeviceServiceCall')}
            </button>
          </div>

          <label className="grid gap-2 text-sm font-bold">
            {t('adminDeviceServiceInput')}
            <textarea
              className="min-h-56 resize-y rounded-md border border-linkflow-border bg-white p-3 font-mono text-sm font-semibold leading-6 outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
              value={inputText}
              onChange={(event) => onInputTextChange(event.target.value)}
              spellCheck={false}
            />
          </label>
        </div>

        <aside className="grid content-start gap-4">
          <div className="rounded-lg border border-linkflow-border bg-slate-50/80 p-3 dark:border-linkflow-dark-border dark:bg-slate-950/30">
            <p className="m-0 mb-2 text-xs font-bold uppercase text-linkflow-subtle dark:text-linkflow-dark-subtle">
              {t('adminDeviceServiceInputDefinition')}
            </p>
            <pre className="m-0 max-h-64 overflow-auto whitespace-pre-wrap break-words rounded-md bg-linkflow-code p-3 font-mono text-xs font-semibold leading-5 text-linkflow-code-text">
              {JSON.stringify(serviceInputDefinition(selectedDefinition), null, 2)}
            </pre>
          </div>

          <dl className="m-0 grid gap-3 rounded-lg border border-linkflow-border bg-white p-3 text-sm dark:border-linkflow-dark-border dark:bg-linkflow-dark-page">
            <div>
              <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
                {t('adminDeviceConnectionStatus')}
              </dt>
              <dd className="m-0 mt-1 font-black">
                {device
                  ? t(
                      device.connection_status === 'online'
                        ? 'adminDeviceConnectionOnline'
                        : 'adminDeviceConnectionOffline',
                    )
                  : t('adminDevicePropertyNoDevice')}
              </dd>
            </div>
            <div>
              <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
                {t('adminThingsModelCurrent')}
              </dt>
              <dd className="m-0 mt-1 break-all font-black">
                {currentModel
                  ? `${currentModel.model_name} v${currentModel.model_version}`
                  : t('adminDeviceServiceNoCurrentModel')}
              </dd>
            </div>
          </dl>
        </aside>
      </div>

      {result ? (
        <article className="grid gap-4 rounded-lg border border-linkflow-border bg-white p-4 shadow-[0_8px_24px_rgba(23,32,51,0.06)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/20">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <p className="m-0 text-xs font-bold uppercase text-linkflow-subtle dark:text-linkflow-dark-subtle">
                {t('adminDeviceServiceResult')}
              </p>
              <h3 className="m-0 mt-1 break-all text-lg font-black text-linkflow-text dark:text-linkflow-dark-text">
                {result.command_id}
              </h3>
            </div>
            <span className="rounded-full bg-emerald-50 px-3 py-1 text-xs font-black text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300">
              {t('adminDeviceServiceDispatched')}
            </span>
          </div>
          <dl className="m-0 grid gap-3 text-sm [grid-template-columns:repeat(auto-fit,minmax(180px,1fr))]">
            <div>
              <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
                {t('adminDeviceServiceTopic')}
              </dt>
              <dd className="m-0 mt-1 break-all font-semibold">
                {result.topic}
              </dd>
            </div>
            <div>
              <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
                {t('adminDeviceServiceSentAt')}
              </dt>
              <dd className="m-0 mt-1 font-semibold">
                {formatDateTime(result.sent_at)}
              </dd>
            </div>
          </dl>
        </article>
      ) : null}
    </section>
  );
};

interface DeviceServiceHistoryPanelProps {
  ackDeadlineSeconds: number;
  calls: DeviceServiceCallHistoryEntry[];
  loading: boolean;
  page: number;
  pageSize: number;
  serviceName: string;
  total: number;
  onAckDeadlineSecondsChange: (seconds: number) => void;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  onReload: () => void;
  onServiceNameChange: (serviceName: string) => void;
}

const DeviceServiceHistoryPanel = ({
  ackDeadlineSeconds,
  calls,
  loading,
  page,
  pageSize,
  serviceName,
  total,
  onAckDeadlineSecondsChange,
  onPageChange,
  onPageSizeChange,
  onReload,
  onServiceNameChange,
}: DeviceServiceHistoryPanelProps) => {
  const { t } = useI18n();
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const columns: AdminDataTableColumn<DeviceServiceCallHistoryEntry>[] = [
    {
      key: 'service_name',
      title: t('adminDeviceServiceName'),
      render: (entry) => (
        <div className="min-w-0">
          <p className="m-0 break-all text-sm font-black text-linkflow-text dark:text-linkflow-dark-text">
            {entry.service_name}
          </p>
          <p className="m-0 mt-1 break-all font-mono text-xs font-semibold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {entry.command_id}
          </p>
        </div>
      ),
    },
    {
      key: 'ack_status',
      title: t('adminDeviceServiceAckStatus'),
      render: (entry) => <AckStatusBadge status={entry.ack_status} />,
    },
    {
      key: 'occurred_at',
      title: t('adminDeviceServiceSentAt'),
      render: (entry) => <TimeCell value={entry.occurred_at} />,
    },
    {
      key: 'ack_occurred_at',
      title: t('adminDeviceServiceAckAt'),
      render: (entry) => (
        <TimeCell value={entry.ack_occurred_at ?? entry.ack_deadline_at} />
      ),
    },
    {
      key: 'input',
      title: t('adminDeviceServiceInput'),
      className: 'min-w-72',
      render: (entry) => <JsonPreview value={entry.input} />,
    },
    {
      key: 'ack_output',
      title: t('adminDeviceServiceAckOutput'),
      className: 'min-w-72',
      render: (entry) => (
        <JsonPreview value={entry.ack_output ?? {}} />
      ),
    },
  ];

  const mobileItems = calls.map((entry) => (
    <article
      key={entry.command_id}
      className="grid gap-3 rounded-lg border border-linkflow-border bg-white p-4 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel"
    >
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="m-0 break-all text-base font-black">
            {entry.service_name}
          </p>
          <p className="m-0 mt-1 break-all font-mono text-xs font-semibold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {entry.command_id}
          </p>
        </div>
        <AckStatusBadge status={entry.ack_status} />
      </div>
      <dl className="m-0 grid gap-3 text-sm">
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDeviceServiceSentAt')}
          </dt>
          <dd className="m-0 mt-1">
            <TimeCell value={entry.occurred_at} />
          </dd>
        </div>
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDeviceServiceAckAt')}
          </dt>
          <dd className="m-0 mt-1">
            <TimeCell value={entry.ack_occurred_at ?? entry.ack_deadline_at} />
          </dd>
        </div>
      </dl>
      <div className="grid gap-3">
        <JsonPreview value={entry.input} />
        <JsonPreview value={entry.ack_output ?? {}} />
      </div>
    </article>
  ));

  return (
    <section className="grid gap-4">
      <div className="flex flex-wrap items-end justify-between gap-3 rounded-lg border border-linkflow-border bg-white p-4 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel">
        <div className="flex flex-wrap items-end gap-3">
          <label className="grid min-w-56 gap-2 text-sm font-bold">
            {t('adminDeviceServiceName')}
            <input
              className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
              value={serviceName}
              onChange={(event) => onServiceNameChange(event.target.value)}
              placeholder={t('adminDeviceServiceAllServices')}
            />
          </label>
          <label className="grid gap-2 text-sm font-bold">
            {t('adminDeviceServiceAckDeadline')}
            <select
              className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
              value={ackDeadlineSeconds}
              onChange={(event) =>
                onAckDeadlineSecondsChange(Number(event.target.value))
              }
            >
              {[30, 60, 90, 180, 300].map((seconds) => (
                <option key={seconds} value={seconds}>
                  {t('adminDeviceServiceAckDeadlineSeconds').replace(
                    '{seconds}',
                    String(seconds),
                  )}
                </option>
              ))}
            </select>
          </label>
        </div>
        <button
          type="button"
          className="inline-flex h-11 items-center gap-2 rounded-md border border-linkflow-primary bg-white/70 px-3 text-sm font-bold text-linkflow-primary shadow-sm transition hover:bg-linkflow-primary-soft disabled:cursor-not-allowed disabled:opacity-60 dark:border-linkflow-dark-primary dark:bg-linkflow-dark-panel/70 dark:text-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-soft"
          onClick={onReload}
          disabled={loading}
        >
          <ReloadOutlined aria-hidden="true" />
          {t('adminDeviceServiceHistoryReload')}
        </button>
      </div>

      <AdminDataTable
        columns={columns}
        emptyDescription={t('adminDeviceServiceHistoryEmpty')}
        emptyTitle={t('adminDeviceServiceHistoryEmpty')}
        getRowKey={(entry) => entry.command_id}
        items={calls}
        minWidthClassName="min-w-[1120px]"
        mobileItems={mobileItems}
        pagination={{
          jumpLabel: t('adminDeviceJump'),
          jumpToLabel: t('adminDeviceJumpTo'),
          loading,
          loadingLabel: t('adminDeviceServiceHistoryLoading'),
          nextLabel: t('adminDeviceNext'),
          onPageChange,
          onPageSizeChange,
          page,
          pageLabel: t('adminDeviceServiceHistoryPage')
            .replace('{page}', String(page))
            .replace('{totalPages}', String(totalPages)),
          pageSize,
          pageSizeLabel: t('adminDevicePageSize'),
          pageSizeOptions: [10, 20, 50, 100],
          prevLabel: t('adminDevicePrev'),
          total,
          totalLabel: t('adminDeviceServiceHistoryTotal').replace(
            '{total}',
            String(total),
          ),
        }}
      />
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
  const [serviceHistoryEntries, setServiceHistoryEntries] = useState<
    DeviceServiceCallHistoryEntry[]
  >([]);
  const [serviceHistoryTotal, setServiceHistoryTotal] = useState(0);
  const [serviceHistoryPage, setServiceHistoryPage] = useState(1);
  const [serviceHistoryPageSize, setServiceHistoryPageSize] = useState(20);
  const [serviceHistoryNameFilter, setServiceHistoryNameFilter] = useState('');
  const [serviceAckDeadlineSeconds, setServiceAckDeadlineSeconds] =
    useState(90);
  const [loadingServiceHistory, setLoadingServiceHistory] = useState(false);
  const [currentModel, setCurrentModel] = useState<ThingsModel | null>(null);
  const [loadingServices, setLoadingServices] = useState(false);
  const [serviceName, setServiceName] = useState('');
  const [serviceInputText, setServiceInputText] = useState('{}');
  const [callingService, setCallingService] = useState(false);
  const [serviceCallResult, setServiceCallResult] =
    useState<DeviceServiceCallResult | null>(null);
  const latestInboxEvent = useEventInboxStore((state) => state.items[0]);
  const selectedDevice = devices.find((device) => device.id === deviceId);
  const selectedProductId = selectedDevice?.product_id || productFilter;

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

  useEffect(() => {
    let cancelled = false;

    const run = async () => {
      if (!tenantId || !selectedProductId) {
        setCurrentModel(null);
        setServiceName('');
        setServiceInputText('{}');
        setServiceCallResult(null);
        return;
      }

      setLoadingServices(true);
      try {
        const result = await listThingsModels({
          tenant_id: tenantId,
          product_id: selectedProductId,
          page: 1,
          page_size: 100,
        });
        if (cancelled) {
          return;
        }

        const nextModel =
          result.items.find((model) => model.is_current) ?? null;
        const services = serviceEntriesOf(nextModel);
        const nextServiceName = services[0]?.[0] ?? '';
        setCurrentModel(nextModel);
        setServiceName(nextServiceName);
        setServiceInputText(
          nextServiceName
            ? defaultServiceInputText(nextModel?.services[nextServiceName])
            : '{}',
        );
        setServiceCallResult(null);
      } catch (error) {
        if (cancelled) {
          return;
        }
        setCurrentModel(null);
        setServiceName('');
        setServiceInputText('{}');
        setServiceCallResult(null);
        notification.error({
          message: t('adminThingsModelLoadFailed'),
          description: error instanceof Error ? error.message : undefined,
          placement: 'topRight',
        });
      } finally {
        if (!cancelled) {
          setLoadingServices(false);
        }
      }
    };

    run();

    return () => {
      cancelled = true;
    };
  }, [selectedProductId, tenantId, t]);

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

  const loadServiceHistory = async (
    nextDeviceId = deviceId,
    options: {
      serviceName?: string;
      ackDeadlineSeconds?: number;
      page?: number;
      pageSize?: number;
      silent?: boolean;
      isCancelled?: () => boolean;
    } = {},
  ) => {
    if (!nextDeviceId) {
      setServiceHistoryEntries([]);
      setServiceHistoryTotal(0);
      return;
    }

    if (!options.silent) {
      setLoadingServiceHistory(true);
    }
    try {
      const result = await listDeviceServiceCalls(nextDeviceId, {
        service_name:
          (options.serviceName ?? serviceHistoryNameFilter).trim() ||
          undefined,
        ack_deadline_seconds:
          options.ackDeadlineSeconds ?? serviceAckDeadlineSeconds,
        page: options.page ?? serviceHistoryPage,
        page_size: options.pageSize ?? serviceHistoryPageSize,
      });
      if (options.isCancelled?.()) {
        return;
      }
      setServiceHistoryEntries(result.items);
      setServiceHistoryTotal(result.total);
      setServiceHistoryPage(result.page);
      setServiceHistoryPageSize(result.page_size);
    } catch (error) {
      if (options.isCancelled?.() || options.silent) {
        return;
      }
      notification.error({
        message: t('adminDeviceServiceHistoryFailed'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    } finally {
      if (!options.silent) {
        setLoadingServiceHistory(false);
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
      if (activeTab !== 'eventHistory') {
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
    let cancelled = false;

    const run = async () => {
      if (activeTab !== 'serviceHistory') {
        return;
      }
      if (!deviceId) {
        setServiceHistoryEntries([]);
        setServiceHistoryTotal(0);
        return;
      }

      setLoadingServiceHistory(true);
      try {
        const result = await listDeviceServiceCalls(deviceId, {
          service_name: serviceHistoryNameFilter.trim() || undefined,
          ack_deadline_seconds: serviceAckDeadlineSeconds,
          page: serviceHistoryPage,
          page_size: serviceHistoryPageSize,
        });
        if (cancelled) {
          return;
        }
        setServiceHistoryEntries(result.items);
        setServiceHistoryTotal(result.total);
        setServiceHistoryPage(result.page);
        setServiceHistoryPageSize(result.page_size);
      } catch (error) {
        if (cancelled) {
          return;
        }
        notification.error({
          message: t('adminDeviceServiceHistoryFailed'),
          description: error instanceof Error ? error.message : undefined,
          placement: 'topRight',
        });
      } finally {
        if (!cancelled) {
          setLoadingServiceHistory(false);
        }
      }
    };

    run();

    return () => {
      cancelled = true;
    };
  }, [
    activeTab,
    deviceId,
    serviceAckDeadlineSeconds,
    serviceHistoryNameFilter,
    serviceHistoryPage,
    serviceHistoryPageSize,
    t,
  ]);

  useEffect(() => {
    if (
      !latestInboxEvent ||
      activeTab !== 'eventHistory' ||
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
        if (activeTab === 'eventHistory') {
          void loadEventHistory(deviceId, { silent: true, isCancelled });
        }
        if (activeTab === 'serviceHistory') {
          void loadServiceHistory(deviceId, { silent: true, isCancelled });
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
    serviceAckDeadlineSeconds,
    serviceHistoryNameFilter,
    serviceHistoryPage,
    serviceHistoryPageSize,
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
    setServiceHistoryEntries([]);
    setServiceHistoryTotal(0);
    setServiceHistoryPage(1);
    setServiceCallResult(null);
  };

  const handleProductFilterChange = (nextProductId: string) => {
    setProductFilter(nextProductId);
    setDeviceId('');
    setLatest(null);
    setEventEntries([]);
    setEventTotal(0);
    setEventPage(1);
    setServiceHistoryEntries([]);
    setServiceHistoryTotal(0);
    setServiceHistoryPage(1);
    setServiceCallResult(null);
  };

  const handleDeviceChange = (nextDeviceId: string) => {
    setDeviceId(nextDeviceId);
    setLatest(null);
    setEventEntries([]);
    setEventTotal(0);
    setEventPage(1);
    setServiceHistoryEntries([]);
    setServiceHistoryTotal(0);
    setServiceHistoryPage(1);
    setServiceCallResult(null);
  };

  const handleEventNameChange = (nextEventName: string) => {
    setEventNameFilter(nextEventName);
    setEventPage(1);
  };

  const handleEventPageSizeChange = (nextPageSize: number) => {
    setEventPageSize(nextPageSize);
    setEventPage(1);
  };

  const handleServiceHistoryNameChange = (nextServiceName: string) => {
    setServiceHistoryNameFilter(nextServiceName);
    setServiceHistoryPage(1);
  };

  const handleServiceHistoryPageSizeChange = (nextPageSize: number) => {
    setServiceHistoryPageSize(nextPageSize);
    setServiceHistoryPage(1);
  };

  const handleServiceAckDeadlineSecondsChange = (nextSeconds: number) => {
    setServiceAckDeadlineSeconds(nextSeconds);
    setServiceHistoryPage(1);
  };

  const handleServiceNameChange = (nextServiceName: string) => {
    setServiceName(nextServiceName);
    setServiceInputText(
      nextServiceName
        ? defaultServiceInputText(currentModel?.services[nextServiceName])
        : '{}',
    );
    setServiceCallResult(null);
  };

  const handleCallService = async () => {
    if (!deviceId || !serviceName) {
      return;
    }

    let input: Record<string, unknown>;
    try {
      const parsed = JSON.parse(serviceInputText.trim() || '{}') as unknown;
      if (!isRecord(parsed)) {
        throw new Error(t('adminDeviceServiceInputMustObject'));
      }
      input = parsed;
    } catch (error) {
      notification.error({
        message: t('adminDeviceServiceInvalidInput'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
      return;
    }

    setCallingService(true);
    try {
      const result = await callDeviceService(deviceId, serviceName, { input });
      setServiceCallResult(result);
      notification.success({
        message: t('adminDeviceServiceCallSuccess'),
        description: result.command_id,
        placement: 'topRight',
      });
      if (activeTab === 'services') {
        void loadServiceHistory(deviceId, {
          page: 1,
          silent: true,
        });
      }
    } catch (error) {
      notification.error({
        message: t('adminDeviceServiceCallFailed'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    } finally {
      setCallingService(false);
    }
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
          aria-selected={activeTab === 'services'}
          className={tabButtonClassName(activeTab === 'services')}
          onClick={() => setActiveTab('services')}
        >
          <ApiOutlined aria-hidden="true" />
          {t('adminDeviceServiceTab')}
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'serviceHistory'}
          className={tabButtonClassName(activeTab === 'serviceHistory')}
          onClick={() => setActiveTab('serviceHistory')}
        >
          <FieldTimeOutlined aria-hidden="true" />
          {t('adminDeviceServiceHistoryTab')}
        </button>
        <button
          type="button"
          role="tab"
          aria-selected={activeTab === 'eventHistory'}
          className={tabButtonClassName(activeTab === 'eventHistory')}
          onClick={() => setActiveTab('eventHistory')}
        >
          <LineChartOutlined aria-hidden="true" />
          {t('adminDeviceEventHistoryTab')}
        </button>
      </div>

      {activeTab === 'latest' ? (
        <LatestPropertiesPanel latest={latest} loading={loadingLatest} />
      ) : activeTab === 'history' ? (
        <HistoryPropertiesPanel />
      ) : activeTab === 'services' ? (
        <DeviceServiceCallPanel
          calling={callingService}
          currentModel={currentModel}
          device={selectedDevice}
          inputText={serviceInputText}
          loadingServices={loadingServices}
          onCall={handleCallService}
          onInputTextChange={setServiceInputText}
          onServiceNameChange={handleServiceNameChange}
          result={serviceCallResult}
          serviceName={serviceName}
        />
      ) : activeTab === 'serviceHistory' ? (
        <DeviceServiceHistoryPanel
          ackDeadlineSeconds={serviceAckDeadlineSeconds}
          calls={serviceHistoryEntries}
          loading={loadingServiceHistory}
          onAckDeadlineSecondsChange={handleServiceAckDeadlineSecondsChange}
          onPageChange={setServiceHistoryPage}
          onPageSizeChange={handleServiceHistoryPageSizeChange}
          onReload={() => loadServiceHistory()}
          onServiceNameChange={handleServiceHistoryNameChange}
          page={serviceHistoryPage}
          pageSize={serviceHistoryPageSize}
          serviceName={serviceHistoryNameFilter}
          total={serviceHistoryTotal}
        />
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
