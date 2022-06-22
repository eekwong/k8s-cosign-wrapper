ARG GO_BUILDER_VERSION

FROM golang:${GO_BUILDER_VERSION} AS go-builder
COPY ./ /k8s-cosign-wrapper
RUN cd /k8s-cosign-wrapper && CGO_ENABLED=0 GOOS=linux make clean binary

FROM gcr.io/distroless/static:debug
COPY --from=go-builder /k8s-cosign-wrapper/dist/k8s-cosign-wrapper /
EXPOSE 8080
ENTRYPOINT [ "/k8s-cosign-wrapper" ]
