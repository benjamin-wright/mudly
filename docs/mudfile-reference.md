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

---

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

---

### CONDITION

Defines the conditions in which the [step](#STEP) or [artefact](#ARTEFACT) should run. Similar to the [COMMAND](#COMMAND) term in that encapsulates a one-line or multi-line shell script.

```
ARTEFACT conditional
  CONDITION [[ "$SOME_VAR" == "should_build" ]]
```

In this example, if the value of the environment variable `SOME_VAR` is not equal to `should_build`, then mudly will skip all the steps in the `conditional` artefact. Multiple conditions are demonstrated in the multi-line version below.

```
ARTEFACT conditional
  CONDITION
    [[ "$CONDITION1" == "passed" ]] /
    && [[ "$CONDITION2" == "passed" ]]
```

When a failing condition is used in a step, only the step is skipped. Any subsequent steps will still run as usual.

---

### CONTEXT

Defines a docker build context, for use with the docker [STEP](#STEP).

```
ARTEFACT <name>
  STEP <name>
    DOCKERFILE my-dockerfile
    CONTEXT <filepath>
```

---

### COMMAND

Defines a shell command, as part of a [STEP](#STEP).

```
ARTEFACT <name>
  STEP <name>
    COMMAND echo "hello world!"
```

or

```
ARTEFACT <name>
  STEP <name>
    COMMAND
      echo "hello"
      echo "world"
```

The `COMMAND` term encapsulates a shell script of one or more lines. If that script returns a non-zero exit code then the step will fail and cancel the rest of the build.

---

### COMPOSE

Defines a `docker-compose` configuration in using the usual YAML syntax.

```
DEVENV <name>
  COMPOSE
    version: "3"
    services:
      db:
        image: my/image
```

---

### DEPENDS ON

Defines a dependency between artefacts, at either the [STEP](#STEP) or [ARTEFACT](#ARTEFACT) level. This allows mudly to:

a) Sequence builds such that any required inputs are available before each artefact build runs.

b) Discover and build the dependencies of any supplied target, so that the entire dependency chain does not need to be supplied to the `mudly` command.

```
ARTEFACT <name>
  DEPENDS ON ./subdir+data
```

In this case, `mudly +<name>` will build both the `<name>` and `data` artefacts, and the artefact `<name>` will not start building until the artefact `data` in the file `./subdir/Mudfile` has finished building. 

```
ARTEFACT <name>
  STEP <name1>
  STEP <name2>
    DEPENDS ON +image
```

This is the same as the previous example, but `<name>` step `<name1>` will start building at the same time as the `+image` target in the same `Mudfile`, while step `<name2>` will not start until `+image` has completely finished.

---

### DEVENV

Defines a development environment, i.e. a selection of services to run in the background for local development / integration testing / etc.

The `DEVENV` term is used at the top level to define the environment, which is then invoked by name at either the [ARTEFACT](#ARTEFACT) or [STEP](#STEP) level.

```
DEVENV test-env
  COMPOSE
    <content>
```

and either

```
ARTEFACT <name>
  DEVENV test-env
  STEP <name>
    COMMAND <content>
```

or

```
ARTEFACT <name>
  STEP <name>
    DEVENV test-env
    COMMAND <content>
```

The services in the named `DEVENV` definition will be spun up as part of the first build containing an artefact or step that references the `DEVENV`. When named by an [ARTEFACT](#ARTEFACT), this happens before the first step is run. When named by a [STEP](#STEP), this happens immediately before that step.

> NB: At the end of the build the development environment is left running, so that it can be re-used between builds. Run `mudly stop` (not implemented yet) to tear down any existing development environments.

supported children:
- [COMPOSE](#COMPOSE)

---

### DOCKERFILE

Defines (at the top level) and references (in a step) a dockerfile, to be used in a docker [STEP](#STEP).

```
DOCKERFILE filename
  FILE
    FROM alpine
    RUN apk add whatever

ARTEFACT <name>
  STEP <name>
    DOCKERFILE filename
```

---

### ENV

---

### FILE

---

### IGNORE

---

### PIPELINE

---

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

It supports two configurations, the command step which encapsulates a shell script of one or more lines:

```
ARTEFACT <name>
  STEP <name>
    CONDITION <content>
    COMMAND <content>
    DEPENDS ON <target>
    DEVENV <name>
    ENV <key>=<value>
    WAIT FOR <content>
    WATCH <filepath>
```

And the docker step which encapsulates a docker build command:

```
ARTEFACT <name>
  STEP <name>
    CONTEXT <filepath>
    DEPENDS ON <target>
    DOCKERFILE <target>/<content>
    IGNORE <filepath>
    TAG <imagetag>
```

supported children:
- [CONDITION](#CONDITION)
- [COMMAND](#COMMAND)
- [CONTEXT](#CONTEXT)
- [DEPENDS ON](#DEPENDS-ON)
- [DEVENV](#DEVENV)
- [DOCKERFILE](#DOCKERFILE)
- [ENV](#ENV)
- [IGNORE](#IGNORE)
- [TAG](#TAG)
- [WAIT FOR](#WAIT-FOR)
- [WATCH](#WATCH)

---

### TAG

---

### WAIT FOR

---

### WATCH
