import {
  ArrowRightOutlined,
  LeftOutlined,
  LoadingOutlined,
  RightOutlined,
} from '@ant-design/icons';
import type { ReactNode } from 'react';
import { useState } from 'react';

export interface AdminDataTableColumn<T> {
  key: string;
  title: ReactNode;
  className?: string;
  headerClassName?: string;
  render: (item: T) => ReactNode;
}

export interface AdminTablePaginationProps {
  jumpLabel: string;
  jumpToLabel: string;
  loading: boolean;
  loadingLabel: string;
  nextLabel: string;
  onPageChange: (page: number) => void;
  onPageSizeChange: (pageSize: number) => void;
  page: number;
  pageLabel: string;
  pageSize: number;
  pageSizeLabel: string;
  pageSizeOptions: number[];
  prevLabel: string;
  total: number;
  totalLabel: string;
}

interface AdminDataTableProps<T> {
  columns: AdminDataTableColumn<T>[];
  emptyDescription: string;
  emptyTitle: string;
  getRowKey: (item: T) => string;
  items: T[];
  minWidthClassName?: string;
  mobileItems: ReactNode;
  pagination: AdminTablePaginationProps;
}

const controlButtonClassName =
  'grid h-10 w-10 place-items-center rounded-md border border-linkflow-border bg-white text-sm font-bold text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary disabled:cursor-not-allowed disabled:opacity-50 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary';

const AdminTablePagination = ({
  jumpLabel,
  jumpToLabel,
  loading,
  loadingLabel,
  nextLabel,
  onPageChange,
  onPageSizeChange,
  page,
  pageLabel,
  pageSize,
  pageSizeLabel,
  pageSizeOptions,
  prevLabel,
  total,
  totalLabel,
}: AdminTablePaginationProps) => {
  const [jumpPage, setJumpPage] = useState('');
  const totalPages = Math.max(1, Math.ceil(total / pageSize));

  const handleJumpPage = () => {
    const nextPage = Number(jumpPage);
    if (!Number.isInteger(nextPage) || nextPage < 1) {
      return;
    }

    onPageChange(Math.min(nextPage, totalPages));
    setJumpPage('');
  };

  return (
    <div className="grid items-center gap-3 border-t border-linkflow-border p-4 dark:border-linkflow-dark-border lg:grid-cols-[1fr_auto_1fr]">
      <div className="flex flex-wrap items-center gap-3 self-center">
        <div className="flex h-10 items-center text-sm font-semibold leading-none text-linkflow-subtle dark:text-linkflow-dark-subtle">
          {loading ? loadingLabel : totalLabel}
        </div>
        <label className="inline-flex h-10 items-center gap-2 text-sm font-semibold text-linkflow-subtle dark:text-linkflow-dark-subtle">
          {pageSizeLabel}
          <select
            className="h-10 rounded-md border border-linkflow-border bg-white px-2 text-sm outline-none dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
            value={pageSize}
            onChange={(event) => onPageSizeChange(Number(event.target.value))}
          >
            {pageSizeOptions.map((option) => (
              <option key={option} value={option}>
                {option}
              </option>
            ))}
          </select>
        </label>
      </div>

      <div className="flex h-10 items-center text-sm font-semibold leading-none text-linkflow-subtle dark:text-linkflow-dark-subtle lg:justify-center">
        <span>{pageLabel}</span>
      </div>

      <div className="flex flex-wrap items-center gap-2 lg:justify-end">
        <label className="inline-flex h-10 items-center gap-2 text-sm font-semibold text-linkflow-subtle dark:text-linkflow-dark-subtle">
          {jumpToLabel}
          <input
            className="h-10 w-20 rounded-md border border-linkflow-border bg-white px-2 text-sm outline-none transition focus:border-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-page"
            inputMode="numeric"
            max={totalPages}
            min={1}
            type="number"
            value={jumpPage}
            onChange={(event) => setJumpPage(event.target.value)}
            onKeyDown={(event) => {
              if (event.key === 'Enter') {
                handleJumpPage();
              }
            }}
          />
        </label>
        <button
          type="button"
          className={controlButtonClassName}
          disabled={loading || jumpPage === ''}
          aria-label={jumpLabel}
          title={jumpLabel}
          onClick={handleJumpPage}
        >
          <ArrowRightOutlined aria-hidden="true" />
        </button>
        <button
          type="button"
          className={controlButtonClassName}
          disabled={page <= 1 || loading}
          aria-label={prevLabel}
          title={prevLabel}
          onClick={() => onPageChange(Math.max(1, page - 1))}
        >
          <LeftOutlined aria-hidden="true" />
        </button>
        <button
          type="button"
          className={controlButtonClassName}
          disabled={page >= totalPages || loading}
          aria-label={nextLabel}
          title={nextLabel}
          onClick={() => onPageChange(Math.min(totalPages, page + 1))}
        >
          <RightOutlined aria-hidden="true" />
        </button>
      </div>
    </div>
  );
};

const AdminDataTable = <T,>({
  columns,
  emptyDescription,
  emptyTitle,
  getRowKey,
  items,
  minWidthClassName = 'min-w-[880px]',
  mobileItems,
  pagination,
}: AdminDataTableProps<T>) => {
  const rowOffset = (pagination.page - 1) * pagination.pageSize;

  return (
    <section className="overflow-hidden rounded-lg border border-linkflow-border bg-white shadow-[0_8px_24px_rgba(23,32,51,0.06)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/20">
      <div className="hidden overflow-x-auto lg:block">
        <table
          className={[
            'w-full border-collapse text-left',
            minWidthClassName,
          ].join(' ')}
        >
          <thead>
            <tr className="border-b border-linkflow-border text-xs uppercase text-linkflow-subtle dark:border-linkflow-dark-border dark:text-linkflow-dark-subtle">
              <th className="w-16 px-4 py-3">#</th>
              {columns.map((column) => (
                <th
                  key={column.key}
                  className={['px-4 py-3', column.headerClassName ?? ''].join(
                    ' ',
                  )}
                >
                  {column.title}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {items.map((item, index) => (
              <tr
                key={getRowKey(item)}
                className="border-b border-linkflow-border last:border-b-0 dark:border-linkflow-dark-border"
              >
                <td className="px-4 py-4 align-middle text-sm font-semibold text-linkflow-subtle dark:text-linkflow-dark-subtle">
                  {rowOffset + index + 1}
                </td>
                {columns.map((column) => (
                  <td
                    key={column.key}
                    className={[
                      'px-4 py-4 align-middle',
                      column.className ?? '',
                    ].join(' ')}
                  >
                    {column.render(item)}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>

      {items.length > 0 ? (
        <div className="grid gap-3 p-4 lg:hidden">{mobileItems}</div>
      ) : null}

      {pagination.loading && items.length === 0 ? (
        <div className="grid min-h-48 place-items-center p-8 text-center text-linkflow-subtle dark:text-linkflow-dark-subtle">
          <div className="grid gap-3 justify-items-center">
            <LoadingOutlined
              className="text-2xl text-linkflow-primary"
              aria-hidden="true"
            />
            <p className="m-0 text-sm font-bold">{pagination.loadingLabel}</p>
          </div>
        </div>
      ) : null}

      {!pagination.loading && items.length === 0 ? (
        <div className="p-8 text-center text-linkflow-subtle dark:text-linkflow-dark-subtle">
          <p className="m-0 text-base font-bold">{emptyTitle}</p>
          <p className="m-0 mt-2">{emptyDescription}</p>
        </div>
      ) : null}

      <AdminTablePagination {...pagination} />
    </section>
  );
};

export default AdminDataTable;
