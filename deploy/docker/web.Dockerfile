FROM node:22-alpine AS builder

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

FROM node:22-alpine AS runtime

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
# Noto Serif SC (OTF) – CJK title & subtitle; Google Sans Code (TTF) – tags & site name.
# Avoid `ADD https://...`: BuildKit resolves those URLs as cache keys and often hits EOF on flaky links to GitHub.
# If still failing, try a mirror prefix, e.g. --build-arg GITHUB_PROXY=https://ghfast.top/
ARG GITHUB_PROXY=
RUN apk add --no-cache curl unzip fontconfig ca-certificates && \
    NOTO_URL="${GITHUB_PROXY}https://github.com/notofonts/noto-cjk/releases/download/Serif2.003/14_NotoSerifSC.zip" && \
    SANS_URL="${GITHUB_PROXY}https://github.com/googlefonts/googlesans-code/releases/download/v6.001/GoogleSansCode-v6.001.zip" && \
    curl -fL --connect-timeout 30 --retry 8 --retry-all-errors --retry-delay 3 -o /tmp/NotoSerifSC.zip "$NOTO_URL" && \
    curl -fL --connect-timeout 30 --retry 8 --retry-all-errors --retry-delay 3 -o /tmp/GoogleSansCode.zip "$SANS_URL" && \
    mkdir -p /usr/share/fonts/og && \
    unzip -q -j /tmp/NotoSerifSC.zip 'SubsetOTF/SC/NotoSerifSC-Regular.otf' 'SubsetOTF/SC/NotoSerifSC-Bold.otf' -d /usr/share/fonts/og/ && \
    unzip -q -j /tmp/GoogleSansCode.zip 'variable/*.ttf' -d /usr/share/fonts/og/ && \
    fc-cache -f && \
    rm -f /tmp/NotoSerifSC.zip /tmp/GoogleSansCode.zip

COPY deploy/docker/renderer-entrypoint.sh /usr/local/bin/renderer-entrypoint.sh
RUN chmod +x /usr/local/bin/renderer-entrypoint.sh

EXPOSE 3000

ENTRYPOINT ["/usr/local/bin/renderer-entrypoint.sh"]
