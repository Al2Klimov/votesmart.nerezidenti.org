FROM golang:alpine as build
RUN ["apk", "add", "upx"]

RUN ["mkdir", "/api"]
COPY go.mod go.sum /api/
WORKDIR /api
RUN ["go", "mod", "download"]

ADD . /api
ENV CGO_ENABLED 0
RUN ["go", "build", "-ldflags", "-s -w", "."]
RUN ["upx", "api"]


FROM scratch
COPY rootfs/ /
COPY --from=build /api/api /api

USER nobody
CMD ["/api"]
