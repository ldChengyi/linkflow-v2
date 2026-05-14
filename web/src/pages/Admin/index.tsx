import { useI18n } from '@/contexts/I18nContext';
import { useTheme } from '@/contexts/ThemeContext';
import type { AdminMessageKey } from '@/i18n/admin';
import type { AppLocale } from '@/i18n/auth';
import { type AdminSectionId, useAdminUiStore } from '@/stores/adminUiStore';
import { useAuthStore } from '@/stores/authStore';
import {
  CloseOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  DeploymentUnitOutlined,
  FileSearchOutlined,
  GlobalOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  MoonOutlined,
  SettingOutlined,
  SunOutlined,
} from '@ant-design/icons';
import { Link, history, useLocation } from '@umijs/max';
import { AnimatePresence, motion, useReducedMotion } from 'motion/react';
import { useEffect, useState } from 'react';
import AuditLogManagement from './AuditLogManagement';
import TenantManagement from './TenantManagement';

interface NavItem {
  icon: React.ReactNode;
  id: AdminSectionId;
  labelKey: AdminMessageKey;
  path: string;
}

interface MetricItem {
  labelKey: AdminMessageKey;
  value: string;
  descriptionKey: AdminMessageKey;
}

interface SectionContent {
  id: AdminSectionId;
  titleKey: AdminMessageKey;
  summaryKey: AdminMessageKey;
  contentKey: AdminMessageKey;
}

const navItems: NavItem[] = [
  {
    icon: <DashboardOutlined aria-hidden="true" />,
    id: 'overview',
    labelKey: 'adminOverviewNav',
    path: '/admin',
  },
  {
    icon: <DeploymentUnitOutlined aria-hidden="true" />,
    id: 'tenants',
    labelKey: 'adminTenantsNav',
    path: '/admin/tenants',
  },
  {
    icon: <FileSearchOutlined aria-hidden="true" />,
    id: 'auditLogs',
    labelKey: 'adminAuditLogsNav',
    path: '/admin/audit-logs',
  },
  {
    icon: <DatabaseOutlined aria-hidden="true" />,
    id: 'deviceManagement',
    labelKey: 'adminDeviceManagementNav',
    path: '/admin/device-management',
  },
  {
    icon: <SettingOutlined aria-hidden="true" />,
    id: 'settings',
    labelKey: 'adminSettingsNav',
    path: '/admin/settings',
  },
];

const metrics: MetricItem[] = [
  {
    labelKey: 'adminMetricOnlineDevices',
    value: '0',
    descriptionKey: 'adminMetricWaitingDevices',
  },
  {
    labelKey: 'adminMetricTenants',
    value: '--',
    descriptionKey: 'adminMetricWaitingTenants',
  },
  {
    labelKey: 'adminMetricActiveProducts',
    value: '--',
    descriptionKey: 'adminMetricWaitingProducts',
  },
];

const sectionContent: Record<AdminSectionId, SectionContent> = {
  overview: {
    id: 'overview',
    titleKey: 'adminOverviewTitle',
    summaryKey: 'adminOverviewSummary',
    contentKey: 'adminOverviewContent',
  },
  tenants: {
    id: 'tenants',
    titleKey: 'adminTenantsTitle',
    summaryKey: 'adminTenantsSummary',
    contentKey: 'adminTenantsContent',
  },
  auditLogs: {
    id: 'auditLogs',
    titleKey: 'adminAuditLogsTitle',
    summaryKey: 'adminAuditLogsSummary',
    contentKey: 'adminAuditLogsContent',
  },
  deviceManagement: {
    id: 'deviceManagement',
    titleKey: 'adminDeviceManagementTitle',
    summaryKey: 'adminDeviceManagementSummary',
    contentKey: 'adminDeviceManagementContent',
  },
  settings: {
    id: 'settings',
    titleKey: 'adminSettingsTitle',
    summaryKey: 'adminSettingsSummary',
    contentKey: 'adminSettingsContent',
  },
};

const getActiveSectionId = (pathname: string): AdminSectionId => {
  if (pathname.startsWith('/admin/tenants')) {
    return 'tenants';
  }

  if (pathname.startsWith('/admin/audit-logs')) {
    return 'auditLogs';
  }

  if (pathname.startsWith('/admin/device-management')) {
    return 'deviceManagement';
  }

  if (pathname.startsWith('/admin/settings')) {
    return 'settings';
  }

  return 'overview';
};

const getNavItemById = (id: AdminSectionId) => {
  return navItems.find((item) => item.id === id) ?? navItems[0];
};

const AdminPage = () => {
  const { locale, setLocale, t } = useI18n();
  const { darkMode, toggleDarkMode } = useTheme();
  const location = useLocation();
  const reduceMotion = useReducedMotion();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const sidebarExpanded = useAdminUiStore((state) => state.sidebarExpanded);
  const setSidebarExpanded = useAdminUiStore(
    (state) => state.setSidebarExpanded,
  );
  const visitedTabs = useAdminUiStore((state) => state.visitedTabs);
  const addVisitedTab = useAdminUiStore((state) => state.addVisitedTab);
  const removeVisitedTab = useAdminUiStore((state) => state.removeVisitedTab);
  const clearSession = useAuthStore((state) => state.clearSession);

  const activeSectionId = getActiveSectionId(location.pathname);
  const activeContent = sectionContent[activeSectionId];

  useEffect(() => {
    addVisitedTab(activeSectionId);
  }, [activeSectionId, addVisitedTab]);

  const handleLogout = () => {
    clearSession();
    history.push('/login');
  };

  const handleCloseTab = (
    event: React.MouseEvent<HTMLButtonElement>,
    tabId: AdminSectionId,
  ) => {
    event.preventDefault();
    event.stopPropagation();

    const nextTabs = visitedTabs.filter((id) => id !== tabId);
    const fallbackTab = getNavItemById(nextTabs.at(-1) ?? 'overview');

    removeVisitedTab(tabId);

    if (tabId === activeSectionId) {
      history.push(fallbackTab.path);
    }
  };

  return (
    <main
      className={[
        darkMode ? 'dark' : '',
        'h-screen overflow-hidden bg-linkflow-page text-linkflow-text dark:bg-linkflow-dark-page dark:text-linkflow-dark-text',
      ].join(' ')}
    >
      <div className="flex h-full min-h-0">
        <aside
          className={[
            'fixed inset-y-0 left-0 z-30 flex flex-col border-r border-linkflow-border bg-linkflow-panel px-3 py-5 shadow-xl shadow-slate-200/70 transition-[transform,width] duration-200 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/40 lg:static lg:h-screen lg:shrink-0 lg:translate-x-0 lg:shadow-none',
            sidebarExpanded
              ? 'w-20 items-center lg:w-64 lg:items-stretch'
              : 'w-20 items-center',
            sidebarOpen ? 'translate-x-0' : '-translate-x-full',
          ].join(' ')}
        >
          <div
            className={[
              'mb-8 flex min-h-11 items-center',
              sidebarExpanded
                ? 'justify-center lg:justify-start lg:gap-3'
                : 'justify-center',
            ].join(' ')}
          >
            <Link
              to="/home"
              className={[
                'grid h-11 w-11 shrink-0 place-items-center rounded-md bg-linkflow-primary text-sm font-bold text-white no-underline shadow-md shadow-linkflow-primary-soft transition hover:bg-linkflow-primary-hover',
                sidebarExpanded ? '' : 'mx-auto',
              ].join(' ')}
              onClick={() => setSidebarOpen(false)}
              aria-label="LinkFlow"
            >
              LF
            </Link>
            {sidebarExpanded ? (
              <div className="hidden min-w-0 flex-1 lg:block">
                <p className="m-0 truncate text-sm font-bold leading-tight">
                  LinkFlow
                </p>
                <p className="m-0 truncate text-xs text-linkflow-muted dark:text-linkflow-dark-muted">
                  {t('adminConsole')}
                </p>
              </div>
            ) : null}
          </div>

          <nav
            className={['grid gap-2', sidebarExpanded ? 'w-full' : ''].join(
              ' ',
            )}
            aria-label={t('adminConsole')}
          >
            {navItems.map((item) => {
              const active = item.id === activeSectionId;

              return (
                <Link
                  key={item.id}
                  to={item.path}
                  aria-current={active ? 'page' : undefined}
                  aria-label={t(item.labelKey)}
                  title={t(item.labelKey)}
                  onClick={() => {
                    addVisitedTab(item.id);
                    setSidebarOpen(false);
                  }}
                  className={[
                    'relative flex h-11 items-center rounded-md text-base no-underline transition',
                    sidebarExpanded
                      ? 'w-11 justify-center lg:w-full lg:justify-start lg:gap-3 lg:px-3'
                      : 'w-11 justify-center',
                    active
                      ? 'bg-linkflow-primary-soft text-linkflow-primary'
                      : 'text-linkflow-muted hover:bg-linkflow-primary-soft hover:text-linkflow-primary dark:text-linkflow-dark-muted dark:hover:text-linkflow-dark-primary',
                  ].join(' ')}
                >
                  {active ? (
                    <motion.span
                      layoutId="admin-active-route"
                      className="absolute inset-y-2 left-0 w-1 rounded-full bg-linkflow-primary"
                      transition={{ duration: 0.18, ease: 'easeOut' }}
                    />
                  ) : null}
                  <span className="grid h-8 w-8 place-items-center text-base text-linkflow-primary">
                    {item.icon}
                  </span>
                  {sidebarExpanded ? (
                    <span className="hidden truncate text-sm font-semibold lg:inline">
                      {t(item.labelKey)}
                    </span>
                  ) : null}
                </Link>
              );
            })}
          </nav>
        </aside>

        {sidebarOpen ? (
          <button
            type="button"
            className="fixed inset-0 z-20 bg-slate-950/30 lg:hidden"
            aria-label="Close sidebar"
            onClick={() => setSidebarOpen(false)}
          />
        ) : null}

        <section className="flex min-h-0 min-w-0 flex-1 flex-col">
          <header className="z-10 flex min-h-16 shrink-0 items-center justify-between border-b border-linkflow-border bg-linkflow-panel/95 px-4 backdrop-blur dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel/95 sm:px-6">
            <div className="flex min-w-0 items-center gap-3">
              <button
                type="button"
                className="grid h-11 w-11 place-items-center rounded-md border border-linkflow-border bg-white text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary lg:hidden"
                aria-label={sidebarOpen ? 'Close sidebar' : 'Open sidebar'}
                onClick={() => setSidebarOpen((current) => !current)}
              >
                {sidebarOpen ? (
                  <MenuFoldOutlined aria-hidden="true" />
                ) : (
                  <MenuUnfoldOutlined aria-hidden="true" />
                )}
              </button>
              <button
                type="button"
                className="hidden h-11 w-11 place-items-center rounded-md border border-linkflow-border bg-white text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary lg:grid"
                aria-label={
                  sidebarExpanded ? t('adminCollapseNav') : t('adminExpandNav')
                }
                title={
                  sidebarExpanded ? t('adminCollapseNav') : t('adminExpandNav')
                }
                onClick={() => setSidebarExpanded((current) => !current)}
              >
                {sidebarExpanded ? (
                  <MenuFoldOutlined aria-hidden="true" />
                ) : (
                  <MenuUnfoldOutlined aria-hidden="true" />
                )}
              </button>
            </div>

            <div className="flex items-center gap-2">
              <button
                type="button"
                className="grid h-11 w-11 place-items-center rounded-md border border-linkflow-border bg-white text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
                aria-label={
                  darkMode ? t('adminToggleLight') : t('adminToggleDark')
                }
                onClick={toggleDarkMode}
              >
                {darkMode ? (
                  <SunOutlined aria-hidden="true" />
                ) : (
                  <MoonOutlined aria-hidden="true" />
                )}
              </button>
              <div className="relative inline-flex h-11 items-center rounded-md border border-linkflow-border bg-white text-linkflow-muted transition focus-within:border-linkflow-primary hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:focus-within:border-linkflow-dark-primary dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary">
                <GlobalOutlined
                  className="pointer-events-none absolute left-3"
                  aria-hidden="true"
                />
                <select
                  aria-label={t('adminLanguage')}
                  className="h-full appearance-none rounded-md bg-transparent pl-9 pr-7 text-sm font-semibold outline-none"
                  value={locale}
                  onChange={(event) =>
                    setLocale(event.target.value as AppLocale)
                  }
                >
                  <option value="zh-CN">中文</option>
                  <option value="en-US">EN</option>
                </select>
              </div>
              <button
                type="button"
                className="grid h-11 w-11 place-items-center rounded-md border border-linkflow-border bg-white text-linkflow-muted transition hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary"
                aria-label={t('adminLogout')}
                onClick={handleLogout}
              >
                <LogoutOutlined aria-hidden="true" />
              </button>
            </div>
          </header>

          <nav
            className="flex min-h-12 shrink-0 items-center gap-2 overflow-x-auto border-b border-linkflow-border bg-linkflow-panel/90 px-4 dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel/90 sm:px-6"
            aria-label={t('adminVisitedTabs')}
          >
            {visitedTabs.map((tabId) => {
              const item = getNavItemById(tabId);
              const active = tabId === activeSectionId;

              return (
                <Link
                  key={tabId}
                  to={item.path}
                  aria-current={active ? 'page' : undefined}
                  className={[
                    'relative inline-flex h-8 shrink-0 items-center gap-2 rounded-md border px-3 text-sm font-semibold no-underline transition',
                    active
                      ? 'border-linkflow-primary bg-linkflow-primary-soft text-linkflow-primary'
                      : 'border-linkflow-border bg-white text-linkflow-muted hover:border-linkflow-primary hover:text-linkflow-primary dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-muted dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-primary',
                  ].join(' ')}
                >
                  {active ? (
                    <motion.span
                      layoutId="admin-active-tab"
                      className="absolute inset-x-3 -bottom-2 h-0.5 rounded-full bg-linkflow-primary"
                      transition={{ duration: 0.18, ease: 'easeOut' }}
                    />
                  ) : null}
                  <span className="text-base leading-none text-linkflow-primary">
                    {item.icon}
                  </span>
                  <span>{t(item.labelKey)}</span>
                  <button
                    type="button"
                    className={[
                      'ml-1 grid h-5 w-5 place-items-center rounded text-xs transition',
                      active
                        ? 'text-linkflow-primary hover:bg-linkflow-primary-soft'
                        : 'text-linkflow-muted hover:bg-linkflow-primary-soft hover:text-linkflow-primary dark:text-linkflow-dark-muted dark:hover:text-linkflow-dark-primary',
                    ].join(' ')}
                    aria-label={`${t('adminCloseTab')}: ${t(item.labelKey)}`}
                    onClick={(event) => handleCloseTab(event, tabId)}
                  >
                    <CloseOutlined aria-hidden="true" />
                  </button>
                </Link>
              );
            })}
          </nav>

          <div className="min-h-0 flex-1 overflow-y-auto p-4 sm:p-6 lg:p-8">
            <AnimatePresence mode="wait">
              <motion.section
                key={activeContent.id}
                initial={reduceMotion ? false : { opacity: 0, y: 8 }}
                animate={reduceMotion ? undefined : { opacity: 1, y: 0 }}
                exit={reduceMotion ? undefined : { opacity: 0, y: -8 }}
                transition={{ duration: 0.18, ease: 'easeOut' }}
              >
                {activeContent.id === 'tenants' ||
                activeContent.id === 'auditLogs' ? null : (
                  <section className="mb-6">
                    <p className="m-0 text-sm font-bold uppercase text-linkflow-primary">
                      {t('adminDashboardEyebrow')}
                    </p>
                    <h2 className="m-0 mt-2 text-[clamp(2rem,5vw,3.5rem)] font-bold leading-none tracking-normal">
                      {t(activeContent.titleKey)}
                    </h2>
                    {t(activeContent.summaryKey) !== '' ? (
                      <p className="m-0 mt-5 max-w-3xl text-base leading-7 text-linkflow-muted dark:text-linkflow-dark-muted sm:text-[17px]">
                        {t(activeContent.summaryKey)}
                      </p>
                    ) : null}
                  </section>
                )}

                {activeContent.id === 'tenants' ? (
                  <TenantManagement />
                ) : activeContent.id === 'auditLogs' ? (
                  <AuditLogManagement />
                ) : (
                  <>
                    <section
                      className="grid gap-4 [grid-template-columns:repeat(auto-fit,minmax(220px,1fr))]"
                      aria-label={t('adminOverviewTitle')}
                    >
                      {metrics.map((item) => (
                        <article
                          key={item.labelKey}
                          className="min-h-32 rounded-lg border border-linkflow-border bg-white p-5 shadow-[0_8px_24px_rgba(23,32,51,0.06)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/20"
                        >
                          <p className="m-0 text-sm font-bold text-linkflow-subtle dark:text-linkflow-dark-subtle">
                            {t(item.labelKey)}
                          </p>
                          <strong className="mt-3 block text-4xl leading-none">
                            {item.value}
                          </strong>
                          <p className="m-0 mt-3 leading-6 text-linkflow-subtle dark:text-linkflow-dark-subtle">
                            {t(item.descriptionKey)}
                          </p>
                        </article>
                      ))}
                    </section>

                    <section
                      className="mt-8"
                      aria-labelledby="admin-entry-title"
                    >
                      <h3
                        id="admin-entry-title"
                        className="m-0 mb-4 text-xl font-bold leading-tight"
                      >
                        {t('adminEntryTitle')}
                      </h3>
                      <div className="grid gap-4 [grid-template-columns:repeat(auto-fit,minmax(240px,1fr))]">
                        {navItems.slice(1).map((item) => (
                          <Link
                            key={item.id}
                            to={item.path}
                            onClick={() => addVisitedTab(item.id)}
                            className="block min-h-36 rounded-lg border border-linkflow-border bg-white p-5 text-linkflow-text no-underline shadow-[0_8px_24px_rgba(23,32,51,0.06)] transition hover:-translate-y-0.5 hover:border-linkflow-primary hover:text-linkflow-text hover:shadow-[0_12px_30px_rgba(18,137,153,0.14)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:text-linkflow-dark-text dark:shadow-black/20 dark:hover:border-linkflow-dark-primary dark:hover:text-linkflow-dark-text dark:hover:shadow-linkflow-dark-primary-soft"
                          >
                            <span className="mb-4 grid h-10 w-10 place-items-center rounded-md bg-linkflow-primary-soft text-xl text-linkflow-primary">
                              {item.icon}
                            </span>
                            <span className="block text-lg font-bold">
                              {t(item.labelKey)}
                            </span>
                            {item.id === 'settings' ? null : (
                              <p className="m-0 mt-3 leading-7 text-linkflow-subtle dark:text-linkflow-dark-subtle">
                                {item.id === 'tenants'
                                  ? t('adminEntryTenantsDescription')
                                  : item.id === 'auditLogs'
                                  ? t('adminEntryAuditLogsDescription')
                                  : t('adminEntryDeviceManagementDescription')}
                              </p>
                            )}
                          </Link>
                        ))}
                      </div>
                    </section>

                    {activeContent.id === 'settings' ? null : (
                      <section className="mt-8 rounded-lg border border-linkflow-border bg-white p-5 shadow-[0_8px_24px_rgba(23,32,51,0.06)] dark:border-linkflow-dark-border dark:bg-linkflow-dark-panel dark:shadow-black/20">
                        <h3 className="m-0 text-lg font-bold">
                          {t('adminContentAreaTitle')}
                        </h3>
                        <p className="m-0 mt-3 max-w-3xl leading-7 text-linkflow-subtle dark:text-linkflow-dark-subtle">
                          {t(activeContent.contentKey)}
                        </p>
                      </section>
                    )}
                  </>
                )}
              </motion.section>
            </AnimatePresence>
          </div>
        </section>
      </div>
    </main>
  );
};

export default AdminPage;
