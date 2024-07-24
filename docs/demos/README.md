# Lunchpail: Getting Started with Demos

Welcome to the Lunchpail demo page. The demo binaries described here
are all products of Lunchpail. Look [here](../build/README.md) for
guidance of building your own binaries.

> Note: the demo commands below currently assume that you have a valid
> Kubernetes context. Cloud VM support and support for bringing up a
> local [Kind](https://github.com/kubernetes-sigs/kind) cluster will
> be documented soon.

- [Demo 1: Hello World](#hello-world-demo)
- [Demo 2: OpenROAD](#openroad-demo)

## Hello World Demo

You can check out the
[source](https://github.com/IBM/lunchpail-demo) or download one of the
[prebuilt
binaries](https://github.com/IBM/lunchpail-demo/releases). For
example, if you are on MacOS with Apple Silicon:

```shell
curl -L https://github.com/IBM/lunchpail-demo/releases/latest/download/lunchpail-demo-darwin-arm64 -o lunchpail-demo && chmod +x lunchpail-demo
./lunchpail-demo up -N
```

## How we built it

To build these binaries, we first downloaded the latest [Lunchpail
release](https://github.com/IBM/lunchpail/releases/latest) and then
ran:

```shell
lunchpail assemble https://github.com/IBM/lunchpail-demo -o /tmp/lunchpail-demo
```

Optionally, if you add the `-A` option, a set of platform binaries
will be generated. Without that flag, a single binary for the current
platform will be generated.

## OpenROAD Demo

[OpenROAD](https://theopenroadproject.org/) is an open-source
electronic design automation (EDA) tool suite. This suite of tools
helps with the optimization of chip designs, including timing and
geometry. The goal of this OpenROAD demo is to sweep a space of chip
design parameters in order to find a design with the smallest chip
area for a given set of timing constraints.
 
You can check out the
[source](https://github.com/IBM/lunchpail-openroad-max-utilization) or download one of the
[prebuilt
binaries](https://github.com/IBM/lunchpail-openroad-max-utilization/releases). For
example, if you are on MacOS with Apple Silicon:

```shell
curl -L https://github.com/IBM/lunchpail-openroad-max-utilization/releases/latest/download/lunchpail-openroad-darwin-arm64 -o lunchpail-openroad && chmod +x lunchpail-openroad
./lunchpail-openroad up -N
```

## How we built it

To build these binaries, we first downloaded the latest [Lunchpail
release](https://github.com/IBM/lunchpail/releases/latest) and then
ran:

```shell
lunchpail assemble https://github.com/IBM/lunchpail-openroad-max-utilization -o /tmp/lunchpail-openroad
```

Optionally, if you add the `-A` option, a set of platform binaries
will be generated. Without that flag, a single binary for the current
platform will be generated.