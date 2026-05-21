FROM node:22-alpine AS build

ARG NPM_REGISTRY=https://registry.npmmirror.com

WORKDIR /src/web

RUN corepack enable && corepack prepare pnpm@latest --activate \
    && npm config set registry "${NPM_REGISTRY}" \
    && pnpm config set registry "${NPM_REGISTRY}"

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile --ignore-scripts

COPY web ./
RUN pnpm run setup && pnpm run build

FROM nginx:1.27-alpine

COPY deploy/docker/web.nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /src/web/dist /usr/share/nginx/html

EXPOSE 80
