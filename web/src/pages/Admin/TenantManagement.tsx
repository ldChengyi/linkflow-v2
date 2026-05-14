import { useI18n } from '@/contexts/I18nContext';
import type { AdminMessageKey } from '@/i18n/admin';
import {
  createTenant,
  deleteTenant,
  listTenants,
  type Tenant,
  type TenantStatus,
  updateTenant,
} from '@/services/tenants';
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

interface TenantFormState {
  tenant_slug: string;
  tenant_name: string;
  status: TenantStatus;
}

type FormMode = 'create' | 'edit';

const emptyForm: TenantFormState = {
  tenant_slug: '',
  tenant_name: '',
  status: 'active',
};

const pageSizeOptions = [10, 20, 50];

const statusLabelKey: Record<TenantStatus, AdminMessageKey> = {
  active: 'adminTenantStatusActive',
  disabled: 'adminTenantStatusDisabled',
};

const tenantStatusBadgeClassName = (status: TenantStatus) => {
  return [
    'inline-flex rounded-full px-2.5 py-1 text-xs font-bold',
    status === 'active'
      ? 'bg-linkflow-primary-soft text-linkflow-primary'
      : 'bg-slate-100 text-linkflow-muted dark:bg-slate-800 dark:text-linkflow-dark-muted',
  ].join(' ');
};

const toFormState = (tenant: Tenant): TenantFormState => ({
  tenant_slug: tenant.tenant_slug,
  tenant_name: tenant.tenant_name,
  status: tenant.status,
});

const TenantManagement = () => {
  const { t } = useI18n();
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [formMode, setFormMode] = useState<FormMode | null>(null);
  const [editingTenant, setEditingTenant] = useState<Tenant | null>(null);
  const [form, setForm] = useState<TenantFormState>(emptyForm);

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  const loadTenants = async (nextPage = page, nextPageSize = pageSize) => {
    setLoading(true);
    try {
      const result = await listTenants({
        page: nextPage,
        page_size: nextPageSize,
      });
      setTenants(result.items);
      setTotal(result.total);
      setPage(result.page);
      setPageSize(result.page_size);
    } catch (error) {
      notification.error({
        message: t('adminTenantLoadFailed'),
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
      setLoading(true);
      try {
        const result = await listTenants({
          page,
          page_size: pageSize,
        });
        if (cancelled) {
          return;
        }
        setTenants(result.items);
        setTotal(result.total);
        setPage(result.page);
        setPageSize(result.page_size);
      } catch (error) {
        if (cancelled) {
          return;
        }
        notification.error({
          message: t('adminTenantLoadFailed'),
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
  }, [page, pageSize, t]);

  const openCreateForm = () => {
    setEditingTenant(null);
    setForm(emptyForm);
    setFormMode('create');
  };

  const openEditForm = (tenant: Tenant) => {
    setEditingTenant(tenant);
    setForm(toFormState(tenant));
    setFormMode('edit');
  };

  const closeForm = () => {
    setFormMode(null);
    setEditingTenant(null);
    setForm(emptyForm);
  };

  const updateFormField = <K extends keyof TenantFormState>(
    key: K,
    value: TenantFormState[K],
  ) => {
    setForm((current) => ({
      ...current,
      [key]: value,
    }));
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    setSaving(true);

    try {
      if (formMode === 'create') {
        await createTenant({
          tenant_slug: form.tenant_slug.trim(),
          tenant_name: form.tenant_name.trim(),
        });
        notification.success({
          message: t('adminTenantCreateSuccess'),
          placement: 'topRight',
        });
        setPage(1);
      }

      if (formMode === 'edit' && editingTenant) {
        await updateTenant(editingTenant.id, {
          tenant_name: form.tenant_name.trim(),
          status: form.status,
        });
        notification.success({
          message: t('adminTenantUpdateSuccess'),
          placement: 'topRight',
        });
      }

      closeForm();
      await loadTenants(formMode === 'create' ? 1 : page, pageSize);
    } catch (error) {
      notification.error({
        message:
          formMode === 'create'
            ? t('adminTenantCreateFailed')
            : t('adminTenantUpdateFailed'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (tenant: Tenant) => {
    if (!window.confirm(t('adminTenantDeleteConfirm'))) {
      return;
    }

    setDeletingId(tenant.id);
    try {
      await deleteTenant(tenant.id);
      notification.success({
        message: t('adminTenantDeleteSuccess'),
        placement: 'topRight',
      });
      if (tenants.length === 1 && page > 1) {
        setPage((current) => current - 1);
      } else {
        await loadTenants();
      }
    } catch (error) {
      notification.error({
        message: t('adminTenantDeleteFailed'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    } finally {
      setDeletingId(null);
    }
  };

  const columns: AdminDataTableColumn<Tenant>[] = [
    {
      key: 'name',
      title: t('adminTenantName'),
      render: (tenant) => (
        <strong className="block text-sm">{tenant.tenant_name}</strong>
      ),
    },
    {
      key: 'slug',
      title: t('adminTenantSlug'),
      className: 'text-sm font-semibold',
      render: (tenant) => tenant.tenant_slug,
    },
    {
      key: 'status',
      title: t('adminTenantStatus'),
      render: (tenant) => (
        <span className={tenantStatusBadgeClassName(tenant.status)}>
          {t(statusLabelKey[tenant.status])}
        </span>
      ),
    },
    {
      key: 'createdAt',
      title: t('adminTenantCreatedAt'),
      className: 'text-sm text-linkflow-subtle dark:text-linkflow-dark-subtle',
      render: (tenant) => formatDateTime(tenant.created_at),
    },
    {
      key: 'updatedAt',
      title: t('adminTenantUpdatedAt'),
      className: 'text-sm text-linkflow-subtle dark:text-linkflow-dark-subtle',
      render: (tenant) => formatDateTime(tenant.updated_at),
    },
    {
      key: 'actions',
      title: t('adminTenantActions'),
      headerClassName: 'text-right',
      render: (tenant) => (
        <div className="flex justify-end gap-2">
          <button
            type="button"
            className="grid h-10 w-10 place-items-center rounded-md border border-linkflow-border bg-white text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
            aria-label={t('adminTenantEdit')}
            onClick={() => openEditForm(tenant)}
          >
            <EditOutlined aria-hidden="true" />
          </button>
          <button
            type="button"
            className="grid h-10 w-10 place-items-center rounded-md border border-linkflow-border bg-white text-red-600 transition hover:border-red-500 hover:bg-red-50 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:hover:border-red-400 dark:hover:bg-red-950/30"
            aria-label={t('adminTenantDelete')}
            onClick={() => handleDelete(tenant)}
            disabled={deletingId === tenant.id}
          >
            <DeleteOutlined aria-hidden="true" />
          </button>
        </div>
      ),
    },
  ];

  const mobileItems = tenants.map((tenant) => (
    <article
      key={tenant.id}
      className="rounded-lg border border-linkflow-border p-4 dark:border-linkflow-dark-border"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h4 className="m-0 truncate text-base font-bold">
            {tenant.tenant_name}
          </h4>
          <p className="m-0 mt-1 text-sm font-semibold text-linkflow-primary">
            {tenant.tenant_slug}
          </p>
        </div>
        <span
          className={[
            'shrink-0',
            tenantStatusBadgeClassName(tenant.status),
          ].join(' ')}
        >
          {t(statusLabelKey[tenant.status])}
        </span>
      </div>
      <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminTenantCreatedAt')}
          </dt>
          <dd className="m-0 mt-1">{formatDateTime(tenant.created_at)}</dd>
        </div>
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminTenantUpdatedAt')}
          </dt>
          <dd className="m-0 mt-1">{formatDateTime(tenant.updated_at)}</dd>
        </div>
      </dl>
      <div className="mt-4 flex gap-2">
        <button
          type="button"
          className="inline-flex h-11 flex-1 items-center justify-center gap-2 rounded-md border border-linkflow-border bg-white text-sm font-bold text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
          onClick={() => openEditForm(tenant)}
        >
          <EditOutlined aria-hidden="true" />
          {t('adminTenantEdit')}
        </button>
        <button
          type="button"
          className="inline-flex h-11 flex-1 items-center justify-center gap-2 rounded-md border border-linkflow-border bg-white text-sm font-bold text-red-600 transition hover:border-red-500 hover:bg-red-50 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:hover:border-red-400 dark:hover:bg-red-950/30"
          onClick={() => handleDelete(tenant)}
          disabled={deletingId === tenant.id}
        >
          <DeleteOutlined aria-hidden="true" />
          {t('adminTenantDelete')}
        </button>
      </div>
    </article>
  ));

  return (
    <section className="grid gap-5">
      <div className="flex justify-start">
        <button
          type="button"
          className="inline-flex h-9 items-center gap-2 rounded-md border border-linkflow-primary bg-white/70 px-3 text-sm font-bold text-linkflow-primary shadow-sm transition hover:bg-linkflow-primary-soft disabled:cursor-not-allowed disabled:opacity-60 dark:border-linkflow-dark-primary dark:bg-linkflow-dark-panel/70 dark:text-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-soft"
          onClick={openCreateForm}
        >
          <PlusOutlined aria-hidden="true" />
          {t('adminTenantCreate')}
        </button>
      </div>

      {formMode ? (
        <div
          className="fixed inset-0 z-50 grid place-items-center bg-slate-950/40 p-4"
          role="dialog"
          aria-modal="true"
          aria-labelledby="tenant-form-title"
        >
          <section className="max-h-[calc(100vh-2rem)] w-full max-w-3xl overflow-y-auto rounded-lg border border-linkflow-border bg-white p-5 shadow-[0_18px_60px_rgba(15,23,42,0.24)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/50">
            <div className="mb-4 flex items-center justify-between gap-3">
              <h3 id="tenant-form-title" className="m-0 text-lg font-bold">
                {formMode === 'create'
                  ? t('adminTenantCreateTitle')
                  : t('adminTenantEditTitle')}
              </h3>
              <button
                type="button"
                className="grid h-11 w-11 place-items-center rounded-md border border-linkflow-border bg-white text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
                aria-label={t('adminTenantCancel')}
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
                {t('adminTenantSlug')}
                <input
                  className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.tenant_slug}
                  onChange={(event) =>
                    updateFormField('tenant_slug', event.target.value)
                  }
                  disabled={formMode === 'edit'}
                  required
                />
              </label>

              <label className="grid gap-2 text-sm font-bold">
                {t('adminTenantName')}
                <input
                  className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.tenant_name}
                  onChange={(event) =>
                    updateFormField('tenant_name', event.target.value)
                  }
                  required
                />
              </label>

              <label className="grid gap-2 text-sm font-bold">
                {t('adminTenantStatus')}
                <select
                  className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.status}
                  onChange={(event) =>
                    updateFormField(
                      'status',
                      event.target.value as TenantStatus,
                    )
                  }
                  disabled={formMode === 'create'}
                >
                  <option value="active">{t('adminTenantStatusActive')}</option>
                  <option value="disabled">
                    {t('adminTenantStatusDisabled')}
                  </option>
                </select>
              </label>

              <div className="flex items-end gap-2 md:col-span-2">
                <button
                  type="submit"
                  className="inline-flex h-11 items-center gap-2 rounded-md bg-linkflow-primary px-4 text-sm font-bold text-white shadow-md shadow-linkflow-primary-soft transition hover:bg-linkflow-primary-hover disabled:cursor-not-allowed disabled:opacity-60 dark:bg-linkflow-dark-primary dark:text-linkflow-dark-page dark:hover:bg-linkflow-dark-primary-hover"
                  disabled={saving}
                >
                  <SaveOutlined aria-hidden="true" />
                  {saving ? t('adminTenantSaving') : t('adminTenantSave')}
                </button>
                <button
                  type="button"
                  className="h-11 rounded-md border border-linkflow-border bg-white px-4 text-sm font-bold text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
                  onClick={closeForm}
                >
                  {t('adminTenantCancel')}
                </button>
              </div>
            </form>
          </section>
        </div>
      ) : null}

      <AdminDataTable
        columns={columns}
        emptyDescription={t('adminTenantEmptyDescription')}
        emptyTitle={t('adminTenantEmpty')}
        getRowKey={(tenant) => tenant.id}
        items={tenants}
        mobileItems={mobileItems}
        pagination={{
          jumpLabel: t('adminTenantJump'),
          jumpToLabel: t('adminTenantJumpTo'),
          loading,
          loadingLabel: t('adminTenantLoading'),
          nextLabel: t('adminTenantNext'),
          onPageChange: setPage,
          onPageSizeChange: (nextPageSize) => {
            setPage(1);
            setPageSize(nextPageSize);
          },
          page,
          pageLabel: t('adminTenantPage')
            .replace('{page}', String(page))
            .replace('{totalPages}', String(totalPages)),
          pageSize,
          pageSizeLabel: t('adminTenantPageSize'),
          pageSizeOptions,
          prevLabel: t('adminTenantPrev'),
          total,
          totalLabel: t('adminTenantTotal').replace('{total}', String(total)),
        }}
      />
    </section>
  );
};

export default TenantManagement;
