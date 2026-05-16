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
type PropertyDataType = 'string' | 'number' | 'boolean';
type PropertyAccessMode = 'read' | 'write' | 'readwrite';
type EventLevel = 'info' | 'warning' | 'error';
type ServiceCallType = 'sync' | 'async';

interface PropertyDraft {
  identifier: string;
  name: string;
  dataType: PropertyDataType;
  accessMode: PropertyAccessMode;
  required: boolean;
}

interface EventDraft {
  identifier: string;
  name: string;
  level: EventLevel;
}

interface ServiceDraft {
  identifier: string;
  name: string;
  callType: ServiceCallType;
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

const emptyPropertyDraft: PropertyDraft = {
  identifier: '',
  name: '',
  dataType: 'string',
  accessMode: 'readwrite',
  required: false,
};

const emptyEventDraft: EventDraft = {
  identifier: '',
  name: '',
  level: 'info',
};

const emptyServiceDraft: ServiceDraft = {
  identifier: '',
  name: '',
  callType: 'sync',
};

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

const objectKeyCount = (value: ThingsModelObject) => {
  return Object.keys(value ?? {}).length;
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
  const [propertyDraft, setPropertyDraft] =
    useState<PropertyDraft>(emptyPropertyDraft);
  const [eventDraft, setEventDraft] = useState<EventDraft>(emptyEventDraft);
  const [serviceDraft, setServiceDraft] =
    useState<ServiceDraft>(emptyServiceDraft);

  const totalPages = Math.max(1, Math.ceil(total / pageSize));
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

  const updatePropertyDraft = <K extends keyof PropertyDraft>(
    key: K,
    value: PropertyDraft[K],
  ) => {
    setPropertyDraft((current) => ({ ...current, [key]: value }));
  };

  const updateEventDraft = <K extends keyof EventDraft>(
    key: K,
    value: EventDraft[K],
  ) => {
    setEventDraft((current) => ({ ...current, [key]: value }));
  };

  const updateServiceDraft = <K extends keyof ServiceDraft>(
    key: K,
    value: ServiceDraft[K],
  ) => {
    setServiceDraft((current) => ({ ...current, [key]: value }));
  };

  const openCreateForm = () => {
    setEditingModel(null);
    setActiveDefinitionTab('properties');
    setPropertyDraft(emptyPropertyDraft);
    setEventDraft(emptyEventDraft);
    setServiceDraft(emptyServiceDraft);
    setForm({
      ...emptyForm,
      product_id: productFilter || products[0]?.id || '',
    });
    setFormMode('create');
  };

  const openEditForm = (model: ThingsModel) => {
    setEditingModel(model);
    setActiveDefinitionTab('properties');
    setPropertyDraft(emptyPropertyDraft);
    setEventDraft(emptyEventDraft);
    setServiceDraft(emptyServiceDraft);
    setForm(toFormState(model));
    setFormMode('edit');
  };

  const closeForm = () => {
    setFormMode(null);
    setEditingModel(null);
    setActiveDefinitionTab('properties');
    setPropertyDraft(emptyPropertyDraft);
    setEventDraft(emptyEventDraft);
    setServiceDraft(emptyServiceDraft);
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

  const handleAddProperty = () => {
    const identifier = propertyDraft.identifier.trim();
    const name = propertyDraft.name.trim();
    if (!identifier || !name) {
      notification.warning({
        message: t('adminThingsModelDefinitionRequired'),
        placement: 'topRight',
      });
      return;
    }

    try {
      updateFormField(
        'properties',
        setObjectEntry(form.properties, identifier, {
          name,
          data_type: propertyDraft.dataType,
          access_mode: propertyDraft.accessMode,
          required: propertyDraft.required,
        }),
      );
      setPropertyDraft(emptyPropertyDraft);
    } catch (error) {
      notification.error({
        message: t('adminThingsModelJSONInvalid'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    }
  };

  const handleAddEvent = () => {
    const identifier = eventDraft.identifier.trim();
    const name = eventDraft.name.trim();
    if (!identifier || !name) {
      notification.warning({
        message: t('adminThingsModelDefinitionRequired'),
        placement: 'topRight',
      });
      return;
    }

    try {
      updateFormField(
        'events',
        setObjectEntry(form.events, identifier, {
          name,
          level: eventDraft.level,
          output: {},
        }),
      );
      setEventDraft(emptyEventDraft);
    } catch (error) {
      notification.error({
        message: t('adminThingsModelJSONInvalid'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    }
  };

  const handleAddService = () => {
    const identifier = serviceDraft.identifier.trim();
    const name = serviceDraft.name.trim();
    if (!identifier || !name) {
      notification.warning({
        message: t('adminThingsModelDefinitionRequired'),
        placement: 'topRight',
      });
      return;
    }

    try {
      updateFormField(
        'services',
        setObjectEntry(form.services, identifier, {
          name,
          call_type: serviceDraft.callType,
          input: {},
          output: {},
        }),
      );
      setServiceDraft(emptyServiceDraft);
    } catch (error) {
      notification.error({
        message: t('adminThingsModelJSONInvalid'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    }
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!tenantId) {
      return;
    }

    setSaving(true);
    try {
      const properties = parseObjectField(form.properties);
      const events = parseObjectField(form.events);
      const services = parseObjectField(form.services);

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
        model.is_current
          ? t('adminThingsModelCurrentYes')
          : t('adminThingsModelCurrentNo'),
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
      className="rounded-lg border border-linkflow-border p-4 dark:border-linkflow-dark-border"
    >
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
          <section className="max-h-[calc(100vh-2rem)] w-full max-w-5xl overflow-y-auto rounded-lg border border-linkflow-border bg-white p-5 shadow-[0_18px_60px_rgba(15,23,42,0.24)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/50">
            <div className="mb-4 flex items-center justify-between gap-3">
              <h3 id="thingsmodel-form-title" className="m-0 text-lg font-bold">
                {formMode === 'create'
                  ? t('adminThingsModelCreateTitle')
                  : t('adminThingsModelEditTitle')}
              </h3>
              <button
                type="button"
                className="grid h-11 w-11 place-items-center rounded-md border border-linkflow-border bg-white text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
                aria-label={t('adminThingsModelCancel')}
                onClick={closeForm}
              >
                <CloseOutlined aria-hidden="true" />
              </button>
            </div>
            <form
              className="grid gap-4 [grid-template-columns:repeat(auto-fit,minmax(220px,1fr))]"
              onSubmit={handleSubmit}
            >
              <label className="grid gap-2 text-sm font-bold">
                {t('adminProductName')}
                <select
                  className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary disabled:opacity-70 dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
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
              <label className="grid gap-2 text-sm font-bold">
                {t('adminThingsModelVersion')}
                <input
                  className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary disabled:opacity-70 dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
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
              <label className="grid gap-2 text-sm font-bold">
                {t('adminThingsModelName')}
                <input
                  className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.model_name}
                  onChange={(event) =>
                    updateFormField('model_name', event.target.value)
                  }
                  required
                />
              </label>
              <label className="grid gap-2 text-sm font-bold">
                {t('adminThingsModelStatus')}
                <select
                  className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
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
              <label className="flex min-h-11 items-center gap-3 rounded-md border border-linkflow-border bg-white px-3 text-sm font-bold dark:border-linkflow-dark-border dark:bg-linkflow-dark-page">
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
              <label className="grid gap-2 text-sm font-bold md:col-span-2">
                {t('adminThingsModelDescription')}
                <textarea
                  className="min-h-24 rounded-md border border-linkflow-border bg-white px-3 py-2 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.description}
                  onChange={(event) =>
                    updateFormField('description', event.target.value)
                  }
                />
              </label>
              <section className="grid gap-4 md:col-span-3">
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
                          'min-h-11 shrink-0 border-b-2 px-3 text-sm font-bold transition',
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
                  <div className="grid gap-4">
                    <p className="m-0 rounded-md border border-linkflow-primary/20 bg-linkflow-primary-soft px-3 py-2 text-sm leading-6 text-linkflow-primary">
                      {t('adminThingsModelPropertiesTip')}
                    </p>
                    <div className="grid gap-3 [grid-template-columns:repeat(auto-fit,minmax(180px,1fr))]">
                      <label className="grid gap-2 text-sm font-bold">
                        {t('adminThingsModelIdentifier')}
                        <input
                          className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                          value={propertyDraft.identifier}
                          onChange={(event) =>
                            updatePropertyDraft(
                              'identifier',
                              event.target.value,
                            )
                          }
                          placeholder="temperature"
                        />
                      </label>
                      <label className="grid gap-2 text-sm font-bold">
                        {t('adminThingsModelDisplayName')}
                        <input
                          className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                          value={propertyDraft.name}
                          onChange={(event) =>
                            updatePropertyDraft('name', event.target.value)
                          }
                          placeholder={t('adminThingsModelPropertyNameExample')}
                        />
                      </label>
                      <label className="grid gap-2 text-sm font-bold">
                        {t('adminThingsModelDataType')}
                        <select
                          className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                          value={propertyDraft.dataType}
                          onChange={(event) =>
                            updatePropertyDraft(
                              'dataType',
                              event.target.value as PropertyDataType,
                            )
                          }
                        >
                          <option value="string">string</option>
                          <option value="number">number</option>
                          <option value="boolean">boolean</option>
                        </select>
                      </label>
                      <label className="grid gap-2 text-sm font-bold">
                        {t('adminThingsModelAccessMode')}
                        <select
                          className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                          value={propertyDraft.accessMode}
                          onChange={(event) =>
                            updatePropertyDraft(
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
                      <label className="flex min-h-11 items-center gap-3 rounded-md border border-linkflow-border bg-white px-3 text-sm font-bold dark:border-linkflow-dark-border dark:bg-linkflow-dark-page">
                        <input
                          className="h-4 w-4"
                          type="checkbox"
                          checked={propertyDraft.required}
                          onChange={(event) =>
                            updatePropertyDraft(
                              'required',
                              event.target.checked,
                            )
                          }
                        />
                        {t('adminThingsModelRequired')}
                      </label>
                      <div className="flex items-end">
                        <button
                          type="button"
                          className="inline-flex h-11 items-center gap-2 rounded-md border border-linkflow-primary bg-white/70 px-3 text-sm font-bold text-linkflow-primary shadow-sm transition hover:bg-linkflow-primary-soft dark:border-linkflow-dark-primary dark:bg-linkflow-dark-panel/70 dark:text-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-soft"
                          onClick={handleAddProperty}
                        >
                          <PlusOutlined aria-hidden="true" />
                          {t('adminThingsModelAddDefinition')}
                        </button>
                      </div>
                    </div>
                    <label className="grid gap-2 text-sm font-bold">
                      {t('adminThingsModelJSONPreview')}
                      <textarea
                        className="min-h-44 rounded-md border border-linkflow-border bg-white px-3 py-2 font-mono text-xs outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                        value={form.properties}
                        onChange={(event) =>
                          updateFormField('properties', event.target.value)
                        }
                        required
                      />
                    </label>
                  </div>
                ) : null}

                {activeDefinitionTab === 'events' ? (
                  <div className="grid gap-4">
                    <p className="m-0 rounded-md border border-linkflow-primary/20 bg-linkflow-primary-soft px-3 py-2 text-sm leading-6 text-linkflow-primary">
                      {t('adminThingsModelEventsTip')}
                    </p>
                    <div className="grid gap-3 [grid-template-columns:repeat(auto-fit,minmax(180px,1fr))]">
                      <label className="grid gap-2 text-sm font-bold">
                        {t('adminThingsModelIdentifier')}
                        <input
                          className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                          value={eventDraft.identifier}
                          onChange={(event) =>
                            updateEventDraft('identifier', event.target.value)
                          }
                          placeholder="overheated"
                        />
                      </label>
                      <label className="grid gap-2 text-sm font-bold">
                        {t('adminThingsModelDisplayName')}
                        <input
                          className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                          value={eventDraft.name}
                          onChange={(event) =>
                            updateEventDraft('name', event.target.value)
                          }
                          placeholder={t('adminThingsModelEventNameExample')}
                        />
                      </label>
                      <label className="grid gap-2 text-sm font-bold">
                        {t('adminThingsModelEventLevel')}
                        <select
                          className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                          value={eventDraft.level}
                          onChange={(event) =>
                            updateEventDraft(
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
                      <div className="flex items-end">
                        <button
                          type="button"
                          className="inline-flex h-11 items-center gap-2 rounded-md border border-linkflow-primary bg-white/70 px-3 text-sm font-bold text-linkflow-primary shadow-sm transition hover:bg-linkflow-primary-soft dark:border-linkflow-dark-primary dark:bg-linkflow-dark-panel/70 dark:text-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-soft"
                          onClick={handleAddEvent}
                        >
                          <PlusOutlined aria-hidden="true" />
                          {t('adminThingsModelAddDefinition')}
                        </button>
                      </div>
                    </div>
                    <label className="grid gap-2 text-sm font-bold">
                      {t('adminThingsModelJSONPreview')}
                      <textarea
                        className="min-h-44 rounded-md border border-linkflow-border bg-white px-3 py-2 font-mono text-xs outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                        value={form.events}
                        onChange={(event) =>
                          updateFormField('events', event.target.value)
                        }
                        required
                      />
                    </label>
                  </div>
                ) : null}

                {activeDefinitionTab === 'services' ? (
                  <div className="grid gap-4">
                    <p className="m-0 rounded-md border border-linkflow-primary/20 bg-linkflow-primary-soft px-3 py-2 text-sm leading-6 text-linkflow-primary">
                      {t('adminThingsModelServicesTip')}
                    </p>
                    <div className="grid gap-3 [grid-template-columns:repeat(auto-fit,minmax(180px,1fr))]">
                      <label className="grid gap-2 text-sm font-bold">
                        {t('adminThingsModelIdentifier')}
                        <input
                          className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                          value={serviceDraft.identifier}
                          onChange={(event) =>
                            updateServiceDraft('identifier', event.target.value)
                          }
                          placeholder="reboot"
                        />
                      </label>
                      <label className="grid gap-2 text-sm font-bold">
                        {t('adminThingsModelDisplayName')}
                        <input
                          className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                          value={serviceDraft.name}
                          onChange={(event) =>
                            updateServiceDraft('name', event.target.value)
                          }
                          placeholder={t('adminThingsModelServiceNameExample')}
                        />
                      </label>
                      <label className="grid gap-2 text-sm font-bold">
                        {t('adminThingsModelCallType')}
                        <select
                          className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                          value={serviceDraft.callType}
                          onChange={(event) =>
                            updateServiceDraft(
                              'callType',
                              event.target.value as ServiceCallType,
                            )
                          }
                        >
                          <option value="sync">sync</option>
                          <option value="async">async</option>
                        </select>
                      </label>
                      <div className="flex items-end">
                        <button
                          type="button"
                          className="inline-flex h-11 items-center gap-2 rounded-md border border-linkflow-primary bg-white/70 px-3 text-sm font-bold text-linkflow-primary shadow-sm transition hover:bg-linkflow-primary-soft dark:border-linkflow-dark-primary dark:bg-linkflow-dark-panel/70 dark:text-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-soft"
                          onClick={handleAddService}
                        >
                          <PlusOutlined aria-hidden="true" />
                          {t('adminThingsModelAddDefinition')}
                        </button>
                      </div>
                    </div>
                    <label className="grid gap-2 text-sm font-bold">
                      {t('adminThingsModelJSONPreview')}
                      <textarea
                        className="min-h-44 rounded-md border border-linkflow-border bg-white px-3 py-2 font-mono text-xs outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                        value={form.services}
                        onChange={(event) =>
                          updateFormField('services', event.target.value)
                        }
                        required
                      />
                    </label>
                  </div>
                ) : null}
              </section>
              <div className="flex items-end gap-2 md:col-span-3">
                <button
                  type="submit"
                  className="inline-flex h-11 items-center gap-2 rounded-md bg-linkflow-primary px-4 text-sm font-bold text-white shadow-md shadow-linkflow-primary-soft transition hover:bg-linkflow-primary-hover disabled:cursor-not-allowed disabled:opacity-60 dark:bg-linkflow-dark-primary dark:text-linkflow-dark-page dark:hover:bg-linkflow-dark-primary-hover"
                  disabled={saving}
                >
                  <SaveOutlined aria-hidden="true" />
                  {saving
                    ? t('adminThingsModelSaving')
                    : t('adminThingsModelSave')}
                </button>
                <button
                  type="button"
                  className="h-11 rounded-md border border-linkflow-border bg-white px-4 text-sm font-bold text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
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
