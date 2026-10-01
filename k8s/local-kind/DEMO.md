# Full-stack demo

The API has an idempotent demo seed mode. It creates the `admin@admin.com` demo
user (password `admin`), a Kubernetes server (`demo-kubernetes` / `demo-secret`),
the `Simple Task Demo` routing and a stable attempt UUID:
`00000000-0000-0000-0000-000000000000`.

Run it after applying database migrations and before starting Clabgate:

```sh
./backend/apiserver --demo
```

Without arguments the built-in single-entry catalog is deployed. To import a
laboratory catalog, pass one JSON file as the positional argument:

```sh
./backend/apiserver --demo ./k8s/local-kind/labs.example.json
```

Both variants go through the same seeding code, but the stable attempt UUID
belongs to the built-in catalog only: with a file the routings are created and no
demo attempt is made.

The file is validated completely and then upserted into `lti_routings` in one
database transaction, matched by the lab `id`. Each entry creates the routing
with that id or overwrites the existing one; entries absent from the file keep
whatever the database already holds. The API serves the catalog from the
database and never reads the JSON file at request time.

```json
{
  "labs": [
    {
      "id": 1,
      "name": "Пример задания: SSH и SNMP",
      "description": "Настройка Linux-based сетевых узлов с автоматической проверкой результата.",
      "labs_path": "https://github.com/cms-lab-core/cms-labs-simple-task.git#main",
      "test_path": "sdn_lab_5"
    }
  ]
}
```

The catalog API is gated by `LAB_CATALOG_ENABLED` and defaults to off: while it
is off, `lab_catalog.list` returns an empty list, `lab_catalog.start` is
rejected, and the frontend hides the laboratory section. Set
`LAB_CATALOG_ENABLED=true` for the backend to serve the imported catalog.

`id` is the routing id and the match key, so it must be positive and unique in
the file. `description` and `test_path` are optional. `labs_path` holds the full
Git link; an optional `#ref` fragment pins a branch or commit, and without it
Clabgate falls back to `CMS_TASK_URL` and `CMS_TASK_BRANCH`. Lab manifests are
read from the repository root, so the link should end with `.git`. Nothing else
is accepted.

For a container deployment use the same image with command `[/app/apiserver, --demo]`
as a one-shot Job. The regular API then starts without `--demo`.
Configure Clabgate with `CMS_LOGIN=demo-kubernetes`, `CMS_PASSWORD=demo-secret`,
`CMS_TASK_URL` remains the fallback for legacy routes. Catalog entries carry
their own repository link. GitHub/GitLab project URLs are supported, and
an arbitrary HTTP(S) Git host can be used with an explicit `.git` URL.

Log in to the frontend with `admin@admin.com` / `admin`, then open the demo
attempt URL. The attempt is intentionally pending until Clabgate acknowledges
the session; this is the same lifecycle used in production.
