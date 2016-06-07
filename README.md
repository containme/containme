## Contain Me

The universal build environment.

## Overview

Build environments are something thats hard to replicate across machines, be it between a developer and build server or even between individual developers.
Containme aims to solve this problem by exploiting Docker to keep these environments identical across machines.
A containeme build consists of two parts, the build spec and profile spec.
A profile is much like a template for a build, it defines the base docker image to use as well as default commands for each stage of the build.
The build spec is typically found at the project root and named `containme.yaml`.
It defines what profile to use and allows you to modify many aspects of the build.

## containme.yaml

Example containme.yaml:
```yaml
environment:
  profile: ./containme.profile.yaml
  after:
  - apt-get install -y git

dependencies:
  override:
    - mvn install: # note the colon here
        timeout: 240

test:
  override:
    - mvn test:
        timeout: 600
```

## containme profile

Example containme profile:
```yaml
image: maven:3.3.3-jdk-9
workspace: /use/src/app
cache_directories:
  - /usr/share/maven
dependencies:
  - mvn install
test:
  - mvn test
```
