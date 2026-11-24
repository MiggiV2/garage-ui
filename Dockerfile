# syntax=docker/dockerfile:1.7

FROM --platform=$BUILDPLATFORM node:22-alpine AS web
WORKDIR /src/web
COPY web/package.json web/package-lock.json ./
RUN --mount=type=cache,target=/root/.npm npm ci --no-audit --no-fund
COPY web/ ./
RUN npm run build

FROM --platform=$BUILDPLATFORM golang:1.27-alpine AS server
ARG TARGETOS TARGETARCH
WORKDIR /src
COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod go mod download
COPY main.go ./
COPY internal/ internal/
COPY --from=web /src/web/dist web/dist
RUN --mount=type=cache,target=/go/pkg/mod --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags="-s -w" -o /out/garage-ui .

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=server /out/garage-ui /garage-ui
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/garage-ui"]
