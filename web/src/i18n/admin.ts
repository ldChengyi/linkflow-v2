export const adminMessages = {
  'zh-CN': {
    adminConsole: '管理控制台',
    adminHeaderDescription: '租户、设备和平台配置的管理入口。',
    adminGoHome: '返回首页',
    adminNotifications: '通知',
    adminLogout: '退出',
    adminToggleDark: '切换到暗色模式',
    adminToggleLight: '切换到亮色模式',
    adminLanguage: '语言',
    adminExpandNav: '展开导航',
    adminCollapseNav: '收起导航',
    adminVisitedTabs: '访问记录',
    adminCloseTab: '关闭标签',
    adminOverviewNav: '概览',
    adminTenantsNav: '租户管理',
    adminDeviceManagementNav: '设备管理',
    adminSettingsNav: '设置',
    adminDashboardEyebrow: 'Dashboard',
    adminOverviewTitle: '平台管理台',
    adminOverviewSummary:
      '这里保持后台的导航和内容区结构。内容卡片采用和学习首页相近的白底、细边框、轻阴影和纯色强调，后续可以逐块替换为真实租户、设备和平台配置数据。',
    adminTenantsTitle: '租户管理',
    adminTenantsSummary:
      '租户管理作为父级入口，后续可以承载租户列表、成员、角色和资源隔离配置。',
    adminDeviceManagementTitle: '设备管理',
    adminDeviceManagementSummary:
      '设备管理作为父级入口，后续可以承载设备列表、产品绑定、在线状态和调试工具。',
    adminSettingsTitle: '平台设置',
    adminSettingsSummary: '',
    adminMetricOnlineDevices: '在线设备',
    adminMetricTenants: '租户数量',
    adminMetricActiveProducts: '活跃产品',
    adminMetricWaitingDevices: '等待接入真实设备状态接口。',
    adminMetricWaitingTenants: '等待接入租户统计接口。',
    adminMetricWaitingProducts: '等待接入产品管理接口。',
    adminEntryTitle: '管理入口',
    adminEntryTenantsDescription: '租户列表、成员和角色权限入口会放在这里。',
    adminEntryDeviceManagementDescription:
      '设备列表、在线状态、产品绑定和调试入口会放在这里。',
    adminEntrySettingsDescription: '',
    adminContentAreaTitle: '当前工作区',
    adminOverviewContent:
      '下一步应该从租户管理或设备管理中选择一个父级入口继续拆子路由。',
    adminTenantsContent:
      '租户管理模块建议先实现租户列表，再增加成员、角色和权限配置。',
    adminDeviceManagementContent:
      '设备管理模块建议先实现设备列表，再增加设备详情、产品绑定和调试入口。',
    adminSettingsContent: '',
  },
  'en-US': {
    adminConsole: 'Admin Console',
    adminHeaderDescription: 'Manage tenants, devices, and settings.',
    adminGoHome: 'Back home',
    adminNotifications: 'Notifications',
    adminLogout: 'Logout',
    adminToggleDark: 'Switch to dark mode',
    adminToggleLight: 'Switch to light mode',
    adminLanguage: 'Language',
    adminExpandNav: 'Expand navigation',
    adminCollapseNav: 'Collapse navigation',
    adminVisitedTabs: 'Visited tabs',
    adminCloseTab: 'Close tab',
    adminOverviewNav: 'Overview',
    adminTenantsNav: 'Tenant management',
    adminDeviceManagementNav: 'Device management',
    adminSettingsNav: 'Settings',
    adminDashboardEyebrow: 'Dashboard',
    adminOverviewTitle: 'Platform dashboard',
    adminOverviewSummary:
      'The admin keeps a dashboard layout while using clean cards with solid colors, thin borders, and light shadows.',
    adminTenantsTitle: 'Tenant management',
    adminTenantsSummary:
      'This parent route will hold tenants, members, roles, and isolation settings.',
    adminDeviceManagementTitle: 'Device management',
    adminDeviceManagementSummary:
      'This parent route will hold devices, product bindings, online status, and debugging tools.',
    adminSettingsTitle: 'Platform settings',
    adminSettingsSummary: '',
    adminMetricOnlineDevices: 'Online devices',
    adminMetricTenants: 'Tenants',
    adminMetricActiveProducts: 'Active products',
    adminMetricWaitingDevices: 'Waiting for the device status API.',
    adminMetricWaitingTenants: 'Waiting for the tenant metrics API.',
    adminMetricWaitingProducts: 'Waiting for the product API.',
    adminEntryTitle: 'Admin entries',
    adminEntryTenantsDescription:
      'Tenant lists, members, and role permissions will live here.',
    adminEntryDeviceManagementDescription:
      'Device lists, status, product binding, and debugging will live here.',
    adminEntrySettingsDescription:
      '',
    adminContentAreaTitle: 'Current workspace',
    adminOverviewContent:
      'Next, choose tenant management or device management and split it into child routes.',
    adminTenantsContent:
      'Start tenant management with a tenant list, then add members, roles, and permission settings.',
    adminDeviceManagementContent:
      'Start device management with a device list, then add details, product binding, and debugging tools.',
    adminSettingsContent: '',
  },
} as const;

export type AdminMessageKey = keyof (typeof adminMessages)['zh-CN'];
