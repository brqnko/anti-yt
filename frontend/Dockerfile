ARG NODE_VERSION=24.12.0
ARG PNPM_VERSION=10.28.0

FROM node:${NODE_VERSION}-alpine AS base

RUN --mount=type=cache,target=/root/.npm \
    npm install -g pnpm@${PNPM_VERSION}

WORKDIR /usr/src/app

FROM base AS dev

ENV NODE_ENV development

RUN --mount=type=bind,source=package.json,target=package.json \
    --mount=type=bind,source=pnpm-lock.yaml,target=pnpm-lock.yaml \
    --mount=type=cache,target=/root/.local/share/pnpm/store \
    pnpm install --dev --frozen-lockfile

EXPOSE 3000

