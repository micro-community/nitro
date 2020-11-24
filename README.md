# Nitro [![License](https://img.shields.io/badge/license-polyform:shield-blue)](https://polyformproject.org/licenses/shield/1.0.0/) [![Docs](https://img.shields.io/badge/godoc-reference-green)](https://go-nitro.dev/docs/v3) 

<img src="https://avatars2.githubusercontent.com/u/73709577" />

**Nitro** is a futuristic blazingly fast framework for distributed app development, IoT, edge and p2p.

Go to [Micro](https://github.com/micro/micro) for Micro. Go to [M3O](https://m3o.com) for a managed version (micro as a service)
## Overview

Nitro will provide the core requirements for distributed app development, IoT, edge and p2p including RPC and Event driven communication. 
The Nitro mantra is in-memory defaults with a pluggable architecture. Blaze with pure in-memory development and swap out as needed 
to go multi-process or multi-host.

Note: Nitro is currently undergoing a complete rewrite and is considered unstable for use.

## Features

Nitro focuses on dapps, IoT, edge and p2p. 
        "time"

Features in development:

- Lightweight RPC based communication
- Event broadcasting and notifications
- CRDT Data synchronisation and storage
- Consensus protocol and execution engine
- WebAssembly target compilation support
- Unique randomized token generation aka BLS
- P2P gossip networking stack in userspace

## Future

In the future there's the potential to launch a live network based on Nitro. More on that soon.

## Docs

See [gonitro.dev/docs](https://gonitro.dev/docs/)

## Discussion

See [nitro/discussions](https://github.com/asim/nitro/discussions) for any discussions, development, etc

## FAQ

### What happened to Go Micro?

Go Micro has now been renamed to Nitro. Go Micro moved back to being a personal project and no longer lives under the organisation github.com/micro. 
The company is now doubling down on Micro itself and has pulled in the needed interfaces to consolidate a Server, Framework and CLI into one tool. 
Go Micro is now no longer maintained by a company. Yet it continued to create confusion even as a personal repo. So for that reason, we're renaming 
to Nitro. Go Micro V2 has been archived at [micro-community/go-micro](https://github.com/micro-community/go-micro) and the plugins at 
[micro-community/go-plugins](https://github.com/micro-community/go-plugins).
### Why has the license changed from Apache 2.0 to Polyform Shield.

Micro-Community/Nitro is open source but makes use of the Polyform Shield license to protect against AWS running it as a managed service.


### Where are all the plugins?

The plugins now live in [github.com/micro-community/nitro-plugins](https://github.com/micro-community/nitro-plugins). This was to reduce the overall size and scope of Go Micro to purely 
a set of interfaces and standard library implementations. Go Plugins is Apache 2.0 licensed but relies on Nitro interfaces and so again can only be used in 
Polyform Shield scope.

### What's the new direction of Nitro?

Nitro will now focus on distributed app development using the Go standard library. It will continue to define abstractions for distributed systems 
but will only do so without external dependencies. In this manner the hope is Nitro can be picked up with minimal overhead for all sorts of new 
applications that have a low memory or low resource footprint. The assumption is there are places which would like to use distributed apps just as 
embedded systems or web assembly, unikernels, and related targets that would benefit from a framework that defined these as primitives for such use.

### How do Nitro and Micro now differ?

Micro is a platform for cloud native development. A complete experience that includes a server, framework and multi-language clients. Beyond that it also 
include environments, multi-tenancy and many more features which push it towards being a hosted Micro as a Service offering. It is a complete platform.

Nitro is more of a embeddable framework for distributed app development and now once again a purely personal project maintained by me and 
perhaps others who still find use for it commercially or noncommercially. It's of sentimental value and something I'd like to carry on for personal projects 
such as things related to edge, IoT, embedded systems, p2p, web assembly, etc.

### I used Go Micro to build microservices. What should I do now?

You should quite honestly go look at [Micro](https://github.com/micro/micro) and then consider the hosted offering at [m3o.com](https://m3o.com) which 
starts as a purely free dev environment in the cloud. Micro continues to address many of the concerns and requirements you had if not more. It is likely 
you managed metrics, tracing, logging and much other boilerplate that needed to be plugged in. Micro will now take this complete platform story approach 
and help you in that journey e.g you're probably running managed kubernetes on a major cloud provider with many other things. We're doing that for you 
instead as a company and platform team.

### I want to use Go Micro version 2.0 for my company. Can I still do that?

Yes. Go Micro 2.0 is still Apache 2.0 licensed which means you can still freely use it for everything you were using before. If you're a new user 
you can do the same. These things are using go modules so you're import path is simply `github.com/micro/go-micro/v2` as it was before. Because 
GitHub handles redirects this should not break. Please continue to use it if you like, but my own support for 2.0 is now end of life.

### Why has the license changed from Apache 2.0 to Polyform Noncommercial

Go Micro was largely a solo maintained effort for the entirety of its lifetime. It has enabled the creation of a company called Micro Services, Inc. which 
now focuses on [Micro](https://github.com/micro/micro) as a Service and has consolidated any interfaces here into a service library in that project. For 
the most part, Go Micro was underfunded and in some ways under appreciated. In version 3.0, going back to something of a personal project of more than 6 years 
I have made the hard decision to relicense as a noncommercial project. 

## License

[Polyform Noncommercial](https://polyformproject.org/licenses/noncommercial/1.0.0/). 
