import { useI18n } from '@/contexts/I18nContext';
import type { AdminMessageKey } from '@/i18n/admin';
import {
  createDevice,
  deleteDevice,
  listDevices,
  updateDevice,
  type Device,
  type DeviceConnectionStatus,
  type DeviceStatus,
} from '@/services/devices';
import { listProducts, type Product } from '@/services/products';
import { listTenants, type Tenant } from '@/services/tenants';
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

interface DeviceFormState {
  product_id: string;
  device_slug: string;
  device_name: string;
  description: string;
  status: DeviceStatus;
  gateway_device_id: string;
}

type FormMode = 'create' | 'edit';

const emptyForm: DeviceFormState = {
  product_id: '',
  device_slug: '',
  device_name: '',
  description: '',
  status: 'active',
  gateway_device_id: '',
};

const pageSizeOptions = [10, 20, 50];

const deviceStatusLabelKey: Record<DeviceStatus, AdminMessageKey> = {
  active: 'adminDeviceStatusActive',
  disabled: 'adminDeviceStatusDisabled',
};

const connectionLabelKey: Record<DeviceConnectionStatus, AdminMessageKey> = {
  online: 'adminDeviceConnectionOnline',
  offline: 'adminDeviceConnectionOffline',
};

const statusBadgeClassName = (status: DeviceStatus) => {
  return [
    'inline-flex rounded-full px-2.5 py-1 text-xs font-bold',
    status === 'active'
      ? 'bg-linkflow-primary-soft text-linkflow-primary'
      : 'bg-slate-100 text-linkflow-muted dark:bg-slate-800 dark:text-linkflow-dark-muted',
  ].join(' ');
};

const connectionBadgeClassName = (status: DeviceConnectionStatus) => {
  return [
    'inline-flex rounded-full px-2.5 py-1 text-xs font-bold',
    status === 'online'
      ? 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300'
      : 'bg-slate-100 text-linkflow-muted dark:bg-slate-800 dark:text-linkflow-dark-muted',
  ].join(' ');
};

const toFormState = (device: Device): DeviceFormState => ({
  product_id: device.product_id,
  device_slug: device.device_slug,
  device_name: device.device_name,
  description: device.description,
  status: device.status,
  gateway_device_id: device.gateway_device_id ?? '',
});

const DeviceManagement = () => {
  const { t } = useI18n();
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [tenantId, setTenantId] = useState('');
  const [products, setProducts] = useState<Product[]>([]);
  const [productFilter, setProductFilter] = useState('');
  const [devices, setDevices] = useState<Device[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [formMode, setFormMode] = useState<FormMode | null>(null);
  const [editingDevice, setEditingDevice] = useState<Device | null>(null);
  const [form, setForm] = useState<DeviceFormState>(emptyForm);
  const [createdSecret, setCreatedSecret] = useState('');

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
        setProductFilter((current) =>
          current && result.items.some((product) => product.id === current)
            ? current
            : '',
        );
        setForm((current) => ({
          ...current,
          product_id:
            current.product_id &&
            result.items.some((product) => product.id === current.product_id)
              ? current.product_id
              : result.items[0]?.id || '',
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

  const loadDevices = async (nextPage = page, nextPageSize = pageSize) => {
    if (!tenantId) {
      setDevices([]);
      setTotal(0);
      return;
    }

    setLoading(true);
    try {
      const result = await listDevices({
        tenant_id: tenantId,
        product_id: productFilter || undefined,
        page: nextPage,
        page_size: nextPageSize,
      });
      setDevices(result.items);
      setTotal(result.total);
      setPage(result.page);
      setPageSize(result.page_size);
    } catch (error) {
      notification.error({
        message: t('adminDeviceLoadFailed'),
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
        setDevices([]);
        setTotal(0);
        return;
      }

      setLoading(true);
      try {
        const result = await listDevices({
          tenant_id: tenantId,
          product_id: productFilter || undefined,
          page,
          page_size: pageSize,
        });
        if (cancelled) {
          return;
        }
        setDevices(result.items);
        setTotal(result.total);
        setPage(result.page);
        setPageSize(result.page_size);
      } catch (error) {
        if (cancelled) {
          return;
        }
        notification.error({
          message: t('adminDeviceLoadFailed'),
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

  const updateFormField = <K extends keyof DeviceFormState>(
    key: K,
    value: DeviceFormState[K],
  ) => {
    setForm((current) => ({ ...current, [key]: value }));
  };

  const openCreateForm = () => {
    setEditingDevice(null);
    setCreatedSecret('');
    setForm({
      ...emptyForm,
      product_id: productFilter || products[0]?.id || '',
    });
    setFormMode('create');
  };

  const openEditForm = (device: Device) => {
    setEditingDevice(device);
    setCreatedSecret('');
    setForm(toFormState(device));
    setFormMode('edit');
  };

  const closeForm = () => {
    setFormMode(null);
    setEditingDevice(null);
    setForm(emptyForm);
  };

  const handleTenantChange = (nextTenantId: string) => {
    setTenantId(nextTenantId);
    setProductFilter('');
    setPage(1);
    closeForm();
  };

  const handleProductFilterChange = (nextProductId: string) => {
    setProductFilter(nextProductId);
    setPage(1);
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!tenantId) {
      return;
    }

    setSaving(true);
    try {
      if (formMode === 'create') {
        const result = await createDevice({
          tenant_id: tenantId,
          product_id: form.product_id,
          device_slug: form.device_slug.trim(),
          device_name: form.device_name.trim(),
          description: form.description.trim(),
          gateway_device_id: form.gateway_device_id.trim(),
        });
        notification.success({
          message: t('adminDeviceCreateSuccess'),
          placement: 'topRight',
        });
        setCreatedSecret(result.device_secret ?? '');
        setPage(1);
      }

      if (formMode === 'edit' && editingDevice) {
        await updateDevice(editingDevice.id, {
          device_name: form.device_name.trim(),
          description: form.description.trim(),
          status: form.status,
          gateway_device_id: form.gateway_device_id.trim(),
        });
        notification.success({
          message: t('adminDeviceUpdateSuccess'),
          placement: 'topRight',
        });
        closeForm();
      }

      await loadDevices(formMode === 'create' ? 1 : page, pageSize);
      if (formMode === 'create') {
        setForm((current) => ({
          ...current,
          device_slug: '',
          device_name: '',
          description: '',
        }));
      }
    } catch (error) {
      notification.error({
        message:
          formMode === 'create'
            ? t('adminDeviceCreateFailed')
            : t('adminDeviceUpdateFailed'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (device: Device) => {
    if (!window.confirm(t('adminDeviceDeleteConfirm'))) {
      return;
    }

    setDeletingId(device.id);
    try {
      await deleteDevice(device.id);
      notification.success({
        message: t('adminDeviceDeleteSuccess'),
        placement: 'topRight',
      });
      if (devices.length === 1 && page > 1) {
        setPage((current) => current - 1);
      } else {
        await loadDevices();
      }
    } catch (error) {
      notification.error({
        message: t('adminDeviceDeleteFailed'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    } finally {
      setDeletingId(null);
    }
  };

  const columns: AdminDataTableColumn<Device>[] = [
    {
      key: 'name',
      title: t('adminDeviceName'),
      render: (device) => (
        <div>
          <strong className="block text-sm">{device.device_name}</strong>
          <span className="mt-1 block text-xs font-semibold text-linkflow-primary">
            {device.device_slug}
          </span>
        </div>
      ),
    },
    {
      key: 'product',
      title: t('adminProductName'),
      className: 'text-sm font-semibold',
      render: (device) => productNameById.get(device.product_id) ?? '--',
    },
    {
      key: 'status',
      title: t('adminDeviceStatus'),
      render: (device) => (
        <span className={statusBadgeClassName(device.status)}>
          {t(deviceStatusLabelKey[device.status])}
        </span>
      ),
    },
    {
      key: 'connection',
      title: t('adminDeviceConnectionStatus'),
      render: (device) => (
        <span className={connectionBadgeClassName(device.connection_status)}>
          {t(connectionLabelKey[device.connection_status])}
        </span>
      ),
    },
    {
      key: 'firmware',
      title: t('adminDeviceFirmwareVersion'),
      className: 'text-sm text-linkflow-subtle dark:text-linkflow-dark-subtle',
      render: (device) => device.firmware_version || '--',
    },
    {
      key: 'updatedAt',
      title: t('adminDeviceUpdatedAt'),
      className: 'text-sm text-linkflow-subtle dark:text-linkflow-dark-subtle',
      render: (device) => formatDateTime(device.updated_at),
    },
    {
      key: 'actions',
      title: t('adminDeviceActions'),
      headerClassName: 'text-right',
      render: (device) => (
        <div className="flex justify-end gap-2">
          <button
            type="button"
            className="grid h-10 w-10 place-items-center rounded-md border border-linkflow-border bg-white text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
            aria-label={t('adminDeviceEdit')}
            onClick={() => openEditForm(device)}
          >
            <EditOutlined aria-hidden="true" />
          </button>
          <button
            type="button"
            className="grid h-10 w-10 place-items-center rounded-md border border-linkflow-border bg-white text-red-600 transition hover:border-red-500 hover:bg-red-50 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:hover:border-red-400 dark:hover:bg-red-950/30"
            aria-label={t('adminDeviceDelete')}
            onClick={() => handleDelete(device)}
            disabled={deletingId === device.id}
          >
            <DeleteOutlined aria-hidden="true" />
          </button>
        </div>
      ),
    },
  ];

  const mobileItems = devices.map((device) => (
    <article
      key={device.id}
      className="rounded-lg border border-linkflow-border p-4 dark:border-linkflow-dark-border"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h4 className="m-0 truncate text-base font-bold">
            {device.device_name}
          </h4>
          <p className="m-0 mt-1 truncate text-sm font-semibold text-linkflow-primary">
            {device.device_slug}
          </p>
        </div>
        <span className={connectionBadgeClassName(device.connection_status)}>
          {t(connectionLabelKey[device.connection_status])}
        </span>
      </div>
      <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminProductName')}
          </dt>
          <dd className="m-0 mt-1">
            {productNameById.get(device.product_id) ?? '--'}
          </dd>
        </div>
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDeviceStatus')}
          </dt>
          <dd className="m-0 mt-1">
            <span className={statusBadgeClassName(device.status)}>
              {t(deviceStatusLabelKey[device.status])}
            </span>
          </dd>
        </div>
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDeviceFirmwareVersion')}
          </dt>
          <dd className="m-0 mt-1">{device.firmware_version || '--'}</dd>
        </div>
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminDeviceUpdatedAt')}
          </dt>
          <dd className="m-0 mt-1">{formatDateTime(device.updated_at)}</dd>
        </div>
      </dl>
      <div className="mt-4 flex gap-2">
        <button
          type="button"
          className="inline-flex h-11 flex-1 items-center justify-center gap-2 rounded-md border border-linkflow-border bg-white text-sm font-bold text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
          onClick={() => openEditForm(device)}
        >
          <EditOutlined aria-hidden="true" />
          {t('adminDeviceEdit')}
        </button>
        <button
          type="button"
          className="inline-flex h-11 flex-1 items-center justify-center gap-2 rounded-md border border-linkflow-border bg-white text-sm font-bold text-red-600 transition hover:border-red-500 hover:bg-red-50 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:hover:border-red-400 dark:hover:bg-red-950/30"
          onClick={() => handleDelete(device)}
          disabled={deletingId === device.id}
        >
          <DeleteOutlined aria-hidden="true" />
          {t('adminDeviceDelete')}
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
        </div>
        <button
          type="button"
          className="inline-flex h-11 items-center gap-2 rounded-md border border-linkflow-primary bg-white/70 px-3 text-sm font-bold text-linkflow-primary shadow-sm transition hover:bg-linkflow-primary-soft disabled:cursor-not-allowed disabled:opacity-60 dark:border-linkflow-dark-primary dark:bg-linkflow-dark-panel/70 dark:text-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-soft"
          onClick={openCreateForm}
          disabled={!tenantId || products.length === 0}
        >
          <PlusOutlined aria-hidden="true" />
          {t('adminDeviceCreate')}
        </button>
      </div>

      {formMode ? (
        <div
          className="fixed inset-0 z-50 grid place-items-center bg-slate-950/40 p-4"
          role="dialog"
          aria-modal="true"
          aria-labelledby="device-form-title"
        >
          <section className="max-h-[calc(100vh-2rem)] w-full max-w-4xl overflow-y-auto rounded-lg border border-linkflow-border bg-white p-5 shadow-[0_18px_60px_rgba(15,23,42,0.24)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/50">
            <div className="mb-4 flex items-center justify-between gap-3">
              <h3 id="device-form-title" className="m-0 text-lg font-bold">
                {formMode === 'create'
                  ? t('adminDeviceCreateTitle')
                  : t('adminDeviceEditTitle')}
              </h3>
              <button
                type="button"
                className="grid h-11 w-11 place-items-center rounded-md border border-linkflow-border bg-white text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
                aria-label={t('adminDeviceCancel')}
                onClick={closeForm}
              >
                <CloseOutlined aria-hidden="true" />
              </button>
            </div>

            {createdSecret ? (
              <section className="mb-4 rounded-lg border border-linkflow-primary bg-linkflow-primary-soft p-4 text-linkflow-primary">
                <h4 className="m-0 text-sm font-bold">
                  {t('adminDeviceSecretTitle')}
                </h4>
                <p className="m-0 mt-2 text-sm leading-6">
                  {t('adminDeviceSecretDescription')}
                </p>
                <code className="mt-3 block overflow-x-auto rounded-md bg-white p-3 text-sm font-bold text-linkflow-text dark:bg-linkflow-dark-page dark:text-linkflow-dark-text">
                  {createdSecret}
                </code>
              </section>
            ) : null}

            <form
              className="grid gap-4 [grid-template-columns:repeat(auto-fit,minmax(220px,1fr))]"
              onSubmit={handleSubmit}
            >
              <label className="grid gap-2 text-sm font-bold">
                {t('adminProductName')}
                <select
                  className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
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
                {t('adminDeviceSlug')}
                <input
                  className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.device_slug}
                  onChange={(event) =>
                    updateFormField('device_slug', event.target.value)
                  }
                  disabled={formMode === 'edit'}
                  required
                />
              </label>
              <label className="grid gap-2 text-sm font-bold">
                {t('adminDeviceName')}
                <input
                  className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.device_name}
                  onChange={(event) =>
                    updateFormField('device_name', event.target.value)
                  }
                  required
                />
              </label>
              {formMode === 'edit' ? (
                <label className="grid gap-2 text-sm font-bold">
                  {t('adminDeviceStatus')}
                  <select
                    className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                    value={form.status}
                    onChange={(event) =>
                      updateFormField(
                        'status',
                        event.target.value as DeviceStatus,
                      )
                    }
                  >
                    <option value="active">
                      {t('adminDeviceStatusActive')}
                    </option>
                    <option value="disabled">
                      {t('adminDeviceStatusDisabled')}
                    </option>
                  </select>
                </label>
              ) : null}
              <label className="grid gap-2 text-sm font-bold">
                {t('adminDeviceGatewayID')}
                <input
                  className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.gateway_device_id}
                  onChange={(event) =>
                    updateFormField('gateway_device_id', event.target.value)
                  }
                />
              </label>
              <label className="grid gap-2 text-sm font-bold md:col-span-2">
                {t('adminDeviceDescription')}
                <textarea
                  className="min-h-24 rounded-md border border-linkflow-border bg-white px-3 py-2 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.description}
                  onChange={(event) =>
                    updateFormField('description', event.target.value)
                  }
                />
              </label>
              <div className="flex items-end gap-2 md:col-span-2">
                <button
                  type="submit"
                  className="inline-flex h-11 items-center gap-2 rounded-md bg-linkflow-primary px-4 text-sm font-bold text-white shadow-md shadow-linkflow-primary-soft transition hover:bg-linkflow-primary-hover disabled:cursor-not-allowed disabled:opacity-60 dark:bg-linkflow-dark-primary dark:text-linkflow-dark-page dark:hover:bg-linkflow-dark-primary-hover"
                  disabled={saving || form.product_id === ''}
                >
                  <SaveOutlined aria-hidden="true" />
                  {saving ? t('adminDeviceSaving') : t('adminDeviceSave')}
                </button>
                <button
                  type="button"
                  className="h-11 rounded-md border border-linkflow-border bg-white px-4 text-sm font-bold text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
                  onClick={closeForm}
                >
                  {t('adminDeviceCancel')}
                </button>
              </div>
            </form>
          </section>
        </div>
      ) : null}

      <AdminDataTable
        columns={columns}
        emptyDescription={t('adminDeviceEmptyDescription')}
        emptyTitle={t('adminDeviceEmpty')}
        getRowKey={(device) => device.id}
        items={devices}
        minWidthClassName="min-w-[980px]"
        mobileItems={mobileItems}
        pagination={{
          jumpLabel: t('adminDeviceJump'),
          jumpToLabel: t('adminDeviceJumpTo'),
          loading,
          loadingLabel: t('adminDeviceLoading'),
          nextLabel: t('adminDeviceNext'),
          onPageChange: setPage,
          onPageSizeChange: (nextPageSize) => {
            setPage(1);
            setPageSize(nextPageSize);
          },
          page,
          pageLabel: t('adminDevicePage')
            .replace('{page}', String(page))
            .replace('{totalPages}', String(totalPages)),
          pageSize,
          pageSizeLabel: t('adminDevicePageSize'),
          pageSizeOptions,
          prevLabel: t('adminDevicePrev'),
          total,
          totalLabel: t('adminDeviceTotal').replace('{total}', String(total)),
        }}
      />
    </section>
  );
};

export default DeviceManagement;
