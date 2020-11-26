package rpc

import (
	"context"
	"fmt"
	"io"
	"net"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gonitro/nitro/app/codec"
	raw "github.com/gonitro/nitro/app/codec/bytes"
	"github.com/gonitro/nitro/app/event"
	"github.com/gonitro/nitro/app/logger"
	"github.com/gonitro/nitro/app/metadata"
	"github.com/gonitro/nitro/app/network"
	"github.com/gonitro/nitro/app/registry"
	"github.com/gonitro/nitro/app/server"
	"github.com/gonitro/nitro/util/addr"
	"github.com/gonitro/nitro/util/backoff"
	mnet "github.com/gonitro/nitro/util/net"
	"github.com/gonitro/nitro/util/socket"
)

type rpcServer struct {
	router *router
	exit   chan chan error

	sync.RWMutex
	opts        server.Options
	handlers    map[string]server.Handler
	subscribers map[server.Subscriber][]event.Subscriber
	// marks the serve as started
	started bool
	// used for first registration
	registered bool
	// subscribe to service name
	subscriber event.Subscriber
	// graceful exit
	wg *sync.WaitGroup

	rsvc *registry.App
}

var (
	log = logger.NewHelper(logger.DefaultLogger).WithFields(map[string]interface{}{"service": "server"})
)

func wait(ctx context.Context) *sync.WaitGroup {
	if ctx == nil {
		return nil
	}
	wg, ok := ctx.Value("wait").(*sync.WaitGroup)
	if !ok {
		return nil
	}
	return wg
}

func newServer(opts ...server.Option) server.Server {
	options := newOptions(opts...)
	router := newRpcRouter()
	router.hdlrWrappers = options.HdlrWrappers
	router.subWrappers = options.SubWrappers

	return &rpcServer{
		opts:        options,
		router:      router,
		handlers:    make(map[string]server.Handler),
		subscribers: make(map[server.Subscriber][]event.Subscriber),
		exit:        make(chan chan error),
		wg:          wait(options.Context),
	}
}

// HandleEvent handles inbound messages to the service directly
// TODO: handle requests from an event. We won't send a response.
func (s *rpcServer) HandleEvent(msg *event.Message) error {
	if msg.Header == nil {
		// create empty map in case of headers empty to avoid panic later
		msg.Header = make(map[string]string)
	}

	// get codec
	ct := msg.Header["Content-Type"]

	// default content type
	if len(ct) == 0 {
		msg.Header["Content-Type"] = DefaultContentType
		ct = DefaultContentType
	}

	// get codec
	cf, err := s.newCodec(ct)
	if err != nil {
		return err
	}

	// copy headers
	hdr := make(map[string]string, len(msg.Header))
	for k, v := range msg.Header {
		hdr[k] = v
	}

	// create context
	ctx := metadata.NewContext(context.Background(), hdr)

	// TODO: inspect message header
	// App means a request
	// Event means a message

	rpcMsg := &rpcMessage{
		event:       msg.Header["Event"],
		contentType: ct,
		payload:     &raw.Frame{Data: msg.Body},
		codec:       cf,
		header:      msg.Header,
		body:        msg.Body,
	}

	// existing router
	r := server.Router(s.router)

	// if the router is present then execute it
	if s.opts.Router != nil {
		// create a wrapped function
		handler := s.opts.Router.ProcessMessage

		// execute the wrapper for it
		for i := len(s.opts.SubWrappers); i > 0; i-- {
			handler = s.opts.SubWrappers[i-1](handler)
		}

		// set the router
		r = rpcRouter{m: handler}
	}

	return r.ProcessMessage(ctx, rpcMsg)
}

// ServeConn serves a single connection
func (s *rpcServer) ServeConn(sock network.Socket) {
	// streams are multiplexed on Stream or Id header
	pool := socket.NewPool()

	// get global waitgroup
	s.Lock()
	gg := s.wg
	s.Unlock()

	// waitgroup to wait for processing to finish
	wg := &waitGroup{
		gg: gg,
	}

	defer func() {
		// wait till done
		wg.Wait()

		// close all the sockets for this connection
		pool.Close()

		// close underlying socket
		sock.Close()

		// recover any panics
		if r := recover(); r != nil {
			if logger.V(logger.ErrorLevel, log) {
				log.Error("panic recovered: ", r)
				log.Error(string(debug.Stack()))
			}
		}
	}()

	for {
		var msg network.Message
		// process inbound messages one at a time
		if err := sock.Recv(&msg); err != nil {
			return
		}

		// check the message header for
		// App is a request
		// Event is a message
		if t := msg.Header["Event"]; len(t) > 0 {
			// TODO: handle the error event
			if err := s.HandleEvent(newMessage(msg)); err != nil {
				msg.Header["Error"] = err.Error()
			}
			// write back some 200
			if err := sock.Send(&network.Message{
				Header: msg.Header,
			}); err != nil {
				break
			}
			// we're done
			continue
		}

		// business as usual

		// use Stream as the stream identifier
		// in the event its blank we'll always process
		// on the same socket
		id := msg.Header["Stream"]

		// if there's no stream id then its a standard request
		// use the Id
		if len(id) == 0 {
			id = msg.Header["Id"]
		}

		// check stream id
		var stream bool

		if v := getHeader("Stream", msg.Header); len(v) > 0 {
			stream = true
		}

		// check if we have an existing socket
		psock, ok := pool.Get(id)

		// if we don't have a socket and its a stream
		if !ok && stream {
			// check if its a last stream EOS error
			err := msg.Header["Error"]
			if err == lastStreamResponseError.Error() {
				pool.Release(psock)
				continue
			}
		}

		// got an existing socket already
		if ok {
			// we're starting processing
			wg.Add(1)

			// pass the message to that existing socket
			if err := psock.Accept(&msg); err != nil {
				// release the socket if there's an error
				pool.Release(psock)
			}

			// done waiting
			wg.Done()

			// continue to the next message
			continue
		}

		// no socket was found so its new
		// set the local and remote values
		psock.SetLocal(sock.Local())
		psock.SetRemote(sock.Remote())

		// load the socket with the current message
		psock.Accept(&msg)

		// now walk the usual path

		// we use this Timeout header to set a server deadline
		to := msg.Header["Timeout"]
		// we use this Content-Type header to identify the codec needed
		ct := msg.Header["Content-Type"]

		// copy the message headers
		hdr := make(map[string]string, len(msg.Header))
		for k, v := range msg.Header {
			hdr[k] = v
		}

		// set local/remote ips
		hdr["Local"] = sock.Local()
		hdr["Remote"] = sock.Remote()

		// create new context with the metadata
		ctx := metadata.NewContext(context.Background(), hdr)

		// set the timeout from the header if we have it
		if len(to) > 0 {
			if n, err := strconv.ParseUint(to, 10, 64); err == nil {
				var cancel context.CancelFunc
				ctx, cancel = context.WithTimeout(ctx, time.Duration(n))
				defer cancel()
			}
		}

		// if there's no content type default it
		if len(ct) == 0 {
			msg.Header["Content-Type"] = DefaultContentType
			ct = DefaultContentType
		}

		// setup old protocol
		cf := setupProtocol(&msg)

		// no legacy codec needed
		if cf == nil {
			var err error
			// try get a new codec
			if cf, err = s.newCodec(ct); err != nil {
				// no codec found so send back an error
				sock.Send(&network.Message{
					Header: map[string]string{
						"Content-Type": "text/plain",
					},
					Body: []byte(err.Error()),
				})

				// release the socket we just created
				pool.Release(psock)
				// now continue
				continue
			}
		}

		// create a new rpc codec based on the pseudo socket and codec
		rcodec := newRpcCodec(&msg, psock, cf)
		// check the protocol as well
		protocol := rcodec.String()

		// internal request
		request := &rpcRequest{
			service:     getHeader("App", msg.Header),
			method:      getHeader("Method", msg.Header),
			endpoint:    getHeader("Endpoint", msg.Header),
			contentType: ct,
			codec:       rcodec,
			header:      msg.Header,
			body:        msg.Body,
			socket:      psock,
			stream:      stream,
		}

		// internal response
		response := &rpcResponse{
			header: make(map[string]string),
			socket: psock,
			codec:  rcodec,
		}

		// set router
		r := server.Router(s.router)

		// if not nil use the router specified
		if s.opts.Router != nil {
			// create a wrapped function
			handler := func(ctx context.Context, req server.Request, rsp interface{}) error {
				return s.opts.Router.ServeRequest(ctx, req, rsp.(server.Response))
			}

			// execute the wrapper for it
			for i := len(s.opts.HdlrWrappers); i > 0; i-- {
				handler = s.opts.HdlrWrappers[i-1](handler)
			}

			// set the router
			r = rpcRouter{h: handler}
		}

		// wait for two coroutines to exit
		// serve the request and process the outbound messages
		wg.Add(2)

		// process the outbound messages from the socket
		go func(id string, psock *socket.Socket) {
			defer func() {
				// TODO: don't hack this but if its grpc just break out of the stream
				// We do this because the underlying connection is h2 and its a stream
				switch protocol {
				case "grpc":
					sock.Close()
				}
				// release the socket
				pool.Release(psock)
				// signal we're done
				wg.Done()

				// recover any panics for outbound process
				if r := recover(); r != nil {
					if logger.V(logger.ErrorLevel, log) {
						log.Error("panic recovered: ", r)
						log.Error(string(debug.Stack()))
					}
				}
			}()

			for {
				// get the message from our internal handler/stream
				m := new(network.Message)
				if err := psock.Process(m); err != nil {
					return
				}

				// send the message back over the socket
				if err := sock.Send(m); err != nil {
					return
				}
			}
		}(id, psock)

		// serve the request in a go routine as this may be a stream
		go func(id string, psock *socket.Socket) {
			defer func() {
				// release the socket
				pool.Release(psock)
				// signal we're done
				wg.Done()

				// recover any panics for call handler
				if r := recover(); r != nil {
					log.Error("panic recovered: ", r)
					log.Error(string(debug.Stack()))
				}
			}()

			// serve the actual request using the request router
			if serveRequestError := r.ServeRequest(ctx, request, response); serveRequestError != nil {
				// write an error response
				writeError := rcodec.Write(&codec.Message{
					Header: msg.Header,
					Error:  serveRequestError.Error(),
					Type:   codec.Error,
				}, nil)

				// if the server request is an EOS error we let the socket know
				// sometimes the socket is already closed on the other side, so we can ignore that error
				alreadyClosed := serveRequestError == lastStreamResponseError && writeError == io.EOF

				// could not write error response
				if writeError != nil && !alreadyClosed {
					log.Debugf("rpc: unable to write error response: %v", writeError)
				}
			}
		}(id, psock)
	}
}

func (s *rpcServer) newCodec(contentType string) (codec.NewCodec, error) {
	if cf, ok := s.opts.Codecs[contentType]; ok {
		return cf, nil
	}
	if cf, ok := DefaultCodecs[contentType]; ok {
		return cf, nil
	}
	return nil, fmt.Errorf("Unsupported Content-Type: %s", contentType)
}

func (s *rpcServer) Options() server.Options {
	s.RLock()
	opts := s.opts
	s.RUnlock()
	return opts
}

func (s *rpcServer) Init(opts ...server.Option) error {
	s.Lock()
	defer s.Unlock()

	for _, opt := range opts {
		opt(&s.opts)
	}
	// update router if its the default
	if s.opts.Router == nil {
		r := newRpcRouter()
		r.hdlrWrappers = s.opts.HdlrWrappers
		r.serviceMap = s.router.serviceMap
		r.subWrappers = s.opts.SubWrappers
		s.router = r
	}

	s.rsvc = nil

	return nil
}

func (s *rpcServer) NewHandler(h interface{}, opts ...server.HandlerOption) server.Handler {
	return s.router.NewHandler(h, opts...)
}

func (s *rpcServer) Handle(h server.Handler) error {
	s.Lock()
	defer s.Unlock()

	if err := s.router.Handle(h); err != nil {
		return err
	}

	s.handlers[h.Name()] = h

	return nil
}

func (s *rpcServer) NewSubscriber(event string, sb interface{}, opts ...server.SubscriberOption) server.Subscriber {
	return s.router.NewSubscriber(event, sb, opts...)
}

func (s *rpcServer) Subscribe(sb server.Subscriber) error {
	s.Lock()
	defer s.Unlock()

	if err := s.router.Subscribe(sb); err != nil {
		return err
	}

	s.subscribers[sb] = nil
	return nil
}

func (s *rpcServer) Add() error {
	s.RLock()
	rsvc := s.rsvc
	config := s.Options()
	s.RUnlock()

	// only register if it exists or is not noop
	if config.Registry == nil || config.Registry.String() == "noop" {
		return nil
	}

	regFunc := func(service *registry.App) error {
		// create registry options
		rOpts := []registry.AddOption{
			registry.AddTTL(config.AddTTL),
			registry.AddDomain(s.opts.Namespace),
		}

		var regErr error

		for i := 0; i < 3; i++ {
			// attempt to register
			if err := config.Registry.Add(service, rOpts...); err != nil {
				// set the error
				regErr = err
				// backoff then retry
				time.Sleep(backoff.Do(i + 1))
				continue
			}
			// success so nil error
			regErr = nil
			break
		}

		return regErr
	}

	// have we registered before?
	if rsvc != nil {
		if err := regFunc(rsvc); err != nil {
			return err
		}
		return nil
	}

	var err error
	var advt, host, port string
	var cacheApp bool

	// check the advertise address first
	// if it exists then use it, otherwise
	// use the address
	if len(config.Advertise) > 0 {
		advt = config.Advertise
	} else {
		advt = config.Address
	}

	if cnt := strings.Count(advt, ":"); cnt >= 1 {
		// ipv6 address in format [host]:port or ipv4 host:port
		host, port, err = net.SplitHostPort(advt)
		if err != nil {
			return err
		}
	} else {
		host = advt
	}

	if ip := net.ParseIP(host); ip != nil {
		cacheApp = true
	}

	addr, err := addr.Extract(host)
	if err != nil {
		return err
	}

	// make copy of metadata
	md := metadata.Copy(config.Metadata)

	// mq-rpc(eg. nats) doesn't need the port. its addr is queue name.
	if port != "" {
		addr = mnet.HostPort(addr, port)
	}

	// register service
	node := &registry.Instance{
		Id:       config.Name + "-" + config.Id,
		Address:  addr,
		Metadata: md,
	}

	node.Metadata["network"] = config.Transport.String()
	node.Metadata["event"] = config.Broker.String()
	node.Metadata["server"] = s.String()
	node.Metadata["registry"] = config.Registry.String()
	node.Metadata["protocol"] = "rpc"

	s.RLock()

	// Maps are ordered randomly, sort the keys for consistency
	var handlerList []string
	for n, e := range s.handlers {
		// Only advertise non internal handlers
		if !e.Options().Internal {
			handlerList = append(handlerList, n)
		}
	}

	sort.Strings(handlerList)

	var subscriberList []server.Subscriber
	for e := range s.subscribers {
		// Only advertise non internal subscribers
		if !e.Options().Internal {
			subscriberList = append(subscriberList, e)
		}
	}

	sort.Slice(subscriberList, func(i, j int) bool {
		return subscriberList[i].Event() > subscriberList[j].Event()
	})

	endpoints := make([]*registry.Endpoint, 0, len(handlerList)+len(subscriberList))

	for _, n := range handlerList {
		endpoints = append(endpoints, s.handlers[n].Endpoints()...)
	}

	for _, e := range subscriberList {
		endpoints = append(endpoints, e.Endpoints()...)
	}

	service := &registry.App{
		Name:      config.Name,
		Version:   config.Version,
		Instances: []*registry.Instance{node},
		Endpoints: endpoints,
	}

	// get registered value
	registered := s.registered

	s.RUnlock()

	if !registered {
		if logger.V(logger.InfoLevel, logger.DefaultLogger) {
			log.Infof("Registry [%s] Adding node: %s", config.Registry.String(), node.Id)
		}
	}

	// register the service
	if err := regFunc(service); err != nil {
		return err
	}

	// already registered? don't need to register subscribers
	if registered {
		return nil
	}

	s.Lock()
	defer s.Unlock()

	// set what we're advertising
	s.opts.Advertise = addr

	// router can exchange messages
	if s.opts.Router != nil {
		// subscribe to the event with own name
		sub, err := s.opts.Broker.Subscribe(config.Name, s.HandleEvent)
		if err != nil {
			return err
		}

		// save the subscriber
		s.subscriber = sub
	}

	// subscribe for all of the subscribers
	for sb := range s.subscribers {
		var opts []event.SubscribeOption
		if queue := sb.Options().Queue; len(queue) > 0 {
			opts = append(opts, event.Queue(queue))
		}

		if cx := sb.Options().Context; cx != nil {
			opts = append(opts, event.SubscribeContext(cx))
		}

		sub, err := config.Broker.Subscribe(sb.Event(), s.HandleEvent, opts...)
		if err != nil {
			return err
		}
		if logger.V(logger.InfoLevel, logger.DefaultLogger) {
			log.Infof("Subscribing to event: %s", sub.Event())
		}
		s.subscribers[sb] = []event.Subscriber{sub}
	}
	if cacheApp {
		s.rsvc = service
	}
	s.registered = true

	return nil
}

func (s *rpcServer) Remove() error {
	var err error
	var advt, host, port string

	s.RLock()
	config := s.Options()
	s.RUnlock()

	// only register if it exists or is not noop
	if config.Registry == nil || config.Registry.String() == "noop" {
		return nil
	}

	// check the advertise address first
	// if it exists then use it, otherwise
	// use the address
	if len(config.Advertise) > 0 {
		advt = config.Advertise
	} else {
		advt = config.Address
	}

	if cnt := strings.Count(advt, ":"); cnt >= 1 {
		// ipv6 address in format [host]:port or ipv4 host:port
		host, port, err = net.SplitHostPort(advt)
		if err != nil {
			return err
		}
	} else {
		host = advt
	}

	addr, err := addr.Extract(host)
	if err != nil {
		return err
	}

	// mq-rpc(eg. nats) doesn't need the port. its addr is queue name.
	if port != "" {
		addr = mnet.HostPort(addr, port)
	}

	node := &registry.Instance{
		Id:      config.Name + "-" + config.Id,
		Address: addr,
	}

	service := &registry.App{
		Name:      config.Name,
		Version:   config.Version,
		Instances: []*registry.Instance{node},
	}

	if logger.V(logger.InfoLevel, logger.DefaultLogger) {
		log.Infof("Registry [%s] Removeing node: %s", config.Registry.String(), node.Id)
	}
	if err := config.Registry.Remove(service, registry.RemoveDomain(s.opts.Namespace)); err != nil {
		return err
	}

	s.Lock()
	s.rsvc = nil

	if !s.registered {
		s.Unlock()
		return nil
	}

	s.registered = false

	// close the subscriber
	if s.subscriber != nil {
		s.subscriber.Unsubscribe()
		s.subscriber = nil
	}

	for sb, subs := range s.subscribers {
		for _, sub := range subs {
			if logger.V(logger.InfoLevel, logger.DefaultLogger) {
				log.Infof("Unsubscribing %s from event: %s", node.Id, sub.Event())
			}
			sub.Unsubscribe()
		}
		s.subscribers[sb] = nil
	}

	s.Unlock()
	return nil
}

func (s *rpcServer) Start() error {
	s.RLock()
	if s.started {
		s.RUnlock()
		return nil
	}
	s.RUnlock()

	config := s.Options()

	// start listening on the network
	ts, err := config.Transport.Listen(config.Address)
	if err != nil {
		return err
	}

	if logger.V(logger.InfoLevel, logger.DefaultLogger) {
		log.Infof("Transport [%s] Listening on %s", config.Transport.String(), ts.Addr())
	}

	// swap address
	s.Lock()
	addr := s.opts.Address
	s.opts.Address = ts.Addr()
	s.Unlock()

	bname := config.Broker.String()

	// connect to the event
	if err := config.Broker.Connect(); err != nil {
		if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
			log.Errorf("Broker [%s] connect error: %v", bname, err)
		}
		return err
	}

	if logger.V(logger.InfoLevel, logger.DefaultLogger) {
		log.Infof("Broker [%s] Connected to %s", bname, config.Broker.Address())
	}

	// use AddCheck func before register
	if err = s.opts.AddCheck(s.opts.Context); err != nil {
		if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
			log.Errorf("Server %s-%s register check error: %s", config.Name, config.Id, err)
		}
	} else {
		// announce self to the world
		if err = s.Add(); err != nil {
			if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
				log.Errorf("Server %s-%s register error: %s", config.Name, config.Id, err)
			}
		}
	}

	exit := make(chan bool)

	go func() {
		for {
			// listen for connections
			err := ts.Accept(s.ServeConn)

			// TODO: listen for messages
			// msg := event.Exchange(service).Consume()

			select {
			// check if we're supposed to exit
			case <-exit:
				return
			// check the error and backoff
			default:
				if err != nil {
					if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
						log.Errorf("Accept error: %v", err)
					}
					time.Sleep(time.Second)
					continue
				}
			}

			// no error just exit
			return
		}
	}()

	go func() {
		t := new(time.Ticker)

		// only process if it exists
		if s.opts.AddInterval > time.Duration(0) {
			// new ticker
			t = time.NewTicker(s.opts.AddInterval)
		}

		// return error chan
		var ch chan error

	Loop:
		for {
			select {
			// register self on interval
			case <-t.C:
				s.RLock()
				registered := s.registered
				s.RUnlock()
				rerr := s.opts.AddCheck(s.opts.Context)
				if rerr != nil && registered {
					if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
						log.Errorf("Server %s-%s register check error: %s, deregister it", config.Name, config.Id, rerr)
					}
					// deregister self in case of error
					if err := s.Remove(); err != nil {
						if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
							log.Errorf("Server %s-%s deregister error: %s", config.Name, config.Id, err)
						}
					}
				} else if rerr != nil && !registered {
					if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
						log.Errorf("Server %s-%s register check error: %s", config.Name, config.Id, rerr)
					}
					continue
				}
				if err := s.Add(); err != nil {
					if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
						log.Errorf("Server %s-%s register error: %s", config.Name, config.Id, err)
					}
				}
			// wait for exit
			case ch = <-s.exit:
				t.Stop()
				close(exit)
				break Loop
			}
		}

		s.RLock()
		registered := s.registered
		s.RUnlock()
		if registered {
			// deregister self
			if err := s.Remove(); err != nil {
				if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
					log.Errorf("Server %s-%s deregister error: %s", config.Name, config.Id, err)
				}
			}
		}

		s.Lock()
		swg := s.wg
		s.Unlock()

		// wait for requests to finish
		if swg != nil {
			swg.Wait()
		}

		// close network listener
		ch <- ts.Close()

		if logger.V(logger.InfoLevel, logger.DefaultLogger) {
			log.Infof("Broker [%s] Disconnected from %s", bname, config.Broker.Address())
		}
		// disconnect the event
		if err := config.Broker.Disconnect(); err != nil {
			if logger.V(logger.ErrorLevel, logger.DefaultLogger) {
				log.Errorf("Broker [%s] Disconnect error: %v", bname, err)
			}
		}

		// swap back address
		s.Lock()
		s.opts.Address = addr
		s.Unlock()
	}()

	// mark the server as started
	s.Lock()
	s.started = true
	s.Unlock()

	return nil
}

func (s *rpcServer) Stop() error {
	s.RLock()
	if !s.started {
		s.RUnlock()
		return nil
	}
	s.RUnlock()

	ch := make(chan error)
	s.exit <- ch

	err := <-ch
	s.Lock()
	s.started = false
	s.Unlock()

	return err
}

func (s *rpcServer) String() string {
	return "rpc"
}
