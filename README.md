# Nitro [![License](https://img.shields.io/badge/license-polyform:shield-blue)](https://polyformproject.org/licenses/shield/1.0.0/) [![Docs](https://img.shields.io/badge/godoc-reference-green)](https://go-nitro.dev/docs/v3) 

<img src="https://avatars2.githubusercontent.com/u/73709577" />

**Nitro** is a futuristic blazingly fast embeddable framework for distributed app development, IoT, edge and p2p.

Go to [Micro](https://github.com/micro/micro) for Micro. Go to [M3O](https://m3o.com) for a managed version (micro as a service)
## Overview

Nitro will provide the core requirements for distributed app development, IoT, edge and p2p including RPC and Event driven communication. 
The Nitro mantra is in-memory defaults with a pluggable architecture. Blaze with pure in-memory development and swap out as needed 
to go multi-process or multi-host.

Note: Nitro is currently undergoing a complete rewrite and is considered unstable for use.

## Features

Nitro focuses on dapps, IoT, edge and p2p. 

Features in development:

- [x] Lightweight RPC based communication
- [x] Event broadcasting and notifications
- [ ] CRDT Data synchronisation and storage
- [ ] Consensus protocol and execution engine
- [ ] WebAssembly target compilation support
- [ ] Unique randomized token generation aka BLS
- [ ] P2P gossip networking stack in userspace

## Future

In the future there's the potential to launch a live network based on Nitro. More on that soon.

## Docs

See [gonitro.dev/docs](https://gonitro.dev/docs/)

## Discussion

See [nitro/discussions](https://github.com/gonitro/nitro/discussions) for any discussions, development, etc.

## FAQ

See the [FAQ](FAQ.md) doc for frequently asked questions.

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
## License

[Polyform Noncommercial](https://polyformproject.org/licenses/noncommercial/1.0.0/). 
