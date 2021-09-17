Return to [Main Page](../README.md)

# Mudly Dev Notes

## Install Mudly from source

```sh
    go build -o bin/mudly ./cmd/mudly
    ln -s $(pwd)/bin/mudly /usr/local/bin/mudly
```

## Visual Studio Code extension

To compile the extension:

```sh
cd extension
npm run package
```

Then right-click the compiled file and select `Install Extension VSIX` to install (at some point we'll do this properly)

## Release Process

1. Create a Github Release through the Github UI
2. Github Actions will automatically build the linux and darwin flavours of the mudly executable and add them to the release