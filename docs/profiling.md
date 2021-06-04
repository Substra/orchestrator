# Profiling the orchestrator

You might need to get a deep insight on what the orchestrator is doing.

This is possible thanks to [pprof](https://golang.org/pkg/runtime/pprof/).
It may need graphviz for some outputs, please refer to pprof's documentation for the installation process.

The orchestrator exposes an HTTP endpoint that pprof can hit to start and retrieve a profile.
The default port is `8484` and it is not exposed as a service by the kubernetes deployment, so you need to rely on port forward.

You'll find below the steps to obtain a profile in a canonical dev environment (adjust values to your own).

## Forward port 8484

```sh
kubectl port-forward owkin-orchestrator-org-1-server-77bddf65b-nfxt7 8484:8484
```

## Start profiling

This will take samples for 30 seconds, adjust to the workload:

```sh
go tool pprof http://localhost:8484/debug/pprof/profile\?seconds\=30
```

## Analyze the profile

Assuming you let pprof write in its default category:

```sh
# render the call graph as png
go tool pprof -png $HOME/pprof/pprof.orchestrator.samples.cpu.001.pb.gz
# or expose a web UI
go tool pprof -http=localhost:8090 $HOME/pprof/pprof.orchestrator.samples.cpu.001.pb.gz
```
