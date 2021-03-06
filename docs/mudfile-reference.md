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

or to reference a devenv in another mudfile

```
ARTEFACT <name>
  DEVENV ./path/to/mudfile test-env
  STEP <name>
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
    TAG image/tag
```

supported children:
- [FILE](#FILE)
- [IGNORE](#IGNORE)

---

### ENV

Defines a new environment variable. This can be used at the top level, in an [ARTEFACT](#ARTEFACT), in a [STEP](#STEP) or in a [PIPELINE](#PIPELINE)

```
ENV VAR_NAME=value
```

or

```
ARTEFACT <name>
  ENV VAR_NAME=value
```

or

```
ARTEFACT <name>
  STEP <name>
    ENV VAR_NAME=value
```

or

```
PIPELINE <name>
  ENV VAR_NAME=value
```

`ENV` values at lower levels in the heirarchy will override those from a higher level, based on the following ordering:

- step
- artefact
- pipeline
- global

So an `ENV` term in a `step` will override an `ENV` term for the same variable name in a `pipeline`, `artefact` or `global`

---

### FILE

Defines the body of a Dockerfile, as the child of a `DOCKERFILE` term. The dockerfile content should be indented below the `FILE` line, but it otherwise in the usual Dockerfile syntax.

```
DOCKERFILE my-image
  FILE
    FROM alpine
    RUN apk add whatevs
```

---

### IGNORE

Defines the content of `.dockerignore` to go with a `Dockerfile` definition, as the child of a `DOCKERFILE` term. The ignore content should be indented below the `IGNORE` line, but it's otherwise in the usual .dockerignore syntax.

```
DOCKERFILE my-image
  IGNORE
    node_modules
    *.tgz
```

---

### PIPELINE

Defines a re-usable pipeline, i.e. a collection of [STEP](#STEP) and [ENV](#ENV) terms.

A simple example:

```
PIPELINE my-pipeline
  ENV VAR_NAME=value
  STEP <name>
    <content>

ARTEFACT <name>
  PIPELINE my-pipeline
```

Optionally, add a relative path to reference a pipeline from another file:

```
ARTEFACT <name>
  PIPELINE ../sibling-dir my-pipeline
```

supported children:
- [STEP](#STEP)
- [ENV](#ENV)

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

Defines the tag that should be applied to an image build in a docker [STEP](#STEP), e.g.

```
ARTEFACT <name>
  STEP <name>
    DOCKERFILE go-image
    TAG my-image
```

---

### WAIT FOR

Defines an additional command for a command-type [STEP](#STEP). The step will defer running the main step command until the wait-for command returns a non-zero code, polling every 0.5 seconds. If the wait-for command never returns a non-zero code then it will run forever, or until you give up and hit `Ctrl-C`.

```
ARTEFACT <name>
  STEP <name>
    WAIT FOR curl $MY_API/status
    COMMAND do something using my api
```

Like [COMMAND](#COMMAND), `WAIT FOR` accepts multi-line arguments:

```
ARTEFACT <name>
  STEP <name>
    WAIT FOR
      result=$(curl $MY_API/status | jq '.some-field' -r)
      validator $result
    COMMAND do something using my api
```

---

### WATCH

Defines a set of filepaths that mudly should monitor for changes. If the timestamps on the monitored files have not changed since the last successful build, mudly will skip the step. This is useful for build tools / generators that don't natively support caching.

Accepts multiple files as space-seperated strings:

```
ARTEFACT <name>
  STEP <name>
    WATCH ./generator-inputs ./generator-data
    COMMAND npm run generate
```