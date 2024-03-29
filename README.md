# My kapp package repo

## What

This is an opinionated way on how you might want to maintain you kapp packages
and package repo. Some important choices:

- It does not use bundles or OCI images to ship the packages, metadata, code,
  ... to the target clusters; it just uses git
- It does maintain the package repo and the packages themselves in the same  git
  repo
- Each package has a specific layout in the filesystem, see below
- Versions of packages are cut based on either tags or branches in the form of
  `<packageName>@<version>`, e.g. `test@0.0.3`
- If we adhere to the below layout, the package repo can be autogenerated
- It should be easy to bring the whole thing over an air-gap (run `./updateRepo`
  with `REPO`  set to your clone)

## Files and directories

```
.
├── ns.rbac.overlay.yml           # helper overlay to create NS & RBAC objects
├── pkgs                          # all them custom packages
│   └── test                      #   ... with a symbolic name 'test'
│       ├── meta.yml              # the package's metadata
│       ├── example-install.yml   # show your users how you might want to use the packages
│       ├── ns.rbac.yml           # all objects, esp. RBAC stuff, this version of the packages needs up front
│       ├── pkg_test.go           # the package's tests
│       ├── pkg.yml               # the package, version specific
│       └── src                   # the actual code of the package we will run through ytt
│           ├── main.yml          #   ... all the file ...
│           └── values.yml        #   ... you might need to do usefull stuff
├── prep-ns.sh                    # helper to generate NS & RBAC objects based on ./pkgs/${pkg}/${ver}/ns.rbac.yml
├── README.md
├── repo                          # autogenerated, with the help of ./updateRepo.sh
│   └── packages
│       └── test
│           ├── 0.0.1.yml
│           └── meta.yml
├── repo.yml                      # example for how to use this repo
├── testing                       # the test helpers
│   ├── jqer.go
│   └── ytt.go
└── updateRepo.sh                 # generate the whole repo metadata
```

## Status

Pretty, pretty early. Don't use. Go away!

## TODO

- where's your concourse?
