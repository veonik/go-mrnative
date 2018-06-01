# go-mrnative
A utility to generate native Go language MapReduce applications using gobind.

This program can transform a Go application into a shared library, wrapped in a JAR with Java
bindings, ready to use on your Hadoop cluster.

### Why?

**This is a toy project.**

A sane person should use hadoop streaming to write MapReduce programs using a non-JVM language.

Applications generated by go-mrnative will work and probably even be performant, however,
because a full Go runtime is embedded in each JVM instance, memory usage is an issue. Additionally,
copying data between the JVM and Go runtime can be extremely costly.


#### How?

**This project is under development.**

It is already very easy to generate Java bindings to a Go shared library using `gobind` and
[gojava](https://github.com/sridharv/gojava).

go-mrnative takes this a step further, creating the boilerplate code necessary to make your
Go structs directly available to Hadoop MapReduce.


## Installation

The simplest way to install go-mrnative is to use `go get`.

```bash
go get -u github.com/veonik/go-mrnative/...
```

This will install the go-mrnative libraries as well as a command line utility.


## Usage

The go-mrnative tool does two things. First, using the `init` command, go-mrnative will
create a skeleton Mapper/Reducer/Combiner in your go package. Second, using the `build`
command, go-mrnative will generate Java boilerplate and bindings for your mapreduce structs,
compile them, and jar them into a file.

### Initializing a Go mapreduce project

Create a new directory on your machine, then go inside that directory and run the go-mrnative
`init` command.

```bash
mkdir gomapred
cd gomapred
go-mrnative init
```

go-mrnative will interactively walk through what you want to create. Currently, its possible to
create Mapper, Reducer, and Combiners (though Combiners are really just Reducers).


### Building a Go mapreduce project

Replace `<pkg>` in the following example with a valid go package name, such as
`github.com/veonik/go-mrnative`.

```bash
go-mrnative build <pkg>
gojava build -s build/java/go <pkg>
```

This process will be streamlined in upcoming changes.