#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Validate GoReleaser build matrix locally.

This script does two things:
  1) Cross-compiles `go build ./...` for the exact GOOS/GOARCH matrix in `.goreleaser.yml`
  2) Runs GoReleaser (defaults to v1.26.2) with `release --snapshot` to generate `dist/`

Usage:
  tools/validate_goreleaser_local.sh [flags]

Flags:
  --go-only           Only run the `go build` cross-compile matrix
  --goreleaser-only   Only run GoReleaser snapshot release
  --skip-before       Skip GoReleaser "before" hooks (e.g., go mod tidy)
  --no-git-check      Do not require a clean git working tree
  -h, --help          Show help

Environment variables:
  GORELEASER_VERSION  GoReleaser version to download/use (default: 1.26.2)
  GORELEASER_BIN      Path to an existing goreleaser binary to use (skips download)

Examples:
  tools/validate_goreleaser_local.sh
  GORELEASER_VERSION=1.26.2 tools/validate_goreleaser_local.sh
  tools/validate_goreleaser_local.sh --go-only
EOF
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"

RUN_GO=1
RUN_GR=1
SKIP_BEFORE=0
GIT_CHECK=1

while [[ $# -gt 0 ]]; do
  case "${1}" in
    --go-only) RUN_GR=0 ;;
    --goreleaser-only) RUN_GO=0 ;;
    --skip-before) SKIP_BEFORE=1 ;;
    --no-git-check) GIT_CHECK=0 ;;
    -h|--help) usage; exit 0 ;;
    *)
      echo "Unknown flag: ${1}" >&2
      echo >&2
      usage >&2
      exit 2
      ;;
  esac
  shift
done

require_no_tracked_changes() {
  [[ "${GIT_CHECK}" -eq 1 ]] || return 0
  if git -C "${REPO_ROOT}" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    if ! git -C "${REPO_ROOT}" diff --quiet || ! git -C "${REPO_ROOT}" diff --cached --quiet; then
      echo "ERROR: tracked files have uncommitted changes. Commit/stash changes or re-run with --no-git-check." >&2
      git -C "${REPO_ROOT}" diff --name-only >&2 || true
      exit 1
    fi
  fi
}

detect_host_asset() {
  local os arch
  os="$(uname -s)"
  arch="$(uname -m)"

  case "${os}" in
    Linux) os="Linux" ;;
    Darwin) os="Darwin" ;;
    *)
      echo "Unsupported host OS for GoReleaser download: ${os}" >&2
      echo "Set GORELEASER_BIN to an existing goreleaser binary instead." >&2
      exit 1
      ;;
  esac

  case "${arch}" in
    x86_64|amd64) arch="x86_64" ;;
    arm64|aarch64) arch="arm64" ;;
    *)
      echo "Unsupported host arch for GoReleaser download: ${arch}" >&2
      echo "Set GORELEASER_BIN to an existing goreleaser binary instead." >&2
      exit 1
      ;;
  esac

  printf "%s %s" "${os}" "${arch}"
}

ensure_goreleaser() {
  if [[ -n "${GORELEASER_BIN:-}" ]]; then
    echo "Using GoReleaser from GORELEASER_BIN: ${GORELEASER_BIN}"
    return 0
  fi

  local version os arch cache_dir bin
  version="${GORELEASER_VERSION:-1.26.2}"
  read -r os arch <<<"$(detect_host_asset)"

  cache_dir="${REPO_ROOT}/.cache/goreleaser/${version}/${os}_${arch}"
  bin="${cache_dir}/goreleaser"

  if [[ -x "${bin}" ]]; then
    GORELEASER_BIN="${bin}"
    echo "Using cached GoReleaser: ${GORELEASER_BIN}"
    return 0
  fi

  mkdir -p "${cache_dir}"

  local url tgz
  url="https://github.com/goreleaser/goreleaser/releases/download/v${version}/goreleaser_${os}_${arch}.tar.gz"
  tgz="${cache_dir}/goreleaser.tar.gz"

  echo "Downloading GoReleaser v${version} (${os}/${arch})"
  curl -fsSL -o "${tgz}" "${url}"
  tar -xzf "${tgz}" -C "${cache_dir}"
  rm -f "${tgz}"

  if [[ ! -x "${bin}" ]]; then
    echo "ERROR: expected goreleaser binary at ${bin}" >&2
    exit 1
  fi

  GORELEASER_BIN="${bin}"
}

run_go_matrix_builds() {
  echo
  echo "== Cross-compiling: go build ./... for GoReleaser matrix =="
  echo "Repo: ${REPO_ROOT}"
  echo

  (cd "${REPO_ROOT}" && go version)

  # Mirror `.goreleaser.yml`:
  local goos=(freebsd windows linux darwin)
  local goarch=(amd64 386 arm arm64)

  for os in "${goos[@]}"; do
    for arch in "${goarch[@]}"; do
      if [[ "${os}" == "darwin" && "${arch}" == "386" ]]; then
        continue
      fi
      if [[ "${os}" == "darwin" && "${arch}" == "arm" ]]; then
        continue
      fi

      echo
      echo "== building ${os}/${arch} =="
      if [[ "${arch}" == "arm" ]]; then
        (cd "${REPO_ROOT}" && CGO_ENABLED=0 GOAMD64=v1 GOOS="${os}" GOARCH="${arch}" GOARM=6 go build ./...)
      else
        (cd "${REPO_ROOT}" && CGO_ENABLED=0 GOAMD64=v1 GOOS="${os}" GOARCH="${arch}" go build ./...)
      fi
    done
  done

  echo
  echo "OK: all GOOS/GOARCH targets compiled."
}

run_goreleaser_snapshot() {
  ensure_goreleaser

  echo
  echo "== GoReleaser snapshot release =="
  "${GORELEASER_BIN}" --version

  local -a args
  args=(release --snapshot --clean --skip=sign)
  if [[ "${SKIP_BEFORE}" -eq 1 ]]; then
    args+=(--skip=before)
  fi

  (cd "${REPO_ROOT}" && "${GORELEASER_BIN}" "${args[@]}")

  echo
  echo "OK: GoReleaser snapshot finished. See: ${REPO_ROOT}/dist/"
}

main() {
  if [[ "${RUN_GO}" -eq 1 ]]; then
    run_go_matrix_builds
  fi

  if [[ "${RUN_GR}" -eq 1 ]]; then
    require_no_tracked_changes
    run_goreleaser_snapshot
    require_no_tracked_changes
  fi
}

main "$@"

