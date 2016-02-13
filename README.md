## Nut

Build [LXC](http://linuxcontainers.org/) containers using [Dockerfile](https://docs.docker.com/engine/reference/builder/) like [DSL](https://en.wikipedia.org/wiki/Domain-specific_language)

### Introduction

Nut is a minimal golang based CLI for building LXC containers. It allows user create containers
using [Dockerfile](https://docs.docker.com/engine/reference/builder/) like syntax, archiving and publishing them in s3.
Nut is intended to be used in CI/CD infrastructure to build & publish container images as artifacts. Nut complements
the lxc cli tools, as well as provide golang based public API to for extensions.

### Usage

All features of nut are implemented as set of sub commands
```
nut --help
```

```
usage: nut [--version] [--help] <command> [<args>]

Available commands are:
    archive    Create tarball images of existing container
    build      Build container from Dockerfile
    fetch      Create container from images stored in s3
    multi      Build multi container environment from docker compose specification
    publish    Publish tarball images of existing container in s3
    restore    Create container from tarball image
    run        Run command/entrypoint inside a container

```
- *build*, *restore* and *fetch* is used create containers. From dockerfile like syntax or images stored in s3 or localdisk 
- *multi* is used to setup a group of containers from [docker-compose](https://docs.docker.com/compose/compose-file/) like specification.
- *archive* and *publish*  allows creating and publishing container images in s3

#### Artifact  Labels

Nut stores container metadata as mainfest.yml file inside the container
directory, right next to rootfs directory. Manifest data stores all labels,
maintainers, exposed ports and entry point details. Labels that starts with "nut_artifact"
are treated differently, their values are considered as build artifacts and
fetched from inside the container to current directory. Following is an example
of building ruby 2.2.3 debian packages for trusty using nut

Example:

- Dockerfile
```sh
FROM trusty
MAINTAINER ranjib@pagerduty.com
RUN apt-get update -y
RUN apt-get install -y build-essential curl git-core libcurl4-openssl-dev libffi-dev libreadline-dev libsqlite3-dev libssl-dev libtool libxml2-dev libxslt1-dev libyaml-dev openssh-server python-software-properties sqlite3 wget zlib1g-dev
RUN mkdir -p /opt/ruby
RUN git clone https://github.com/sstephenson/ruby-build.git /opt/ruby-build
RUN /opt/ruby-build/bin/ruby-build 2.2.3 /opt/ruby/2.2.3
RUN /opt/ruby/2.2.3/bin/gem install bundler fpm --no-ri --no-rdoc
RUN /opt/ruby/2.2.3/bin/fpm -s dir -t deb -n ruby-2.2.3 -v 1.0.0 /opt/ruby/2.2.3
LABEL nut_artifact_ruby=/root/ruby-2.2.3_1.0.0_amd64.deb
```
And then nut can be invoked as:
```
nut build -ephemeral
```
Upon invocation nut will clone a new container from `trusty`, execute the RUN statement, which in turn will build ruby debian package, and then copy theresulting debian from /root/ruby-2.2.3_1.0.0_amd64.deb to current directory.

Since vanilla LXC is not aware of image repositories, all containers are created from cloning existing container(s).
A trusty (ubuntu 14.04) container can be created as
```
lxc-create -n trusty -t download -- -d ubuntu -a amd64 -r trusty
```
This in turn, can be used inside a Dockerfile DSL via the `FROM` instruction.
Nut converts the image name from  `org/repo:version` to 'org-repo_version' as the container
from which the new container will be built.
For example `FROM pagerduty/ruby:2.2.3` will instruct Nut to create a container by cloning
an existing container named `pagerduty-ruby_2.2.3`.

### Development

We use vagrant for development purpose. Following will setup a development vagrant instance
as well as kick a test job.

```
vagrant up
vagrant reload
vagrant ssh -c "nut build -specfile gopath/src/github.com/PagerDuty/nut/Dockerfile -ephemeral"
```

### LICENSE

[Apache 2](http://www.apache.org/licenses/LICENSE-2.0)
