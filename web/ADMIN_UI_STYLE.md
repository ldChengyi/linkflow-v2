# LinkFlow Admin UI Style

This document records the current admin UI direction so future frontend work can continue in the same style.

## Stack

- Framework: Umi Max with React and TypeScript.
- Styling: Tailwind CSS utilities. Prefer Tailwind classes in components for the admin UI instead of adding page-specific Less modules.
- Icons: `@ant-design/icons`, imported by icon name.
- Animation: `motion/react` for small route/tab transitions. Keep animations light and respect `useReducedMotion`.
- UI state: Zustand persist store in `src/stores/adminUiStore.ts`.

## Layout

- Admin pages use a fixed shell:
  - left sidebar for primary navigation,
  - top navbar for global controls,
  - tab bar below the navbar for visited route records,
  - right content area as the only scrollable region.
- The root admin shell should stay `h-screen overflow-hidden`.
- The content area should stay `min-h-0 flex-1 overflow-y-auto`.
- The sidebar and navbar should not scroll with large content pages.

## Sidebar

- Default sidebar is narrow and icon-only.
- The expand/collapse button lives at the far left of the top navbar, not inside the sidebar.
- Desktop behavior:
  - collapsed: icon-only sidebar,
  - expanded: wider sidebar with menu names.
- Mobile behavior:
  - top-left button opens/closes the sidebar drawer,
  - sidebar remains icon-only to avoid squeezing mobile content.
- Active route uses a solid primary accent and a small animated left indicator.

## Navbar

- Do not show user email/role text in the navbar.
- Right-side navbar controls are icon-style controls only:
  - dark mode toggle,
  - i18n language dropdown,
  - logout.
- Navbar controls should use 44px touch targets.
- Keep navbar visual style quiet: white/panel background, thin border, muted icons, primary hover color.

## Tabs

- A visited-tabs bar sits directly under the navbar.
- Clicking a sidebar route or admin entry adds the route to visited tabs.
- Tabs are persisted through Zustand.
- Tabs display the route icon, route label, and a close `x` button.
- Closing a tab should not trigger tab navigation.
- If the active tab is closed, navigate to the last remaining tab. If no useful tab remains, fall back to overview.
- Tabs should support horizontal scrolling instead of wrapping or compressing.

## Routes

Current admin route model:

- `/admin`: overview
- `/admin/tenants`: tenant management parent route
- `/admin/tenants/products`: product management route
- `/admin/tenants/thingsmodels`: thing model management route
- `/admin/tenants/devices`: device management route
- `/admin/audit-logs`: audit log list route
- `/admin/settings`: settings route

Do not restore the old standalone event route unless product scope changes.

## Visual Style

- The admin is a management console, not a landing page.
- Keep the shell operational and dense enough for repeated use.
- Cards should inherit the learning homepage's card feeling:
  - solid white or dark panel background,
  - 8px radius,
  - thin border,
  - light shadow,
  - primary color accents,
  - no decorative gradient/orb backgrounds.
- Page content should use restrained headings and practical copy.
- Use responsive grids with `repeat(auto-fit,minmax(...))` for card groups.

## Dark Mode And I18n

- Dark mode and locale are persisted in `linkflow.admin-ui`.
- Use `useTheme()` and `useI18n()` instead of local component state.
- Admin text should be added to `src/i18n/admin.ts`.
- Keep Chinese and English keys in sync.

## State Boundaries

- Persist global UI preferences in `src/stores/adminUiStore.ts`:
  - dark mode,
  - locale,
  - sidebar expanded state,
  - visited tabs.
- Keep temporary mobile drawer open/close state local to `AdminPage`; it should not persist across refreshes.
- Do not use server-state patterns until real API data is introduced.

## Current Build Note

`npm run build` requires Node 20+ because Umi utoopack enforces that. The project `.nvmrc` is `22`.
