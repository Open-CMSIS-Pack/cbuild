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

## Usage

```bash
cbuild: Build Invocation 1.3.0 (C) 2022 Arm Ltd. and Contributors

Usage:
  cbuild [command] <project.cprj|csolution.yml> [flags]

Available Commands:
  help            Help about any command
  list-contexts   Print list of contexts in a csolution.yml
  list-toolchains Print list of installed toolchains

Flags:
  -c, --clean              Remove intermediate and output directories
      --context string     Input context name e.g. project.buildType+targetType
  -d, --debug              Enable debug messages
  -g, --generator string   Select build system generator (default "Ninja")
  -h, --help               Print usage
  -i, --intdir string      Set directory for intermediate files
  -j, --jobs int           Number of job slots for parallel execution
      --load string        Set policy for packs loading [latest|all|required]
  -l, --log string         Save output messages in a log file
  -o, --outdir string      Set directory for output files
  -p, --packs              Download missing software packs with cpackget
  -q, --quiet              Suppress output messages except build invocations
  -r, --rebuild            Remove intermediate and output directories and rebuild
  -s, --schema             Validate project input file(s) against schema
  -t, --target string      Optional CMake target name
  -u, --update string      Generate *.cprj file for reproducing current build
      --update-rte         Update the RTE directory and files
  -v, --verbose            Enable verbose messages from toolchain builds
  -V, --version            Print version

Use "cbuild [command] --help" for more information about a command.
```
