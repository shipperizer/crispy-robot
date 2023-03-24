FROM --platform=$BUILDPLATFORM golang:1.19 AS builder

ARG SKAFFOLD_GO_GCFLAGS
ARG TARGETOS
ARG TARGETARCH

ENV GO111MODULE=on
ENV CGO_ENABLED=0
ENV GOOS=$TARGETOS
ENV GOARCH=$TARGETARCH
ENV GO_BIN=/go/bin/app

RUN apt-get update
RUN apt-get install -y build-essential git unzip curl wget file

WORKDIR /var/app

COPY . .

ARG app_name=web

ENV APP_NAME=$app_name

RUN make build

FROM gcr.io/distroless/base

COPY --from=builder /go/bin/app /app

CMD ["/app"]
