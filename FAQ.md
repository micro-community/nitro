# FAQ

## What happened to Go Micro?

Go Micro has now been renamed to Nitro. Go Micro moved back to being a personal project and no longer lives under the organisation github.com/micro. 
The company is now doubling down on Micro itself and has pulled in the needed interfaces to consolidate a Server, Framework and CLI into one tool. 
Go Micro is now no longer maintained by a company. Yet it continued to create confusion even as a personal repo. So for that reason, we're renaming 
to Nitro. Go Micro V2 has been archived at [microhq/go-micro](https://github.com/microhq/go-micro) and the plugins at 
[microhq/plugins](https://github.com/microhq/go-plugins).

## What's the new direction of Nitro?

Nitro will now focus on distributed app development using the Go standard library. It will continue to define abstractions for distributed systems 
but will only do so without external dependencies. In this manner the hope is Nitro can be picked up with minimal overhead for all sorts of new 
applications that have a low memory or low resource footprint. The assumption is there are places which would like to use distributed apps just as 
embedded systems or web assembly, unikernels, and related targets that would benefit from a framework that defined these as primitives for such use.

## How does this relate to AWS Nitro?

It doesn't. We're a more popular, better, faster framework with the same name. 

## How do Nitro and Micro now differ?

Micro is a platform for cloud native development. A complete experience that includes a server, framework and multi-language clients. Beyond that it also 
include environments, multi-tenancy and many more features which push it towards being a hosted Micro as a Service offering. It is a complete platform.

Nitro is more of a embeddable framework for distributed app development and now once again a purely personal project maintained by me and 
perhaps others who still find use for it commercially or noncommercially. It's of sentimental value and something I'd like to carry on for personal projects 
such as things related to edge, IoT, embedded systems, p2p, web assembly, etc.

## I used Go Micro to build microservices. What should I do now?

You should quite honestly go look at [Micro](https://github.com/micro/micro) and then consider the hosted offering at [m3o.com](https://m3o.com) which 
starts as a purely free dev environment in the cloud. Micro continues to address many of the concerns and requirements you had if not more. It is likely 
you managed metrics, tracing, logging and much other boilerplate that needed to be plugged in. Micro will now take this complete platform story approach 
and help you in that journey e.g you're probably running managed kubernetes on a major cloud provider with many other things. We're doing that for you 
instead as a company and platform team.

## I want to use Go Micro version 2.0 for my company. Can I still do that?

Yes. Go Micro 2.0 is still Apache 2.0 licensed which means you can still freely use it for everything you were using before. If you're a new user 
you can do the same. These things are using go modules so you're import path is simply `github.com/micro/go-micro/v2` as it was before. Because 
GitHub handles redirects this should not break. Please continue to use it if you like, but my own support for 2.0 is now end of life.

## Why has the license changed from Apache 2.0 to Polyform Noncommercial

Go Micro was largely a solo maintained effort for the entirety of its lifetime. It has enabled the creation of a company called Micro Services, Inc. which 
now focuses on [Micro](https://github.com/micro/micro) as a Service and has consolidated any interfaces here into a service library in that project. For 
the most part, Go Micro was underfunded and in some ways under appreciated. In version 3.0, going back to something of a personal project of more than 6 years 
I have made the hard decision to relicense as a noncommercial project. 
