# GitHub Actions CI/CD Discussion — Summary

Date: 2026-09-22

## 1. Goal

CI/CD architecture for `task-api Go` application:

- **Feature branch push**
  - Build and test the application.
  - Build the Docker image.
  - Publish the image to GHCR.
  - Automatically deploy that exact image to the local/self-hosted environment.
- **Main or release branch push**
  - Build and test.
  - Publish the Docker image to GHCR.
  - Do not automatically deploy locally.
- **Promotion**
  - Promote an existing GHCR image to a release tag without rebuilding it.
- **Manual deployment**
  - Select an existing GHCR image/tag and deploy it locally.
- **Cloud deployment**
  - Leave a future extension point, cloud is not being used yet.

The important architectural principle is:

> GHCR is the artifact repository/source of truth. Local deployment should pull the image that was actually built and published.

---

## 2. Workflow Files

The intended `.github/workflows/` structure is:

```text
ci.yml
build-and-publish.yml
deploy-local.yml
deploy-local-manual.yml
promote.yml
deploy-cloud.yml
```

### Responsibilities

| Workflow | Responsibility |
|---|---|
| `ci.yml` | Orchestrates push and manual operations |
| `build-and-publish.yml` | Build/test image and push it to GHCR |
| `deploy-local.yml` | Reusable workflow that pulls an existing GHCR image and deploys it |
| `deploy-local-manual.yml` | Manual local deployment entry point |
| `promote.yml` | Retag/promote an existing GHCR artifact |
| `deploy-cloud.yml` | Future cloud deployment |

---

# 3. Feature Branch Flow

```text
feature/my-new-feature
```

A push should result in:

```text
feature/my-new-feature
        |
        v
build-and-publish.yml
        |
        +-- go vet
        +-- go test
        +-- docker build
        +-- docker push
        |
        v
GHCR
ghcr.io/<owner>/<repo>:feature-my-new-feature
        |
        v
deploy-local.yml
        |
        +-- docker pull
        +-- docker compose up
```

The branch name is converted to a valid Docker tag by replacing `/` with `-`.

Examples:

```text
feature/foo       -> feature-foo
feature/a/b       -> feature-a-b
release/1.2.0     -> release-1.2.0
main              -> main
```

---

# 4. Main / Release Flow

A push to:

```text
main
```

publishes:

```text
ghcr.io/<owner>/<repo>:main
```

A push to:

```text
release/1.2.0
```

publishes:

```text
ghcr.io/<owner>/<repo>:release-1.2.0
```

Neither automatically deployed locally (or cloud).

They are available for later manual deployment or promotion.

---

# 5. `build-and-publish.yml`

Builds and Pushes the Docker image into GHCR.

image:

```yaml
env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}
```

permissions:

```yaml
permissions:
  contents: read
  packages: write
```

result:

```yaml
      - name: Build and push
        uses: docker/build-push-action@v6
        with:
          context: .
          push: true
          tags: |
            ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}:${{ steps.image.outputs.tag }}
```

---

# 6. `ci.yml`

`ci.yml` is the default for any branch:

```yaml
name: CI

on:
  push:
    branches:
      - '**'

...
...

  deploy:
    if: >
      github.event_name == 'push' &&
      github.ref_name != 'main' &&
      !startsWith(github.ref_name, 'release/')

    needs: build

...
```

Workflow:

### Feature push

```text
push feature/foo
    |
    v
build-and-publish
    |
    v
GHCR: feature-foo
    |
    v
deploy-local
```

### Main push

```text
push main
    |
    v
build-and-publish
    |
    v
GHCR: main
```

No automatic local deployment.

### Release push

```text
push release/1.2.0
    |
    v
build-and-publish
    |
    v
GHCR: release-1.2.0
```

No automatic local deployment.

---

# 7. `deploy-local.yml`

`deploy-local.yml` is a reusable workflow.

It should consume an existing image:

```yaml
...
...

      - name: Deploy
        run: |
          export IMAGE_REF="${REGISTRY}/${IMAGE_NAME}:${{ steps.image.outputs.tag }}"

          echo "Deploying: ${IMAGE_REF}"

          docker compose pull task-api
          docker compose up -d --remove-orphans

      - name: Cleanup
        run: docker image prune -f
```

The Docker Compose file uses:

```yaml
services:
  task-api:
    image: ${IMAGE_REF}
```

---

# 8. Manual Deployment

Manual deployment is intended to deploy an existing artifact:

```text
image_tag = main
```

or:

```text
image_tag = release-1.2.0
```

or:

```text
image_tag = feature-my-feature
```

Manual workflow:

```yaml
name: Deploy Local Manual

on:
  workflow_dispatch:
    inputs:
      image_tag:
        description: Image tag to deploy
        required: true
        type: string

jobs:
  deploy:
    uses: ./.github/workflows/deploy-local.yml

    with:
      image_tag: ${{ inputs.image_tag }}

    permissions:
      contents: read
      packages: read

    secrets: inherit
```

---

# 9. GHCR Authentication

For GHCR authentication, we use the built-in token:

```yaml
- name: Login to GHCR
  run: |
    echo "${{ github.token }}" |
      docker login "${REGISTRY}" \
        -u "${{ github.actor }}" \
        --password-stdin
```

The permissions should be:

```text
build-and-publish.yml:
    packages: write

promote.yml:
    packages: write

deploy-local.yml:
    packages: read
```

The caller also needs to grant the appropriate permission because reusable workflow permissions are constrained by the calling workflow.

---

# 10. Promotion

Promotion should not rebuild the application.

For example:

```text
feature-my-feature
        |
        | promote
        v
release-1.2.0
```

Both tags should refer to the same image artifact.

The conceptual flow is:

```text
GHCR:
    ghcr.io/owner/repo:feature-my-feature
                  |
                  | retag
                  v
    ghcr.io/owner/repo:release-1.2.0
```

The promotion workflow needs:

```yaml
permissions:
  packages: write
```

because it creates/pushes the release tag.

A simple implementation can pull, retag, and push:

```yaml
docker pull ghcr.io/owner/repo:${SOURCE_TAG}
docker tag \
  ghcr.io/owner/repo:${SOURCE_TAG} \
  ghcr.io/owner/repo:${RELEASE_TAG}
docker push ghcr.io/owner/repo:${RELEASE_TAG}
```

A registry-side tag operation can also be used later if avoiding the image pull is desired.

---

# 11. Final Intended Architecture

The overall lifecycle is:

```text
                        FEATURE DEVELOPMENT

feature/foo
     |
     v
+----------------------+
| build-and-publish    |
| - checkout           |
| - go vet             |
| - go test            |
| - docker build       |
| - docker push        |
+----------+-----------+
           |
           v
    GHCR: feature-foo
           |
           v
+----------------------+
| deploy-local         |
| - docker pull        |
| - docker compose     |
+----------------------+
           |
           v
     LOCAL ENVIRONMENT


                        RELEASE PROCESS

feature-foo / main
       |
       v
   promote.yml
       |
       v
GHCR: release-1.2.0
       |
       +------------------+
       |                  |
       v                  v
deploy-local        future cloud
       |
       v
LOCAL RELEASE
```

## Key rules

1. **Every push builds and publishes an image.**
2. **Feature branch pushes automatically deploy locally.**
3. **Main/release pushes publish only; they do not automatically deploy locally.**
4. **Deployment always pulls an already-published GHCR image.**
5. **Promotion retags an existing image; it does not rebuild it.**
6. **`GITHUB_TOKEN` is used for GHCR authentication.**
7. **Build/promotion require `packages: write`; deployment requires `packages: read`.**
8. **Branch names are converted into valid Docker tags by replacing `/` with `-`.**
