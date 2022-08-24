[![Release](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/release.yml/badge.svg)](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/release.yml)
[![Build](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/build.yml/badge.svg)](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/build.yml)
[![Test](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/test.yml/badge.svg)](https://github.com/Open-CMSIS-Pack/cbuild/actions/workflows/test.yml)

# cbuild: Open-CMSIS-Pack Build Invocation Utility

This utility allows embedded developers to orchestrate the build of CPRJ projects using `cbuildgen`, `cpackget`, `cmake` and `ninja.`

## Usage

```bash
cbuild: Build Invocation 0.11.0-dev0 (C) 2022 ARM

Usage:
  cbuild <project.cprj> [flags]

Flags:
  -c, --clean              Remove intermediate and output directories
  -d, --debug              Enable debug messages
  -g, --generator string   Select build system generator (default "Ninja")
  -h, --help               Print usage
  -i, --intdir string      Set directory for intermediate files
  -j, --jobs int           Number of job slots for parallel execution
  -l, --log string         Save output messages in a log file
  -o, --outdir string      Set directory for output files
  -p, --packs              Download missing software packs with cpackget
  -q, --quiet              Suppress output messages except build invocations
  -r, --rebuild            Remove intermediate and output directories and rebuild
  -s, --schema             Check *.cprj file against CPRJ.xsd schema
  -t, --target string      Optional CMake target name
  -u, --update string      Generate *.cprj file for reproducing current build
  -v, --version            Print version
```
