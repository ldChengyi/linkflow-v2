import { useI18n } from '@/contexts/I18nContext';
import type { AdminMessageKey } from '@/i18n/admin';
import { listProducts, type Product } from '@/services/products';
import { listTenants, type Tenant } from '@/services/tenants';
import {
  createThingsModel,
  deleteThingsModel,
  listThingsModels,
  type ThingsModel,
  type ThingsModelObject,
  type ThingsModelStatus,
  updateThingsModel,
} from '@/services/thingsmodels';
import { formatDateTime } from '@/utils/date';
import {
  CloseOutlined,
  DeleteOutlined,
  EditOutlined,
  PlusOutlined,
  SaveOutlined,
} from '@ant-design/icons';
import { notification } from 'antd';
import type { FormEvent } from 'react';
import { useEffect, useState } from 'react';
import AdminDataTable, {
  type AdminDataTableColumn,
} from './components/AdminDataTable';

interface ThingsModelFormState {
  product_id: string;
  model_version: string;
  model_name: string;
  description: string;
  status: ThingsModelStatus;
  is_current: boolean;
  properties: string;
  events: string;
  services: string;
}

type FormMode = 'create' | 'edit';
type ModelDefinitionTab = 'properties' | 'events' | 'services';
type PropertyDataType = 'int' | 'float' | 'double' | 'bool' | 'string';
type PropertyAccessMode = 'read' | 'write' | 'readwrite';
type EventLevel = 'info' | 'warning' | 'error';
type ServiceCallType = 'sync' | 'async';
type DefinitionField = 'properties' | 'events' | 'services';

interface ParamDraft {
  identifier: string;
  name: string;
  dataType: PropertyDataType;
  required: boolean;
  min: string;
  max: string;
  step: string;
  unit: string;
  precision: string;
}

interface ParamDraftRow extends ParamDraft {
  localId: string;
}

interface PropertyDraft extends ParamDraft {
  accessMode: PropertyAccessMode;
}

interface PropertyDraftRow extends PropertyDraft {
  localId: string;
}

interface EventDraft {
  identifier: string;
  name: string;
  desc: string;
  level: EventLevel;
}

interface EventDraftRow extends EventDraft {
  localId: string;
  outputParams: ParamDraftRow[];
}

interface ServiceDraft {
  identifier: string;
  name: string;
  desc: string;
  callType: ServiceCallType;
}

interface ServiceDraftRow extends ServiceDraft {
  localId: string;
  inputParams: ParamDraftRow[];
  outputParams: ParamDraftRow[];
}

const emptyObjectText = '{}';

const emptyForm: ThingsModelFormState = {
  product_id: '',
  model_version: '1',
  model_name: '',
  description: '',
  status: 'draft',
  is_current: false,
  properties: emptyObjectText,
  events: emptyObjectText,
  services: emptyObjectText,
};

const emptyParamDraft: ParamDraft = {
  identifier: '',
  name: '',
  dataType: 'string',
  required: false,
  min: '',
  max: '',
  step: '',
  unit: '',
  precision: '',
};

const createLocalId = () => {
  return (
    globalThis.crypto?.randomUUID?.() ??
    `${Date.now().toString(36)}-${Math.random().toString(36).slice(2)}`
  );
};

const createParamDraftRow = (): ParamDraftRow => ({
  ...emptyParamDraft,
  localId: createLocalId(),
});

const createPropertyDraftRow = (
  draft: PropertyDraft = emptyPropertyDraft,
): PropertyDraftRow => ({
  ...draft,
  localId: createLocalId(),
});

const emptyPropertyDraft: PropertyDraft = {
  ...emptyParamDraft,
  accessMode: 'readwrite',
};

const emptyEventDraft: EventDraft = {
  identifier: '',
  name: '',
  desc: '',
  level: 'info',
};

const emptyServiceDraft: ServiceDraft = {
  identifier: '',
  name: '',
  desc: '',
  callType: 'sync',
};

const createEventDraftRow = (
  draft: EventDraft = emptyEventDraft,
  outputParams: ParamDraftRow[] = [],
): EventDraftRow => ({
  ...draft,
  localId: createLocalId(),
  outputParams,
});

const createServiceDraftRow = (
  draft: ServiceDraft = emptyServiceDraft,
  inputParams: ParamDraftRow[] = [],
  outputParams: ParamDraftRow[] = [],
): ServiceDraftRow => ({
  ...draft,
  localId: createLocalId(),
  inputParams,
  outputParams,
});

const pageSizeOptions = [10, 20, 50];

const definitionTabs: Array<{
  id: ModelDefinitionTab;
  labelKey: AdminMessageKey;
}> = [
  { id: 'properties', labelKey: 'adminThingsModelPropertiesTab' },
  { id: 'events', labelKey: 'adminThingsModelEventsTab' },
  { id: 'services', labelKey: 'adminThingsModelServicesTab' },
];

const statusLabelKey: Record<ThingsModelStatus, AdminMessageKey> = {
  draft: 'adminThingsModelStatusDraft',
  published: 'adminThingsModelStatusPublished',
  deprecated: 'adminThingsModelStatusDeprecated',
};

const statusBadgeClassName = (status: ThingsModelStatus) => {
  return [
    'inline-flex rounded-full px-2.5 py-1 text-xs font-bold',
    status === 'published'
      ? 'bg-linkflow-primary-soft text-linkflow-primary'
      : status === 'deprecated'
      ? 'bg-slate-100 text-linkflow-muted dark:bg-slate-800 dark:text-linkflow-dark-muted'
      : 'bg-amber-50 text-amber-700 dark:bg-amber-950/30 dark:text-amber-300',
  ].join(' ');
};

const currentModelRowClassName =
  'border-l-4 border-l-linkflow-primary bg-linkflow-primary-soft/45 dark:border-l-linkflow-dark-primary dark:bg-linkflow-dark-primary/10';

const currentModelRibbonClassName =
  'pointer-events-none absolute right-3 top-3 z-10 flex h-6 items-center justify-center rounded-md bg-linkflow-primary px-2.5 text-[10px] font-black uppercase leading-none text-white shadow-sm dark:bg-linkflow-dark-primary dark:text-slate-950';

const currentModelBadgeClassName =
  'inline-flex items-center rounded-md border border-linkflow-primary bg-linkflow-primary px-2.5 py-1 text-xs font-black leading-none text-white shadow-sm dark:border-linkflow-dark-primary dark:bg-linkflow-dark-primary dark:text-slate-950';

const stringifyObject = (value: ThingsModelObject) => {
  return JSON.stringify(value ?? {}, null, 2);
};

const toFormState = (model: ThingsModel): ThingsModelFormState => ({
  product_id: model.product_id,
  model_version: String(model.model_version),
  model_name: model.model_name,
  description: model.description,
  status: model.status,
  is_current: model.is_current,
  properties: stringifyObject(model.properties),
  events: stringifyObject(model.events),
  services: stringifyObject(model.services),
});

const isPlainObject = (value: unknown): value is ThingsModelObject => {
  return typeof value === 'object' && value !== null && !Array.isArray(value);
};

const parseObjectField = (value: string) => {
  const parsed: unknown = JSON.parse(value);
  if (!isPlainObject(parsed)) {
    throw new Error('json object required');
  }
  return parsed;
};

const setObjectEntry = (raw: string, key: string, value: ThingsModelObject) => {
  const current = parseObjectField(raw);
  return stringifyObject({
    ...current,
    [key]: value,
  });
};

const mergeObjectTemplate = (raw: string, template: ThingsModelObject) => {
  const current = parseObjectField(raw);
  return stringifyObject({
    ...current,
    ...template,
  });
};

const objectEntry = (value: ThingsModelObject, key: string) => {
  const entry = value[key];
  return isPlainObject(entry) ? entry : {};
};

const numberOrUndefined = (value: string) => {
  const trimmed = value.trim();
  if (trimmed === '') {
    return undefined;
  }
  const nextValue = Number(trimmed);
  if (!Number.isFinite(nextValue)) {
    throw new Error('numeric spec fields must be valid numbers');
  }
  return nextValue;
};

const intOrUndefined = (value: string) => {
  const trimmed = value.trim();
  if (trimmed === '') {
    return undefined;
  }
  const nextValue = Number(trimmed);
  if (!Number.isInteger(nextValue) || nextValue < 0) {
    throw new Error('precision must be a non-negative integer');
  }
  return nextValue;
};

const stringFromUnknown = (value: unknown) => {
  return typeof value === 'string' ? value : '';
};

const boolFromUnknown = (value: unknown) => {
  return typeof value === 'boolean' ? value : false;
};

const dataTypeFromUnknown = (value: unknown): PropertyDataType => {
  switch (value) {
    case 'int':
    case 'float':
    case 'double':
    case 'bool':
    case 'string':
      return value;
    default:
      return 'string';
  }
};

const accessModeFromUnknown = (value: unknown): PropertyAccessMode => {
  switch (value) {
    case 'read':
    case 'write':
    case 'readwrite':
      return value;
    default:
      return 'readwrite';
  }
};

const eventLevelFromUnknown = (value: unknown): EventLevel => {
  switch (value) {
    case 'info':
    case 'warning':
    case 'error':
      return value;
    default:
      return 'info';
  }
};

const callTypeFromUnknown = (value: unknown): ServiceCallType => {
  switch (value) {
    case 'sync':
    case 'async':
      return value;
    default:
      return 'sync';
  }
};

const draftFromParamDefinition = (
  identifier: string,
  definition: unknown,
): ParamDraftRow => {
  const body = isPlainObject(definition) ? definition : {};
  const spec = isPlainObject(body.spec) ? body.spec : {};
  return {
    localId: createLocalId(),
    identifier,
    name: stringFromUnknown(body.name),
    dataType: dataTypeFromUnknown(body.data_type),
    required: boolFromUnknown(body.required),
    min: spec.min === undefined ? '' : String(spec.min),
    max: spec.max === undefined ? '' : String(spec.max),
    step: spec.step === undefined ? '' : String(spec.step),
    unit: stringFromUnknown(spec.unit),
    precision: spec.precision === undefined ? '' : String(spec.precision),
  };
};

const propertyRowsFromObject = (
  value: ThingsModelObject,
): PropertyDraftRow[] => {
  return Object.entries(value ?? {}).map(([identifier, definition]) => {
    const param = draftFromParamDefinition(identifier, definition);
    const body = isPlainObject(definition) ? definition : {};
    return {
      ...param,
      accessMode: accessModeFromUnknown(body.access_mode),
    };
  });
};

const eventRowsFromObject = (value: ThingsModelObject): EventDraftRow[] => {
  return Object.entries(value ?? {}).map(([identifier, definition]) => {
    const body = isPlainObject(definition) ? definition : {};
    const output = isPlainObject(body.output) ? body.output : {};
    return createEventDraftRow(
      {
        identifier,
        name: stringFromUnknown(body.name),
        desc: stringFromUnknown(body.desc),
        level: eventLevelFromUnknown(body.level),
      },
      Object.entries(output).map(([paramIdentifier, paramDefinition]) =>
        draftFromParamDefinition(paramIdentifier, paramDefinition),
      ),
    );
  });
};

const serviceRowsFromObject = (value: ThingsModelObject): ServiceDraftRow[] => {
  return Object.entries(value ?? {}).map(([identifier, definition]) => {
    const body = isPlainObject(definition) ? definition : {};
    const input = isPlainObject(body.input) ? body.input : {};
    const output = isPlainObject(body.output) ? body.output : {};
    return createServiceDraftRow(
      {
        identifier,
        name: stringFromUnknown(body.name),
        desc: stringFromUnknown(body.desc),
        callType: callTypeFromUnknown(body.call_type),
      },
      Object.entries(input).map(([paramIdentifier, paramDefinition]) =>
        draftFromParamDefinition(paramIdentifier, paramDefinition),
      ),
      Object.entries(output).map(([paramIdentifier, paramDefinition]) =>
        draftFromParamDefinition(paramIdentifier, paramDefinition),
      ),
    );
  });
};

const paramSpecFromDraft = (draft: ParamDraft): ThingsModelObject => {
  const spec: ThingsModelObject = {};
  const min = numberOrUndefined(draft.min);
  const max = numberOrUndefined(draft.max);
  const step = numberOrUndefined(draft.step);
  const precision = intOrUndefined(draft.precision);
  const unit = draft.unit.trim();

  if (min !== undefined) {
    spec.min = min;
  }
  if (max !== undefined) {
    spec.max = max;
  }
  if (step !== undefined) {
    spec.step = step;
  }
  if (unit !== '') {
    spec.unit = unit;
  }
  if (precision !== undefined) {
    spec.precision = precision;
  }

  return spec;
};

const paramDefinitionFromDraft = (draft: ParamDraft): ThingsModelObject => {
  const spec = paramSpecFromDraft(draft);
  const definition: ThingsModelObject = {
    name: draft.name.trim(),
    data_type: draft.dataType,
    required: draft.required,
  };

  if (Object.keys(spec).length > 0) {
    definition.spec = spec;
  }

  return definition;
};

const propertyDefinitionFromDraft = (
  draft: PropertyDraft,
): ThingsModelObject => {
  return {
    ...paramDefinitionFromDraft(draft),
    access_mode: draft.accessMode,
  };
};

const objectFromPropertyRows = (
  rows: PropertyDraftRow[],
): ThingsModelObject => {
  return rows.reduce<ThingsModelObject>((result, row) => {
    const identifier = row.identifier.trim();
    if (!identifier) {
      return result;
    }
    result[identifier] = propertyDefinitionFromDraft(row);
    return result;
  }, {});
};

const objectFromEventRows = (rows: EventDraftRow[]): ThingsModelObject => {
  return rows.reduce<ThingsModelObject>((result, row) => {
    const identifier = row.identifier.trim();
    if (!identifier) {
      return result;
    }
    const output = row.outputParams.reduce<ThingsModelObject>(
      (outputResult, param) => {
        const paramIdentifier = param.identifier.trim();
        if (!paramIdentifier) {
          return outputResult;
        }
        outputResult[paramIdentifier] = paramDefinitionFromDraft(param);
        return outputResult;
      },
      {},
    );
    result[identifier] = {
      name: row.name.trim(),
      level: row.level,
      desc: row.desc.trim(),
      output,
    };
    return result;
  }, {});
};

const objectFromServiceRows = (rows: ServiceDraftRow[]): ThingsModelObject => {
  return rows.reduce<ThingsModelObject>((result, row) => {
    const identifier = row.identifier.trim();
    if (!identifier) {
      return result;
    }
    const input = row.inputParams.reduce<ThingsModelObject>(
      (inputResult, param) => {
        const paramIdentifier = param.identifier.trim();
        if (!paramIdentifier) {
          return inputResult;
        }
        inputResult[paramIdentifier] = paramDefinitionFromDraft(param);
        return inputResult;
      },
      {},
    );
    const output = row.outputParams.reduce<ThingsModelObject>(
      (outputResult, param) => {
        const paramIdentifier = param.identifier.trim();
        if (!paramIdentifier) {
          return outputResult;
        }
        outputResult[paramIdentifier] = paramDefinitionFromDraft(param);
        return outputResult;
      },
      {},
    );
    result[identifier] = {
      name: row.name.trim(),
      call_type: row.callType,
      desc: row.desc.trim(),
      input,
      output,
    };
    return result;
  }, {});
};

const safeDefinitionPreview = (build: () => ThingsModelObject) => {
  try {
    return stringifyObject(build());
  } catch (error) {
    return stringifyObject({
      error: error instanceof Error ? error.message : 'invalid definition',
    });
  }
};

const mergeEventDefinition = (
  raw: string,
  eventKey: string,
  event: EventDraft,
  outputParams: ParamDraft[] = [],
) => {
  const current = parseObjectField(raw);
  const existing = objectEntry(current, eventKey);
  const output = objectEntry(existing, 'output');
  const nextOutput = outputParams.reduce<ThingsModelObject>(
    (result, outputParam) => ({
      ...result,
      [outputParam.identifier.trim()]: paramDefinitionFromDraft(outputParam),
    }),
    output,
  );

  return stringifyObject({
    ...current,
    [eventKey]: {
      ...existing,
      name: event.name.trim(),
      level: event.level,
      desc: event.desc.trim(),
      output: nextOutput,
    },
  });
};

const mergeServiceDefinition = (
  raw: string,
  serviceKey: string,
  service: ServiceDraft,
  inputParams: ParamDraft[] = [],
  outputParams: ParamDraft[] = [],
) => {
  const current = parseObjectField(raw);
  const existing = objectEntry(current, serviceKey);
  const input = objectEntry(existing, 'input');
  const output = objectEntry(existing, 'output');
  const nextInput = inputParams.reduce<ThingsModelObject>(
    (result, inputParam) => ({
      ...result,
      [inputParam.identifier.trim()]: paramDefinitionFromDraft(inputParam),
    }),
    input,
  );
  const nextOutput = outputParams.reduce<ThingsModelObject>(
    (result, outputParam) => ({
      ...result,
      [outputParam.identifier.trim()]: paramDefinitionFromDraft(outputParam),
    }),
    output,
  );

  return stringifyObject({
    ...current,
    [serviceKey]: {
      ...existing,
      name: service.name.trim(),
      call_type: service.callType,
      desc: service.desc.trim(),
      input: nextInput,
      output: nextOutput,
    },
  });
};

const objectKeyCount = (value: ThingsModelObject) => {
  return Object.keys(value ?? {}).length;
};

interface DefinitionPreset {
  labelKey: AdminMessageKey;
  template: ThingsModelObject;
}

const propertyPresets: DefinitionPreset[] = [
  {
    labelKey: 'adminThingsModelPresetTemperature',
    template: {
      temperature: {
        name: '温度',
        data_type: 'float',
        access_mode: 'read',
        required: false,
        spec: { min: -40, max: 125, step: 0.1, unit: 'celsius', precision: 1 },
      },
    },
  },
  {
    labelKey: 'adminThingsModelPresetHumidity',
    template: {
      humidity: {
        name: '湿度',
        data_type: 'float',
        access_mode: 'read',
        required: false,
        spec: { min: 0, max: 100, step: 0.1, unit: 'percent', precision: 1 },
      },
    },
  },
  {
    labelKey: 'adminThingsModelPresetBattery',
    template: {
      battery_level: {
        name: '电量',
        data_type: 'int',
        access_mode: 'read',
        required: false,
        spec: { min: 0, max: 100, step: 1, unit: 'percent' },
      },
    },
  },
  {
    labelKey: 'adminThingsModelPresetSwitch',
    template: {
      switch_state: {
        name: '开关状态',
        data_type: 'bool',
        access_mode: 'readwrite',
        required: false,
      },
    },
  },
];

const eventPresets: DefinitionPreset[] = [
  {
    labelKey: 'adminThingsModelPresetTemperatureAlarm',
    template: {
      temperature_alarm: {
        name: '温度报警',
        level: 'warning',
        desc: '设备检测到温度超过安全阈值',
        output: {
          temperature: {
            name: '当前温度',
            data_type: 'float',
            required: true,
            spec: {
              min: -40,
              max: 125,
              step: 0.1,
              unit: 'celsius',
              precision: 1,
            },
          },
          threshold: {
            name: '报警阈值',
            data_type: 'float',
            required: false,
            spec: { unit: 'celsius', precision: 1 },
          },
        },
      },
    },
  },
  {
    labelKey: 'adminThingsModelPresetLowBattery',
    template: {
      low_battery: {
        name: '低电量告警',
        level: 'warning',
        desc: '设备电量低于设定阈值',
        output: {
          battery_level: {
            name: '当前电量',
            data_type: 'int',
            required: true,
            spec: { min: 0, max: 100, step: 1, unit: 'percent' },
          },
        },
      },
    },
  },
  {
    labelKey: 'adminThingsModelPresetFaultReport',
    template: {
      fault_report: {
        name: '故障上报',
        level: 'error',
        desc: '设备上报故障码和故障信息',
        output: {
          code: { name: '故障码', data_type: 'string', required: true },
          message: { name: '故障信息', data_type: 'string', required: false },
        },
      },
    },
  },
];

const servicePresets: DefinitionPreset[] = [
  {
    labelKey: 'adminThingsModelPresetReboot',
    template: {
      reboot: {
        name: '重启设备',
        call_type: 'async',
        desc: '平台下发重启命令，设备返回执行结果',
        input: {},
        output: {
          accepted: { name: '是否接受', data_type: 'bool', required: true },
        },
      },
    },
  },
  {
    labelKey: 'adminThingsModelPresetReportInterval',
    template: {
      set_report_interval: {
        name: '设置上报周期',
        call_type: 'sync',
        desc: '设置设备属性上报间隔',
        input: {
          interval_seconds: {
            name: '上报间隔',
            data_type: 'int',
            required: true,
            spec: { min: 1, max: 86400, step: 1, unit: 'second' },
          },
        },
        output: {
          applied: { name: '是否生效', data_type: 'bool', required: true },
        },
      },
    },
  },
  {
    labelKey: 'adminThingsModelPresetSwitchService',
    template: {
      set_switch_state: {
        name: '设置开关',
        call_type: 'sync',
        desc: '平台设置设备开关状态',
        input: {
          switch_state: {
            name: '目标开关状态',
            data_type: 'bool',
            required: true,
          },
        },
        output: {
          applied: { name: '是否生效', data_type: 'bool', required: true },
        },
      },
    },
  },
];

interface DefinitionPresetBarProps {
  onApply: (template: ThingsModelObject) => void;
  presets: DefinitionPreset[];
  title: string;
}

const DefinitionPresetBar = ({
  onApply,
  presets,
  title,
}: DefinitionPresetBarProps) => {
  const { t } = useI18n();

  return (
    <section className="grid gap-2 rounded-lg border border-linkflow-border bg-slate-50/70 px-3 py-2 dark:border-linkflow-dark-border dark:bg-slate-950/20">
      <div className="text-xs font-black uppercase tracking-wide text-linkflow-subtle dark:text-linkflow-dark-subtle">
        {title}
      </div>
      <div className="flex flex-wrap gap-2">
        {presets.map((preset) => (
          <button
            key={preset.labelKey}
            type="button"
            className="inline-flex h-8 items-center rounded-md border border-linkflow-primary/30 bg-white px-2.5 text-xs font-bold text-linkflow-primary transition hover:border-linkflow-primary hover:bg-linkflow-primary-soft dark:border-linkflow-dark-primary/40 dark:bg-linkflow-dark-panel/70 dark:text-linkflow-dark-primary dark:hover:border-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-soft"
            onClick={(event) => {
              event.preventDefault();
              event.stopPropagation();
              onApply(preset.template);
            }}
          >
            {t(preset.labelKey)}
          </button>
        ))}
      </div>
    </section>
  );
};

interface JSONPreviewFieldProps {
  label: string;
  value: string;
}

const JSONPreviewField = ({ label, value }: JSONPreviewFieldProps) => {
  return (
    <label className="group grid overflow-hidden rounded-lg border border-linkflow-border bg-slate-50/70 text-sm font-bold shadow-sm ring-1 ring-slate-950/5 transition dark:border-linkflow-dark-border dark:bg-slate-950/30 dark:ring-white/5">
      <span className="flex min-h-10 items-center justify-between gap-3 border-b border-linkflow-border bg-white/80 px-3 py-2 dark:border-linkflow-dark-border dark:bg-linkflow-dark-page/80">
        <span>{label}</span>
        <span className="rounded bg-slate-100 px-2 py-0.5 font-mono text-[11px] font-bold uppercase tracking-wide text-linkflow-subtle dark:bg-slate-800 dark:text-linkflow-dark-subtle">
          JSON
        </span>
      </span>
      <pre
        className="m-0 max-h-60 min-h-36 overflow-auto whitespace-pre-wrap break-words border-0 bg-transparent px-3 py-3 font-mono text-xs font-medium leading-5 text-linkflow-text outline-none transition dark:text-linkflow-dark-text"
        title={value}
      >
        {value}
      </pre>
      <textarea className="sr-only" value={value} readOnly required />
    </label>
  );
};

interface ParamDraftListProps {
  addLabel: string;
  onAddRow: () => void;
  onChangeRow: <K extends keyof ParamDraft>(
    localId: string,
    key: K,
    value: ParamDraft[K],
  ) => void;
  onRemoveRow: (localId: string) => void;
  rows: ParamDraftRow[];
  title: string;
}

const ParamDraftList = ({
  addLabel,
  onAddRow,
  onChangeRow,
  onRemoveRow,
  rows,
  title,
}: ParamDraftListProps) => {
  const { t } = useI18n();

  return (
    <section className="grid gap-3 border-t border-linkflow-primary/15 pt-3 dark:border-linkflow-dark-primary/20">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h4 className="m-0 text-sm font-black text-linkflow-primary dark:text-linkflow-dark-primary">
          {title}
        </h4>
        <button
          type="button"
          className="inline-flex h-9 items-center gap-2 rounded-md border border-linkflow-primary bg-white/70 px-3 text-xs font-bold text-linkflow-primary shadow-sm transition hover:bg-linkflow-primary-soft dark:border-linkflow-dark-primary dark:bg-linkflow-dark-panel/70 dark:text-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-soft"
          onClick={(event) => {
            event.preventDefault();
            event.stopPropagation();
            onAddRow();
          }}
        >
          <PlusOutlined aria-hidden="true" />
          {addLabel}
        </button>
      </div>
      {rows.length > 0 ? (
        <div className="grid gap-3">
          {rows.map((draft, index) => (
            <div
              key={draft.localId}
              className="grid gap-3 rounded-md border border-linkflow-border bg-slate-50/70 p-3 dark:border-linkflow-dark-border dark:bg-slate-950/20"
            >
              <div className="flex items-center justify-between gap-2">
                <span className="text-xs font-black text-linkflow-subtle dark:text-linkflow-dark-subtle">
                  #{index + 1}
                </span>
                <button
                  type="button"
                  className="inline-flex h-8 items-center gap-1 rounded-md border border-rose-200 bg-white px-2 text-xs font-bold text-rose-600 transition hover:bg-rose-50 dark:border-rose-900/70 dark:bg-linkflow-dark-panel dark:text-rose-300 dark:hover:bg-rose-950/30"
                  onClick={(event) => {
                    event.preventDefault();
                    event.stopPropagation();
                    onRemoveRow(draft.localId);
                  }}
                >
                  <DeleteOutlined aria-hidden="true" />
                  {t('adminThingsModelRemoveParam')}
                </button>
              </div>
              <div className="grid gap-3 [grid-template-columns:repeat(auto-fit,minmax(150px,1fr))]">
                <label className="grid gap-1.5 text-sm font-bold">
                  {t('adminThingsModelParamIdentifier')}
                  <input
                    className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                    value={draft.identifier}
                    onChange={(event) =>
                      onChangeRow(
                        draft.localId,
                        'identifier',
                        event.target.value,
                      )
                    }
                    placeholder="temperature"
                  />
                </label>
                <label className="grid gap-1.5 text-sm font-bold">
                  {t('adminThingsModelParamName')}
                  <input
                    className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                    value={draft.name}
                    onChange={(event) =>
                      onChangeRow(draft.localId, 'name', event.target.value)
                    }
                    placeholder={t('adminThingsModelPropertyNameExample')}
                  />
                </label>
                <label className="grid gap-1.5 text-sm font-bold">
                  {t('adminThingsModelDataType')}
                  <select
                    className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                    value={draft.dataType}
                    onChange={(event) =>
                      onChangeRow(
                        draft.localId,
                        'dataType',
                        event.target.value as PropertyDataType,
                      )
                    }
                  >
                    <option value="string">string</option>
                    <option value="int">int</option>
                    <option value="float">float</option>
                    <option value="double">double</option>
                    <option value="bool">bool</option>
                  </select>
                </label>
                <label className="inline-flex h-10 w-fit self-end items-center gap-2 rounded-full border border-linkflow-border bg-white px-3 text-sm font-bold text-linkflow-muted dark:border-linkflow-dark-border dark:bg-slate-900/40 dark:text-linkflow-dark-muted">
                  <input
                    className="h-4 w-4"
                    type="checkbox"
                    checked={draft.required}
                    onChange={(event) =>
                      onChangeRow(
                        draft.localId,
                        'required',
                        event.target.checked,
                      )
                    }
                  />
                  {t('adminThingsModelRequired')}
                </label>
                <label className="grid gap-1.5 text-sm font-bold">
                  {t('adminThingsModelSpecMin')}
                  <input
                    className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                    inputMode="decimal"
                    value={draft.min}
                    onChange={(event) =>
                      onChangeRow(draft.localId, 'min', event.target.value)
                    }
                    placeholder="-40"
                  />
                </label>
                <label className="grid gap-1.5 text-sm font-bold">
                  {t('adminThingsModelSpecMax')}
                  <input
                    className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                    inputMode="decimal"
                    value={draft.max}
                    onChange={(event) =>
                      onChangeRow(draft.localId, 'max', event.target.value)
                    }
                    placeholder="125"
                  />
                </label>
                <label className="grid gap-1.5 text-sm font-bold">
                  {t('adminThingsModelSpecStep')}
                  <input
                    className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                    inputMode="decimal"
                    value={draft.step}
                    onChange={(event) =>
                      onChangeRow(draft.localId, 'step', event.target.value)
                    }
                    placeholder="0.1"
                  />
                </label>
                <label className="grid gap-1.5 text-sm font-bold">
                  {t('adminThingsModelSpecUnit')}
                  <input
                    className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                    value={draft.unit}
                    onChange={(event) =>
                      onChangeRow(draft.localId, 'unit', event.target.value)
                    }
                    placeholder="celsius"
                  />
                </label>
                <label className="grid gap-1.5 text-sm font-bold">
                  {t('adminThingsModelSpecPrecision')}
                  <input
                    className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                    inputMode="numeric"
                    value={draft.precision}
                    onChange={(event) =>
                      onChangeRow(
                        draft.localId,
                        'precision',
                        event.target.value,
                      )
                    }
                    placeholder="1"
                  />
                </label>
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="rounded-md border border-dashed border-linkflow-border bg-slate-50/60 px-3 py-2 text-sm font-medium text-linkflow-subtle dark:border-linkflow-dark-border dark:bg-slate-950/20 dark:text-linkflow-dark-subtle">
          {t('adminThingsModelNoParams')}
        </div>
      )}
    </section>
  );
};

const ThingsModelManagement = () => {
  const { t } = useI18n();
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [tenantId, setTenantId] = useState('');
  const [products, setProducts] = useState<Product[]>([]);
  const [productFilter, setProductFilter] = useState('');
  const [models, setModels] = useState<ThingsModel[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [formMode, setFormMode] = useState<FormMode | null>(null);
  const [editingModel, setEditingModel] = useState<ThingsModel | null>(null);
  const [form, setForm] = useState<ThingsModelFormState>(emptyForm);
  const [activeDefinitionTab, setActiveDefinitionTab] =
    useState<ModelDefinitionTab>('properties');
  const [propertyRows, setPropertyRows] = useState<PropertyDraftRow[]>([]);
  const [eventRows, setEventRows] = useState<EventDraftRow[]>([]);
  const [serviceRows, setServiceRows] = useState<ServiceDraftRow[]>([]);

  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const propertyDefinitionsText = safeDefinitionPreview(() =>
    objectFromPropertyRows(propertyRows),
  );
  const eventDefinitionsText = safeDefinitionPreview(() =>
    objectFromEventRows(eventRows),
  );
  const serviceDefinitionsText = safeDefinitionPreview(() =>
    objectFromServiceRows(serviceRows),
  );
  const productNameById = new Map(
    products.map((product) => [product.id, product.product_name]),
  );

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
        setForm((current) => ({
          ...current,
          product_id: current.product_id || result.items[0]?.id || '',
        }));
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

  const loadModels = async (nextPage = page, nextPageSize = pageSize) => {
    if (!tenantId) {
      setModels([]);
      setTotal(0);
      return;
    }

    setLoading(true);
    try {
      const result = await listThingsModels({
        tenant_id: tenantId,
        product_id: productFilter || undefined,
        page: nextPage,
        page_size: nextPageSize,
      });
      setModels(result.items);
      setTotal(result.total);
      setPage(result.page);
      setPageSize(result.page_size);
    } catch (error) {
      notification.error({
        message: t('adminThingsModelLoadFailed'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    let cancelled = false;

    const run = async () => {
      if (!tenantId) {
        setModels([]);
        setTotal(0);
        return;
      }

      setLoading(true);
      try {
        const result = await listThingsModels({
          tenant_id: tenantId,
          product_id: productFilter || undefined,
          page,
          page_size: pageSize,
        });
        if (cancelled) {
          return;
        }
        setModels(result.items);
        setTotal(result.total);
        setPage(result.page);
        setPageSize(result.page_size);
      } catch (error) {
        if (cancelled) {
          return;
        }
        notification.error({
          message: t('adminThingsModelLoadFailed'),
          description: error instanceof Error ? error.message : undefined,
          placement: 'topRight',
        });
      } finally {
        if (!cancelled) {
          setLoading(false);
        }
      }
    };

    run();

    return () => {
      cancelled = true;
    };
  }, [page, pageSize, productFilter, tenantId, t]);

  const updateFormField = <K extends keyof ThingsModelFormState>(
    key: K,
    value: ThingsModelFormState[K],
  ) => {
    setForm((current) => ({ ...current, [key]: value }));
  };

  const updatePropertyRow = <K extends keyof PropertyDraft>(
    localId: string,
    key: K,
    value: PropertyDraft[K],
  ) => {
    setPropertyRows((current) =>
      current.map((row) =>
        row.localId === localId ? { ...row, [key]: value } : row,
      ),
    );
  };

  const updateEventRow = <K extends keyof EventDraft>(
    localId: string,
    key: K,
    value: EventDraft[K],
  ) => {
    setEventRows((current) =>
      current.map((row) =>
        row.localId === localId ? { ...row, [key]: value } : row,
      ),
    );
  };
  const addEventOutputParam = (eventLocalId: string) => {
    setEventRows((current) =>
      current.map((row) =>
        row.localId === eventLocalId
          ? {
              ...row,
              outputParams: [...row.outputParams, createParamDraftRow()],
            }
          : row,
      ),
    );
  };
  const updateEventOutputParam = <K extends keyof ParamDraft>(
    eventLocalId: string,
    localId: string,
    key: K,
    value: ParamDraft[K],
  ) => {
    setEventRows((current) =>
      current.map((row) =>
        row.localId === eventLocalId
          ? {
              ...row,
              outputParams: row.outputParams.map((param) =>
                param.localId === localId ? { ...param, [key]: value } : param,
              ),
            }
          : row,
      ),
    );
  };
  const removeEventOutputParam = (eventLocalId: string, localId: string) => {
    setEventRows((current) =>
      current.map((row) =>
        row.localId === eventLocalId
          ? {
              ...row,
              outputParams: row.outputParams.filter(
                (param) => param.localId !== localId,
              ),
            }
          : row,
      ),
    );
  };

  const updateServiceRow = <K extends keyof ServiceDraft>(
    localId: string,
    key: K,
    value: ServiceDraft[K],
  ) => {
    setServiceRows((current) =>
      current.map((row) =>
        row.localId === localId ? { ...row, [key]: value } : row,
      ),
    );
  };
  const addServiceInputParam = (serviceLocalId: string) => {
    setServiceRows((current) =>
      current.map((row) =>
        row.localId === serviceLocalId
          ? { ...row, inputParams: [...row.inputParams, createParamDraftRow()] }
          : row,
      ),
    );
  };
  const updateServiceInputParam = <K extends keyof ParamDraft>(
    serviceLocalId: string,
    localId: string,
    key: K,
    value: ParamDraft[K],
  ) => {
    setServiceRows((current) =>
      current.map((row) =>
        row.localId === serviceLocalId
          ? {
              ...row,
              inputParams: row.inputParams.map((param) =>
                param.localId === localId ? { ...param, [key]: value } : param,
              ),
            }
          : row,
      ),
    );
  };
  const removeServiceInputParam = (serviceLocalId: string, localId: string) => {
    setServiceRows((current) =>
      current.map((row) =>
        row.localId === serviceLocalId
          ? {
              ...row,
              inputParams: row.inputParams.filter(
                (param) => param.localId !== localId,
              ),
            }
          : row,
      ),
    );
  };
  const addServiceOutputParam = (serviceLocalId: string) => {
    setServiceRows((current) =>
      current.map((row) =>
        row.localId === serviceLocalId
          ? {
              ...row,
              outputParams: [...row.outputParams, createParamDraftRow()],
            }
          : row,
      ),
    );
  };
  const updateServiceOutputParam = <K extends keyof ParamDraft>(
    serviceLocalId: string,
    localId: string,
    key: K,
    value: ParamDraft[K],
  ) => {
    setServiceRows((current) =>
      current.map((row) =>
        row.localId === serviceLocalId
          ? {
              ...row,
              outputParams: row.outputParams.map((param) =>
                param.localId === localId ? { ...param, [key]: value } : param,
              ),
            }
          : row,
      ),
    );
  };
  const removeServiceOutputParam = (
    serviceLocalId: string,
    localId: string,
  ) => {
    setServiceRows((current) =>
      current.map((row) =>
        row.localId === serviceLocalId
          ? {
              ...row,
              outputParams: row.outputParams.filter(
                (param) => param.localId !== localId,
              ),
            }
          : row,
      ),
    );
  };

  const openCreateForm = () => {
    setEditingModel(null);
    setActiveDefinitionTab('properties');
    setPropertyRows([createPropertyDraftRow()]);
    setEventRows([createEventDraftRow()]);
    setServiceRows([createServiceDraftRow()]);
    setForm({
      ...emptyForm,
      product_id: productFilter || products[0]?.id || '',
    });
    setFormMode('create');
  };

  const openEditForm = (model: ThingsModel) => {
    setEditingModel(model);
    setActiveDefinitionTab('properties');
    setPropertyRows(propertyRowsFromObject(model.properties));
    setEventRows(eventRowsFromObject(model.events));
    setServiceRows(serviceRowsFromObject(model.services));
    setForm(toFormState(model));
    setFormMode('edit');
  };

  const closeForm = () => {
    setFormMode(null);
    setEditingModel(null);
    setActiveDefinitionTab('properties');
    setPropertyRows([]);
    setEventRows([]);
    setServiceRows([]);
    setForm(emptyForm);
  };

  const handleTenantChange = (nextTenantId: string) => {
    setTenantId(nextTenantId);
    setProductFilter('');
    setPage(1);
    closeForm();
  };

  const handleStatusChange = (status: ThingsModelStatus) => {
    setForm((current) => ({
      ...current,
      status,
      is_current: status === 'published' ? current.is_current : false,
    }));
  };

  const validateParamRows = (params: ParamDraftRow[]) => {
    return params.every(
      (param) => param.identifier.trim() !== '' && param.name.trim() !== '',
    );
  };

  const validateDefinitionRows = () => {
    return (
      propertyRows.every(
        (row) => row.identifier.trim() !== '' && row.name.trim() !== '',
      ) &&
      eventRows.every(
        (row) =>
          row.identifier.trim() !== '' &&
          row.name.trim() !== '' &&
          validateParamRows(row.outputParams),
      ) &&
      serviceRows.every(
        (row) =>
          row.identifier.trim() !== '' &&
          row.name.trim() !== '' &&
          validateParamRows(row.inputParams) &&
          validateParamRows(row.outputParams),
      )
    );
  };

  const handleApplyPreset = (
    field: DefinitionField,
    template: ThingsModelObject,
  ) => {
    if (field === 'properties') {
      setPropertyRows((current) => [
        ...current,
        ...propertyRowsFromObject(template),
      ]);
      return;
    }
    if (field === 'events') {
      setEventRows((current) => [...current, ...eventRowsFromObject(template)]);
      return;
    }
    setServiceRows((current) => [
      ...current,
      ...serviceRowsFromObject(template),
    ]);
  };

  const handleAddProperty = () => {
    setPropertyRows((current) => [...current, createPropertyDraftRow()]);
  };

  const handleRemoveProperty = (localId: string) => {
    setPropertyRows((current) =>
      current.filter((row) => row.localId !== localId),
    );
  };

  const handleAddEvent = () => {
    setEventRows((current) => [...current, createEventDraftRow()]);
  };

  const handleRemoveEvent = (localId: string) => {
    setEventRows((current) => current.filter((row) => row.localId !== localId));
  };

  const handleAddService = () => {
    setServiceRows((current) => [...current, createServiceDraftRow()]);
  };

  const handleRemoveService = (localId: string) => {
    setServiceRows((current) =>
      current.filter((row) => row.localId !== localId),
    );
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!tenantId) {
      return;
    }

    if (!validateDefinitionRows()) {
      notification.warning({
        message: t('adminThingsModelDefinitionRequired'),
        placement: 'topRight',
      });
      return;
    }

    setSaving(true);
    try {
      const properties = objectFromPropertyRows(propertyRows);
      const events = objectFromEventRows(eventRows);
      const services = objectFromServiceRows(serviceRows);

      if (formMode === 'create') {
        await createThingsModel({
          tenant_id: tenantId,
          product_id: form.product_id,
          model_version: Number(form.model_version),
          model_name: form.model_name.trim(),
          description: form.description.trim(),
          status: form.status,
          is_current: form.status === 'published' && form.is_current,
          properties,
          events,
          services,
        });
        notification.success({
          message: t('adminThingsModelCreateSuccess'),
          placement: 'topRight',
        });
        setPage(1);
      }

      if (formMode === 'edit' && editingModel) {
        await updateThingsModel(editingModel.id, {
          model_name: form.model_name.trim(),
          description: form.description.trim(),
          status: form.status,
          is_current: form.status === 'published' && form.is_current,
          properties,
          events,
          services,
        });
        notification.success({
          message: t('adminThingsModelUpdateSuccess'),
          placement: 'topRight',
        });
      }

      closeForm();
      await loadModels(formMode === 'create' ? 1 : page, pageSize);
    } catch (error) {
      notification.error({
        message:
          error instanceof SyntaxError
            ? t('adminThingsModelJSONInvalid')
            : formMode === 'create'
            ? t('adminThingsModelCreateFailed')
            : t('adminThingsModelUpdateFailed'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (model: ThingsModel) => {
    if (!window.confirm(t('adminThingsModelDeleteConfirm'))) {
      return;
    }

    setDeletingId(model.id);
    try {
      await deleteThingsModel(model.id);
      notification.success({
        message: t('adminThingsModelDeleteSuccess'),
        placement: 'topRight',
      });
      if (models.length === 1 && page > 1) {
        setPage((current) => current - 1);
      } else {
        await loadModels();
      }
    } catch (error) {
      notification.error({
        message: t('adminThingsModelDeleteFailed'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    } finally {
      setDeletingId(null);
    }
  };

  const columns: AdminDataTableColumn<ThingsModel>[] = [
    {
      key: 'name',
      title: t('adminThingsModelName'),
      render: (model) => (
        <div>
          <strong className="block text-sm">{model.model_name}</strong>
          <span className="mt-1 block text-xs font-semibold text-linkflow-primary">
            v{model.model_version}
          </span>
        </div>
      ),
    },
    {
      key: 'product',
      title: t('adminProductName'),
      render: (model) =>
        productNameById.get(model.product_id) ??
        t('adminThingsModelUnknownProduct'),
    },
    {
      key: 'status',
      title: t('adminThingsModelStatus'),
      render: (model) => (
        <span className={statusBadgeClassName(model.status)}>
          {t(statusLabelKey[model.status])}
        </span>
      ),
    },
    {
      key: 'current',
      title: t('adminThingsModelCurrent'),
      render: (model) =>
        model.is_current ? (
          <span className={currentModelBadgeClassName}>
            {t('adminThingsModelCurrent')}
          </span>
        ) : (
          <span className="text-sm text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminThingsModelCurrentNo')}
          </span>
        ),
    },
    {
      key: 'schema',
      title: t('adminThingsModelSchema'),
      className: 'text-sm text-linkflow-subtle dark:text-linkflow-dark-subtle',
      render: (model) =>
        t('adminThingsModelSchemaSummary')
          .replace('{properties}', String(objectKeyCount(model.properties)))
          .replace('{events}', String(objectKeyCount(model.events)))
          .replace('{services}', String(objectKeyCount(model.services))),
    },
    {
      key: 'updatedAt',
      title: t('adminThingsModelUpdatedAt'),
      className: 'text-sm text-linkflow-subtle dark:text-linkflow-dark-subtle',
      render: (model) => formatDateTime(model.updated_at),
    },
    {
      key: 'actions',
      title: t('adminThingsModelActions'),
      headerClassName: 'text-right',
      render: (model) => (
        <div className="flex justify-end gap-2">
          <button
            type="button"
            className="grid h-10 w-10 place-items-center rounded-md border border-linkflow-border bg-white text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
            aria-label={t('adminThingsModelEdit')}
            onClick={() => openEditForm(model)}
          >
            <EditOutlined aria-hidden="true" />
          </button>
          <button
            type="button"
            className="grid h-10 w-10 place-items-center rounded-md border border-linkflow-border bg-white text-red-600 transition hover:border-red-500 hover:bg-red-50 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:hover:border-red-400 dark:hover:bg-red-950/30"
            aria-label={t('adminThingsModelDelete')}
            onClick={() => handleDelete(model)}
            disabled={deletingId === model.id}
          >
            <DeleteOutlined aria-hidden="true" />
          </button>
        </div>
      ),
    },
  ];

  const mobileItems = models.map((model) => (
    <article
      key={model.id}
      className={[
        'relative overflow-hidden rounded-lg border border-linkflow-border p-4 dark:border-linkflow-dark-border',
        model.is_current ? currentModelRowClassName : '',
      ].join(' ')}
    >
      {model.is_current ? (
        <span className={currentModelRibbonClassName} aria-hidden="true">
          {t('adminThingsModelCurrent')}
        </span>
      ) : null}
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h4 className="m-0 truncate text-base font-bold">
            {model.model_name}
          </h4>
          <p className="m-0 mt-1 truncate text-sm font-semibold text-linkflow-primary">
            {productNameById.get(model.product_id) ??
              t('adminThingsModelUnknownProduct')}{' '}
            / v{model.model_version}
          </p>
        </div>
        <span className={statusBadgeClassName(model.status)}>
          {t(statusLabelKey[model.status])}
        </span>
      </div>
      <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminThingsModelCurrent')}
          </dt>
          <dd className="m-0 mt-1">
            {model.is_current
              ? t('adminThingsModelCurrentYes')
              : t('adminThingsModelCurrentNo')}
          </dd>
        </div>
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminThingsModelUpdatedAt')}
          </dt>
          <dd className="m-0 mt-1">{formatDateTime(model.updated_at)}</dd>
        </div>
        <div className="col-span-2">
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminThingsModelSchema')}
          </dt>
          <dd className="m-0 mt-1">
            {t('adminThingsModelSchemaSummary')
              .replace('{properties}', String(objectKeyCount(model.properties)))
              .replace('{events}', String(objectKeyCount(model.events)))
              .replace('{services}', String(objectKeyCount(model.services)))}
          </dd>
        </div>
      </dl>
      <div className="mt-4 flex gap-2">
        <button
          type="button"
          className="inline-flex h-11 flex-1 items-center justify-center gap-2 rounded-md border border-linkflow-border bg-white text-sm font-bold text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
          onClick={() => openEditForm(model)}
        >
          <EditOutlined aria-hidden="true" />
          {t('adminThingsModelEdit')}
        </button>
        <button
          type="button"
          className="inline-flex h-11 flex-1 items-center justify-center gap-2 rounded-md border border-linkflow-border bg-white text-sm font-bold text-red-600 transition hover:border-red-500 hover:bg-red-50 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:hover:border-red-400 dark:hover:bg-red-950/30"
          onClick={() => handleDelete(model)}
          disabled={deletingId === model.id}
        >
          <DeleteOutlined aria-hidden="true" />
          {t('adminThingsModelDelete')}
        </button>
      </div>
    </article>
  ));

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
              onChange={(event) => {
                setProductFilter(event.target.value);
                setPage(1);
              }}
            >
              <option value="">{t('adminThingsModelAllProducts')}</option>
              {products.map((product) => (
                <option key={product.id} value={product.id}>
                  {product.product_name}
                </option>
              ))}
            </select>
          </label>
        </div>
        <button
          type="button"
          className="inline-flex h-11 items-center gap-2 rounded-md border border-linkflow-primary bg-white/70 px-3 text-sm font-bold text-linkflow-primary shadow-sm transition hover:bg-linkflow-primary-soft disabled:cursor-not-allowed disabled:opacity-60 dark:border-linkflow-dark-primary dark:bg-linkflow-dark-panel/70 dark:text-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-soft"
          onClick={openCreateForm}
          disabled={!tenantId || products.length === 0}
        >
          <PlusOutlined aria-hidden="true" />
          {t('adminThingsModelCreate')}
        </button>
      </div>

      {formMode ? (
        <div
          className="fixed inset-0 z-50 grid place-items-center bg-slate-950/40 p-4"
          role="dialog"
          aria-modal="true"
          aria-labelledby="thingsmodel-form-title"
        >
          <section className="max-h-[calc(100vh-2rem)] w-full max-w-6xl overflow-y-auto rounded-lg border border-linkflow-border bg-white p-4 shadow-[0_18px_60px_rgba(15,23,42,0.24)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/50">
            <div className="mb-3 flex items-center justify-between gap-3">
              <h3 id="thingsmodel-form-title" className="m-0 text-lg font-bold">
                {formMode === 'create'
                  ? t('adminThingsModelCreateTitle')
                  : t('adminThingsModelEditTitle')}
              </h3>
              <button
                type="button"
                className="grid h-10 w-10 place-items-center rounded-md border border-linkflow-border bg-white text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
                aria-label={t('adminThingsModelCancel')}
                onClick={closeForm}
              >
                <CloseOutlined aria-hidden="true" />
              </button>
            </div>
            <form
              className="grid gap-3 md:grid-cols-2 xl:grid-cols-4"
              onSubmit={handleSubmit}
            >
              <label className="grid gap-1.5 text-sm font-bold">
                {t('adminProductName')}
                <select
                  className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary disabled:opacity-70 dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.product_id}
                  onChange={(event) =>
                    updateFormField('product_id', event.target.value)
                  }
                  disabled={formMode === 'edit'}
                  required
                >
                  {products.map((product) => (
                    <option key={product.id} value={product.id}>
                      {product.product_name}
                    </option>
                  ))}
                </select>
              </label>
              <label className="grid gap-1.5 text-sm font-bold">
                {t('adminThingsModelVersion')}
                <input
                  className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary disabled:opacity-70 dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  min={1}
                  type="number"
                  value={form.model_version}
                  onChange={(event) =>
                    updateFormField('model_version', event.target.value)
                  }
                  disabled={formMode === 'edit'}
                  required
                />
              </label>
              <label className="grid gap-1.5 text-sm font-bold">
                {t('adminThingsModelName')}
                <input
                  className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.model_name}
                  onChange={(event) =>
                    updateFormField('model_name', event.target.value)
                  }
                  required
                />
              </label>
              <label className="grid gap-1.5 text-sm font-bold">
                {t('adminThingsModelStatus')}
                <select
                  className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.status}
                  onChange={(event) =>
                    handleStatusChange(event.target.value as ThingsModelStatus)
                  }
                >
                  <option value="draft">
                    {t('adminThingsModelStatusDraft')}
                  </option>
                  <option value="published">
                    {t('adminThingsModelStatusPublished')}
                  </option>
                  <option value="deprecated">
                    {t('adminThingsModelStatusDeprecated')}
                  </option>
                </select>
              </label>
              <label className="grid gap-1.5 text-sm font-bold md:col-span-2 xl:col-span-3">
                {t('adminThingsModelDescription')}
                <textarea
                  className="min-h-16 rounded-md border border-linkflow-border bg-white px-3 py-2 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.description}
                  onChange={(event) =>
                    updateFormField('description', event.target.value)
                  }
                />
              </label>
              <label className="inline-flex h-10 w-fit self-end items-center gap-2 rounded-full border border-linkflow-border bg-slate-50 px-3 text-sm font-bold text-linkflow-muted dark:border-linkflow-dark-border dark:bg-slate-900/40 dark:text-linkflow-dark-muted">
                <input
                  className="h-4 w-4"
                  type="checkbox"
                  checked={form.is_current}
                  disabled={form.status !== 'published'}
                  onChange={(event) =>
                    updateFormField('is_current', event.target.checked)
                  }
                />
                {t('adminThingsModelCurrent')}
              </label>
              <section className="col-span-full grid gap-3">
                <div
                  className="flex gap-2 overflow-x-auto border-b border-linkflow-border dark:border-linkflow-dark-border"
                  role="tablist"
                  aria-label={t('adminThingsModelDefinitionTabs')}
                >
                  {definitionTabs.map((tab) => {
                    const active = activeDefinitionTab === tab.id;

                    return (
                      <button
                        key={tab.id}
                        type="button"
                        role="tab"
                        aria-selected={active}
                        className={[
                          'min-h-10 shrink-0 border-b-2 px-3 text-sm font-bold transition',
                          active
                            ? 'border-linkflow-primary text-linkflow-primary'
                            : 'border-transparent text-linkflow-muted hover:text-linkflow-primary dark:text-linkflow-dark-muted dark:hover:text-linkflow-dark-primary',
                        ].join(' ')}
                        onClick={() => setActiveDefinitionTab(tab.id)}
                      >
                        {t(tab.labelKey)}
                      </button>
                    );
                  })}
                </div>

                {activeDefinitionTab === 'properties' ? (
                  <div className="grid gap-3">
                    <p className="m-0 rounded-md border border-linkflow-primary/20 bg-linkflow-primary-soft px-3 py-1.5 text-sm leading-5 text-linkflow-primary">
                      {t('adminThingsModelPropertiesTip')}
                    </p>
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <DefinitionPresetBar
                        onApply={(template) =>
                          handleApplyPreset('properties', template)
                        }
                        presets={propertyPresets}
                        title={t('adminThingsModelCommonPresets')}
                      />
                      <button
                        type="button"
                        className="inline-flex h-10 items-center gap-2 rounded-md border border-linkflow-primary bg-white/70 px-3 text-sm font-bold text-linkflow-primary shadow-sm transition hover:bg-linkflow-primary-soft dark:border-linkflow-dark-primary dark:bg-linkflow-dark-panel/70 dark:text-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-soft"
                        onClick={handleAddProperty}
                      >
                        <PlusOutlined aria-hidden="true" />
                        {t('adminThingsModelAddProperty')}
                      </button>
                    </div>
                    <div className="grid gap-3">
                      {propertyRows.map((property, index) => (
                        <section
                          key={property.localId}
                          className="grid gap-3 rounded-lg border border-linkflow-primary/25 bg-white/75 p-3 shadow-sm ring-1 ring-linkflow-primary/10 dark:border-linkflow-dark-primary/25 dark:bg-linkflow-dark-page/40 dark:ring-linkflow-dark-primary/10"
                        >
                          <div className="flex flex-wrap items-center justify-between gap-2">
                            <h4 className="m-0 text-sm font-black text-linkflow-primary dark:text-linkflow-dark-primary">
                              {property.identifier.trim() ||
                                `${t('adminThingsModelPropertiesTab')} #${
                                  index + 1
                                }`}
                            </h4>
                            <button
                              type="button"
                              className="inline-flex h-8 items-center gap-1 rounded-md border border-rose-200 bg-white px-2 text-xs font-bold text-rose-600 transition hover:bg-rose-50 dark:border-rose-900/70 dark:bg-linkflow-dark-panel dark:text-rose-300 dark:hover:bg-rose-950/30"
                              onClick={() =>
                                handleRemoveProperty(property.localId)
                              }
                            >
                              <DeleteOutlined aria-hidden="true" />
                              {t('adminThingsModelRemoveDefinition')}
                            </button>
                          </div>
                          <div className="grid gap-3 [grid-template-columns:repeat(auto-fit,minmax(180px,1fr))]">
                            <label className="grid gap-1.5 text-sm font-bold">
                              {t('adminThingsModelIdentifier')}
                              <input
                                className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                                value={property.identifier}
                                onChange={(event) =>
                                  updatePropertyRow(
                                    property.localId,
                                    'identifier',
                                    event.target.value,
                                  )
                                }
                                placeholder="temperature"
                              />
                            </label>
                            <label className="grid gap-1.5 text-sm font-bold">
                              {t('adminThingsModelDisplayName')}
                              <input
                                className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                                value={property.name}
                                onChange={(event) =>
                                  updatePropertyRow(
                                    property.localId,
                                    'name',
                                    event.target.value,
                                  )
                                }
                                placeholder={t(
                                  'adminThingsModelPropertyNameExample',
                                )}
                              />
                            </label>
                            <label className="grid gap-1.5 text-sm font-bold">
                              {t('adminThingsModelDataType')}
                              <select
                                className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                                value={property.dataType}
                                onChange={(event) =>
                                  updatePropertyRow(
                                    property.localId,
                                    'dataType',
                                    event.target.value as PropertyDataType,
                                  )
                                }
                              >
                                <option value="string">string</option>
                                <option value="int">int</option>
                                <option value="float">float</option>
                                <option value="double">double</option>
                                <option value="bool">bool</option>
                              </select>
                            </label>
                            <label className="grid gap-1.5 text-sm font-bold">
                              {t('adminThingsModelAccessMode')}
                              <select
                                className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                                value={property.accessMode}
                                onChange={(event) =>
                                  updatePropertyRow(
                                    property.localId,
                                    'accessMode',
                                    event.target.value as PropertyAccessMode,
                                  )
                                }
                              >
                                <option value="read">
                                  {t('adminThingsModelAccessRead')}
                                </option>
                                <option value="write">
                                  {t('adminThingsModelAccessWrite')}
                                </option>
                                <option value="readwrite">
                                  {t('adminThingsModelAccessReadWrite')}
                                </option>
                              </select>
                            </label>
                            <label className="inline-flex h-10 w-fit self-end items-center gap-2 rounded-full border border-linkflow-border bg-slate-50 px-3 text-sm font-bold text-linkflow-muted dark:border-linkflow-dark-border dark:bg-slate-900/40 dark:text-linkflow-dark-muted">
                              <input
                                className="h-4 w-4"
                                type="checkbox"
                                checked={property.required}
                                onChange={(event) =>
                                  updatePropertyRow(
                                    property.localId,
                                    'required',
                                    event.target.checked,
                                  )
                                }
                              />
                              {t('adminThingsModelRequired')}
                            </label>
                            <label className="grid gap-1.5 text-sm font-bold">
                              {t('adminThingsModelSpecMin')}
                              <input
                                className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                                inputMode="decimal"
                                value={property.min}
                                onChange={(event) =>
                                  updatePropertyRow(
                                    property.localId,
                                    'min',
                                    event.target.value,
                                  )
                                }
                                placeholder="-40"
                              />
                            </label>
                            <label className="grid gap-1.5 text-sm font-bold">
                              {t('adminThingsModelSpecMax')}
                              <input
                                className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                                inputMode="decimal"
                                value={property.max}
                                onChange={(event) =>
                                  updatePropertyRow(
                                    property.localId,
                                    'max',
                                    event.target.value,
                                  )
                                }
                                placeholder="125"
                              />
                            </label>
                            <label className="grid gap-1.5 text-sm font-bold">
                              {t('adminThingsModelSpecStep')}
                              <input
                                className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                                inputMode="decimal"
                                value={property.step}
                                onChange={(event) =>
                                  updatePropertyRow(
                                    property.localId,
                                    'step',
                                    event.target.value,
                                  )
                                }
                                placeholder="0.1"
                              />
                            </label>
                            <label className="grid gap-1.5 text-sm font-bold">
                              {t('adminThingsModelSpecUnit')}
                              <input
                                className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                                value={property.unit}
                                onChange={(event) =>
                                  updatePropertyRow(
                                    property.localId,
                                    'unit',
                                    event.target.value,
                                  )
                                }
                                placeholder="celsius"
                              />
                            </label>
                            <label className="grid gap-1.5 text-sm font-bold">
                              {t('adminThingsModelSpecPrecision')}
                              <input
                                className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                                inputMode="numeric"
                                value={property.precision}
                                onChange={(event) =>
                                  updatePropertyRow(
                                    property.localId,
                                    'precision',
                                    event.target.value,
                                  )
                                }
                                placeholder="1"
                              />
                            </label>
                          </div>
                        </section>
                      ))}
                    </div>
                    <JSONPreviewField
                      label={t('adminThingsModelJSONPreview')}
                      value={propertyDefinitionsText}
                    />
                  </div>
                ) : null}

                {activeDefinitionTab === 'events' ? (
                  <div className="grid gap-3">
                    <p className="m-0 rounded-md border border-linkflow-primary/20 bg-linkflow-primary-soft px-3 py-1.5 text-sm leading-5 text-linkflow-primary">
                      {t('adminThingsModelEventsTip')}
                    </p>
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <DefinitionPresetBar
                        onApply={(template) =>
                          handleApplyPreset('events', template)
                        }
                        presets={eventPresets}
                        title={t('adminThingsModelCommonPresets')}
                      />
                      <button
                        type="button"
                        className="inline-flex h-10 items-center gap-2 rounded-md border border-linkflow-primary bg-white/70 px-3 text-sm font-bold text-linkflow-primary shadow-sm transition hover:bg-linkflow-primary-soft dark:border-linkflow-dark-primary dark:bg-linkflow-dark-panel/70 dark:text-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-soft"
                        onClick={handleAddEvent}
                      >
                        <PlusOutlined aria-hidden="true" />
                        {t('adminThingsModelAddEvent')}
                      </button>
                    </div>
                    <div className="grid gap-3">
                      {eventRows.map((eventRow, index) => (
                        <section
                          key={eventRow.localId}
                          className="grid gap-3 rounded-lg border border-linkflow-primary/25 bg-white/75 p-3 shadow-sm ring-1 ring-linkflow-primary/10 dark:border-linkflow-dark-primary/25 dark:bg-linkflow-dark-page/40 dark:ring-linkflow-dark-primary/10"
                        >
                          <div className="flex flex-wrap items-center justify-between gap-2">
                            <h4 className="m-0 text-sm font-black text-linkflow-primary dark:text-linkflow-dark-primary">
                              {eventRow.identifier.trim() ||
                                `${t('adminThingsModelEventsTab')} #${
                                  index + 1
                                }`}
                            </h4>
                            <button
                              type="button"
                              className="inline-flex h-8 items-center gap-1 rounded-md border border-rose-200 bg-white px-2 text-xs font-bold text-rose-600 transition hover:bg-rose-50 dark:border-rose-900/70 dark:bg-linkflow-dark-panel dark:text-rose-300 dark:hover:bg-rose-950/30"
                              onClick={() =>
                                handleRemoveEvent(eventRow.localId)
                              }
                            >
                              <DeleteOutlined aria-hidden="true" />
                              {t('adminThingsModelRemoveDefinition')}
                            </button>
                          </div>
                          <div className="grid gap-3 [grid-template-columns:repeat(auto-fit,minmax(180px,1fr))]">
                            <label className="grid gap-1.5 text-sm font-bold">
                              {t('adminThingsModelIdentifier')}
                              <input
                                className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                                value={eventRow.identifier}
                                onChange={(event) =>
                                  updateEventRow(
                                    eventRow.localId,
                                    'identifier',
                                    event.target.value,
                                  )
                                }
                                placeholder="temperature_alarm"
                              />
                            </label>
                            <label className="grid gap-1.5 text-sm font-bold">
                              {t('adminThingsModelDisplayName')}
                              <input
                                className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                                value={eventRow.name}
                                onChange={(event) =>
                                  updateEventRow(
                                    eventRow.localId,
                                    'name',
                                    event.target.value,
                                  )
                                }
                                placeholder={t(
                                  'adminThingsModelEventNameExample',
                                )}
                              />
                            </label>
                            <label className="grid gap-1.5 text-sm font-bold">
                              {t('adminThingsModelEventLevel')}
                              <select
                                className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                                value={eventRow.level}
                                onChange={(event) =>
                                  updateEventRow(
                                    eventRow.localId,
                                    'level',
                                    event.target.value as EventLevel,
                                  )
                                }
                              >
                                <option value="info">info</option>
                                <option value="warning">warning</option>
                                <option value="error">error</option>
                              </select>
                            </label>
                          </div>
                          <label className="grid gap-1.5 text-sm font-bold">
                            {t('adminThingsModelDescriptionField')}
                            <textarea
                              className="min-h-20 resize-y rounded-md border border-linkflow-border bg-white px-3 py-2 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                              value={eventRow.desc}
                              onChange={(event) =>
                                updateEventRow(
                                  eventRow.localId,
                                  'desc',
                                  event.target.value,
                                )
                              }
                              placeholder={t(
                                'adminThingsModelEventDescExample',
                              )}
                            />
                          </label>
                          <ParamDraftList
                            addLabel={t('adminThingsModelNewParam')}
                            onAddRow={() =>
                              addEventOutputParam(eventRow.localId)
                            }
                            onChangeRow={(localId, key, value) =>
                              updateEventOutputParam(
                                eventRow.localId,
                                localId,
                                key,
                                value,
                              )
                            }
                            onRemoveRow={(localId) =>
                              removeEventOutputParam(eventRow.localId, localId)
                            }
                            rows={eventRow.outputParams}
                            title={t('adminThingsModelEventOutput')}
                          />
                        </section>
                      ))}
                    </div>
                    <JSONPreviewField
                      label={t('adminThingsModelJSONPreview')}
                      value={eventDefinitionsText}
                    />
                  </div>
                ) : null}

                {activeDefinitionTab === 'services' ? (
                  <div className="grid gap-3">
                    <p className="m-0 rounded-md border border-linkflow-primary/20 bg-linkflow-primary-soft px-3 py-1.5 text-sm leading-5 text-linkflow-primary">
                      {t('adminThingsModelServicesTip')}
                    </p>
                    <div className="flex flex-wrap items-center justify-between gap-2">
                      <DefinitionPresetBar
                        onApply={(template) =>
                          handleApplyPreset('services', template)
                        }
                        presets={servicePresets}
                        title={t('adminThingsModelCommonPresets')}
                      />
                      <button
                        type="button"
                        className="inline-flex h-10 items-center gap-2 rounded-md border border-linkflow-primary bg-white/70 px-3 text-sm font-bold text-linkflow-primary shadow-sm transition hover:bg-linkflow-primary-soft dark:border-linkflow-dark-primary dark:bg-linkflow-dark-panel/70 dark:text-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-soft"
                        onClick={handleAddService}
                      >
                        <PlusOutlined aria-hidden="true" />
                        {t('adminThingsModelAddService')}
                      </button>
                    </div>
                    <div className="grid gap-3">
                      {serviceRows.map((service, index) => (
                        <section
                          key={service.localId}
                          className="grid gap-3 rounded-lg border border-linkflow-primary/25 bg-white/75 p-3 shadow-sm ring-1 ring-linkflow-primary/10 dark:border-linkflow-dark-primary/25 dark:bg-linkflow-dark-page/40 dark:ring-linkflow-dark-primary/10"
                        >
                          <div className="flex flex-wrap items-center justify-between gap-2">
                            <h4 className="m-0 text-sm font-black text-linkflow-primary dark:text-linkflow-dark-primary">
                              {service.identifier.trim() ||
                                `${t('adminThingsModelServicesTab')} #${
                                  index + 1
                                }`}
                            </h4>
                            <button
                              type="button"
                              className="inline-flex h-8 items-center gap-1 rounded-md border border-rose-200 bg-white px-2 text-xs font-bold text-rose-600 transition hover:bg-rose-50 dark:border-rose-900/70 dark:bg-linkflow-dark-panel dark:text-rose-300 dark:hover:bg-rose-950/30"
                              onClick={() =>
                                handleRemoveService(service.localId)
                              }
                            >
                              <DeleteOutlined aria-hidden="true" />
                              {t('adminThingsModelRemoveDefinition')}
                            </button>
                          </div>
                          <div className="grid gap-3 [grid-template-columns:repeat(auto-fit,minmax(180px,1fr))]">
                            <label className="grid gap-1.5 text-sm font-bold">
                              {t('adminThingsModelIdentifier')}
                              <input
                                className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                                value={service.identifier}
                                onChange={(event) =>
                                  updateServiceRow(
                                    service.localId,
                                    'identifier',
                                    event.target.value,
                                  )
                                }
                                placeholder="reboot"
                              />
                            </label>
                            <label className="grid gap-1.5 text-sm font-bold">
                              {t('adminThingsModelDisplayName')}
                              <input
                                className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                                value={service.name}
                                onChange={(event) =>
                                  updateServiceRow(
                                    service.localId,
                                    'name',
                                    event.target.value,
                                  )
                                }
                                placeholder={t(
                                  'adminThingsModelServiceNameExample',
                                )}
                              />
                            </label>
                            <label className="grid gap-1.5 text-sm font-bold">
                              {t('adminThingsModelCallType')}
                              <select
                                className="h-10 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                                value={service.callType}
                                onChange={(event) =>
                                  updateServiceRow(
                                    service.localId,
                                    'callType',
                                    event.target.value as ServiceCallType,
                                  )
                                }
                              >
                                <option value="sync">sync</option>
                                <option value="async">async</option>
                              </select>
                            </label>
                          </div>
                          <label className="grid gap-1.5 text-sm font-bold">
                            {t('adminThingsModelDescriptionField')}
                            <textarea
                              className="min-h-20 resize-y rounded-md border border-linkflow-border bg-white px-3 py-2 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                              value={service.desc}
                              onChange={(event) =>
                                updateServiceRow(
                                  service.localId,
                                  'desc',
                                  event.target.value,
                                )
                              }
                              placeholder={t(
                                'adminThingsModelServiceDescExample',
                              )}
                            />
                          </label>
                          <ParamDraftList
                            addLabel={t('adminThingsModelNewParam')}
                            onAddRow={() =>
                              addServiceInputParam(service.localId)
                            }
                            onChangeRow={(localId, key, value) =>
                              updateServiceInputParam(
                                service.localId,
                                localId,
                                key,
                                value,
                              )
                            }
                            onRemoveRow={(localId) =>
                              removeServiceInputParam(service.localId, localId)
                            }
                            rows={service.inputParams}
                            title={t('adminThingsModelServiceInput')}
                          />
                          <ParamDraftList
                            addLabel={t('adminThingsModelNewParam')}
                            onAddRow={() =>
                              addServiceOutputParam(service.localId)
                            }
                            onChangeRow={(localId, key, value) =>
                              updateServiceOutputParam(
                                service.localId,
                                localId,
                                key,
                                value,
                              )
                            }
                            onRemoveRow={(localId) =>
                              removeServiceOutputParam(service.localId, localId)
                            }
                            rows={service.outputParams}
                            title={t('adminThingsModelServiceOutput')}
                          />
                        </section>
                      ))}
                    </div>
                    <JSONPreviewField
                      label={t('adminThingsModelJSONPreview')}
                      value={serviceDefinitionsText}
                    />
                  </div>
                ) : null}
              </section>
              <div className="col-span-full flex items-end gap-2 pt-1">
                <button
                  type="submit"
                  className="inline-flex h-10 items-center gap-2 rounded-md bg-linkflow-primary px-4 text-sm font-bold text-white shadow-md shadow-linkflow-primary-soft transition hover:bg-linkflow-primary-hover disabled:cursor-not-allowed disabled:opacity-60 dark:bg-linkflow-dark-primary dark:text-linkflow-dark-page dark:hover:bg-linkflow-dark-primary-hover"
                  disabled={saving}
                >
                  <SaveOutlined aria-hidden="true" />
                  {saving
                    ? t('adminThingsModelSaving')
                    : t('adminThingsModelSave')}
                </button>
                <button
                  type="button"
                  className="h-10 rounded-md border border-linkflow-border bg-white px-4 text-sm font-bold text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
                  onClick={closeForm}
                >
                  {t('adminThingsModelCancel')}
                </button>
              </div>
            </form>
          </section>
        </div>
      ) : null}

      <AdminDataTable
        columns={columns}
        emptyDescription={t('adminThingsModelEmptyDescription')}
        emptyTitle={t('adminThingsModelEmpty')}
        getRowClassName={(model) =>
          model.is_current ? currentModelRowClassName : ''
        }
        getRowKey={(model) => model.id}
        items={models}
        mobileItems={mobileItems}
        pagination={{
          jumpLabel: t('adminThingsModelJump'),
          jumpToLabel: t('adminThingsModelJumpTo'),
          loading,
          loadingLabel: t('adminThingsModelLoading'),
          nextLabel: t('adminThingsModelNext'),
          onPageChange: setPage,
          onPageSizeChange: (nextPageSize) => {
            setPage(1);
            setPageSize(nextPageSize);
          },
          page,
          pageLabel: t('adminThingsModelPage')
            .replace('{page}', String(page))
            .replace('{totalPages}', String(totalPages)),
          pageSize,
          pageSizeLabel: t('adminThingsModelPageSize'),
          pageSizeOptions,
          prevLabel: t('adminThingsModelPrev'),
          total,
          totalLabel: t('adminThingsModelTotal').replace(
            '{total}',
            String(total),
          ),
        }}
      />
    </section>
  );
};

export default ThingsModelManagement;
