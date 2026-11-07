# syntax=docker/dockerfile:1

FROM --platform=$BUILDPLATFORM node:26-alpine AS web-deps
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN npm ci

FROM web-deps AS web
COPY web/ ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS go-deps
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download

FROM go-deps AS backend
ARG TARGETOS TARGETARCH
COPY backend/ ./
COPY --from=web /src/backend/web/dist ./web/dist
RUN --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/autovolume .

FROM gcr.io/distroless/static-debian13:nonroot
WORKDIR /app
COPY --from=backend /out/autovolume ./
COPY backend/config.json ./
EXPOSE 8080
ENTRYPOINT ["/app/autovolume"]
