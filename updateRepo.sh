#!/usr/bin/env bash

set -e
set -u
set -o pipefail

readonly HERE="$( cd "$(dirname "${BASH_SOURCE[0]}")" && pwd )"
readonly PKG_DIR="${HERE}/pkgs"
readonly REPO_DIR="${HERE}/repo"

handlePkg() {
    local pkgName="$1"

    local pkgDir="${PKG_DIR}/${pkgName}"
    local repoDir="${REPO_DIR}/packages/${pkgName}"
    local meta="${pkgDir}/meta.yml"

    mkdir -p "${repoDir}"
    echo "## ${pkgName}: writing metadata"
    cp "$meta" "${repoDir}/meta.yml"

    local ver
    while read -d $'\0' -r ver
    do
        echo "## ${pkgName}: writing package for version ${ver}"
        ytt \
            --data-values-file "$meta" \
            -v version="${ver}" \
            -v repoPath="pkgs/${pkgName}/${ver}/src" \
            --data-values-file <(
                ytt -f "${pkgDir}/${ver}/src/values.yml" --data-values-schema-inspect -o openapi-v3
            ) \
            -f "${pkgDir}/${ver}/pkg.yml" \
            > "${repoDir}/${ver}.yml"
    done < <(find "$pkgDir" -mindepth 1 -maxdepth 1 -type d -printf '%P\0')
}

main() {
    local pkgName

    while read -d $'\0' -r pkgName
    do
        handlePkg "$pkgName"
    done < <(find "$PKG_DIR" -mindepth 1 -maxdepth 1 -type d -printf '%P\0')
}

main "$@"
