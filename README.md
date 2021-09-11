# MUDLY BUILD TOOL

Because sometimes you don't have room to containerise everything

## What is Mudly?

A cheap and cheerful knock-off of the much more impressive https://earthly.dev. It's a build tool that will orchestrate build tasks for you, allow you to in-line and share your dockerfiles and take care of spinning up and shutting down your development environment.

Mudly works entirely out of sub-processes, in your own local environment. This has obvious drawbacks in propagating the "it works on my machine" effect, but avoids the cpu / storage / memory consumption of container-based alternatives.

## Docs

Reference:
- [Command](./docs/command.md)

## Installation

### 1. Mudly Command

```sh
    go build -o bin/mudly ./cmd/mudly
    ln -s $(pwd)/bin/mudly /usr/local/bin/mudly
```

### 2. Visual Studio Code extension

To compile the extension:

```sh
cd extension
npm run package
```

Then right-click the compiled file and select `Install Extension VSIX` to install (at some point we'll do this properly)

### 3. Release Process

1. Create a Github Release through the Github UI
2. Github Actions will automatically build the linux and darwin flavours of the mudly executable and add them to the release
