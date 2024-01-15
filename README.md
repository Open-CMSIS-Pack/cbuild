[![Release](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/release.yml/badge.svg)](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/release.yml)
[![Build](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/build.yml/badge.svg)](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/build.yml)
[![Test](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/test.yml/badge.svg)](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/test.yml)
[![TPIP Check](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/tpip-check.yml/badge.svg)](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/tpip-check.yml)
[![Markdown](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/markdown.yml/badge.svg)](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/markdown.yml)

[![Maintainability](https://api.codeclimate.com/v1/badges/53904fe8cbd887f3d5b0/maintainability)](https://codeclimate.com/github/Open-CMSIS-Pack/cbuild/maintainability)
[![Test Coverage](https://api.codeclimate.com/v1/badges/53904fe8cbd887f3d5b0/test_coverage)](https://codeclimate.com/github/Open-CMSIS-Pack/cbuild/test_coverage)
[![OpenSSF Scorecard](https://api.securityscorecards.dev/projects/github.com/Open-CMSIS-Pack/cbuild/badge)](https://securityscorecards.dev/viewer/?uri=github.com/Open-CMSIS-Pack/cbuild)

# cbuild: Open-CMSIS-Pack Build Invocation Utility

This utility allows embedded developers to build **CPRJ** and **csolution** projects by orchestrating the following tools:

- `cbuildgen`
- `csolution`
- `cpackget`
- `cmake`
- `ninja`

By default, `cbuild` expects a fully configured solution (*.csolution.yml) or context (*.cprj).
As such, it will not create, copy or update any files in the RTE directories. In case such modifications are required,
use the command line option: `--update-rte`.

## Usage

```bash
cbuild: Build Invocation 2.0.0 (C) 2024 Arm Ltd. and Contributors

Usage:
  cbuild [command] <name>.csolution.yml [options]

Commands:
  buildcprj   Use a *.CPRJ file as build input
  help        Help about any command
  list        List information about environment, toolchains, and contexts

Options:
  -C, --clean              Remove intermediate and output directories
  -c, --context arg [...]  Input context names [<project-name>][.<build-type>][+<target-type>]
  -S, --context-set        Use context set
  -d, --debug              Enable debug messages
      --frozen-packs       The list of packs from cbuild-pack.yml is frozen and raises error if not up-to-date
  -g, --generator arg      Select build system generator (default "Ninja")
  -h, --help               Print usage
  -j, --jobs int           Number of job slots for parallel execution
  -l, --load arg           Set policy for packs loading [latest | all | required]
      --log arg            Save output messages in a log file
  -O, --output arg         Set directory for all output files
  -p, --packs              Download missing software packs with cpackget
  -q, --quiet              Suppress output messages except build invocations
  -r, --rebuild            Remove intermediate and output directories and rebuild
  -s, --schema             Validate project input file(s) against schema
  -t, --target arg         Optional CMake target name
      --toolchain arg      Input toolchain to be used
      --update-rte         Update the RTE directory and files
  -v, --verbose            Enable verbose messages from toolchain builds
  -V, --version            Print version

Use "cbuild [command] --help" for more information about a command.
```
