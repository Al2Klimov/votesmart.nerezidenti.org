FROM golang:alpine as build
RUN ["apk", "add", "upx"]
ADD . /api

WORKDIR /api
ENV CGO_ENABLED 0

RUN ["go", "build", "-ldflags", "-s -w", "."]
RUN ["upx", "api"]


FROM scratch
COPY rootfs/ /
COPY --from=build /api/api /api

USER nobody
CMD ["/api"]
