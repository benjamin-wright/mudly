Return to [Main Page](../README.md)

# Mudly Command Reference

The main way of using the mudly command line utility is to pass it a list of targets to build:

```
mudly <target1> <target2> ...
```

## Targets

A typical target consists of;

- The path to a `Mudfile` relative to the current working directory (optional)
- A `+` separator character
- The name of the artefact

so

`mudly ../other+thing1` would build the `thing1` artefact from either a) a `Mudfile` in the sibling directory `other`, or b) an `other.Mudfile` in the parent directory. 

`mudly +image` would build the `image` target from a Mudfile in the current working directory.

## Configuration

### Log Level

You can configure the logging level of mudly by the `--log-level` flag:

```
mudly --log-level debug <target>
```

or by the `MUDLY_LOG_LEVEL` environment variable

```
MUDLY_LOG_LEVEL=debug mudly <target>
```

## Options

### --deps, --dependencies

Build only the dependencies of the provided targets, but not the targets themselves.

### --no-deps, --no-dependencies

Build only the specified targets, ignoring any dependencies.