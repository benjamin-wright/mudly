Return to [Main Page](../README.md)

# Mudly Config Reference

The `Mudfile` file format is used to define the build process for one or more artefacts.

## Overall Structure

If you've done any work with Docker or Earthly then the general style of a `Mudfile` will be somewhat familiar. A typical `Mudfile` might look like the following:

```
DOCKERFILE server-image
  FILE
    FROM node:alpine
    
    RUN npm install
    
    ENTRYPOINT [ "node" ]
    CMD [ "run", "start" ]
  IGNORE
    node_modules

ARTEFACT server
  STEP build
    COMMAND npm install

  STEP test
    COMMAND npm run test

  STEP image
    DOCKERFILE server-image
    TAG my-image
```

The file above first defines a Dockerfile for building a node application and assigns the name `server-image` to that file. The artefact `server` is then defined which runs `npm install` then `npm run test`, and finally builds the docker image.

Each of the terms in the file is described below.

## Reference

### ARTEFACT

Defines a buildable artefact. An artefact can be anything, creating a file, building a golang binary, publishing a node module, etc.

```
ARTEFACT <name>
  <children>
```

supported children:
- [CONDITION](#CONDITION)
- [DEPENDS ON](#DEPENDS-ON)
- [DEVENV](#DEVENV)
- [ENV](#ENV)
- [PIPELINE](#PIPELINE)
- [STEP](#STEP)

### CONDITION

Defines the conditions in which the step or artefact should run:

```
ARTEFACT conditional
  CONDITION [[ "$SOME_VAR" == "should_build" ]]
```

In this example, if the value of the environment variable `SOME_VAR` is not equal to `should_build`, then mudly will skip all the steps in the `conditional` artefact.

`CONDITION` can also take multi-line arguments:

```
ARTEFACT conditional
  CONDITION
    [[ "$CONDITION1" == "passed" ]] /
    && [[ "$CONDITION2" == "passed" ]]
```

When a failing condition is used in a step, only the step is skipped. Any subsequent steps will still run as usual.

### COMMAND



### COMPOSE

### DEPENDS ON

### DEVENV

### ENV

### FILE

### IGNORE

### PIPELINE

### STEP

The `STEP` term can be used directly inside an `ARTEFACT`:

```
ARTEFACT <name>
  STEP <name>
    <content>
```

or inside a re-usable `PIPELINE`:

```
PIPELINE my-pipeline
  STEP <name>
    <content>

ARTEFACT <name>
  PIPELINE my-pipeline
```

supported children:
- [TAG](#TAG)

### TAG

### WAIT FOR

### WATCH
