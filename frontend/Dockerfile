FROM node:26.4.0-alpine3.23 AS build

WORKDIR /usr/src/app

RUN --mount=type=bind,source=package.json,target=package.json \
    --mount=type=bind,source=package-lock.json,target=package-lock.json \
    --mount=type=cache,target=/root/.npm \
    npm ci

COPY . .
RUN npm run build

FROM busybox:1.38.0-musl
COPY --from=build /usr/src/app/dist /dist
CMD ["sh", "-c", "rm -rf /output/* && cp -r /dist/. /output/"]
