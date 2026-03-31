ARG NODE_IMAGE=
FROM ${NODE_IMAGE:-node}:22-alpine AS builder

# Corepack downloads pnpm via undici → often ECONNRESET behind flaky TLS (e.g. to registry.npmjs.org).
# Install pnpm with npm instead; set registry when building in CN:
#   docker build --build-arg NPM_REGISTRY=https://registry.npmmirror.com ...
ARG NPM_REGISTRY=https://registry.npmjs.org
ARG PNPM_VERSION=10.33.0
ENV NPM_CONFIG_REGISTRY=${NPM_REGISTRY}

WORKDIR /app

RUN npm install -g pnpm@${PNPM_VERSION}

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile --prod=false

COPY web/. .
COPY shared /shared

ARG APP_VERSION=dev
ARG BUILD_COMMIT=unknown
ENV APP_VERSION=${APP_VERSION} \
    BUILD_COMMIT=${BUILD_COMMIT}

RUN pnpm build

ARG NODE_IMAGE=
FROM ${NODE_IMAGE:-node}:22-alpine AS runtime

ARG NPM_REGISTRY=https://registry.npmjs.org
ARG PNPM_VERSION=10.33.0
ENV NPM_CONFIG_REGISTRY=${NPM_REGISTRY}

WORKDIR /app

ARG APP_VERSION=dev
ARG BUILD_COMMIT=unknown
ENV NODE_ENV=production \
    APP_VERSION=${APP_VERSION} \
    BUILD_COMMIT=${BUILD_COMMIT}

RUN npm install -g pnpm@${PNPM_VERSION}

COPY web/package.json web/pnpm-lock.yaml web/pnpm-workspace.yaml ./
RUN pnpm install --frozen-lockfile --prod

COPY --from=builder /app/build /app/build

# Install fonts for OG image rendering (resvg on musl cannot load woff2).
# Fonts are bundled in repo under deploy/font/ to avoid slow/flaky downloads on servers.
# Put .otf/.ttf files there (e.g. NotoSerifSC + GoogleSansCode) before building.
COPY deploy/font/ /usr/share/fonts/og/
RUN apk add --no-cache fontconfig ca-certificates \
  && fc-cache -f

COPY deploy/docker/renderer-entrypoint.sh /usr/local/bin/renderer-entrypoint.sh
RUN sed -i 's/\r$//' /usr/local/bin/renderer-entrypoint.sh \
  && chmod +x /usr/local/bin/renderer-entrypoint.sh

EXPOSE 3000

ENTRYPOINT ["/usr/local/bin/renderer-entrypoint.sh"]
