import { defineConfig } from '@umijs/max';

export default defineConfig({
  antd: {},
  access: {},
  model: {},
  initialState: {},
  request: {},
  layout: {
    title: '@umijs/max',
  },
  routes: [
    {
      path: '/',
      redirect: '/admin',
    },
    {
      name: '管理',
      path: '/admin',
      component: './Admin',
      layout: false,
      wrappers: ['@/wrappers/AuthGuard'],
    },
    {
      name: '租户管理',
      path: '/admin/tenants',
      component: './Admin',
      layout: false,
      wrappers: ['@/wrappers/AuthGuard'],
    },
    {
      name: '设备管理',
      path: '/admin/device-management',
      component: './Admin',
      layout: false,
      wrappers: ['@/wrappers/AuthGuard'],
    },
    {
      name: '平台设置',
      path: '/admin/settings',
      component: './Admin',
      layout: false,
      wrappers: ['@/wrappers/AuthGuard'],
    },
    {
      name: '首页',
      path: '/home',
      component: './Home',
      layout: false,
      wrappers: ['@/wrappers/AuthGuard'],
    },
    {
      name: '登录',
      path: '/login',
      component: './Auth/Login',
      layout: false,
    },
    {
      name: '注册',
      path: '/register',
      component: './Auth/Register',
      layout: false,
    },
    {
      name: 'API文档',
      path: '/apidocs',
      component: './ApiDocs',
      layout: false,
      wrappers: ['@/wrappers/AuthGuard'],
    },
  ],
  proxy: {
    '/api': {
      target: 'http://127.0.0.1:18080',
      changeOrigin: true,
    },
  },
  npmClient: 'pnpm',
  utoopack: {},
});
