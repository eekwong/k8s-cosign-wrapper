match[{"msg": msg}] {
    input.request.operation == "CREATE"
    input.request.kind.kind == "Pod"
    input.request.resource.resource == "pods"
    distroless_images := [ container | container := input.request.object.spec.containers[_]
        startswith(container.image, "gcr.io/distroless/") ]
    count(distroless_images) > 0
    verified :=  [ container | container := distroless_images[_]
        body := { "image": container.image }
        response := http.send({
            "method": "POST",
            "url": "http://cosign.k8s-cosign-wrapper/verify",
            "body": body})
        response.status_code == 200 ]
    count(verified) < count(distroless_images)
    msg := sprintf("number of verified %v < number of distroless images %v", [count(verified), count(distroless_images)])
}
