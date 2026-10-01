# Local kind smoke environment

## Develop the complete stack from source

Kind is managed by its CLI rather than as a Compose service. The existing
`backend/docker-compose-dev.yml` starts MySQL and Vault, while Backend,
Clabgate and Vite run directly from the working tree for fast rebuilds and
debugger support. An nginx gateway inside kind keeps the browser, API,
JupyterLab and ttyd on one origin.

Install Docker, kind, kubectl and Helm, then run from the repository root:

```bash
make dev-up
```

This starts the existing Compose dependencies, creates (or reuses)
`cms-labs-local`, installs the CMS Labs Clabernetes chart `0.8.0-4`, creates
the `cms-labs-system` namespace and installs the development gateway. The two
equivalent direct cluster commands are:

```bash
./k8s/local-kind/up.sh --dev
./k8s/local-kind/up.sh --local
```

Start the applications in separate terminals:

```bash
make dev-backend
make dev-seed
make dev-clabgate
make dev-front
```

`make dev-seed` is a one-shot command and may be run before or after the regular
Backend process. Clabgate first tries its in-cluster ServiceAccount and then the
standard `KUBECONFIG`/`~/.kube/config`, so the same binary works locally, in CI,
and inside Kubernetes. The bootstrap writes an isolated development kubeconfig
to `.local/kubeconfig`; `make dev-clabgate` selects it automatically. Its default
local endpoints are Backend on port `5000`, Clabgate on port `5001` and Vite
on port `5183`.

Open the application through `http://127.0.0.1:18080`, not directly through
Vite. The kind NodePort reaches nginx, which routes `/api/` and
`/clabgate/api/` to the host processes and resolves per-session JupyterLab and
ttyd Services inside Kubernetes. This preserves cookies, iframes and WebSocket
upgrades on the same browser origin.

The gateway does not maintain a copy of the production proxy rules. During
`up.sh --dev`, its ConfigMap is generated from
`nextui-dashboard/nginx/templates/default.conf.template`; only the frontend
include is replaced with a Vite proxy. Re-run `up.sh --dev` after changing the
nginx template. Running `up.sh` without either flag leaves the gateway out and
is suitable for the image-based smoke environment.

To remove the local dependencies and disposable cluster:

```bash
make dev-down
```

The local Bearer token is validated by calling CMS
`/api/v1/sso/userinfo`; Clabgate does not need a copy of the CMS JWT public key.
The workspace cookie check remains separate because Jupyter and ttyd traffic is
proxied directly to namespace Services and does not pass through the Clabgate
JSON-RPC handlers.

This is a disposable self-hosted Kubernetes environment for the complete
Clabgate session lifecycle. It runs two real Clabgate replicas against the
Kubernetes API and the upstream Clabernetes controller. A stateful boundary
mock replaces CMS and GitLab; small protocol-compatible images replace the
production Jupyter and checker images.

The same lifecycle runs for every pull request in
`.github/workflows/e2e.yml` on an ephemeral GitHub-hosted kind cluster. No
production Kubernetes credentials are used.

The smoke topology includes a `client` node that must access the Docker daemon
inside the kind node. Clabernetes therefore mounts the host socket into the
generated pod via:

```yaml
node:
  kind: linux
  image: ghcr.io/srl-labs/alpine
  clabernetes:
    mountDockerSock: true
```

The fixed attempt is `00000000-0000-0000-0000-000000000000`. The frontend is
published at `http://127.0.0.1:18080`.

## Tested versions

- kind `0.33.0`;
- Kubernetes `1.33.12`;
- CMS Labs Clabernetes chart `0.8.0-4`;
- Docker Desktop on `linux/arm64`.

The same commands work on `amd64` when `GOARCH` and Docker `--platform` are
changed accordingly.

## Bootstrap the cluster

From the repository root:

```bash
kind create cluster \
  --config k8s/local-kind/kind.yaml \
  --image kindest/node:v1.33.12@sha256:3f5c8443c620245e4d355cfe09e96a91ead32ceaa569d3f1ca9edf0cb2fe2ff4

helm upgrade --install clabernetes \
  oci://ghcr.io/cms-lab-core/cms-labs-clabernetes/clabernetes \
  --version 0.8.0-4 \
  --namespace c9s \
  --create-namespace \
  --kube-context kind-cms-labs-local

kubectl wait --for=condition=Available deployment \
  --all -n c9s --context kind-cms-labs-local --timeout=5m
```

Deployment readiness is checked explicitly after Helm installation so the same
sequence works in local development and CI.

If the node inherits a host proxy bound to `127.0.0.1`, containerd cannot use
that address from inside the node container. Configure Docker Desktop with a
proxy reachable as `host.docker.internal`, or set the corresponding systemd
environment on `cms-labs-local-control-plane` before pulling images.

## Build and load local images

The following commands are for Apple Silicon / ARM64:

```bash
mkdir -p clabgate/build clabgate/e2e/mock/build clabgate/e2e/jupyter/build

(cd clabgate && \
  CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o build/clabgate main.go && \
  cp start.sh build/start.sh)
(cd clabgate/e2e/mock && \
  CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o build/smoke-mock .)
(cd clabgate/e2e/jupyter && \
  CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o build/jupyter-smoke .)
(cd nextui-dashboard && ./node_modules/.bin/vite build)

docker build --platform linux/arm64 -t cms-labs/clabgate:local clabgate
docker build --platform linux/arm64 -t cms-labs/smoke-mock:local clabgate/e2e/mock
docker build --platform linux/arm64 -t cms-labs/jupyter-smoke:local clabgate/e2e/jupyter
docker build --platform linux/arm64 -t cms-labs/checker-smoke:local clabgate/e2e/checker
docker build --platform linux/arm64 -t cms-labs/front:local nextui-dashboard

kind load docker-image --name cms-labs-local \
  cms-labs/clabgate:local \
  cms-labs/smoke-mock:local \
  cms-labs/jupyter-smoke:local \
  cms-labs/checker-smoke:local \
  cms-labs/front:local

kubectl apply --context kind-cms-labs-local -f k8s/local-kind/smoke.yaml
```

After rebuilding images in an existing cluster, reset only the disposable
smoke state and restart its deployments:

```bash
./k8s/local-kind/reset-smoke.sh
```

The manifest uses `imagePullPolicy: Never` for top-level smoke components, so
these images cannot accidentally be pulled into a production cluster. The
session controller uses `IfNotPresent`; therefore rebuilds under the same tag
must be followed by `kind load docker-image` and replacement of the old Pod.

## Verify end to end

The verifier sends a short-lived test Bearer token; the mock CMS resolves it to
student 42 through `/api/v1/sso/userinfo`. It does not persist or print the
token, workspace grant or cookie.

```bash
./k8s/local-kind/verify.sh
```

It checks all production-relevant boundaries:

- two Clabgate replicas are Ready and one holder owns Lease
  `clabgate-session-reconciler`;
- namespace `lab-00000000-0000-0000-0000-000000000000`, Jupyter, PVC, Service
  and Clabernetes Topology are Ready;
- mock CMS moves from `pending` to `active` only after readiness;
- direct workspace access without a grant returns HTTP 401;
- `session.open` exchanges a short-lived grant for a scoped cookie and reaches
  the namespace-local Jupyter Service;
- `session.check` creates a Job and CMS receives a new `check_id`, score,
  report, structured per-task logs and bounded raw Pod logs.

The successful run on 2026-09-22 returned score `9/10`, report
`kind end-to-end smoke passed`, and checker Pod logs. Clabernetes Topology
`smoke` reported `TopologyReady=True` and `topologyState=running`.

## Inspect and remove

```bash
kubectl get pods -A --context kind-cms-labs-local
kubectl get topology -A --context kind-cms-labs-local
kubectl get lease -n cms-labs-system --context kind-cms-labs-local
curl --noproxy '*' http://127.0.0.1:18080/api/state

kind delete cluster --name cms-labs-local
```

The last command removes the entire disposable cluster, including the session
PVC and all smoke data.
