import { useI18n } from '@/contexts/I18nContext';
import {
  listAuditLogs,
  type AuditLog,
  type AuditResult,
} from '@/services/auditLogs';
import { formatDateTime } from '@/utils/date';
import { CloseOutlined, EyeOutlined } from '@ant-design/icons';
import { notification } from 'antd';
import type { ReactNode } from 'react';
import { useEffect, useState } from 'react';
import AdminDataTable, {
  type AdminDataTableColumn,
} from './components/AdminDataTable';

const pageSizeOptions = [10, 20, 50];

const resultClassName: Record<AuditResult, string> = {
  failure: 'bg-red-50 text-red-600 dark:bg-red-950/30 dark:text-red-300',
  success: 'bg-linkflow-primary-soft text-linkflow-primary',
};

const methodClassName: Record<string, string> = {
  DELETE: 'bg-red-50 text-red-600 dark:bg-red-950/30 dark:text-red-300',
  GET: 'bg-sky-50 text-sky-700 dark:bg-sky-950/40 dark:text-sky-300',
  POST: 'bg-emerald-50 text-emerald-700 dark:bg-emerald-950/40 dark:text-emerald-300',
  PUT: 'bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-300',
};

const getStatusClassName = (statusCode: number) => {
  if (statusCode >= 500) {
    return 'bg-red-50 text-red-600 dark:bg-red-950/30 dark:text-red-300';
  }

  if (statusCode >= 400) {
    return 'bg-amber-50 text-amber-700 dark:bg-amber-950/40 dark:text-amber-300';
  }

  if (statusCode >= 300) {
    return 'bg-violet-50 text-violet-700 dark:bg-violet-950/40 dark:text-violet-300';
  }

  return 'bg-linkflow-primary-soft text-linkflow-primary';
};

const tagBaseClassName =
  'inline-flex whitespace-nowrap rounded-full px-2.5 py-1 text-xs font-bold leading-none';

const renderMethodTag = (method: string) => {
  const normalizedMethod = method.toUpperCase();

  return (
    <span
      className={[
        tagBaseClassName,
        methodClassName[normalizedMethod] ??
          'bg-slate-100 text-linkflow-muted dark:bg-slate-800 dark:text-linkflow-dark-muted',
      ].join(' ')}
    >
      {normalizedMethod}
    </span>
  );
};

const renderStatusTag = (statusCode: number) => {
  return (
    <span
      className={[tagBaseClassName, getStatusClassName(statusCode)].join(' ')}
    >
      {statusCode}
    </span>
  );
};

const formatMetadata = (metadata: Record<string, unknown>) => {
  if (Object.keys(metadata).length === 0) {
    return '--';
  }

  return JSON.stringify(metadata, null, 2);
};

interface AuditDetailFieldProps {
  children: ReactNode;
  description: string;
  label: string;
  wide?: boolean;
}

const AuditDetailField = ({
  children,
  description,
  label,
  wide = false,
}: AuditDetailFieldProps) => {
  return (
    <div
      className={[
        'min-w-0 rounded-md border border-linkflow-border bg-slate-50/70 p-3 dark:border-linkflow-dark-border dark:bg-slate-900/30',
        wide ? 'md:col-span-2' : '',
      ].join(' ')}
    >
      <dt className="text-sm font-bold text-linkflow-text dark:text-linkflow-dark-text">
        {label}
      </dt>
      <dd className="m-0 mt-1 text-sm font-semibold text-linkflow-muted dark:text-linkflow-dark-muted">
        {children}
      </dd>
      <p className="m-0 mt-2 text-xs leading-5 text-linkflow-subtle dark:text-linkflow-dark-subtle">
        {description}
      </p>
    </div>
  );
};

const AuditLogManagement = () => {
  const { t } = useI18n();
  const [logs, setLogs] = useState<AuditLog[]>([]);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [loading, setLoading] = useState(false);
  const [selectedLog, setSelectedLog] = useState<AuditLog | null>(null);

  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  useEffect(() => {
    let cancelled = false;

    const run = async () => {
      setLoading(true);
      try {
        const result = await listAuditLogs({
          page,
          page_size: pageSize,
        });
        if (cancelled) {
          return;
        }
        setLogs(result.items);
        setTotal(result.total);
        setPage(result.page);
        setPageSize(result.page_size);
      } catch (error) {
        if (cancelled) {
          return;
        }
        notification.error({
          message: t('adminAuditLoadFailed'),
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

  const columns: AdminDataTableColumn<AuditLog>[] = [
    {
      key: 'action',
      title: t('adminAuditAction'),
      render: (log) => (
        <div>
          <strong className="block text-sm">{log.action}</strong>
          <span className="mt-1 block text-xs text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {log.resource_type}
            {log.resource_id ? ` / ${log.resource_id}` : ''}
          </span>
        </div>
      ),
    },
    {
      key: 'result',
      title: t('adminAuditResult'),
      render: (log) => (
        <span
          className={[
            tagBaseClassName,
            resultClassName[log.result] ?? resultClassName.failure,
          ].join(' ')}
        >
          {log.result === 'success'
            ? t('adminAuditResultSuccess')
            : t('adminAuditResultFailure')}
        </span>
      ),
    },
    {
      key: 'method',
      title: t('adminAuditMethod'),
      render: (log) => renderMethodTag(log.method),
    },
    {
      key: 'path',
      title: t('adminAuditPath'),
      className:
        'max-w-[22rem] text-sm text-linkflow-subtle dark:text-linkflow-dark-subtle',
      render: (log) => <span className="block truncate">{log.path}</span>,
    },
    {
      key: 'status',
      title: t('adminAuditStatusCode'),
      render: (log) => renderStatusTag(log.status_code),
    },
    {
      key: 'duration',
      title: t('adminAuditDuration'),
      className: 'text-sm text-linkflow-subtle dark:text-linkflow-dark-subtle',
      render: (log) =>
        t('adminAuditDurationValue').replace(
          '{duration}',
          String(log.duration_ms),
        ),
    },
    {
      key: 'createdAt',
      title: t('adminAuditCreatedAt'),
      className: 'text-sm text-linkflow-subtle dark:text-linkflow-dark-subtle',
      render: (log) => formatDateTime(log.created_at),
    },
    {
      key: 'actions',
      title: t('adminAuditActions'),
      headerClassName: 'text-right',
      render: (log) => (
        <div className="flex justify-end">
          <button
            type="button"
            className="grid h-10 w-10 place-items-center rounded-md border border-linkflow-border bg-white text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
            aria-label={t('adminAuditViewDetail')}
            title={t('adminAuditViewDetail')}
            onClick={() => setSelectedLog(log)}
          >
            <EyeOutlined aria-hidden="true" />
          </button>
        </div>
      ),
    },
  ];

  const mobileItems = logs.map((log) => (
    <article
      key={log.id}
      className="rounded-lg border border-linkflow-border p-4 dark:border-linkflow-dark-border"
    >
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <h4 className="m-0 truncate text-base font-bold">{log.action}</h4>
          <p className="m-0 mt-1 truncate text-sm font-semibold text-linkflow-primary">
            {log.method} {log.path}
          </p>
        </div>
        <span
          className={[
            'shrink-0',
            tagBaseClassName,
            resultClassName[log.result] ?? resultClassName.failure,
          ].join(' ')}
        >
          {log.result === 'success'
            ? t('adminAuditResultSuccess')
            : t('adminAuditResultFailure')}
        </span>
      </div>
      <dl className="mt-4 grid grid-cols-2 gap-3 text-sm">
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminAuditMethod')}
          </dt>
          <dd className="m-0 mt-1">{renderMethodTag(log.method)}</dd>
        </div>
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminAuditResource')}
          </dt>
          <dd className="m-0 mt-1">
            {log.resource_type}
            {log.resource_id ? ` / ${log.resource_id}` : ''}
          </dd>
        </div>
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminAuditStatusCode')}
          </dt>
          <dd className="m-0 mt-1">{renderStatusTag(log.status_code)}</dd>
        </div>
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminAuditDuration')}
          </dt>
          <dd className="m-0 mt-1">
            {t('adminAuditDurationValue').replace(
              '{duration}',
              String(log.duration_ms),
            )}
          </dd>
        </div>
        <div>
          <dt className="font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
            {t('adminAuditCreatedAt')}
          </dt>
          <dd className="m-0 mt-1">{formatDateTime(log.created_at)}</dd>
        </div>
      </dl>
      <button
        type="button"
        className="mt-4 inline-flex h-10 items-center gap-2 rounded-md border border-linkflow-border bg-white px-3 text-sm font-bold text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
        onClick={() => setSelectedLog(log)}
      >
        <EyeOutlined aria-hidden="true" />
        {t('adminAuditViewDetail')}
      </button>
    </article>
  ));

  return (
    <section className="grid gap-5">
      <AdminDataTable
        columns={columns}
        emptyDescription={t('adminAuditEmptyDescription')}
        emptyTitle={t('adminAuditEmpty')}
        getRowKey={(log) => log.id}
        items={logs}
        minWidthClassName="min-w-[1080px]"
        mobileItems={mobileItems}
        pagination={{
          jumpLabel: t('adminAuditJump'),
          jumpToLabel: t('adminAuditJumpTo'),
          loading,
          loadingLabel: t('adminAuditLoading'),
          nextLabel: t('adminAuditNext'),
          onPageChange: setPage,
          onPageSizeChange: (nextPageSize) => {
            setPage(1);
            setPageSize(nextPageSize);
          },
          page,
          pageLabel: t('adminAuditPage')
            .replace('{page}', String(page))
            .replace('{totalPages}', String(totalPages)),
          pageSize,
          pageSizeLabel: t('adminAuditPageSize'),
          pageSizeOptions,
          prevLabel: t('adminAuditPrev'),
          total,
          totalLabel: t('adminAuditTotal').replace('{total}', String(total)),
        }}
      />

      {selectedLog ? (
        <div
          className="fixed inset-0 z-50 grid place-items-center bg-slate-950/40 p-4"
          role="dialog"
          aria-modal="true"
          aria-labelledby="audit-log-detail-title"
        >
          <section className="max-h-[calc(100vh-2rem)] w-full max-w-4xl overflow-y-auto rounded-lg border border-linkflow-border bg-white p-5 shadow-[0_18px_60px_rgba(15,23,42,0.24)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/50">
            <div className="mb-4 flex items-center justify-between gap-3">
              <h3 id="audit-log-detail-title" className="m-0 text-lg font-bold">
                {t('adminAuditDetailTitle')}
              </h3>
              <button
                type="button"
                className="grid h-11 w-11 place-items-center rounded-md border border-linkflow-border bg-white text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
                aria-label={t('adminAuditCloseDetail')}
                onClick={() => setSelectedLog(null)}
              >
                <CloseOutlined aria-hidden="true" />
              </button>
            </div>

            <div className="grid gap-5">
              <section>
                <h4 className="m-0 mb-3 text-sm font-bold text-linkflow-primary">
                  {t('adminAuditDetailBusiness')}
                </h4>
                <dl className="grid gap-3 md:grid-cols-2">
                  <AuditDetailField
                    label={t('adminAuditAction')}
                    description={t('adminAuditActionDescription')}
                  >
                    {selectedLog.action}
                  </AuditDetailField>
                  <AuditDetailField
                    label={t('adminAuditResult')}
                    description={t('adminAuditResultDescription')}
                  >
                    <span
                      className={[
                        tagBaseClassName,
                        resultClassName[selectedLog.result] ??
                          resultClassName.failure,
                      ].join(' ')}
                    >
                      {selectedLog.result === 'success'
                        ? t('adminAuditResultSuccess')
                        : t('adminAuditResultFailure')}
                    </span>
                  </AuditDetailField>
                  <AuditDetailField
                    label={t('adminAuditResource')}
                    description={t('adminAuditResourceDescription')}
                    wide
                  >
                    <span className="break-all">
                      {selectedLog.resource_type}
                      {selectedLog.resource_id
                        ? ` / ${selectedLog.resource_id}`
                        : ''}
                    </span>
                  </AuditDetailField>
                </dl>
              </section>

              <section>
                <h4 className="m-0 mb-3 text-sm font-bold text-linkflow-primary">
                  {t('adminAuditDetailHttp')}
                </h4>
                <dl className="grid gap-3 md:grid-cols-2">
                  <AuditDetailField
                    label={t('adminAuditMethod')}
                    description={t('adminAuditMethodDescription')}
                  >
                    {renderMethodTag(selectedLog.method)}
                  </AuditDetailField>
                  <AuditDetailField
                    label={t('adminAuditStatusCode')}
                    description={t('adminAuditStatusDescription')}
                  >
                    {renderStatusTag(selectedLog.status_code)}
                  </AuditDetailField>
                  <AuditDetailField
                    label={t('adminAuditPath')}
                    description={t('adminAuditPathDescription')}
                    wide
                  >
                    <span className="break-all">{selectedLog.path}</span>
                  </AuditDetailField>
                  <AuditDetailField
                    label={t('adminAuditDuration')}
                    description={t('adminAuditDurationDescription')}
                  >
                    {t('adminAuditDurationValue').replace(
                      '{duration}',
                      String(selectedLog.duration_ms),
                    )}
                  </AuditDetailField>
                  <AuditDetailField
                    label={t('adminAuditCreatedAt')}
                    description={t('adminAuditCreatedAtDescription')}
                  >
                    {formatDateTime(selectedLog.created_at)}
                  </AuditDetailField>
                </dl>
              </section>

              <section>
                <h4 className="m-0 mb-3 text-sm font-bold text-linkflow-primary">
                  {t('adminAuditDetailTrace')}
                </h4>
                <dl className="grid gap-3 md:grid-cols-2">
                  <AuditDetailField
                    label={t('adminAuditRequestID')}
                    description={t('adminAuditRequestIDDescription')}
                  >
                    <span className="break-all">
                      {selectedLog.request_id || '--'}
                    </span>
                  </AuditDetailField>
                  <AuditDetailField
                    label={t('adminAuditTraceID')}
                    description={t('adminAuditTraceIDDescription')}
                  >
                    <span className="break-all">
                      {selectedLog.trace_id || '--'}
                    </span>
                  </AuditDetailField>
                  <AuditDetailField
                    label={t('adminAuditIP')}
                    description={t('adminAuditIPDescription')}
                  >
                    <span className="break-all">{selectedLog.ip || '--'}</span>
                  </AuditDetailField>
                  <AuditDetailField
                    label={t('adminAuditUserAgent')}
                    description={t('adminAuditUserAgentDescription')}
                  >
                    <span className="break-all">
                      {selectedLog.user_agent || '--'}
                    </span>
                  </AuditDetailField>
                </dl>
              </section>

              <section>
                <h4 className="m-0 mb-3 text-sm font-bold text-linkflow-primary">
                  {t('adminAuditMetadata')}
                </h4>
                <p className="m-0 mb-2 text-xs leading-5 text-linkflow-subtle dark:text-linkflow-dark-subtle">
                  {t('adminAuditMetadataDescription')}
                </p>
                <pre className="m-0 max-h-72 overflow-auto rounded-md border border-linkflow-border bg-slate-50 p-3 text-xs leading-6 text-linkflow-text dark:border-linkflow-dark-border dark:bg-slate-900/60 dark:text-linkflow-dark-text">
                  {formatMetadata(selectedLog.metadata)}
                </pre>
              </section>
            </div>
          </section>
        </div>
      ) : null}
    </section>
  );
};

export default AuditLogManagement;
