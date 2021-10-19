Return to [Main Page](../README.md)

# Mudly Installation

## OSX / Linux

Check the releases page for the latest release version and compatible `$OS` / `$ARCH` combinations.

```sh
curl -L -o /usr/local/bin/mudly https://github.com/benjamin-wright/mudly/releases/download/$RELEASE/mudly-$OS-$ARCH

chmod +x /usr/local/bin/mudly
```

## Windows

Watch this space...

# VSCode Extension Installation

```
curl -L -o ./mudly-formatter.vsix https://github.com/benjamin-wright/mudly/releases/download/$RELEASE/mudly-formatter.vsix
```

In VS Code, Open `extensions` and select the `...` option from the top-right of the explorer tab. Click `Install from VSIX...` and select the downloaded `mudly-formatter.vsix` file. Profit.
