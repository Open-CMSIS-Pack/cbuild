[![Release](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/release.yml/badge.svg)](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/release.yml)
[![Build](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/build.yml/badge.svg)](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/build.yml)
[![Test](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/test.yml/badge.svg)](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/test.yml)

# cbuild: Open-CMSIS-Pack Build Invocation Utility

This utility allows embedded developers to build **CPRJ** and **csolution** projects by orchestrating the following tools:

- `cbuildgen`
- `csolution`
- `cpackget`
- `cmake`
- `ninja`

By default, `cbuild` expects a fully configured solution (*.csolution.yml) or context (*.cprj). As such, it will not create, copy or update any files in the RTE directories. In case such modifications are required, use the command line option: `--update-rte`.

## Usage

```bash
cbuild: Build Invocation 2.0.0-dev2 (C) 2023 Arm Ltd. and Contributors

Usage:
  cbuild [command] <csolution.yml> [flags]

Available Commands:
  buildcprj   Generate output
  help        Help about any command
  list        List information

Flags:
  -C, --clean              Remove intermediate and output directories
  -c, --context strings    Input context name e.g. [<cproject>][.<build-type>][+<target-type>]
  -d, --debug              Enable debug messages
  -g, --generator string   Select build system generator (default "Ninja")
  -h, --help               Print usage
  -j, --jobs int           Number of job slots for parallel execution
  -l, --load string        Set policy for packs loading [latest|all|required]
      --log string         Save output messages in a log file
  -O, --output string      Set directory for all output files
  -p, --packs              Download missing software packs with cpackget
  -q, --quiet              Suppress output messages except build invocations
  -r, --rebuild            Remove intermediate and output directories and rebuild
  -s, --schema             Validate project input file(s) against schema
  -t, --target string      Optional CMake target name
      --toolchain string   Input toolchain to be used
      --update-rte         Update the RTE directory and files
  -v, --verbose            Enable verbose messages from toolchain builds
  -V, --version            Print version

Use "cbuild [command] --help" for more information about a command.
```
