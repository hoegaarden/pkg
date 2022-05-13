#!/usr/bin/env bash

set -e
set -u
set -o pipefail

readonly DIR="$( cd "$(dirname "${BASH_SOURCE[0]}")" && pwd )"
readonly OVERLAY='ns.rbac.overlay.yml'

getFileFromVersion() {
    local repoPath="$1"
    local ver="$2"

    if [[ "$ver" == "-" ]] ; then
        cat "${repoPath}"
        return
    fi

    git show "${ver}:${repoPath}"
}

maybeApply() {
    if test -t 1 ; then
        kubectl apply -f -
    else
        cat
    fi
}

main() {
    local pkg="${1?must be the app/pkg we are preparing for, one of those in ${DIR}/pkgs}"
    local name="${2?must be the app\'s/pkg\'s instance name}"
    local ns="${3?must be the app\'s/pkg\'s namespace}"
    local ver="${4?must be the app\'s/pkg\'s version, can be - for the currently checked out revision}"

    local rbacFile="pkgs/${pkg}/ns.rbac.yml"

    cd "${DIR}"

    {
        getFileFromVersion "${rbacFile}" "${ver}"
        echo '---'
        getFileFromVersion "${OVERLAY}" "${ver}"
    } | ytt -f - \
        --data-values-env 'DV' \
        --data-value "instance=${name}" \
        --data-value "ns=${ns}" \
      | maybeApply
}

main "$@"
