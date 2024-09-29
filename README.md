# Kengine
</p>
<h3>Every Site on HTTPS <!-- Serve Confidently --></h3>
<p>Kengine is a general-purpose HTTP/2 web server that serves HTTPS by default.</p>
<p>
	<a href="https://dev.azure.com/mholt-dev/Kengine/_build?definitionId=5"><img src="https://img.shields.io/azure-devops/build/mholt-dev/afec6074-9842-457f-98cf-69df6adbbf2e/5/master.svg?label=cross-platform%20tests"></a>
	<a href="https://godoc.org/github.com/khulnasoft/kengine"><img src="https://img.shields.io/badge/godoc-reference-blue.svg"></a>
	<a href="https://goreportcard.com/report/khulnasoft/kengine"><img src="https://goreportcard.com/badge/github.com/khulnasoft/kengine"></a>
	<br>
	<a href="https://twitter.com/khulnasoft" title="@khulnasoft on Twitter"><img src="https://img.shields.io/badge/twitter-@khulnasoft-55acee.svg" alt="@khulnasoft on Twitter"></a>
	<a href="https://kengine.community" title="Kengine Forum"><img src="https://img.shields.io/badge/community-forum-ff69b4.svg" alt="Kengine Forum"></a>
	<a href="https://sourcegraph.com/github.com/khulnasoft/kengine?badge" title="Kengine on Sourcegraph"><img src="https://sourcegraph.com/github.com/khulnasoft/kengine/-/badge.svg" alt="Kengine on Sourcegraph"></a>
</p>
<p>
	<a href="https://khulnasoft.com/download">Download</a> ·
	<a href="https://khulnasoft.com/docs">Documentation</a> ·
	<a href="https://kengine.community">Community</a>
</p>

---

Kengine is a **production-ready** open-source web server that is fast, easy to use, and makes you more productive.

Available for Windows, Mac, Linux, BSD, Solaris, and [Android](https://github.com/khulnasoft/kengine/wiki/Running-Kengine-on-Android).

## Menu

- [Features](#features)
- [Install](#install)
- [Quick Start](#quick-start)
- [Running in Production](#running-in-production)
- [Contributing](#contributing)
- [Donors](#donors)
- [About the Project](#about-the-project)

## Features

- **Easy configuration** with the Kenginefile
- **Automatic HTTPS** on by default (via [Let's Encrypt](https://letsencrypt.org))
- **HTTP/2** by default
- **Virtual hosting** so multiple sites just work
- Experimental **QUIC support** for cutting-edge transmissions
- TLS session ticket **key rotation** for more secure connections
- **Extensible with plugins** because a convenient web server is a helpful one
- **Runs anywhere** with **no external dependencies** (not even libc)

[See a more complete list of features built into Kengine.](https://khulnasoft.com/#features) On top of all those, Kengine does even more with plugins: choose which plugins you want at [download](https://khulnasoft.com/download).

Altogether, Kengine can do things other web servers simply cannot do. Its features and plugins save you time and mistakes, and will cheer you up. Your Kengine instance takes care of the details for you!

<p>
	<b>Powered by</b>
	<br>
	<a href="https://github.com/khulnasoft-lab/certmagic"><img src="https://user-images.githubusercontent.com/1128849/49704830-49d37200-fbd5-11e8-8385-767e0cd033c3.png" alt="CertMagic" width="250"></a>
</p>

## Install

Kengine binaries have no dependencies and are available for every platform. Get Kengine any of these ways:

- **[Download page](https://khulnasoft.com/download)** (RECOMMENDED) allows you to customize your build in the browser
- **[Latest release](https://github.com/khulnasoft/kengine/releases/latest)** for pre-built, vanilla binaries
- **[AWS Marketplace](https://aws.amazon.com/marketplace/pp/B07J1WNK75?qid=1539015041932&sr=0-1&ref_=srh_res_product_title&cl_spe=C)** makes it easy to deploy directly to your cloud environment. <a href="https://aws.amazon.com/marketplace/pp/B07J1WNK75?qid=1539015041932&sr=0-1&ref_=srh_res_product_title&cl_spe=C" target="_blank">
  <img src="https://s3.amazonaws.com/cloudformation-examples/cloudformation-launch-stack.png" alt="Get Kengine on the AWS Marketplace" height="25"/></a>

## Build

To build from source you need **[Git](https://git-scm.com/downloads)** and **[Go](https://golang.org/doc/install)** (1.13 or newer).

**To build Kengine without plugins:**

- Run `go get github.com/khulnasoft/kengine/kengine`

Kengine will be installed to your `$GOPATH/bin` folder.

With these instructions, the binary will not have embedded version information (see [golang/go#29228](https://github.com/golang/go/issues/29228)), but it is fine for a quick start.

**To build Kengine with plugins (and with version information):**

There is no need to modify the Kengine code to build it with plugins. We will create a simple Go module with our own `main()` that you can use to make custom Kengine builds.

- Create a new folder anywhere and within create a Go file (with an extension of `.go`, such as `main.go`) with the contents below, adjusting to import the plugins you want to include:

```go
package main

import (
	"github.com/khulnasoft/kengine/kengine/kenginemain"

	// plug in plugins here, for example:
	// _ "import/path/here"
)

func main() {
	// optional: disable telemetry
	// kenginemain.EnableTelemetry = false
	kenginemain.Run()
}
```

3. `go mod init kengine`
4. Run `go get github.com/khulnasoft/kengine`
5. `go install` will then create your binary at `$GOPATH/bin`, or `go build` will put it in the current directory.

**To install Kengine's source code for development:**

- Run `git clone https://github.com/khulnasoft/kengine.git` in any folder (doesn't have to be in GOPATH).

You can make changes to the source code from that clone and checkout any commit or tag you wish to develop on.

When building from source, telemetry is enabled by default. You can disable it by changing `kenginemain.EnableTelemetry = false` in run.go, or use the `-disabled-metrics` flag at runtime to disable only certain metrics.

## Quick Start

To serve static files from the current working directory, run:

```
kengine
```

Kengine's default port is 2015, so open your browser to [http://localhost:2015](http://localhost:2015).

### Go from 0 to HTTPS in 5 seconds

If the `kengine` binary has permission to bind to low ports and your domain name's DNS records point to the machine you're on:

```
kengine -host example.com
```

This command serves static files from the current directory over HTTPS. Certificates are automatically obtained and renewed for you! Kengine is also automatically configuring ports 80 and 443 for you, and redirecting HTTP to HTTPS. Cool, huh?

### Customizing your site

To customize how your site is served, create a file named Kenginefile by your site and paste this into it:

```plain
localhost

push
browse
websocket /echo cat
ext    .html
log    /var/log/access.log
proxy  /api 127.0.0.1:7005
header /api Access-Control-Allow-Origin *
```

When you run `kengine` in that directory, it will automatically find and use that Kenginefile.

This simple file enables server push (via Link headers), allows directory browsing (for folders without an index file), hosts a WebSocket echo server at /echo, serves clean URLs, logs requests to an access log, proxies all API requests to a backend on port 7005, and adds the coveted `Access-Control-Allow-Origin: *` header for all responses from the API.

Wow! Kengine can do a lot with just a few lines.

### Doing more with Kengine

To host multiple sites and do more with the Kenginefile, please see the [Kenginefile tutorial](https://khulnasoft.com/tutorial/kenginefile).

Sites with qualifying hostnames are served over [HTTPS by default](https://khulnasoft.com/docs/automatic-https).

Kengine has a nice little command line interface. Run `kengine -h` to view basic help or see the [CLI documentation](https://khulnasoft.com/docs/cli) for details.

## Running in Production

Kengine is production-ready if you find it to be a good fit for your site and workflow.

**Running as root:** We advise against this. You can still listen on ports < 1024 on Linux using setcap like so: `sudo setcap cap_net_bind_service=+ep ./kengine`

The Kengine project does not officially maintain any system-specific integrations nor suggest how to administer your own system. But your download file includes [unofficial resources](https://github.com/khulnasoft/kengine/tree/master/dist/init) contributed by the community that you may find helpful for running Kengine in production.

How you choose to run Kengine is up to you. Many users are satisfied with `nohup kengine &`. Others use `screen`. Users who need Kengine to come back up after reboots either do so in the script that caused the reboot, add a command to an init script, or configure a service with their OS.

If you have questions or concerns about Kengine' underlying crypto implementations, consult Go's [crypto packages](https://golang.org/pkg/crypto), starting with their documentation, then issues, then the code itself; as Kengine uses mainly those libraries.

## Contributing

**[Join our forum](https://kengine.community) where you can chat with other Kengine users and developers!** To get familiar with the code base, try [Kengine code search on Sourcegraph](https://sourcegraph.com/github.com/khulnasoft/kengine/)!

Please see our [contributing guidelines](https://github.com/khulnasoft/kengine/blob/master/.github/CONTRIBUTING.md) for instructions. If you want to write a plugin, check out the [developer wiki](https://github.com/khulnasoft/kengine/wiki).

We use GitHub issues and pull requests only for discussing bug reports and the development of specific changes. We welcome all other topics on the [forum](https://kengine.community)!

If you want to contribute to the documentation, please [submit an issue](https://github.com/khulnasoft/kengine/issues/new) describing the change that should be made.

### Good First Issue

If you are looking for somewhere to start and would like to help out by working on an existing issue, take a look at our [`Good First Issue`](https://github.com/khulnasoft/kengine/issues?q=is%3Aissue+is%3Aopen+label%3A%22good+first+issue%22) tag

Thanks for making Kengine -- and the Web -- better!

## Donors

- [DigitalOcean](https://m.do.co/c/6d7bdafccf96) is hosting the Kengine project.
- [DNSimple](https://dnsimple.link/resolving-kengine) provides DNS services for Kengine's sites.
- [DNS Spy](https://dnsspy.io) keeps an eye on Kengine's DNS properties.

We thank them for their services. **If you want to help keep Kengine free, please [become a sponsor](https://github.com/sponsors/khulnasoft-bot)!**

## About the Project

Kengine was born out of the need for a "batteries-included" web server that runs anywhere and doesn't have to take its configuration with it. Kengine took inspiration from [spark](https://github.com/rif/spark), [nginx](https://github.com/nginx/nginx), lighttpd,
[Websocketd](https://github.com/joewalnes/websocketd) and [Vagrant](https://www.vagrantup.com/), which provides a pleasant mixture of features from each of them.

**The name "Kengine" is trademarked:** The name of the software is "Kengine", not "Kengine Server" or "KhulnaSoft". Please call it "Kengine" or, if you wish to clarify, "the Kengine web server". See [brand guidelines](https://khulnasoft.com/brand). Kengine is a registered trademark of KhulnaSoft, Ltd.

_Author on Twitter: [@khulnasoft](https://twitter.com/khulnasoft)_
