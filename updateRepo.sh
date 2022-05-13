#!/usr/bin/env bash

set -e
set -u
set -o pipefail

readonly HERE="$( cd "$(dirname "${BASH_SOURCE[0]}")" && pwd )"
readonly PKG_DIR="pkgs"
readonly REPO_DIR="repo"

# if you want to clone the whole thing to a different repo:
readonly REPO="${REPO:-https://github.com/hoegaarden/pkg}"

currentTip() {
    git rev-parse --abbrev-ref HEAD
}

getPkgRevs() {
    git tag -l "${1}@*"
    git branch -l "${1}@*"
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

    local currentTip
    currentTip="$( currentTip )"
    local pkgDir="${PKG_DIR}/${pkgName}"
    local repoDir="${REPO_DIR}/packages/${pkgName}"
    mkdir -p "${repoDir}"

    # get the meta from the current tip
    local meta="${pkgDir}/meta.yml"
    echo "## ${pkgName}: writing package meta data"
    getVersionedFile "$meta" "$currentTip" > "${repoDir}/meta.yml"

    local rev ver
    while read -r rev
    do
        ver="${rev#*@}"
        echo "## ${pkgName}/${ver}: writing package from ${rev} (verison from git revision)"
        genPkgForVersion "$pkgDir" "$rev" "$ver" "$meta" \
            > "${repoDir}/${ver}.yml"
    done < <(getPkgRevs "$pkgName")

    # the current tip as additional "dev" version
    local versionFile="${pkgDir}/next_version"
    ver="$( getVersionedFile "$versionFile" "$currentTip" )"
    echo "## ${pkgName}/${ver}: writing package from ${currentTip} (version from $versionFile)"
    genPkgForVersion "$pkgDir" "$currentTip" "$ver" "$meta" \
        > "${repoDir}/${ver}.yml"
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
