ARG NODE_VERSION=24.12.0

FROM node:${NODE_VERSION}-alpine AS base

# 作業ディレクトリの設定
WORKDIR /usr/src/app

FROM base AS dev

ENV NODE_ENV development

RUN --mount=type=bind,source=package.json,target=package.json \
    --mount=type=bind,source=package-lock.json,target=package-lock.json \
    --mount=type=cache,target=/root/.npm \
    npm ci

EXPOSE 3000

CMD ["sleep", "infinity" ]