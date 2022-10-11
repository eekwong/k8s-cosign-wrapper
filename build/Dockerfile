ARG GO_BUILDER_VERSION

FROM golang:${GO_BUILDER_VERSION} AS go-builder
ARG COSIGN_VERSION
COPY ./ /k8s-cosign-wrapper
RUN cd /k8s-cosign-wrapper && \
    CGO_ENABLED=0 GOOS=linux make clean binary && \
    wget -O /cosign https://github.com/sigstore/cosign/releases/download/v${COSIGN_VERSION}/cosign-linux-amd64 && \
    chmod a+x /cosign

FROM alpine:3.16.2
COPY --from=go-builder /cosign /
COPY --from=go-builder /k8s-cosign-wrapper/dist/k8s-cosign-wrapper /
EXPOSE 8080
ENTRYPOINT [ "/k8s-cosign-wrapper" ]
