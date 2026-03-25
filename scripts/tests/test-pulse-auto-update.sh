#!/usr/bin/env bash
#
# Smoke tests for scripts/pulse-auto-update.sh helper behavior.

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
AUTO_UPDATE_SCRIPT="${ROOT_DIR}/scripts/pulse-auto-update.sh"

if [[ ! -f "${AUTO_UPDATE_SCRIPT}" ]]; then
  echo "pulse-auto-update.sh not found at ${AUTO_UPDATE_SCRIPT}" >&2
  exit 1
fi

# shellcheck disable=SC1090
source "${AUTO_UPDATE_SCRIPT}"

failures=0

assert_success() {
  local desc="$1"
  shift
  if "$@"; then
    echo "[PASS] ${desc}"
    return 0
  else
    echo "[FAIL] ${desc}" >&2
    ((failures++))
    return 1
  fi
}

test_wait_for_service_active_succeeds_after_retry() {
  local calls=0

  systemctl() {
    if [[ "$1" == "is-active" ]]; then
      ((calls += 1))
      if (( calls >= 3 )); then
        return 0
      fi
      return 1
    fi
    return 1
  }

  sleep() { :; }

  wait_for_service_active pulse 5
}

test_perform_update_restores_backup_when_service_stays_down() {
  local tmpdir
  tmpdir="$(mktemp -d)"
  local status=0

  INSTALL_DIR="${tmpdir}/opt/pulse"
  CONFIG_DIR="${tmpdir}/etc/pulse"
  mkdir -p "${INSTALL_DIR}/bin" "${CONFIG_DIR}"

  printf 'v5.1.24\n' > "${INSTALL_DIR}/VERSION"
  cat > "${INSTALL_DIR}/bin/pulse" <<'EOF'
#!/usr/bin/env bash
echo "v5.1.24"
EOF
  chmod +x "${INSTALL_DIR}/bin/pulse"

  export INSTALL_DIR
  export FAKE_NEW_VERSION="v5.1.25"

  get_current_version() {
    tr -d '\r\n' < "${INSTALL_DIR}/VERSION"
  }

  detect_service_name() {
    echo "pulse"
  }

  local is_active_calls=0
  local restart_called=0

  systemctl() {
    if [[ "$1" == "is-active" ]]; then
      ((is_active_calls += 1))
      if (( is_active_calls == 1 )); then
        return 0
      fi
      return 1
    fi
    if [[ "$1" == "start" ]]; then
      return 1
    fi
    return 0
  }

  restart_service_if_needed() {
    restart_called=1
    return 0
  }

  sleep() { :; }

  curl() {
    cat <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$FAKE_NEW_VERSION" > "$INSTALL_DIR/VERSION"
exit 0
EOF
  }

  if perform_update "v5.1.25"; then
    echo "perform_update unexpectedly succeeded" >&2
    status=1
  fi

  if [[ "${status}" -eq 0 ]] && [[ "$(tr -d '\r\n' < "${INSTALL_DIR}/VERSION")" != "v5.1.24" ]]; then
    echo "expected VERSION to be restored after failed restart" >&2
    status=1
  fi

  if [[ "${status}" -eq 0 ]] && (( restart_called != 1 )); then
    echo "expected restart_service_if_needed to be called" >&2
    status=1
  fi

  rm -rf "${tmpdir}"
  return "${status}"
}

main() {
  assert_success "wait_for_service_active retries until active" test_wait_for_service_active_succeeds_after_retry
  assert_success "perform_update restores backup when service stays down" test_perform_update_restores_backup_when_service_stays_down

  if (( failures > 0 )); then
    echo "Total failures: ${failures}" >&2
    return 1
  fi

  echo "All pulse-auto-update smoke tests passed."
}

main "$@"
