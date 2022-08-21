# k8s-chaos-monkey
Kubernetes CRD Chaos Monkey Testing

Publish a new version:
```
vi Makefile
--> change the version in IMG (v0.0.1)
make generate
make manifests
make docker-build
make docker-push
```

Debug steps
```
--Login to Kubernetes
make generate
make manifests
make install
make run
```

Generate deploy yaml:
```
make generate
make manifests
./bin/kustomize build config/default > deploy/podchaosmonkeys.yaml
```

Install steps:
```
--Login to Kubernetes
kubectl apply -f deploy/podchaosmonkeys.yaml
```

Uninstall steps:
```
--Login to Kubernetes
make undeploy
```

Demo sample:
```
--Login to Kubernetes
kubectl create namespace demo-chaos
kubectl apply -f examples/hamster.yaml -n demo-chaos
kubectl apply -f examples/hamster-chaos.yaml -n demo-chaos
```

Demo CRD:
```yaml
apiVersion: "testing.chaos.io/v1"
kind: PodChaosMonkey
metadata:
  name: hamster-chaos
spec:
  targetRef:
    apiVersion: "apps/v1"
    kind: Deployment
    name: hamster
  conditions:
    # Minimal number of pods which need to be alive
    minPods: 2
    # Minimal time in minutes that has to be running
    minTimeRunning: 10
    # Minimal Random time in minutes between two consecutive pod evictions
    minTimeRandom: 10
    # Maximum Random time in minutes between two consecutive pod evictions 
    maxTimeRandom: 30
```
