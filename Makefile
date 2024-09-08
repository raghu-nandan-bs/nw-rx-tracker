.PHONY: objects
objects:
	# comment import _ "github.com/cilium/ebpf/cmd/bpf2go" in pkg/bpf/gen.nw_mb.go
	sed -i 's/\/\/import _ "github.com\/cilium\/ebpf\/cmd\/bpf2go"/import _ "github.com\/cilium\/ebpf\/cmd\/bpf2go"/' pkg/bpf/gen.nw_mb.go
	go mod vendor
	go generate ./pkg/bpf
	sed -i 's/^import _ "github.com\/cilium\/ebpf\/cmd\/bpf2go"/\/\/import _ "github.com\/cilium\/ebpf\/cmd\/bpf2go"/' pkg/bpf/gen.nw_mb.go

.PHONY: nw-rx-tracker
nw-rx-tracker: objects
	go mod vendor
	go build -o nw-rx-tracker


.PHONY: release
release:
	docker build -t nw-rx-tracker .
	mkdir -p release
	DOCKER_ID=$$(docker create nw-rx-tracker) && \
		docker cp $${DOCKER_ID}:/app/nw-rx-tracker release/ && \
		docker rm $${DOCKER_ID}

