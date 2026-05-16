import { useI18n } from '@/contexts/I18nContext';
import type { AdminMessageKey } from '@/i18n/admin';
import {
  createProduct,
  deleteProduct,
  listProducts,
  type Product,
  type ProductAuthType,
  type ProductNodeType,
  type ProductProtocolType,
  type ProductStatus,
  updateProduct,
} from '@/services/products';
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

interface ProductFormState {
  product_key: string;
  product_name: string;
  description: string;
  node_type: ProductNodeType;
  auth_type: ProductAuthType;
  protocol_type: ProductProtocolType;
  status: ProductStatus;
}

type FormMode = 'create' | 'edit';

const emptyForm: ProductFormState = {
  product_key: '',
  product_name: '',
  description: '',
  node_type: 'direct',
  auth_type: 'secret',
  protocol_type: 'mqtt',
  status: 'active',
};

const pageSizeOptions = [10, 20, 50];

const productStatusLabelKey: Record<ProductStatus, AdminMessageKey> = {
  active: 'adminProductStatusActive',
  disabled: 'adminProductStatusDisabled',
};

const nodeTypeLabelKey: Record<ProductNodeType, AdminMessageKey> = {
  direct: 'adminProductNodeDirect',
  gateway: 'adminProductNodeGateway',
  sub_device: 'adminProductNodeSubDevice',
};

const authTypeLabelKey: Record<ProductAuthType, AdminMessageKey> = {
  secret: 'adminProductAuthSecret',
  certificate: 'adminProductAuthCertificate',
  anonymous: 'adminProductAuthAnonymous',
};

const statusBadgeClassName = (status: ProductStatus) => {
  return [
    'inline-flex rounded-full px-2.5 py-1 text-xs font-bold',
    status === 'active'
      ? 'bg-linkflow-primary-soft text-linkflow-primary'
      : 'bg-slate-100 text-linkflow-muted dark:bg-slate-800 dark:text-linkflow-dark-muted',
  ].join(' ');
};

const toFormState = (product: Product): ProductFormState => ({
  product_key: product.product_key,
  product_name: product.product_name,
  description: product.description,
  node_type: product.node_type,
  auth_type: product.auth_type,
  protocol_type: product.protocol_type,
  status: product.status,
});

const ProductManagement = () => {
  const { t } = useI18n();
  const [tenants, setTenants] = useState<Tenant[]>([]);
  const [tenantId, setTenantId] = useState('');
  const [products, setProducts] = useState<Product[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [deletingId, setDeletingId] = useState<string | null>(null);
  const [formMode, setFormMode] = useState<FormMode | null>(null);
  const [editingProduct, setEditingProduct] = useState<Product | null>(null);
  const [form, setForm] = useState<ProductFormState>(emptyForm);

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

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

  const loadProducts = async (nextPage = page, nextPageSize = pageSize) => {
    if (!tenantId) {
      setProducts([]);
      setTotal(0);
      return;
    }

    setLoading(true);
    try {
      const result = await listProducts({
        tenant_id: tenantId,
        page: nextPage,
        page_size: nextPageSize,
      });
      setProducts(result.items);
      setTotal(result.total);
      setPage(result.page);
      setPageSize(result.page_size);
    } catch (error) {
      notification.error({
        message: t('adminProductLoadFailed'),
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
        setProducts([]);
        setTotal(0);
        return;
      }

      setLoading(true);
      try {
        const result = await listProducts({
          tenant_id: tenantId,
          page,
          page_size: pageSize,
        });
        if (cancelled) {
          return;
        }
        setProducts(result.items);
        setTotal(result.total);
        setPage(result.page);
        setPageSize(result.page_size);
      } catch (error) {
        if (cancelled) {
          return;
        }
        notification.error({
          message: t('adminProductLoadFailed'),
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
  }, [page, pageSize, tenantId, t]);

  const updateFormField = <K extends keyof ProductFormState>(
    key: K,
    value: ProductFormState[K],
  ) => {
    setForm((current) => ({ ...current, [key]: value }));
  };

  const openCreateForm = () => {
    setEditingProduct(null);
    setForm(emptyForm);
    setFormMode('create');
  };

  const openEditForm = (product: Product) => {
    setEditingProduct(product);
    setForm(toFormState(product));
    setFormMode('edit');
  };

  const closeForm = () => {
    setFormMode(null);
    setEditingProduct(null);
    setForm(emptyForm);
  };

  const handleTenantChange = (nextTenantId: string) => {
    setTenantId(nextTenantId);
    setPage(1);
    closeForm();
  };

  const handleSubmit = async (event: FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    if (!tenantId) {
      return;
    }

    setSaving(true);
    try {
      if (formMode === 'create') {
        await createProduct({
          tenant_id: tenantId,
          product_key: form.product_key.trim(),
          product_name: form.product_name.trim(),
          description: form.description.trim(),
          node_type: form.node_type,
          auth_type: form.auth_type,
          protocol_type: form.protocol_type,
        });
        notification.success({
          message: t('adminProductCreateSuccess'),
          placement: 'topRight',
        });
        setPage(1);
      }

      if (formMode === 'edit' && editingProduct) {
        await updateProduct(editingProduct.id, {
          product_name: form.product_name.trim(),
          description: form.description.trim(),
          node_type: form.node_type,
          auth_type: form.auth_type,
          protocol_type: form.protocol_type,
          status: form.status,
        });
        notification.success({
          message: t('adminProductUpdateSuccess'),
          placement: 'topRight',
        });
      }

      closeForm();
      await loadProducts(formMode === 'create' ? 1 : page, pageSize);
    } catch (error) {
      notification.error({
        message:
          formMode === 'create'
            ? t('adminProductCreateFailed')
            : t('adminProductUpdateFailed'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (product: Product) => {
    if (!window.confirm(t('adminProductDeleteConfirm'))) {
      return;
    }

    setDeletingId(product.id);
    try {
      await deleteProduct(product.id);
      notification.success({
        message: t('adminProductDeleteSuccess'),
        placement: 'topRight',
      });
      if (products.length === 1 && page > 1) {
        setPage((current) => current - 1);
      } else {
        await loadProducts();
      }
    } catch (error) {
      notification.error({
        message: t('adminProductDeleteFailed'),
        description: error instanceof Error ? error.message : undefined,
        placement: 'topRight',
      });
    } finally {
      setDeletingId(null);
    }
  };

  const columns: AdminDataTableColumn<Product>[] = [
    {
      key: 'name',
      title: t('adminProductName'),
      render: (product) => (
        <div>
          <strong className="block text-sm">{product.product_name}</strong>
          <span className="mt-1 block text-xs font-semibold text-linkflow-primary">
            {product.product_key}
          </span>
        </div>
      ),
    },
    {
      key: 'nodeType',
      title: t('adminProductNodeType'),
      render: (product) => t(nodeTypeLabelKey[product.node_type]),
    },
    {
      key: 'authType',
      title: t('adminProductAuthType'),
      render: (product) => t(authTypeLabelKey[product.auth_type]),
    },
    {
      key: 'protocol',
      title: t('adminProductProtocolType'),
      className: 'text-sm font-semibold uppercase',
      render: (product) => product.protocol_type,
    },
    {
      key: 'status',
      title: t('adminProductStatus'),
      render: (product) => (
        <span className={statusBadgeClassName(product.status)}>
          {t(productStatusLabelKey[product.status])}
        </span>
      ),
    },
    {
      key: 'updatedAt',
      title: t('adminProductUpdatedAt'),
      className: 'text-sm text-linkflow-subtle dark:text-linkflow-dark-subtle',
      render: (product) => formatDateTime(product.updated_at),
    },
    {
      key: 'actions',
      title: t('adminProductActions'),
      headerClassName: 'text-right',
      render: (product) => (
        <div className="flex justify-end gap-2">
          <button
            type="button"
            className="grid h-10 w-10 place-items-center rounded-md border border-linkflow-border bg-white text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
            aria-label={t('adminProductEdit')}
            onClick={() => openEditForm(product)}
          >
            <EditOutlined aria-hidden="true" />
          </button>
          <button
            type="button"
            className="grid h-10 w-10 place-items-center rounded-md border border-linkflow-border bg-white text-red-600 transition hover:border-red-500 hover:bg-red-50 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:hover:border-red-400 dark:hover:bg-red-950/30"
            aria-label={t('adminProductDelete')}
            onClick={() => handleDelete(product)}
            disabled={deletingId === product.id}
          >
            <DeleteOutlined aria-hidden="true" />
          </button>
        </div>
      ),
    },
  ];

  const mobileItems = products.map((product) => (
    <article
      key={product.id}
      className="rounded-lg border border-linkflow-border p-4 dark:border-linkflow-dark-border"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h4 className="m-0 truncate text-base font-bold">
            {product.product_name}
          </h4>
          <p className="m-0 mt-1 truncate text-sm font-semibold text-linkflow-primary">
            {product.product_key}
          </p>
        </div>
        <span className={statusBadgeClassName(product.status)}>
          {t(productStatusLabelKey[product.status])}
        </span>
      </div>
      <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminProductNodeType')}
          </dt>
          <dd className="m-0 mt-1">{t(nodeTypeLabelKey[product.node_type])}</dd>
        </div>
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminProductAuthType')}
          </dt>
          <dd className="m-0 mt-1">{t(authTypeLabelKey[product.auth_type])}</dd>
        </div>
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminProductProtocolType')}
          </dt>
          <dd className="m-0 mt-1 uppercase">{product.protocol_type}</dd>
        </div>
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminProductUpdatedAt')}
          </dt>
          <dd className="m-0 mt-1">{formatDateTime(product.updated_at)}</dd>
        </div>
      </dl>
      <div className="mt-4 flex gap-2">
        <button
          type="button"
          className="inline-flex h-11 flex-1 items-center justify-center gap-2 rounded-md border border-linkflow-border bg-white text-sm font-bold text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
          onClick={() => openEditForm(product)}
        >
          <EditOutlined aria-hidden="true" />
          {t('adminProductEdit')}
        </button>
        <button
          type="button"
          className="inline-flex h-11 flex-1 items-center justify-center gap-2 rounded-md border border-linkflow-border bg-white text-sm font-bold text-red-600 transition hover:border-red-500 hover:bg-red-50 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:hover:border-red-400 dark:hover:bg-red-950/30"
          onClick={() => handleDelete(product)}
          disabled={deletingId === product.id}
        >
          <DeleteOutlined aria-hidden="true" />
          {t('adminProductDelete')}
        </button>
      </div>
    </article>
  ));

  return (
    <section className="grid gap-5">
      <div className="flex flex-wrap items-end justify-between gap-3">
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
        <button
          type="button"
          className="inline-flex h-11 items-center gap-2 rounded-md border border-linkflow-primary bg-white/70 px-3 text-sm font-bold text-linkflow-primary shadow-sm transition hover:bg-linkflow-primary-soft disabled:cursor-not-allowed disabled:opacity-60 dark:border-linkflow-dark-primary dark:bg-linkflow-dark-panel/70 dark:text-linkflow-dark-primary dark:hover:bg-linkflow-dark-primary-soft"
          onClick={openCreateForm}
          disabled={!tenantId}
        >
          <PlusOutlined aria-hidden="true" />
          {t('adminProductCreate')}
        </button>
      </div>

      {formMode ? (
        <div
          className="fixed inset-0 z-50 grid place-items-center bg-slate-950/40 p-4"
          role="dialog"
          aria-modal="true"
          aria-labelledby="product-form-title"
        >
          <section className="max-h-[calc(100vh-2rem)] w-full max-w-3xl overflow-y-auto rounded-lg border border-linkflow-border bg-white p-5 shadow-[0_18px_60px_rgba(15,23,42,0.24)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/50">
            <div className="mb-4 flex items-center justify-between gap-3">
              <h3 id="product-form-title" className="m-0 text-lg font-bold">
                {formMode === 'create'
                  ? t('adminProductCreateTitle')
                  : t('adminProductEditTitle')}
              </h3>
              <button
                type="button"
                className="grid h-11 w-11 place-items-center rounded-md border border-linkflow-border bg-white text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
                aria-label={t('adminProductCancel')}
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
                {t('adminProductKey')}
                <input
                  className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.product_key}
                  onChange={(event) =>
                    updateFormField('product_key', event.target.value)
                  }
                  disabled={formMode === 'edit'}
                  required
                />
              </label>
              <label className="grid gap-2 text-sm font-bold">
                {t('adminProductName')}
                <input
                  className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.product_name}
                  onChange={(event) =>
                    updateFormField('product_name', event.target.value)
                  }
                  required
                />
              </label>
              <label className="grid gap-2 text-sm font-bold">
                {t('adminProductNodeType')}
                <select
                  className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.node_type}
                  onChange={(event) =>
                    updateFormField(
                      'node_type',
                      event.target.value as ProductNodeType,
                    )
                  }
                >
                  <option value="direct">{t('adminProductNodeDirect')}</option>
                  <option value="gateway">
                    {t('adminProductNodeGateway')}
                  </option>
                  <option value="sub_device">
                    {t('adminProductNodeSubDevice')}
                  </option>
                </select>
              </label>
              <label className="grid gap-2 text-sm font-bold">
                {t('adminProductAuthType')}
                <select
                  className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.auth_type}
                  onChange={(event) =>
                    updateFormField(
                      'auth_type',
                      event.target.value as ProductAuthType,
                    )
                  }
                >
                  <option value="secret">{t('adminProductAuthSecret')}</option>
                  <option value="certificate">
                    {t('adminProductAuthCertificate')}
                  </option>
                  <option value="anonymous">
                    {t('adminProductAuthAnonymous')}
                  </option>
                </select>
              </label>
              <label className="grid gap-2 text-sm font-bold">
                {t('adminProductProtocolType')}
                <select
                  className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.protocol_type}
                  onChange={(event) =>
                    updateFormField(
                      'protocol_type',
                      event.target.value as ProductProtocolType,
                    )
                  }
                >
                  <option value="mqtt">MQTT</option>
                  <option value="http">HTTP</option>
                  <option value="coap">CoAP</option>
                  <option value="modbus">Modbus</option>
                  <option value="opcua">OPC UA</option>
                  <option value="lora">LoRa</option>
                </select>
              </label>
              <label className="grid gap-2 text-sm font-bold">
                {t('adminProductStatus')}
                <select
                  className="h-11 rounded-md border border-linkflow-border bg-white px-3 text-sm font-medium outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
                  value={form.status}
                  onChange={(event) =>
                    updateFormField(
                      'status',
                      event.target.value as ProductStatus,
                    )
                  }
                  disabled={formMode === 'create'}
                >
                  <option value="active">
                    {t('adminProductStatusActive')}
                  </option>
                  <option value="disabled">
                    {t('adminProductStatusDisabled')}
                  </option>
                </select>
              </label>
              <label className="grid gap-2 text-sm font-bold md:col-span-2">
                {t('adminProductDescription')}
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
                  disabled={saving}
                >
                  <SaveOutlined aria-hidden="true" />
                  {saving ? t('adminProductSaving') : t('adminProductSave')}
                </button>
                <button
                  type="button"
                  className="h-11 rounded-md border border-linkflow-border bg-white px-4 text-sm font-bold text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
                  onClick={closeForm}
                >
                  {t('adminProductCancel')}
                </button>
              </div>
            </form>
          </section>
        </div>
      ) : null}

      <AdminDataTable
        columns={columns}
        emptyDescription={t('adminProductEmptyDescription')}
        emptyTitle={t('adminProductEmpty')}
        getRowKey={(product) => product.id}
        items={products}
        mobileItems={mobileItems}
        pagination={{
          jumpLabel: t('adminProductJump'),
          jumpToLabel: t('adminProductJumpTo'),
          loading,
          loadingLabel: t('adminProductLoading'),
          nextLabel: t('adminProductNext'),
          onPageChange: setPage,
          onPageSizeChange: (nextPageSize) => {
            setPage(1);
            setPageSize(nextPageSize);
          },
          page,
          pageLabel: t('adminProductPage')
            .replace('{page}', String(page))
            .replace('{totalPages}', String(totalPages)),
          pageSize,
          pageSizeLabel: t('adminProductPageSize'),
          pageSizeOptions,
          prevLabel: t('adminProductPrev'),
          total,
          totalLabel: t('adminProductTotal').replace('{total}', String(total)),
        }}
      />
    </section>
  );
};

export default ProductManagement;
