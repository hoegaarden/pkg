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
readonly PKG_NS="${PKG_NS:-hoegaarden.github.io}"

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

getVersionedValueSchema() {
    local file="$1"
    local rev="$2"

    ytt -f - --data-values-schema-inspect -o openapi-v3 < <(
        getVersionedFile "${file}" "${rev}" 2>/dev/null || {
            echo >&2 "## values file '${file}' @ ${rev} not existant, skipping values schema generation"
        }
    )
}

genPkgForVersion() {
    local pkgDir="$1"
    local gitRev="$2"
    local ver="$3"
    local meta="$4"

    local date
    if [[ "$gitRev" == "$TRUNK" ]] ; then
        date="$( date -d '@0' --iso-8601=s )"
    else
        date="$( getCommitDate "$gitRev" )"
    fi

    getVersionedFile "${pkgDir}/pkg.yml" "$gitRev" \
        | ytt \
            --data-values-file "$meta" \
            -v version="${ver}" \
            -v repoPath="${pkgDir}/src" \
            -v repo="${REPO}" \
            -v repoRef="${gitRev}" \
            -v date="${date}" \
            --data-values-file <(
                getVersionedValueSchema "${pkgDir}/src/values.yml" "$gitRev"
            ) \
            -f -
}

handlePkg() {
    local pkgName="$1"

    local pkgDir="${PKG_DIR}/${pkgName}"
    local repoDir="${REPO_DIR}/packages/${pkgName}"
    mkdir -p "${repoDir}"

    # get the meta from the trunk
    local metaSrc="${pkgDir}/meta.yml"
    local metaDest="${repoDir}/meta.yml"
    echo >&2 "## ${pkgName}: writing package meta data from ${TRUNK}"
    getVersionedFile "$metaSrc" "$TRUNK" \
        | ytt -v pkgName="$pkgName" -v pkgNS="$PKG_NS" -f - \
        > "${metaDest}"

    local rev ver
    while read -r rev
    do
        ver="${rev#*@}"
        echo >&2 "## ${pkgName}/${ver}: writing package from ${rev} (verison from git revision)"
        genPkgForVersion "$pkgDir" "$rev" "$ver" "$metaDest" \
            > "${repoDir}/${ver}.yml"
    done < <(getPkgRevs "$pkgName")

    # "release" the trunk as a floating dev release
    ver="0.0.0-dev"
    echo >&2 "## ${pkgName}/${ver}: writing package from ${TRUNK}"
    genPkgForVersion "$pkgDir" "$TRUNK" "$ver" "$metaDest" \
        > "${repoDir}/next.yml"
}

commitRepo() {
    git add "$1"

    if git diff --cached --quiet ; then
        echo >&2 "## no change in ${1}, nothing to commit"
    else
        echo >&2 "## committing updated packages in ${1}"
        git commit -m '[repo] Update repo for all packages'
    fi
}

main() {
    cd "$HERE"

    local pkgName

    while read -d $'\0' -r pkgName
    do
        handlePkg "$pkgName"
    done < <(find "$PKG_DIR" -mindepth 1 -maxdepth 1 -type d -printf '%P\0')

    commitRepo "$REPO_DIR"
}

main "$@"
