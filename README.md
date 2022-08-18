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
make install
make run
```

Install steps:
```
--Login to Kubernetes
make deploy
```

Uninstall steps:
```
--Login to Kubernetes
make undeploy
```

