# Clabgate: production plan

This document is the implementation contract for replacing JupyterHub with one
Kubernetes namespace per CMS/LTI attempt. It describes the first production
release; later collaboration work must not weaken its isolation model.

## Accepted decisions

- The existing CMS server/client with type `k8s` remains the owner of attempts.
- Clabgate is responsible for the complete namespace: task manifests,
  Clabernetes topology, JupyterLab, checker jobs, status and deletion.
- The CMS routing `labs_path` carries a full GitLab or GitHub repository URL,
  optionally followed by `#ref`; `CMS_TASK_URL` and `CMS_TASK_BRANCH` remain
  fallbacks. `test_path` is optional, is never fetched or executed by Clabgate,
  and is forwarded to the trusted checker image as a compile-time registry
  selector.
- A CMS attempt stays `pending` while resources are being provisioned. It becomes
  `active` only after Clabgate observes the topology (when present) and Jupyter
  Deployment as ready. Kubernetes is the runtime source of truth.
- One attempt has one owner and one JupyterLab in this release. A future
  `collaboration` value greater than one will map several authorized CMS users
  to the same attempt; it is intentionally not implemented yet.
- Students and instructors can start the checker. A checker result contains a
  stable `check_id`, score, human-readable report and bounded logs. CMS remains
  responsible for LTI/Moodle grade synchronization.
- The production deployment keeps two Clabgate replicas. Only the elected
  leader runs background reconciliation; both replicas serve HTTP requests.

## Session state machine

```text
CMS pending
    -> namespace/provisioning
    -> namespace/ready + CMS active
    -> CMS terminating
    -> namespace deleted + CMS completed
```

The namespace has `labs.cmslabs.ru/session-phase` and
`labs.cmslabs.ru/ready-at` annotations. Namespace existence alone never means
that an attempt is active. A failed/degraded topology also never activates it.
All transitions are monotonic and retryable. Repeated reconciliation or a
restarted Clabgate must converge on the same objects.

## Task source and allowed manifests

Clabgate uses the GitLab or GitHub Repository API selected from the routing URL.
It resolves the selected ref, recursively lists `*.template.yaml` and
`*.template.yml`, sorts file paths, downloads the exact files and records the
resolved commit SHA in the namespace. The suffix prevents unrelated CI and
local demo manifests in the same repository from becoming session resources.

The namespaced allow-list is deliberately small:

- `v1/ConfigMap`, `ServiceAccount`, `Service`;
- `apps/v1/Deployment`;
- `rbac.authorization.k8s.io/v1` `Role` and `RoleBinding`;
- `networking.k8s.io/v1/NetworkPolicy`;
- exactly one `c9s.run/v1alpha1/Topology`.

Cluster-scoped objects, Secrets and unknown GVK are rejected. Deployments may
not use host namespaces, hostPath, hostPort, privileged containers, privilege
escalation or added capabilities. Services must remain ClusterIP; RBAC is
limited to same-namespace ServiceAccounts and Roles without wildcards or
non-resource URLs. Clabgate validates the complete bundle before the first
write, overwrites namespace and ownership labels, and tracks supplied
Deployments as part of session readiness. It does not know any
`cms-labs-terminal` configuration schema.

Private projects use optional `TASK_REPOSITORY_TOKEN` (`GITLAB_TOKEN` remains a
legacy fallback); the token is sent only to the
host from `CMS_TASK_URL` and is never copied into a laboratory namespace.

## JupyterLab access

JupyterLab listens only on its ClusterIP Service and runs without its own bearer
token. Browser access is protected by the application gateway:

1. an authenticated CMS user calls `session.open`;
2. Clabgate verifies attempt ownership (or instructor/admin role) and returns a
   short-lived, single-purpose signed grant;
3. the browser exchanges the grant for a `Secure`, `HttpOnly`, `SameSite=Lax`
   workspace cookie and is redirected to a clean URL;
4. nginx checks every HTTP and WebSocket request through Clabgate
   `auth_request` before proxying to the namespace Service.

This reuses the current CMS JWT/OIDC login without exposing a long-lived token
in URLs, browser history or Jupyter logs. `WORKSPACE_AUTH_SECRET` is shared by
both Clabgate replicas and must come from Vault.

Jupyter Pods and checker Pods set `automountServiceAccountToken: false`.
Jupyter gets only its PVC. Checker is a network/topology check in this release
and does not mount the RWO notebook PVC.

Laboratory checks live in the separate `github.com/cms-lab-core/cms-labs-checker`
repository. Each `labs/<name>` package implements the stable `LabChecker`
interface and owns its `_test.go` files. The packages are linked into one static
binary through an explicit registry; Go plugins are deliberately avoided because
their compiler/dependency ABI requirements make container releases fragile.

## Checker result contract

The checker writes one JSON object to `/dev/termination-log`:

```json
{
  "max_score": 10,
  "current_score": 8,
  "result_display": "8/10 checks passed",
  "report": "Optional Markdown summary",
  "tasks": [{
    "title": "SSH connectivity",
    "description": "Router accepts SSH connections",
    "logs": [{"node": "r1", "message": "connected"}],
    "complete": true
  }]
}
```

Clabgate adds the generated `check_id` and bounded Pod logs, then sends the
result to CMS. CMS stores these fields in `LTIAttemptResult` and performs the
existing LTI grade sync. Re-delivery with the same `check_id` is idempotent.
Checker Jobs are retained temporarily for inspection and then removed by the
Kubernetes TTL controller.

## Production gates

- Unit tests cover GitLab/GitHub path handling, manifest ordering/allow-list, readiness
  transitions, workspace grant validation, checker selection and each registered
  laboratory implementation.
- `go test ./...`, frontend typecheck/lint/build and Helm rendering pass.
- RBAC includes only namespaced session resources, Namespace lifecycle and the
  Lease used for leader election.
- Resource requests/limits and probes exist for Clabgate, Jupyter and checker.
- Logs never contain repository tokens, CMS passwords, workspace grants/cookies or
  notebook contents.
- Before cutover, finish or explicitly terminate old JupyterHub attempts owned
  by the same `k8s` client. Otherwise the new reconciler cannot distinguish an
  old Hub attempt with no namespace from a deleted Clabgate session.

## Deferred after the first release

- collaboration greater than one;
- idle timeout and storage retention policy;
- per-course images/resources and GPU scheduling;
- richer checker artifact storage beyond bounded CMS report/log fields;
- informer caches and horizontal scale beyond the current two replicas.
