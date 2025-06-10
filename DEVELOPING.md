# Developing cbuild

Follow the steps below to start developing for `cbuild`:

1. Requirements:

- [Install Make](https://www.gnu.org/software/make/)
- [Install Golang](https://golang.org/doc/install)
- [Install GolangCI-Lint](https://golangci-lint.run/welcome/install/#local-installation)

2. Clone the repo:
`$ git clone https://github.com/Open-CMSIS-Pack/cbuild.git`

3. Enter the checked source
`cd cbuild`

4. Configure your local environment
`make config`

5. Make sure all tests are passing
`make test-all`

6. Make sure it builds
`make build/cbuild`

7. Done! You can now start modifying the source code. Please refer to [contributing guide](CONTRIBUTING.md)
for guidelines.

# Releasing

If you have the right to push to the `main` branch of this repo, you might be entitled to
make releases. Do that by running:
`make release`

*NOTE*: We use [Semantic Versioning](https://semver.org/) for versioning cbuild.
