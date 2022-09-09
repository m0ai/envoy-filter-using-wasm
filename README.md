# WASM Based Envoy Filter Example

request 헤더에 포함된 `x-username` 을 확인하여 이를 response 본문에 추가하는 예제입니다.

### Requirements
- Docker
- [Taskfile](https://taskfile.dev/installation/)
- Go >= 1.9

### Quick Start

**terminal session 1**

```shell
$ task build
task: [build] tinygo build -o main.wasm -scheduler=none -target=wasi ./main.go

$ file main.wasm
main.wasm: WebAssembly (wasm) binary module version 0x1 (MVP)

$ task run
task: [build] tinygo build -o main.wasm -scheduler=none -target=wasi ./main.go
task: [run] docker run -it --rm  ...

[2022-09-09 18:13:19.034][1][info][main] [source/server/server.cc:368] initializing epoch 0 (base id=0, hot restart version=11.120)
[2022-09-09 18:13:19.034][1][info][main] [source/server/server.cc:370] statically linked extensions
```

**terminal session 2**

```shell
$ curl localhost:18000 --data "Hello World"
Hello World, To Anonymous
~~~~
$ curl localhost:18000 --data "Hello World" -H "x-username: m0ai"
Hello World, To m0ai
```

## TODO
- [ ] Add Github Actions
  - [ ] build -> run -> test pipeline
  - [ ] test code ( test.main.go )
  - [ ] e2e test ( running on envoy container )
- [ ] Write Kubernetes Example
  - [ ] k8s/istio sidecar volume injection