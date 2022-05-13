#!/usr/bin/env bash

set -e
set -u
set -o pipefail

readonly HERE="$( cd "$(dirname "${BASH_SOURCE[0]}")" && pwd )"
readonly PKG_DIR="pkgs"
readonly REPO_DIR="repo"
readonly TRUNK='main'

# if you want to clone the whole thing to a different repo:
readonly REPO="${REPO:-https://github.com/hoegaarden/pkg}"

getPkgRevs() {
    git for-each-ref --format='%(refname:short)' \
        --sort='refname:short' \
        "refs/heads/${1}@*" \
        "refs/tags/${1}@*"
}

getVersionedFile() {
    local repoPath="$1"
    local rev="${2:-}"

    git show "${rev}:${repoPath}"
}

getCommitDate() {
    git show -s --format=%cI "${1}"
}

genPkgForVersion() {
    local pkgDir="$1"
    local gitRev="$2"
    local ver="$3"
    local meta="$4"

    getVersionedFile "${pkgDir}/pkg.yml" "$gitRev" \
        | ytt \
            --data-values-file "$meta" \
            -v version="${ver}" \
            -v repoPath="${pkgDir}/src" \
            -v repo="${REPO}" \
            -v repoRef="${gitRev}" \
            -v date="$( getCommitDate "$gitRev" )" \
            --data-values-file <(
                getVersionedFile "${pkgDir}/src/values.yml" "$gitRev" | ytt -f - --data-values-schema-inspect -o openapi-v3
            ) \
            -f -
}

handlePkg() {
    local pkgName="$1"

    local pkgDir="${PKG_DIR}/${pkgName}"
    local repoDir="${REPO_DIR}/packages/${pkgName}"
    mkdir -p "${repoDir}"

    # get the meta from the trunk
    local meta="${pkgDir}/meta.yml"
    echo "## ${pkgName}: writing package meta data from ${TRUNK}"
    getVersionedFile "$meta" "$TRUNK" > "${repoDir}/meta.yml"

    local rev ver
    while read -r rev
    do
        ver="${rev#*@}"
        echo "## ${pkgName}/${ver}: writing package from ${rev} (verison from git revision)"
        genPkgForVersion "$pkgDir" "$rev" "$ver" "$meta" \
            > "${repoDir}/${ver}.yml"
    done < <(getPkgRevs "$pkgName")

    # "release" the trunk as a floating dev release
    ver="0.0.0-dev"
    echo "## ${pkgName}/${ver}: writing package from ${TRUNK}"
    genPkgForVersion "$pkgDir" "$TRUNK" "$ver" "$meta" \
        > "${repoDir}/next.yml"
}

main() {
    cd "$HERE"

    local pkgName

    while read -d $'\0' -r pkgName
    do
        handlePkg "$pkgName"
    done < <(find "$PKG_DIR" -mindepth 1 -maxdepth 1 -type d -printf '%P\0')
}

main "$@"
