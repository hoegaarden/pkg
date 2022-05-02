#!/usr/bin/env bash

set -e
set -u
set -o pipefail

readonly DIR="$( cd "$(dirname "${BASH_SOURCE[0]}")" && pwd )"
readonly OVERLAY="${DIR}/ns.rbac.overlay.yml"

main() {
    local pkg="${1?must be the app/pkg we are preparing for, one of those in ${DIR}/pkgs}"
    local ver="${2?must be the app\'s/pkg\'s version}"
    local name="${3?must be the app\'s/pkg\'s instance name}"
    local ns="${4?must be the app\'s/pkg\'s namespace}"

    local objFile="${DIR}/pkgs/${pkg}/${ver}/ns.rbac.yml"

    [ -e "$objFile" ] || {
        echo >&2 "no namespace/rbac config for package '$pkg/$ver' found, expected it in '$objFile'"
        return 1
    }

    ytt -f "$objFile" -f "${OVERLAY}" --data-values-env 'DV' --data-value "instance=${name}" --data-value "ns=${ns}" | {
        if test -t 1 ; then
            kubectl apply -f -
        else
            cat
        fi
    }
}

main "$@"
