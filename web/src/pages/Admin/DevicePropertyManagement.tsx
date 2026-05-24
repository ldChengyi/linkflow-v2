import { useI18n } from '@/contexts/I18nContext';
import { useDevicesRealtime } from '@/hooks/useDevicesRealtime';
import {
  callDeviceService,
  getDeviceLatestProperties,
  getDevicePropertyTrend,
  listDeviceEvents,
  listDevicePropertySets,
  listDeviceServiceCalls,
  listDevices,
  setDeviceProperties,
  type Device,
  type DeviceEventEntry,
  type DeviceLatestProperties,
  type DevicePropertySetHistoryEntry,
  type DevicePropertySetResult,
  type DevicePropertyTrend,
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
  ApiOutlined,
  AppstoreOutlined,
  CheckCircleOutlined,
  CloseOutlined,
  DisconnectOutlined,
  DownOutlined,
  FieldTimeOutlined,
  LineChartOutlined,
  PlusOutlined,
  ReloadOutlined,
  SendOutlined,
} from '@ant-design/icons';
import { Tooltip, notification } from 'antd';
import { LineChart, type LineSeriesOption } from 'echarts/charts';
import {
  DataZoomComponent,
  GridComponent,
  LegendComponent,
  TooltipComponent,
  type DataZoomComponentOption,
  type GridComponentOption,
  type LegendComponentOption,
  type TooltipComponentOption,
} from 'echarts/components';
import {
  init as initECharts,
  use as useECharts,
  type ComposeOption,
  type ECharts,
} from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { useEffect, useRef, useState } from 'react';
import AdminDataTable, {
  type AdminDataTableColumn,
} from './components/AdminDataTable';
import RealtimeStatusBadge, {
  isRealtimeFallbackStatus,
  realtimeFallbackRefreshMs,
} from './components/RealtimeStatusBadge';

useECharts([
  LineChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent,
  CanvasRenderer,
]);

type TrendChartOption = ComposeOption<
  | LineSeriesOption
  | GridComponentOption
  | TooltipComponentOption
  | LegendComponentOption
  | DataZoomComponentOption
>;

type DevicePropertyTab =
  | 'latest'
  | 'history'
  | 'set'
  | 'services'
  | 'propertySetHistory'
  | 'serviceHistory'
  | 'eventHistory';

type DevicePropertyTrendAggregate = 'avg' | 'min' | 'max' | 'last';

interface DeviceTrendPropertyOption {
  key: string;
  label: string;
  unit?: string;
}

type ServiceInputValueType =
  | 'string'
  | 'int'
  | 'float'
  | 'double'
  | 'bool'
  | 'json';

interface ServiceInputParamDraft {
  id: string;
  name: string;
  value: string;
  valueType: ServiceInputValueType;
}

const trendLineColors = [
  '#128999',
  '#7c3aed',
  '#059669',
  '#f59e0b',
  '#dc2626',
  '#2563eb',
  '#9333ea',
  '#0f766e',
];

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
  compact?: boolean;
  maxHeightClassName?: string;
}

const JsonPreview = ({
  value,
  compact = false,
  maxHeightClassName = 'max-h-36',
}: JsonPreviewProps) => {
  const formatted = JSON.stringify(value ?? {}, null, 2);
  const summary = JSON.stringify(value ?? {});

  if (compact) {
    return (
      <details className="group rounded-md border border-linkflow-border bg-white p-2 shadow-sm dark:border-linkflow-dark-border dark:bg-linkflow-dark-page">
        <summary
          className="relative flex min-h-8 min-w-0 cursor-pointer list-none items-center pr-10 text-xs font-bold text-linkflow-muted outline-none transition hover:text-linkflow-primary dark:text-linkflow-dark-muted dark:hover:text-linkflow-dark-primary [&::-webkit-details-marker]:hidden"
          title={formatted}
        >
          <span className="flex h-8 min-w-0 flex-1 items-center truncate font-mono leading-none">
            {summary}
          </span>
          <span className="absolute right-0 top-1/2 grid h-8 w-8 -translate-y-1/2 place-items-center rounded-md border border-linkflow-border bg-white text-linkflow-muted transition group-open:rotate-180 group-open:border-linkflow-primary group-open:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:group-open:border-linkflow-dark-primary dark:group-open:text-linkflow-dark-primary">
            <DownOutlined aria-hidden="true" />
          </span>
        </summary>
        <pre className="m-0 mt-2 whitespace-pre-wrap break-words rounded-md border border-linkflow-border bg-white p-3 font-mono text-xs font-semibold leading-5 text-linkflow-text dark:border-linkflow-dark-border dark:bg-linkflow-dark-page dark:text-linkflow-dark-text">
          {formatted}
        </pre>
      </details>
    );
  }

  return (
    <pre
      className={[
        'm-0 overflow-auto whitespace-pre-wrap break-words rounded-md border border-linkflow-border bg-white p-3 font-mono text-xs font-semibold leading-5 text-linkflow-text dark:border-linkflow-dark-border dark:bg-linkflow-dark-page dark:text-linkflow-dark-text',
        maxHeightClassName,
      ].join(' ')}
      title={formatted}
    >
      {formatted}
    </pre>
  );
};

const shortenID = (value: string) => {
  if (value.length <= 18) {
    return value;
  }
  return `${value.slice(0, 8)}...${value.slice(-6)}`;
};

const ShortID = ({ value }: { value: string }) => (
  <Tooltip title={value}>
    <span className="inline-flex max-w-full rounded-md bg-slate-50 px-2 py-1 font-mono text-xs font-bold text-linkflow-muted dark:bg-slate-900/60 dark:text-linkflow-dark-muted">
      {shortenID(value)}
    </span>
  </Tooltip>
);

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

const connectionStatusClassName = (status?: Device['connection_status']) => {
  if (status === 'online') {
    return 'border-emerald-200 bg-emerald-50 text-emerald-700 dark:border-emerald-900/70 dark:bg-emerald-950/40 dark:text-emerald-300';
  }

  return 'border-rose-200 bg-rose-50 text-rose-700 dark:border-rose-900/70 dark:bg-rose-950/40 dark:text-rose-300';
};

const ConnectionStatusBadge = ({ device }: { device?: Device }) => {
  const { t } = useI18n();
  const online = device?.connection_status === 'online';

  return (
    <span
      className={[
        'inline-flex h-10 items-center gap-2 rounded-md border px-3 text-xs font-black leading-none',
        connectionStatusClassName(device?.connection_status),
      ].join(' ')}
    >
      {online ? (
        <CheckCircleOutlined aria-hidden="true" />
      ) : (
        <DisconnectOutlined aria-hidden="true" />
      )}
      {device
        ? t(
            online
              ? 'adminDeviceConnectionOnline'
              : 'adminDeviceConnectionOffline',
          )
        : t('adminDevicePropertyNoDevice')}
    </span>
  );
};

const DeviceContextStatusBar = ({
  currentModel,
  device,
}: {
  currentModel: ThingsModel | null;
  device?: Device;
}) => {
  const { t } = useI18n();

  return (
    <div className="grid gap-3 rounded-lg border border-linkflow-border bg-white p-3 text-sm shadow-sm dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel [grid-template-columns:repeat(auto-fit,minmax(260px,1fr))]">
      <div className="flex h-14 items-center justify-between gap-3 rounded-md border border-linkflow-border bg-slate-50/70 px-3 dark:border-linkflow-dark-border dark:bg-slate-950/25">
        <span className="inline-flex h-10 shrink-0 items-center text-xs font-black uppercase leading-none tracking-wide text-linkflow-subtle dark:text-linkflow-dark-subtle">
          {t('adminDeviceConnectionStatus')}
        </span>
        <span className="inline-flex h-10 min-w-0 items-center justify-end">
          <ConnectionStatusBadge device={device} />
        </span>
      </div>
      <div className="flex h-14 min-w-0 items-center justify-between gap-3 rounded-md border border-linkflow-border bg-slate-50/70 px-3 dark:border-linkflow-dark-border dark:bg-slate-950/25">
        <span className="inline-flex h-10 shrink-0 items-center text-xs font-black uppercase leading-none tracking-wide text-linkflow-subtle dark:text-linkflow-dark-subtle">
          {t('adminThingsModelCurrent')}
        </span>
        <span
          className="inline-flex h-10 min-w-0 items-center justify-end gap-2"
          title={
            currentModel
              ? `${currentModel.model_name} v${currentModel.model_version}`
              : t('adminDeviceServiceNoCurrentModel')
          }
        >
          {currentModel ? (
            <>
              <span className="inline-flex h-10 min-w-0 max-w-full items-center truncate rounded-md bg-white px-3 font-black leading-none text-linkflow-text shadow-sm dark:bg-linkflow-dark-page dark:text-linkflow-dark-text">
                {currentModel.model_name}
              </span>
              <span className="inline-flex h-10 shrink-0 items-center rounded-md border border-linkflow-primary/25 bg-linkflow-primary-soft px-2.5 font-mono text-xs font-black leading-none text-linkflow-primary dark:border-linkflow-dark-primary/30 dark:bg-linkflow-dark-primary-soft dark:text-linkflow-dark-primary">
                v{currentModel.model_version}
              </span>
            </>
          ) : (
            <span className="inline-flex h-10 min-w-0 items-center truncate rounded-md bg-white px-3 font-black leading-none text-linkflow-subtle shadow-sm dark:bg-linkflow-dark-page dark:text-linkflow-dark-subtle">
              {t('adminDeviceServiceNoCurrentModel')}
            </span>
          )}
        </span>
      </div>
    </div>
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

const createDraftID = () => {
  if (typeof crypto !== 'undefined' && 'randomUUID' in crypto) {
    return crypto.randomUUID();
  }

  return `${Date.now()}-${Math.random().toString(16).slice(2)}`;
};

const serviceInputValueType = (definition: unknown): ServiceInputValueType => {
  const dataType = propertyDataType(definition);
  switch (dataType) {
    case 'int':
    case 'float':
    case 'double':
    case 'bool':
    case 'string':
      return dataType;
    default:
      return 'json';
  }
};

const defaultServiceInputValue = (
  definition: unknown,
  valueType = serviceInputValueType(definition),
) => {
  switch (valueType) {
    case 'int':
    case 'float':
    case 'double':
      return '0';
    case 'bool':
      return 'false';
    case 'string':
      return '';
    case 'json':
      return 'null';
    default:
      return '';
  }
};

const defaultServiceInputParams = (
  definition: unknown,
): ServiceInputParamDraft[] => {
  const input = serviceInputDefinition(definition);

  return Object.entries(input)
    .filter(([, rawField]) => isRecord(rawField) && rawField.required === true)
    .map(([key, rawField]) => {
      const valueType = serviceInputValueType(rawField);
      return {
        id: createDraftID(),
        name: key,
        value: defaultServiceInputValue(rawField, valueType),
        valueType,
      };
    });
};

const serviceInputParamsToRecord = (
  params: ServiceInputParamDraft[],
): Record<string, unknown> => {
  const values: Record<string, unknown> = {};
  const seen = new Set<string>();

  for (const param of params) {
    const name = param.name.trim();
    if (!name) {
      throw new Error('Service input parameter name is required');
    }
    if (seen.has(name)) {
      throw new Error(`Duplicate service input parameter: ${name}`);
    }
    seen.add(name);

    switch (param.valueType) {
      case 'int': {
        const value = Number(param.value);
        if (!Number.isInteger(value)) {
          throw new Error(`${name} must be an integer`);
        }
        values[name] = value;
        break;
      }
      case 'float':
      case 'double': {
        const value = Number(param.value);
        if (!Number.isFinite(value)) {
          throw new Error(`${name} must be a number`);
        }
        values[name] = value;
        break;
      }
      case 'bool':
        values[name] = param.value === 'true';
        break;
      case 'json':
        values[name] = JSON.parse(param.value || 'null') as unknown;
        break;
      default:
        values[name] = param.value;
        break;
    }
  }

  return values;
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

const propertyDisplayName = (name: string, definition: unknown) => {
  if (!isRecord(definition) || typeof definition.name !== 'string') {
    return name;
  }

  return definition.name.trim() ? definition.name : name;
};

const propertyDataType = (definition: unknown) => {
  if (!isRecord(definition) || typeof definition.data_type !== 'string') {
    return undefined;
  }

  return definition.data_type;
};

const propertyUnit = (definition: unknown) => {
  if (!isRecord(definition)) {
    return undefined;
  }

  if (typeof definition.unit === 'string' && definition.unit.trim()) {
    return definition.unit;
  }

  if (!isRecord(definition.spec) || typeof definition.spec.unit !== 'string') {
    return undefined;
  }

  return definition.spec.unit.trim() ? definition.spec.unit : undefined;
};

const isNumericPropertyDefinition = (definition: unknown) => {
  const dataType = propertyDataType(definition);
  return dataType === 'int' || dataType === 'float' || dataType === 'double';
};

const isWritablePropertyDefinition = (definition: unknown) => {
  if (!isRecord(definition) || typeof definition.access_mode !== 'string') {
    return false;
  }

  return (
    definition.access_mode === 'write' || definition.access_mode === 'readwrite'
  );
};

const writablePropertyEntriesOf = (model: ThingsModel | null) => {
  return Object.entries(model?.properties ?? {}).filter(([, definition]) =>
    isWritablePropertyDefinition(definition),
  );
};

const defaultPropertyValue = (definition: unknown) => {
  switch (propertyDataType(definition)) {
    case 'int':
    case 'float':
    case 'double':
      return 0;
    case 'bool':
      return false;
    case 'string':
      return '';
    default:
      return null;
  }
};

const defaultPropertySetText = (model: ThingsModel | null) => {
  const entries = writablePropertyEntriesOf(model);
  const first = entries[0];
  if (!first) {
    return '{}';
  }

  return JSON.stringify(
    {
      [first[0]]: defaultPropertyValue(first[1]),
    },
    null,
    2,
  );
};

const trendPropertyOptionsOf = (
  currentModel: ThingsModel | null,
  latest: DeviceLatestProperties | null,
) => {
  const modelProperties = Object.entries(currentModel?.properties ?? {})
    .filter(([, definition]) => isNumericPropertyDefinition(definition))
    .map(([key, definition]) => ({
      key,
      label: propertyDisplayName(key, definition),
      unit: propertyUnit(definition),
    }));

  if (modelProperties.length > 0) {
    return modelProperties;
  }

  return Object.entries(latest?.properties ?? {})
    .filter(([, value]) => typeof value === 'number')
    .map(([key]) => ({ key, label: key }));
};

const defaultTrendRange = () => {
  const to = new Date();
  const from = new Date(to.getTime() - 6 * 60 * 60 * 1000);
  return {
    from: dateTimeLocalValue(from),
    to: dateTimeLocalValue(to),
  };
};

const dateTimeLocalValue = (date: Date) => {
  const offsetMs = date.getTimezoneOffset() * 60 * 1000;
  return new Date(date.getTime() - offsetMs).toISOString().slice(0, 16);
};

const dateTimeLocalToISOString = (value: string) => {
  if (!value) {
    return '';
  }

  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return '';
  }

  return date.toISOString();
};

const formatTrendBucketLabel = (
  seconds: number,
  t: ReturnType<typeof useI18n>['t'],
) => {
  if (seconds >= 60 && seconds % 60 === 0) {
    return t('adminDevicePropertyTrendBucketMinutes').replace(
      '{minutes}',
      String(seconds / 60),
    );
  }

  return t('adminDevicePropertyTrendBucketSeconds').replace(
    '{seconds}',
    String(seconds),
  );
};

interface DevicePropertyCardProps {
  name: string;
  value: unknown;
  definition?: unknown;
  occurredAt?: string;
}

const DevicePropertyCard = ({
  name,
  value,
  definition,
  occurredAt,
}: DevicePropertyCardProps) => {
  const valueType = valueTypeOf(value);
  const displayType = propertyDataType(definition) ?? valueType;
  const displayName = propertyDisplayName(name, definition);
  const hasDisplayAlias = displayName !== name;
  const formattedValue = formatPropertyValue(value);
  const structuredValue = isStructuredPropertyValue(value);
  const unit = propertyUnit(definition);
  const occurredAtText = formatDateTime(occurredAt ?? '');
  const occurredAtFull = formatFullDateTime(occurredAt);

  return (
    <article className="grid min-h-32 gap-3 rounded-lg border border-linkflow-border bg-white p-4 shadow-[0_6px_18px_rgba(23,32,51,0.05)] transition hover:border-linkflow-primary hover:shadow-[0_10px_24px_rgba(18,137,153,0.1)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/20 dark:hover:border-linkflow-dark-primary">
      <div className="flex min-w-0 items-start justify-between gap-3">
        <div className="min-w-0">
          <h4
            className="m-0 truncate text-sm font-black text-linkflow-text dark:text-linkflow-dark-text"
            title={displayName}
          >
            {displayName}
          </h4>
          {hasDisplayAlias ? (
            <p
              className="m-0 mt-1 truncate font-mono text-xs font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle"
              title={name}
            >
              {name}
            </p>
          ) : null}
        </div>
        <span className={valueTypeBadgeClassName(displayType)}>
          {displayType}
        </span>
      </div>

      {structuredValue ? (
        <JsonPreview value={value} compact />
      ) : (
        <div className="flex min-h-12 min-w-0 items-end gap-2 rounded-md bg-slate-50 px-3 py-2 dark:bg-slate-950/30">
          <p
            className="m-0 min-w-0 flex-1 break-words text-2xl font-black leading-8 text-linkflow-text dark:text-linkflow-dark-text"
            title={formattedValue}
          >
            {formattedValue}
          </p>
          {unit ? (
            <span className="mb-1 shrink-0 text-xs font-black text-linkflow-muted dark:text-linkflow-dark-muted">
              {unit}
            </span>
          ) : null}
        </div>
      )}

      <time
        className="block truncate text-xs font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle"
        dateTime={occurredAt}
        title={occurredAtFull}
      >
        {occurredAtText}
      </time>
    </article>
  );
};

interface LatestPropertiesPanelProps {
  latest: DeviceLatestProperties | null;
  loading: boolean;
  currentModel: ThingsModel | null;
}

const LatestPropertiesPanel = ({
  latest,
  loading,
  currentModel,
}: LatestPropertiesPanelProps) => {
  const { t } = useI18n();
  const entries = Object.entries(latest?.properties ?? {});
  const propertyDefinitions = currentModel?.properties ?? {};

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

      <div className="grid gap-3 [grid-template-columns:repeat(auto-fit,minmax(190px,1fr))]">
        {entries.map(([key, value]) => (
          <DevicePropertyCard
            key={key}
            definition={propertyDefinitions[key]}
            name={key}
            occurredAt={latest.occurred_at}
            value={value}
          />
        ))}
      </div>
    </section>
  );
};

interface TrendChartProps {
  aggregate: DevicePropertyTrendAggregate;
  bucketSeconds: number;
  propertyOptions: DeviceTrendPropertyOption[];
  trend: DevicePropertyTrend | null;
}

const aggregateLabel = (
  aggregate: DevicePropertyTrendAggregate,
  t: ReturnType<typeof useI18n>['t'],
) => {
  switch (aggregate) {
    case 'min':
      return t('adminDevicePropertyTrendAggregateMin');
    case 'max':
      return t('adminDevicePropertyTrendAggregateMax');
    case 'last':
      return t('adminDevicePropertyTrendAggregateLast');
    default:
      return t('adminDevicePropertyTrendAggregateAvg');
  }
};

const formatAxisTime = (value: number | string) => {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return String(value);
  }

  return date.toLocaleString(undefined, {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  });
};

const formatTooltipTime = (value: string) => {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }

  return date.toLocaleString(undefined, {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
    second: '2-digit',
  });
};

const formatTrendNumber = (value: unknown) => {
  return typeof value === 'number' && Number.isFinite(value)
    ? value.toFixed(2)
    : '--';
};

const trendSeriesLabel = (
  property: string,
  propertyOptions: DeviceTrendPropertyOption[],
) => {
  const option = propertyOptions.find((item) => item.key === property);
  if (!option) {
    return property;
  }
  return option.unit ? `${option.label} (${option.unit})` : option.label;
};

const trendTooltipFormatter = (
  params: unknown,
  propertyOptions: DeviceTrendPropertyOption[],
) => {
  const items = Array.isArray(params) ? params : [params];
  const rows = items
    .filter(isRecord)
    .map((item) => {
      const data = Array.isArray(item.data) ? item.data : [];
      const property = typeof data[5] === 'string' ? data[5] : '';
      const label =
        typeof item.seriesName === 'string'
          ? item.seriesName
          : trendSeriesLabel(property, propertyOptions);
      const color =
        typeof item.color === 'string'
          ? item.color
          : 'var(--color-linkflow-primary)';
      return [
        `<div style="display:flex;align-items:center;gap:8px;margin-top:8px;">`,
        `<span style="width:9px;height:9px;border-radius:999px;background:${color};display:inline-block;"></span>`,
        `<strong>${label}</strong>`,
        `</div>`,
        `<div style="padding-left:17px;line-height:1.7;">`,
        `value: ${formatTrendNumber(data[1])}<br/>`,
        `min: ${formatTrendNumber(data[2])} / max: ${formatTrendNumber(
          data[3],
        )}<br/>`,
        `count: ${typeof data[4] === 'number' ? data[4] : '--'}`,
        `</div>`,
      ].join('');
    })
    .join('');

  const first = items.find(isRecord);
  const data = first && Array.isArray(first.data) ? first.data : [];
  const time = typeof data[6] === 'string' ? data[6] : '';

  return [
    `<div style="font-weight:800;margin-bottom:4px;">${formatTooltipTime(
      time,
    )}</div>`,
    rows,
  ].join('');
};

const TrendChart = ({
  aggregate,
  bucketSeconds,
  propertyOptions,
  trend,
}: TrendChartProps) => {
  const { t } = useI18n();
  const chartRef = useRef<HTMLDivElement | null>(null);
  const chartInstanceRef = useRef<ECharts | null>(null);
  const allPoints =
    trend?.series.flatMap((series) =>
      series.points.map((point) => ({
        ...point,
        property: series.property,
        time: new Date(point.bucket_at).getTime(),
      })),
    ) ?? [];
  const validPoints = allPoints.filter(
    (point) => Number.isFinite(point.time) && Number.isFinite(point.value),
  );

  useEffect(() => {
    if (!chartRef.current || validPoints.length === 0) {
      return undefined;
    }

    const chart =
      chartInstanceRef.current ??
      initECharts(chartRef.current, undefined, {
        renderer: 'canvas',
      });
    chartInstanceRef.current = chart;
    const option: TrendChartOption = {
      color: trendLineColors,
      animationDuration: 450,
      grid: {
        top: 56,
        right: 34,
        bottom: 84,
        left: 68,
        containLabel: true,
      },
      legend: {
        top: 14,
        left: 18,
        itemGap: 18,
        textStyle: {
          color: '#64748b',
          fontWeight: 700,
        },
        type: 'scroll',
      },
      tooltip: {
        trigger: 'axis',
        confine: true,
        axisPointer: {
          type: 'cross',
          label: {
            backgroundColor: '#128999',
          },
        },
        borderColor: '#d8e1ea',
        borderWidth: 1,
        padding: 12,
        textStyle: {
          color: '#172033',
          fontWeight: 600,
        },
        formatter: (params) => trendTooltipFormatter(params, propertyOptions),
      },
      xAxis: {
        type: 'time',
        name: `${aggregateLabel(aggregate, t)} · ${formatTrendBucketLabel(
          bucketSeconds,
          t,
        )}`,
        nameLocation: 'middle',
        nameGap: 42,
        axisLabel: {
          color: '#64748b',
          fontWeight: 700,
          formatter: (value: number | string) => formatAxisTime(value),
        },
        axisLine: {
          lineStyle: {
            color: '#cbd5e1',
          },
        },
        splitLine: {
          show: true,
          lineStyle: {
            color: '#e2e8f0',
            type: 'dashed',
          },
        },
      },
      yAxis: {
        type: 'value',
        name: t('adminDevicePropertyValue'),
        nameGap: 48,
        scale: true,
        splitNumber: 8,
        axisLabel: {
          color: '#64748b',
          fontWeight: 700,
          formatter: (value: number) => Number(value).toFixed(1),
        },
        axisLine: {
          show: true,
          lineStyle: {
            color: '#cbd5e1',
          },
        },
        splitLine: {
          lineStyle: {
            color: '#e2e8f0',
            type: 'dashed',
          },
        },
      },
      dataZoom: [
        {
          type: 'inside',
          throttle: 80,
        },
        {
          type: 'slider',
          height: 22,
          bottom: 22,
          borderColor: '#d8e1ea',
          fillerColor: 'rgba(18, 137, 153, 0.16)',
          handleStyle: {
            color: '#128999',
          },
          textStyle: {
            color: '#64748b',
            fontWeight: 700,
          },
        },
      ],
      series:
        trend?.series.map((series) => ({
          type: 'line',
          name: trendSeriesLabel(series.property, propertyOptions),
          smooth: true,
          showSymbol: true,
          symbol: 'circle',
          symbolSize: 6,
          lineStyle: {
            width: 3,
          },
          emphasis: {
            focus: 'series',
          },
          areaStyle: {
            opacity: 0.08,
          },
          data: series.points.map((point) => [
            new Date(point.bucket_at).getTime(),
            point.value,
            point.min,
            point.max,
            point.count,
            series.property,
            point.bucket_at,
          ]),
        })) ?? [],
    };

    chart.setOption(option, true);

    const resizeObserver = new ResizeObserver(() => chart.resize());
    resizeObserver.observe(chartRef.current);

    return () => {
      resizeObserver.disconnect();
    };
  }, [aggregate, bucketSeconds, propertyOptions, t, trend, validPoints.length]);

  useEffect(() => {
    return () => {
      chartInstanceRef.current?.dispose();
      chartInstanceRef.current = null;
    };
  }, []);

  if (validPoints.length === 0) {
    return (
      <div className="grid min-h-80 place-items-center rounded-lg border border-linkflow-border bg-slate-50 p-6 text-center dark:border-linkflow-dark-border dark:bg-slate-900/30">
        <div>
          <LineChartOutlined
            className="text-3xl text-linkflow-subtle dark:text-linkflow-dark-subtle"
            aria-hidden="true"
          />
          <p className="m-0 mt-3 text-sm font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDevicePropertyHistoryEmpty')}
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="rounded-lg border border-linkflow-border bg-white p-3 dark:border-linkflow-dark-border dark:bg-linkflow-dark-page">
      <div
        ref={chartRef}
        className="h-[420px] min-w-0"
        role="img"
        aria-label={t('adminDevicePropertyHistoryTitle')}
      />
    </div>
  );
};

interface HistoryPropertiesPanelProps {
  aggregate: DevicePropertyTrendAggregate;
  bucketSeconds: number;
  from: string;
  loading: boolean;
  onAggregateChange: (value: DevicePropertyTrendAggregate) => void;
  onBucketSecondsChange: (value: number) => void;
  onFromChange: (value: string) => void;
  onPropertyToggle: (property: string) => void;
  onReload: () => void;
  onToChange: (value: string) => void;
  propertyOptions: DeviceTrendPropertyOption[];
  selectedProperties: string[];
  to: string;
  trend: DevicePropertyTrend | null;
}

const HistoryPropertiesPanel = ({
  aggregate,
  bucketSeconds,
  from,
  loading,
  onAggregateChange,
  onBucketSecondsChange,
  onFromChange,
  onPropertyToggle,
  onReload,
  onToChange,
  propertyOptions,
  selectedProperties,
  to,
  trend,
}: HistoryPropertiesPanelProps) => {
  const { t } = useI18n();
  const pointCount =
    trend?.series.reduce((total, series) => total + series.points.length, 0) ??
    0;

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

      <div className="grid gap-3 rounded-lg border border-linkflow-border bg-slate-50 p-3 dark:border-linkflow-dark-border dark:bg-slate-950/30">
        <div className="grid gap-3 [grid-template-columns:repeat(auto-fit,minmax(180px,1fr))]">
          <label className="grid gap-2 text-sm font-bold">
            {t('adminDevicePropertyTrendFrom')}
            <input
              className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
              type="datetime-local"
              value={from}
              onChange={(event) => onFromChange(event.target.value)}
            />
          </label>
          <label className="grid gap-2 text-sm font-bold">
            {t('adminDevicePropertyTrendTo')}
            <input
              className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
              type="datetime-local"
              value={to}
              onChange={(event) => onToChange(event.target.value)}
            />
          </label>
          <label className="grid gap-2 text-sm font-bold">
            {t('adminDevicePropertyTrendBucket')}
            <select
              className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
              value={bucketSeconds}
              onChange={(event) =>
                onBucketSecondsChange(Number(event.target.value))
              }
            >
              {[60, 300, 900, 3600].map((seconds) => (
                <option key={seconds} value={seconds}>
                  {formatTrendBucketLabel(seconds, t)}
                </option>
              ))}
            </select>
          </label>
          <label className="grid gap-2 text-sm font-bold">
            {t('adminDevicePropertyTrendAggregate')}
            <select
              className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
              value={aggregate}
              onChange={(event) =>
                onAggregateChange(
                  event.target.value as DevicePropertyTrendAggregate,
                )
              }
            >
              <option value="avg">
                {t('adminDevicePropertyTrendAggregateAvg')}
              </option>
              <option value="min">
                {t('adminDevicePropertyTrendAggregateMin')}
              </option>
              <option value="max">
                {t('adminDevicePropertyTrendAggregateMax')}
              </option>
              <option value="last">
                {t('adminDevicePropertyTrendAggregateLast')}
              </option>
            </select>
          </label>
        </div>

        <div className="grid gap-2">
          <p className="m-0 text-sm font-bold">
            {t('adminDevicePropertyTrendProperties')}
          </p>
          {propertyOptions.length === 0 ? (
            <p className="m-0 text-sm font-semibold text-linkflow-subtle dark:text-linkflow-dark-subtle">
              {t('adminDevicePropertyTrendNoProperties')}
            </p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {propertyOptions.map((property, index) => {
                const selected = selectedProperties.includes(property.key);
                return (
                  <label
                    key={property.key}
                    className={[
                      'inline-flex min-h-10 cursor-pointer items-center gap-2 rounded-md border px-3 text-sm font-bold transition',
                      selected
                        ? 'border-linkflow-primary bg-white text-linkflow-primary shadow-sm dark:border-linkflow-dark-primary dark:bg-linkflow-dark-page dark:text-linkflow-dark-primary'
                        : 'border-linkflow-border bg-white/70 text-linkflow-muted hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page/70 dark:text-linkflow-dark-muted',
                    ].join(' ')}
                  >
                    <input
                      checked={selected}
                      className="h-4 w-4 accent-linkflow-primary"
                      type="checkbox"
                      onChange={() => onPropertyToggle(property.key)}
                    />
                    <span
                      className="h-2.5 w-2.5 rounded-full"
                      style={{
                        backgroundColor:
                          trendLineColors[index % trendLineColors.length],
                      }}
                    />
                    <span>{property.label}</span>
                    {property.unit ? (
                      <span className="text-xs text-linkflow-subtle dark:text-linkflow-dark-subtle">
                        {property.unit}
                      </span>
                    ) : null}
                  </label>
                );
              })}
            </div>
          )}
        </div>

        <div className="flex flex-wrap items-center justify-between gap-3">
          <p className="m-0 text-xs font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDevicePropertyTrendPoints').replace(
              '{count}',
              String(pointCount),
            )}
          </p>
          <button
            type="button"
            className="inline-flex h-11 items-center gap-2 rounded-md bg-linkflow-primary px-4 text-sm font-bold text-white shadow-sm transition hover:bg-linkflow-primary-hover disabled:cursor-not-allowed disabled:opacity-60 dark:bg-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-hover"
            disabled={
              loading ||
              selectedProperties.length === 0 ||
              propertyOptions.length === 0
            }
            onClick={onReload}
          >
            <ReloadOutlined aria-hidden="true" />
            {loading
              ? t('adminDevicePropertyTrendLoading')
              : t('adminDevicePropertyTrendReload')}
          </button>
        </div>
      </div>

      <TrendChart
        aggregate={aggregate}
        bucketSeconds={bucketSeconds}
        propertyOptions={propertyOptions}
        trend={trend}
      />
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
      className: 'min-w-56',
      render: (entry) => <JsonPreview value={entry.params} compact />,
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
      <JsonPreview value={entry.params} compact />
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
  inputParams: ServiceInputParamDraft[];
  loadingServices: boolean;
  result: DeviceServiceCallResult | null;
  serviceName: string;
  onAddInputParam: (name?: string, definition?: unknown) => void;
  onCall: () => void;
  onInputParamChange: (
    id: string,
    patch: Partial<Omit<ServiceInputParamDraft, 'id'>>,
  ) => void;
  onRemoveInputParam: (id: string) => void;
  onServiceNameChange: (serviceName: string) => void;
}

interface DevicePropertySetPanelProps {
  currentModel: ThingsModel | null;
  device: Device | undefined;
  loadingModel: boolean;
  propertiesText: string;
  result: DevicePropertySetResult | null;
  setting: boolean;
  onPropertiesTextChange: (text: string) => void;
  onSetProperties: () => void;
  onUseProperty: (propertyName: string) => void;
}

const DevicePropertySetPanel = ({
  currentModel,
  device,
  loadingModel,
  propertiesText,
  result,
  setting,
  onPropertiesTextChange,
  onSetProperties,
  onUseProperty,
}: DevicePropertySetPanelProps) => {
  const { t } = useI18n();
  const writableProperties = writablePropertyEntriesOf(currentModel);
  const canSet =
    Boolean(device) &&
    device?.status === 'active' &&
    device?.connection_status === 'online' &&
    writableProperties.length > 0 &&
    !setting;

  return (
    <section className="grid gap-3">
      <div className="grid gap-3 rounded-lg border border-linkflow-border bg-white p-4 shadow-[0_8px_24px_rgba(23,32,51,0.06)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/20">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <p className="m-0 text-sm font-black text-linkflow-text dark:text-linkflow-dark-text">
              {t('adminDevicePropertySetTitle')}
            </p>
            <p className="m-0 mt-1 text-xs font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
              {writableProperties.length > 0
                ? t('adminDevicePropertyWritableCount').replace(
                    '{count}',
                    String(writableProperties.length),
                  )
                : currentModel
                ? t('adminDevicePropertySetNoWritable')
                : t('adminDeviceServiceNoCurrentModel')}
            </p>
          </div>
          <button
            type="button"
            className="inline-flex h-11 items-center gap-2 rounded-md border border-linkflow-primary bg-linkflow-primary px-4 text-sm font-black text-white shadow-sm transition hover:bg-linkflow-primary-hover disabled:cursor-not-allowed disabled:opacity-60 dark:border-linkflow-dark-primary dark:bg-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-hover"
            onClick={onSetProperties}
            disabled={!canSet}
          >
            <SendOutlined aria-hidden="true" />
            {setting
              ? t('adminDevicePropertySetting')
              : t('adminDevicePropertySetSend')}
          </button>
        </div>

        <div className="grid gap-2">
          <p className="m-0 text-xs font-bold uppercase text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDevicePropertyWritableProperties')}
          </p>
          {writableProperties.length === 0 ? (
            <p className="m-0 rounded-md border border-dashed border-linkflow-border bg-slate-50 px-3 py-2 text-sm font-bold text-linkflow-subtle dark:border-linkflow-dark-border dark:bg-slate-950/30 dark:text-linkflow-dark-subtle">
              {t('adminDevicePropertySetNoWritable')}
            </p>
          ) : (
            <div className="flex gap-2 overflow-x-auto pb-1">
              {writableProperties.map(([name, definition]) => (
                <button
                  key={name}
                  type="button"
                  className="grid min-h-11 min-w-44 shrink-0 gap-1 rounded-md border border-linkflow-border bg-white px-3 py-2 text-left transition hover:border-linkflow-primary hover:bg-linkflow-primary-soft dark:border-linkflow-dark-border dark:bg-linkflow-dark-page dark:hover:border-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-soft"
                  onClick={() => onUseProperty(name)}
                >
                  <span className="truncate text-sm font-black text-linkflow-text dark:text-linkflow-dark-text">
                    {propertyDisplayName(name, definition)}
                  </span>
                  <span className="truncate font-mono text-xs font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
                    {name} / {propertyDataType(definition) ?? 'unknown'}
                  </span>
                </button>
              ))}
            </div>
          )}
        </div>

        <label className="grid gap-2 text-sm font-bold">
          {t('adminDevicePropertySetPayload')}
          <textarea
            className="min-h-44 resize-y rounded-md border border-linkflow-border bg-white p-3 font-mono text-sm font-semibold leading-6 outline-none transition focus:border-linkflow-primary disabled:opacity-70 dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
            value={propertiesText}
            onChange={(event) => onPropertiesTextChange(event.target.value)}
            disabled={loadingModel || writableProperties.length === 0}
            spellCheck={false}
          />
        </label>
      </div>

      {result ? (
        <article className="grid gap-4 rounded-lg border border-linkflow-border bg-white p-4 shadow-[0_8px_24px_rgba(23,32,51,0.06)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/20">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <p className="m-0 text-xs font-bold uppercase text-linkflow-subtle dark:text-linkflow-dark-subtle">
                {t('adminDevicePropertySetResult')}
              </p>
              <div className="mt-2">
                <ShortID value={result.command_id} />
              </div>
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
          <JsonPreview value={result.properties} compact />
        </article>
      ) : null}
    </section>
  );
};

const DeviceServiceCallPanel = ({
  calling,
  currentModel,
  device,
  inputParams,
  loadingServices,
  result,
  serviceName,
  onAddInputParam,
  onCall,
  onInputParamChange,
  onRemoveInputParam,
  onServiceNameChange,
}: DeviceServiceCallPanelProps) => {
  const { t } = useI18n();
  const services = serviceEntriesOf(currentModel);
  const selectedDefinition = serviceName
    ? currentModel?.services[serviceName]
    : undefined;
  const inputDefinition = serviceInputDefinition(selectedDefinition);
  const inputDefinitionEntries = Object.entries(inputDefinition);
  const generatedInput = (() => {
    try {
      return serviceInputParamsToRecord(inputParams);
    } catch {
      return {};
    }
  })();
  const canCall =
    Boolean(device) &&
    device?.status === 'active' &&
    device?.connection_status === 'online' &&
    Boolean(serviceName) &&
    !calling;

  return (
    <section className="grid gap-3">
      <div className="grid gap-3 rounded-lg border border-linkflow-border bg-white p-4 shadow-[0_8px_24px_rgba(23,32,51,0.06)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/20">
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

        <div className="grid gap-2">
          <p className="m-0 text-xs font-bold uppercase text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDeviceServiceInputDefinition')}
          </p>
          {inputDefinitionEntries.length === 0 ? (
            <p className="m-0 rounded-md border border-dashed border-linkflow-border bg-slate-50 px-3 py-2 text-sm font-bold text-linkflow-subtle dark:border-linkflow-dark-border dark:bg-slate-950/30 dark:text-linkflow-dark-subtle">
              {t('adminDeviceServiceNoInputDefinition')}
            </p>
          ) : (
            <div className="flex gap-2 overflow-x-auto pb-1">
              {inputDefinitionEntries.map(([name, definition]) => (
                <button
                  key={name}
                  type="button"
                  className="grid min-h-11 min-w-44 shrink-0 gap-1 rounded-md border border-linkflow-border bg-white px-3 py-2 text-left transition hover:border-linkflow-primary hover:bg-linkflow-primary-soft dark:border-linkflow-dark-border dark:bg-linkflow-dark-page dark:hover:border-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-soft"
                  onClick={() => onAddInputParam(name, definition)}
                >
                  <span className="truncate text-sm font-black text-linkflow-text dark:text-linkflow-dark-text">
                    {propertyDisplayName(name, definition)}
                  </span>
                  <span className="truncate font-mono text-xs font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
                    {name} / {propertyDataType(definition) ?? 'json'}
                  </span>
                </button>
              ))}
            </div>
          )}
        </div>

        <div className="grid gap-3 lg:grid-cols-[minmax(0,1fr)_minmax(18rem,0.72fr)]">
          <div className="grid content-start gap-3">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <p className="m-0 text-sm font-black text-linkflow-text dark:text-linkflow-dark-text">
                {t('adminDeviceServiceInputParams')}
              </p>
              <button
                type="button"
                className="inline-flex h-10 items-center gap-2 rounded-md border border-linkflow-border bg-white px-3 text-sm font-bold text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
                onClick={() => onAddInputParam()}
              >
                <PlusOutlined aria-hidden="true" />
                {t('adminDeviceServiceAddInputParam')}
              </button>
            </div>

            {inputParams.length === 0 ? (
              <div className="rounded-lg border border-dashed border-linkflow-border bg-slate-50 p-4 text-sm font-bold text-linkflow-subtle dark:border-linkflow-dark-border dark:bg-slate-950/30 dark:text-linkflow-dark-subtle">
                {t('adminDeviceServiceInputParamsEmpty')}
              </div>
            ) : (
              <div className="grid gap-3">
                {inputParams.map((param) => (
                  <div
                    key={param.id}
                    className="grid gap-3 rounded-lg border border-linkflow-border bg-white p-3 dark:border-linkflow-dark-border dark:bg-linkflow-dark-page lg:grid-cols-[minmax(9rem,0.8fr)_minmax(8rem,0.5fr)_minmax(12rem,1fr)_2.75rem]"
                  >
                    <label className="grid gap-2 text-sm font-bold">
                      {t('adminDeviceServiceParamName')}
                      <input
                        className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel"
                        value={param.name}
                        onChange={(event) =>
                          onInputParamChange(param.id, {
                            name: event.target.value,
                          })
                        }
                      />
                    </label>
                    <label className="grid gap-2 text-sm font-bold">
                      {t('adminDeviceServiceParamType')}
                      <select
                        className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel"
                        value={param.valueType}
                        onChange={(event) => {
                          const nextType = event.target
                            .value as ServiceInputValueType;
                          onInputParamChange(param.id, {
                            valueType: nextType,
                            value: defaultServiceInputValue(
                              undefined,
                              nextType,
                            ),
                          });
                        }}
                      >
                        {(
                          [
                            'string',
                            'int',
                            'float',
                            'double',
                            'bool',
                            'json',
                          ] satisfies ServiceInputValueType[]
                        ).map((type) => (
                          <option key={type} value={type}>
                            {type}
                          </option>
                        ))}
                      </select>
                    </label>
                    {param.valueType === 'bool' ? (
                      <label className="grid gap-2 text-sm font-bold">
                        {t('adminDeviceServiceParamValue')}
                        <select
                          className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel"
                          value={param.value}
                          onChange={(event) =>
                            onInputParamChange(param.id, {
                              value: event.target.value,
                            })
                          }
                        >
                          <option value="false">false</option>
                          <option value="true">true</option>
                        </select>
                      </label>
                    ) : (
                      <label className="grid gap-2 text-sm font-bold">
                        {t('adminDeviceServiceParamValue')}
                        <input
                          className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel"
                          value={param.value}
                          onChange={(event) =>
                            onInputParamChange(param.id, {
                              value: event.target.value,
                            })
                          }
                        />
                      </label>
                    )}
                    <button
                      type="button"
                      className="grid h-11 w-11 place-items-center self-end rounded-md border border-linkflow-border bg-white text-linkflow-muted transition hover:border-rose-400 hover:text-rose-600 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted"
                      aria-label={t('adminDeviceServiceRemoveInputParam')}
                      onClick={() => onRemoveInputParam(param.id)}
                    >
                      <CloseOutlined aria-hidden="true" />
                    </button>
                  </div>
                ))}
              </div>
            )}
          </div>
          <div className="rounded-lg border border-linkflow-border bg-slate-50/80 p-3 dark:border-linkflow-dark-border dark:bg-slate-950/30">
            <p className="m-0 mb-2 text-xs font-bold uppercase text-linkflow-subtle dark:text-linkflow-dark-subtle">
              {t('adminDeviceServiceGeneratedInput')}
            </p>
            <JsonPreview value={generatedInput} maxHeightClassName="max-h-64" />
          </div>
        </div>
      </div>

      {result ? (
        <article className="grid gap-4 rounded-lg border border-linkflow-border bg-white p-4 shadow-[0_8px_24px_rgba(23,32,51,0.06)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/20">
          <div className="flex flex-wrap items-start justify-between gap-3">
            <div>
              <p className="m-0 text-xs font-bold uppercase text-linkflow-subtle dark:text-linkflow-dark-subtle">
                {t('adminDeviceServiceResult')}
              </p>
              <div className="mt-2">
                <ShortID value={result.command_id} />
              </div>
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
          <div className="mt-2">
            <ShortID value={entry.command_id} />
          </div>
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
      className: 'min-w-56',
      render: (entry) => <JsonPreview value={entry.input} compact />,
    },
    {
      key: 'ack_output',
      title: t('adminDeviceServiceAckOutput'),
      className: 'min-w-56',
      render: (entry) => <JsonPreview value={entry.ack_output ?? {}} compact />,
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
          <div className="mt-2">
            <ShortID value={entry.command_id} />
          </div>
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
        <JsonPreview value={entry.input} compact />
        <JsonPreview value={entry.ack_output ?? {}} compact />
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

interface DevicePropertySetHistoryPanelProps {
  ackDeadlineSeconds: number;
  loading: boolean;
  page: number;
  pageSize: number;
  propertyName: string;
  sets: DevicePropertySetHistoryEntry[];
  total: number;
  onAckDeadlineSecondsChange: (seconds: number) => void;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  onPropertyNameChange: (propertyName: string) => void;
  onReload: () => void;
}

const DevicePropertySetHistoryPanel = ({
  ackDeadlineSeconds,
  loading,
  page,
  pageSize,
  propertyName,
  sets,
  total,
  onAckDeadlineSecondsChange,
  onPageChange,
  onPageSizeChange,
  onPropertyNameChange,
  onReload,
}: DevicePropertySetHistoryPanelProps) => {
  const { t } = useI18n();
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const columns: AdminDataTableColumn<DevicePropertySetHistoryEntry>[] = [
    {
      key: 'command_id',
      title: t('adminDevicePropertySetCommand'),
      render: (entry) => (
        <div className="min-w-0">
          <ShortID value={entry.command_id} />
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
      key: 'properties',
      title: t('adminDevicePropertySetPayload'),
      className: 'min-w-56',
      render: (entry) => <JsonPreview value={entry.properties} compact />,
    },
    {
      key: 'ack_properties',
      title: t('adminDevicePropertySetAckProperties'),
      className: 'min-w-56',
      render: (entry) => (
        <JsonPreview value={entry.ack_properties ?? {}} compact />
      ),
    },
  ];

  const mobileItems = sets.map((entry) => (
    <article
      key={entry.command_id}
      className="grid gap-3 rounded-lg border border-linkflow-border bg-white p-4 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel"
    >
      <div className="flex flex-wrap items-start justify-between gap-3">
        <ShortID value={entry.command_id} />
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
        <JsonPreview value={entry.properties} compact />
        <JsonPreview value={entry.ack_properties ?? {}} compact />
      </div>
    </article>
  ));

  return (
    <section className="grid gap-4">
      <div className="flex flex-wrap items-end justify-between gap-3 rounded-lg border border-linkflow-border bg-white p-4 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel">
        <div className="flex flex-wrap items-end gap-3">
          <label className="grid min-w-56 gap-2 text-sm font-bold">
            {t('adminDevicePropertySetPropertyName')}
            <input
              className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
              value={propertyName}
              onChange={(event) => onPropertyNameChange(event.target.value)}
              placeholder={t('adminDevicePropertySetAllProperties')}
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
          {t('adminDevicePropertySetHistoryReload')}
        </button>
      </div>

      <AdminDataTable
        columns={columns}
        emptyDescription={t('adminDevicePropertySetHistoryEmpty')}
        emptyTitle={t('adminDevicePropertySetHistoryEmpty')}
        getRowKey={(entry) => entry.command_id}
        items={sets}
        minWidthClassName="min-w-[1120px]"
        mobileItems={mobileItems}
        pagination={{
          jumpLabel: t('adminDeviceJump'),
          jumpToLabel: t('adminDeviceJumpTo'),
          loading,
          loadingLabel: t('adminDevicePropertySetHistoryLoading'),
          nextLabel: t('adminDeviceNext'),
          onPageChange,
          onPageSizeChange,
          page,
          pageLabel: t('adminDevicePropertySetHistoryPage')
            .replace('{page}', String(page))
            .replace('{totalPages}', String(totalPages)),
          pageSize,
          pageSizeLabel: t('adminDevicePageSize'),
          pageSizeOptions: [10, 20, 50, 100],
          prevLabel: t('adminDevicePrev'),
          total,
          totalLabel: t('adminDevicePropertySetHistoryTotal').replace(
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
  const [trend, setTrend] = useState<DevicePropertyTrend | null>(null);
  const [loadingTrend, setLoadingTrend] = useState(false);
  const [trendProperties, setTrendProperties] = useState<string[]>([]);
  const [trendRange, setTrendRange] = useState(defaultTrendRange);
  const [trendBucketSeconds, setTrendBucketSeconds] = useState(300);
  const [trendAggregate, setTrendAggregate] =
    useState<DevicePropertyTrendAggregate>('avg');
  const [eventEntries, setEventEntries] = useState<DeviceEventEntry[]>([]);
  const [eventTotal, setEventTotal] = useState(0);
  const [eventPage, setEventPage] = useState(1);
  const [eventPageSize, setEventPageSize] = useState(20);
  const [eventNameFilter, setEventNameFilter] = useState('');
  const [loadingEvents, setLoadingEvents] = useState(false);
  const [propertySetHistoryEntries, setPropertySetHistoryEntries] = useState<
    DevicePropertySetHistoryEntry[]
  >([]);
  const [propertySetHistoryTotal, setPropertySetHistoryTotal] = useState(0);
  const [propertySetHistoryPage, setPropertySetHistoryPage] = useState(1);
  const [propertySetHistoryPageSize, setPropertySetHistoryPageSize] =
    useState(20);
  const [propertySetHistoryNameFilter, setPropertySetHistoryNameFilter] =
    useState('');
  const [propertySetAckDeadlineSeconds, setPropertySetAckDeadlineSeconds] =
    useState(90);
  const [loadingPropertySetHistory, setLoadingPropertySetHistory] =
    useState(false);
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
  const [serviceInputParams, setServiceInputParams] = useState<
    ServiceInputParamDraft[]
  >([]);
  const [callingService, setCallingService] = useState(false);
  const [serviceCallResult, setServiceCallResult] =
    useState<DeviceServiceCallResult | null>(null);
  const [propertySetText, setPropertySetText] = useState('{}');
  const [settingProperties, setSettingProperties] = useState(false);
  const [propertySetResult, setPropertySetResult] =
    useState<DevicePropertySetResult | null>(null);
  const latestInboxEvent = useEventInboxStore((state) => state.items[0]);
  const selectedDevice = devices.find((device) => device.id === deviceId);
  const selectedProductId = selectedDevice?.product_id || productFilter;
  const trendPropertyOptions = trendPropertyOptionsOf(currentModel, latest);

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
        setServiceInputParams([]);
        setServiceCallResult(null);
        setPropertySetText('{}');
        setPropertySetResult(null);
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
        setPropertySetText(defaultPropertySetText(nextModel));
        setPropertySetResult(null);
        setServiceName(nextServiceName);
        setServiceInputParams(
          nextServiceName
            ? defaultServiceInputParams(nextModel?.services[nextServiceName])
            : [],
        );
        setServiceCallResult(null);
      } catch (error) {
        if (cancelled) {
          return;
        }
        setCurrentModel(null);
        setServiceName('');
        setServiceInputParams([]);
        setServiceCallResult(null);
        setPropertySetText('{}');
        setPropertySetResult(null);
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

  const loadPropertyTrend = async (
    nextDeviceId = deviceId,
    options: {
      properties?: string[];
      from?: string;
      to?: string;
      bucketSeconds?: number;
      aggregate?: DevicePropertyTrendAggregate;
      silent?: boolean;
      isCancelled?: () => boolean;
    } = {},
  ) => {
    const nextProperties = options.properties ?? trendProperties;
    if (!nextDeviceId || nextProperties.length === 0) {
      setTrend(null);
      return;
    }

    const fromISO = dateTimeLocalToISOString(options.from ?? trendRange.from);
    const toISO = dateTimeLocalToISOString(options.to ?? trendRange.to);
    if (!fromISO || !toISO) {
      setTrend(null);
      return;
    }

    if (!options.silent) {
      setLoadingTrend(true);
    }
    try {
      const result = await getDevicePropertyTrend(nextDeviceId, {
        properties: nextProperties.join(','),
        from: fromISO,
        to: toISO,
        bucket_seconds: options.bucketSeconds ?? trendBucketSeconds,
        agg: options.aggregate ?? trendAggregate,
      });
      if (options.isCancelled?.()) {
        return;
      }
      setTrend(result);
    } catch (error) {
      if (options.isCancelled?.() || options.silent) {
        return;
      }
      notification.error({
        message: t('adminDevicePropertyTrendFailed'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    } finally {
      if (!options.silent) {
        setLoadingTrend(false);
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
          (options.serviceName ?? serviceHistoryNameFilter).trim() || undefined,
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

  const loadPropertySetHistory = async (
    nextDeviceId = deviceId,
    options: {
      propertyName?: string;
      ackDeadlineSeconds?: number;
      page?: number;
      pageSize?: number;
      silent?: boolean;
      isCancelled?: () => boolean;
    } = {},
  ) => {
    if (!nextDeviceId) {
      setPropertySetHistoryEntries([]);
      setPropertySetHistoryTotal(0);
      return;
    }

    if (!options.silent) {
      setLoadingPropertySetHistory(true);
    }
    try {
      const result = await listDevicePropertySets(nextDeviceId, {
        property_name:
          (options.propertyName ?? propertySetHistoryNameFilter).trim() ||
          undefined,
        ack_deadline_seconds:
          options.ackDeadlineSeconds ?? propertySetAckDeadlineSeconds,
        page: options.page ?? propertySetHistoryPage,
        page_size: options.pageSize ?? propertySetHistoryPageSize,
      });
      if (options.isCancelled?.()) {
        return;
      }
      setPropertySetHistoryEntries(result.items);
      setPropertySetHistoryTotal(result.total);
      setPropertySetHistoryPage(result.page);
      setPropertySetHistoryPageSize(result.page_size);
    } catch (error) {
      if (options.isCancelled?.() || options.silent) {
        return;
      }
      notification.error({
        message: t('adminDevicePropertySetHistoryFailed'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    } finally {
      if (!options.silent) {
        setLoadingPropertySetHistory(false);
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
    const available = new Set(trendPropertyOptions.map((option) => option.key));
    setTrendProperties((current) => {
      const kept = current.filter((property) => available.has(property));
      if (
        kept.length === current.length &&
        kept.every((property, index) => property === current[index])
      ) {
        return current;
      }
      if (kept.length > 0) {
        return kept;
      }
      return trendPropertyOptions.slice(0, 2).map((option) => option.key);
    });
  }, [currentModel, latest]);

  useEffect(() => {
    let cancelled = false;

    const run = async () => {
      if (activeTab !== 'history') {
        return;
      }
      if (!deviceId || trendProperties.length === 0) {
        setTrend(null);
        return;
      }

      setLoadingTrend(true);
      try {
        const fromISO = dateTimeLocalToISOString(trendRange.from);
        const toISO = dateTimeLocalToISOString(trendRange.to);
        if (!fromISO || !toISO) {
          setTrend(null);
          return;
        }
        const result = await getDevicePropertyTrend(deviceId, {
          properties: trendProperties.join(','),
          from: fromISO,
          to: toISO,
          bucket_seconds: trendBucketSeconds,
          agg: trendAggregate,
        });
        if (cancelled) {
          return;
        }
        setTrend(result);
      } catch (error) {
        if (cancelled) {
          return;
        }
        notification.error({
          message: t('adminDevicePropertyTrendFailed'),
          description: error instanceof Error ? error.message : undefined,
          placement: 'topRight',
        });
      } finally {
        if (!cancelled) {
          setLoadingTrend(false);
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
    t,
    trendAggregate,
    trendBucketSeconds,
    trendProperties,
    trendRange.from,
    trendRange.to,
  ]);

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
      if (activeTab !== 'propertySetHistory') {
        return;
      }
      if (!deviceId) {
        setPropertySetHistoryEntries([]);
        setPropertySetHistoryTotal(0);
        return;
      }

      setLoadingPropertySetHistory(true);
      try {
        const result = await listDevicePropertySets(deviceId, {
          property_name: propertySetHistoryNameFilter.trim() || undefined,
          ack_deadline_seconds: propertySetAckDeadlineSeconds,
          page: propertySetHistoryPage,
          page_size: propertySetHistoryPageSize,
        });
        if (cancelled) {
          return;
        }
        setPropertySetHistoryEntries(result.items);
        setPropertySetHistoryTotal(result.total);
        setPropertySetHistoryPage(result.page);
        setPropertySetHistoryPageSize(result.page_size);
      } catch (error) {
        if (cancelled) {
          return;
        }
        notification.error({
          message: t('adminDevicePropertySetHistoryFailed'),
          description: error instanceof Error ? error.message : undefined,
          placement: 'topRight',
        });
      } finally {
        if (!cancelled) {
          setLoadingPropertySetHistory(false);
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
    propertySetAckDeadlineSeconds,
    propertySetHistoryNameFilter,
    propertySetHistoryPage,
    propertySetHistoryPageSize,
    t,
  ]);

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
        if (activeTab === 'history') {
          void loadPropertyTrend(deviceId, { silent: true, isCancelled });
        }
        if (activeTab === 'eventHistory') {
          void loadEventHistory(deviceId, { silent: true, isCancelled });
        }
        if (activeTab === 'propertySetHistory') {
          void loadPropertySetHistory(deviceId, { silent: true, isCancelled });
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
    propertySetAckDeadlineSeconds,
    propertySetHistoryNameFilter,
    propertySetHistoryPage,
    propertySetHistoryPageSize,
    realtimeStatus,
    serviceAckDeadlineSeconds,
    serviceHistoryNameFilter,
    serviceHistoryPage,
    serviceHistoryPageSize,
    tenantId,
    trendAggregate,
    trendBucketSeconds,
    trendProperties,
    trendRange.from,
    trendRange.to,
  ]);

  const handleTenantChange = (nextTenantId: string) => {
    setTenantId(nextTenantId);
    setProductFilter('');
    setDeviceId('');
    setLatest(null);
    setTrend(null);
    setTrendProperties([]);
    setEventEntries([]);
    setEventTotal(0);
    setEventPage(1);
    setPropertySetHistoryEntries([]);
    setPropertySetHistoryTotal(0);
    setPropertySetHistoryPage(1);
    setServiceHistoryEntries([]);
    setServiceHistoryTotal(0);
    setServiceHistoryPage(1);
    setServiceCallResult(null);
    setPropertySetResult(null);
  };

  const handleProductFilterChange = (nextProductId: string) => {
    setProductFilter(nextProductId);
    setDeviceId('');
    setLatest(null);
    setTrend(null);
    setTrendProperties([]);
    setEventEntries([]);
    setEventTotal(0);
    setEventPage(1);
    setPropertySetHistoryEntries([]);
    setPropertySetHistoryTotal(0);
    setPropertySetHistoryPage(1);
    setServiceHistoryEntries([]);
    setServiceHistoryTotal(0);
    setServiceHistoryPage(1);
    setServiceCallResult(null);
    setPropertySetResult(null);
  };

  const handleDeviceChange = (nextDeviceId: string) => {
    setDeviceId(nextDeviceId);
    setLatest(null);
    setTrend(null);
    setTrendProperties([]);
    setEventEntries([]);
    setEventTotal(0);
    setEventPage(1);
    setPropertySetHistoryEntries([]);
    setPropertySetHistoryTotal(0);
    setPropertySetHistoryPage(1);
    setServiceHistoryEntries([]);
    setServiceHistoryTotal(0);
    setServiceHistoryPage(1);
    setServiceCallResult(null);
    setPropertySetResult(null);
  };

  const handleEventNameChange = (nextEventName: string) => {
    setEventNameFilter(nextEventName);
    setEventPage(1);
  };

  const handleEventPageSizeChange = (nextPageSize: number) => {
    setEventPageSize(nextPageSize);
    setEventPage(1);
  };

  const handlePropertySetHistoryNameChange = (nextPropertyName: string) => {
    setPropertySetHistoryNameFilter(nextPropertyName);
    setPropertySetHistoryPage(1);
  };

  const handlePropertySetHistoryPageSizeChange = (nextPageSize: number) => {
    setPropertySetHistoryPageSize(nextPageSize);
    setPropertySetHistoryPage(1);
  };

  const handlePropertySetAckDeadlineSecondsChange = (nextSeconds: number) => {
    setPropertySetAckDeadlineSeconds(nextSeconds);
    setPropertySetHistoryPage(1);
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

  const handleTrendPropertyToggle = (property: string) => {
    setTrendProperties((current) => {
      if (current.includes(property)) {
        return current.filter((item) => item !== property);
      }
      if (current.length >= 8) {
        return current;
      }
      return [...current, property];
    });
  };

  const handleTrendFromChange = (nextFrom: string) => {
    setTrendRange((current) => ({ ...current, from: nextFrom }));
  };

  const handleTrendToChange = (nextTo: string) => {
    setTrendRange((current) => ({ ...current, to: nextTo }));
  };

  const handleServiceNameChange = (nextServiceName: string) => {
    setServiceName(nextServiceName);
    setServiceInputParams(
      nextServiceName
        ? defaultServiceInputParams(currentModel?.services[nextServiceName])
        : [],
    );
    setServiceCallResult(null);
  };

  const handleAddServiceInputParam = (name = '', definition?: unknown) => {
    const valueType = serviceInputValueType(definition);
    setServiceInputParams((current) => [
      ...current,
      {
        id: createDraftID(),
        name,
        value: defaultServiceInputValue(definition, valueType),
        valueType,
      },
    ]);
  };

  const handleServiceInputParamChange = (
    id: string,
    patch: Partial<Omit<ServiceInputParamDraft, 'id'>>,
  ) => {
    setServiceInputParams((current) =>
      current.map((param) =>
        param.id === id
          ? {
              ...param,
              ...patch,
            }
          : param,
      ),
    );
  };

  const handleRemoveServiceInputParam = (id: string) => {
    setServiceInputParams((current) =>
      current.filter((param) => param.id !== id),
    );
  };

  const handleUseWritableProperty = (propertyName: string) => {
    const definition = currentModel?.properties[propertyName];
    setPropertySetText(
      JSON.stringify(
        {
          [propertyName]: defaultPropertyValue(definition),
        },
        null,
        2,
      ),
    );
    setPropertySetResult(null);
  };

  const handleSetProperties = async () => {
    if (!deviceId) {
      return;
    }

    let properties: Record<string, unknown>;
    try {
      const parsed = JSON.parse(propertySetText.trim() || '{}') as unknown;
      if (!isRecord(parsed)) {
        throw new Error(t('adminDevicePropertySetPayloadMustObject'));
      }
      properties = parsed;
    } catch (error) {
      notification.error({
        message: t('adminDevicePropertySetInvalidPayload'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
      return;
    }

    setSettingProperties(true);
    try {
      const result = await setDeviceProperties(deviceId, { properties });
      setPropertySetResult(result);
      notification.success({
        message: t('adminDevicePropertySetSuccess'),
        description: result.command_id,
        placement: 'topRight',
      });
      void loadPropertySetHistory(deviceId, {
        page: 1,
        silent: true,
      });
    } catch (error) {
      notification.error({
        message: t('adminDevicePropertySetFailed'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    } finally {
      setSettingProperties(false);
    }
  };

  const handleCallService = async () => {
    if (!deviceId || !serviceName) {
      return;
    }

    let input: Record<string, unknown>;
    try {
      input = serviceInputParamsToRecord(serviceInputParams);
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

      <DeviceContextStatusBar
        currentModel={currentModel}
        device={selectedDevice}
      />

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
          aria-selected={activeTab === 'set'}
          className={tabButtonClassName(activeTab === 'set')}
          onClick={() => setActiveTab('set')}
        >
          <SendOutlined aria-hidden="true" />
          {t('adminDevicePropertySetTab')}
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
          aria-selected={activeTab === 'propertySetHistory'}
          className={tabButtonClassName(activeTab === 'propertySetHistory')}
          onClick={() => setActiveTab('propertySetHistory')}
        >
          <FieldTimeOutlined aria-hidden="true" />
          {t('adminDevicePropertySetHistoryTab')}
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
        <LatestPropertiesPanel
          currentModel={currentModel}
          latest={latest}
          loading={loadingLatest}
        />
      ) : activeTab === 'history' ? (
        <HistoryPropertiesPanel
          aggregate={trendAggregate}
          bucketSeconds={trendBucketSeconds}
          from={trendRange.from}
          loading={loadingTrend}
          onAggregateChange={setTrendAggregate}
          onBucketSecondsChange={setTrendBucketSeconds}
          onFromChange={handleTrendFromChange}
          onPropertyToggle={handleTrendPropertyToggle}
          onReload={() => loadPropertyTrend()}
          onToChange={handleTrendToChange}
          propertyOptions={trendPropertyOptions}
          selectedProperties={trendProperties}
          to={trendRange.to}
          trend={trend}
        />
      ) : activeTab === 'set' ? (
        <DevicePropertySetPanel
          currentModel={currentModel}
          device={selectedDevice}
          loadingModel={loadingServices}
          onPropertiesTextChange={setPropertySetText}
          onSetProperties={handleSetProperties}
          onUseProperty={handleUseWritableProperty}
          propertiesText={propertySetText}
          result={propertySetResult}
          setting={settingProperties}
        />
      ) : activeTab === 'services' ? (
        <DeviceServiceCallPanel
          calling={callingService}
          currentModel={currentModel}
          device={selectedDevice}
          inputParams={serviceInputParams}
          loadingServices={loadingServices}
          onAddInputParam={handleAddServiceInputParam}
          onCall={handleCallService}
          onInputParamChange={handleServiceInputParamChange}
          onRemoveInputParam={handleRemoveServiceInputParam}
          onServiceNameChange={handleServiceNameChange}
          result={serviceCallResult}
          serviceName={serviceName}
        />
      ) : activeTab === 'propertySetHistory' ? (
        <DevicePropertySetHistoryPanel
          ackDeadlineSeconds={propertySetAckDeadlineSeconds}
          loading={loadingPropertySetHistory}
          onAckDeadlineSecondsChange={handlePropertySetAckDeadlineSecondsChange}
          onPageChange={setPropertySetHistoryPage}
          onPageSizeChange={handlePropertySetHistoryPageSizeChange}
          onPropertyNameChange={handlePropertySetHistoryNameChange}
          onReload={() => loadPropertySetHistory()}
          page={propertySetHistoryPage}
          pageSize={propertySetHistoryPageSize}
          propertyName={propertySetHistoryNameFilter}
          sets={propertySetHistoryEntries}
          total={propertySetHistoryTotal}
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
