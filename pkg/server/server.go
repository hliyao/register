package grpcServer

import (
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"path"
	"strings"
	"syscall"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc"
	channelz "google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/reflection"
	"log"

	_ "net/http/pprof"
	"register/pkg/config"
	"register/pkg/monitor/web"
	grpclb "register/pkg/register"
)

type Adapter string

const (
	AdapterJSON = "json"
	AdapterINI  = "ini"
	AdapterYAML = "yaml"
	AdapterXML  = "xml"
)

type GRPCServerIntializer func(ctx context.Context, s *grpc.Server, sc *config.ServerConfig) error
type GRPCServerCloser func(ctx context.Context, s *grpc.Server) error

type Options struct {
	defaultConfigPath    string
	defaultConfigAdaptor Adapter

	gRPCServerOptions     []grpc.ServerOption
	gRPCServerInitializer []GRPCServerIntializer
	gRPCServerCloser      GRPCServerCloser
}

var defaultServerOption = Options{
	gRPCServerInitializer: []GRPCServerIntializer{},
	gRPCServerOptions:     []grpc.ServerOption{},
	gRPCServerCloser:      func(ctx context.Context, s *grpc.Server) error { return nil },
}

// A ServerOption sets options such as credentials, codec and keepalive parameters, etc.
type ServerOption func(*Options)

func WithDefaultConfig(path string, adapter Adapter) ServerOption {
	return func(o *Options) {
		o.defaultConfigPath = path
		o.defaultConfigAdaptor = adapter
	}
}

// WithGRPCServerInitializer return a ServerOption which pends a
// server initializer to initialize a gRPC server.
func WithGRPCServerInitializer(initializer GRPCServerIntializer) ServerOption {
	return func(o *Options) {
		o.gRPCServerInitializer = append(o.gRPCServerInitializer, initializer)
	}
}

func WithGRPCServerCloser(closer GRPCServerCloser) ServerOption {
	return func(o *Options) {
		o.gRPCServerCloser = closer
	}
}

func WithGRPCServerOptions(opts ...grpc.ServerOption) ServerOption {
	return func(o *Options) {
		o.gRPCServerOptions = append(o.gRPCServerOptions, opts...)
	}
}

type ServerFlags struct {
	*flag.FlagSet

	ConfigPath  string
	configType  string
	listen      string
	logLevel    string
	debugListen string
	addr        string
	name        string
	domain      string

	RecoverPanic, LogRequests, LogResponses, withNameRegistration bool
}

func ParseServerFlags(args []string) *ServerFlags {
	fs := flag.NewFlagSet(args[0], flag.ExitOnError)

	var (
		configPathFlag           = fs.String("config", "", "The path of server config")
		configTypeFlag           = fs.String("config_type", "", "The type of server config file")
		listenFlag               = fs.String("listen", "", "The address which gRPC logic server listens on.")
		logLevelFlag             = fs.String("log_level", "", "The level of log output.")
		debugListenFlag          = fs.String("debug", "", "The address which debug/pprof server listens on.")
		addrFlag                 = fs.String("addr", "", "The addr which clients should connect, the listen address will be used by default.")
		nameFlag                 = fs.String("name", "", "Name of the server.")
		domainFlag               = fs.String("domain", "", "Domain of the server.")
		versionFlag              = fs.Bool("v", false, "Display this screen.")
		dontRecoverPanicFlag     = fs.Bool("dont-recover-panic", false, "Don't recover from panic.")
		withNameRegistrationFlag = fs.Bool("with-name-registration", true, "Register to name server")
		logRequestsFlag          = fs.Bool("log-requests", false, "log requests or not")
		logResponsesFlag         = fs.Bool("log-responses", false, "log responses or not")
	)

	if err := fs.Parse(args[1:]); err != nil {
		return nil
	}

	if *versionFlag {
		fmt.Fprintf(os.Stdout, "Usage of %s:\n", os.Args[0])
		fs.SetOutput(os.Stdout)
		fs.PrintDefaults()
		os.Exit(0)
	}

	return &ServerFlags{
		ConfigPath:           *configPathFlag,
		configType:           *configTypeFlag,
		listen:               *listenFlag,
		logLevel:             *logLevelFlag,
		debugListen:          *debugListenFlag,
		addr:                 *addrFlag,
		name:                 *nameFlag,
		domain:               *domainFlag,
		withNameRegistration: *withNameRegistrationFlag,
		RecoverPanic:         !(*dontRecoverPanicFlag),
		LogRequests:          *logRequestsFlag,
		LogResponses:         *logResponsesFlag,

		FlagSet: fs,
	}
}

type Server struct {
	svr *grpc.Server

	ctx    context.Context
	cancel context.CancelFunc

	opt             *Options
	sc              *config.ServerConfig
	sf              *ServerFlags
	initializeError error
}

// NewServer instantiate a gRPC server with opts
func NewServer(flags *ServerFlags, opts ...ServerOption) *Server {
	opt := defaultServerOption
	// apply ServerOptions
	for _, o := range opts {
		o(&opt)
	}

	var sc = config.EmptyServerConfig()
	var err error

	if len(flags.ConfigPath) == 0 {
		flags.ConfigPath = opt.defaultConfigPath
	}
	if len(flags.configType) == 0 {
		flags.configType = string(opt.defaultConfigAdaptor)
	}
	if len(flags.ConfigPath) > 0 {
		configType := flags.configType
		if len(configType) == 0 {
			ext := path.Ext(flags.ConfigPath)
			if len(ext) > 0 {
				configType = ext[1:]
			}
		}

		err = sc.InitWithPath(configType, flags.ConfigPath)
		if err != nil {
			log.Printf("Warning: Failed to load config file: %s [err: %+v]", flags.ConfigPath, err)
		}
	}

	// The priority of arguments in command line is higher than those in config file
	setValue := func(target *string, val *string) {
		if val != nil && len(*val) > 0 {
			*target = *val
		}
	}
	setValue(&sc.Listen, &flags.listen)
	setValue(&sc.LogLevel, &flags.logLevel)
	setValue(&sc.DebugListen, &flags.debugListen)
	setValue(&sc.Addr, &flags.addr)
	setValue(&sc.Name, &flags.name)
	setValue(&sc.Domain, &flags.domain)
	if sc.Addr == "" {
		sc.Addr = sc.Listen
	}
	if len(flags.listen) > 0 {
		sc.HadManualSetListen = true
	}

	return &Server{
		opt: &opt,
		sc:  sc,
		sf:  flags,
	}
}

// Serve on context.Background()
func (s *Server) Serve() error {
	return s.ServeContext(context.Background())
}

// ServeContext runs server on a specific context
func (s *Server) ServeContext(ctx context.Context) (err error) {
	/* init logger */
	logModuleName := s.sc.Name
	if len(logModuleName) == 0 {
		_, logModuleName = path.Split(os.Args[0])
	}

	if !s.sc.HadManualSetListen {
		ipAddr, err := s.getIp()
		if err != nil {
			return err
		}

		if ipAddr != "" {
			s.sc.Listen = ipAddr + ":0"
		}
		log.Printf("Server auto set listen address(%v), prepare to find a idle port...", s.sc.Listen)
	} else {
		log.Printf("Server manual set listen address(%v)", s.sc.Listen)
	}

	/* init gRPC server */
	s.ctx, s.cancel = context.WithCancel(ctx)
	s.svr, err = s.createGRPCServer(s.ctx)
	if err != nil {
		log.Fatalf("Failed to create gRPC Server, err %v", err)
		return
	}
	reflection.Register(s.svr)
	channelz.RegisterChannelzServiceToServer(s.svr)

	var l net.Listener
	l, err = net.Listen("tcp", s.sc.Listen)
	if err != nil {
		log.Fatalf("Failed to listen %s, err %v", s.sc.Listen, err)
		return
	}
	s.sc.Addr = l.Addr().String()

	debugSvr := s.runDebugServer()

	shutdown := make(chan interface{})
	go s.signalHandler(shutdown)
	go s.registerToNameServer(s.ctx)

	log.Printf("Running [%s] on %s", s.sc.Name, l.Addr())
	s.svr.Serve(l)
	<-shutdown // block untils GracefulStop
	log.Printf("Stopped [%s].", s.sc.Name)

	s.cancel()

	debugSvr.Close()

	return
}

func (s *Server) getIp() (string, error) {
	addr, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	var ipAddress []string
	for _, address := range addr {
		// 检查ip地址判断是否回环地址
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ipAddress = append(ipAddress, ipNet.IP.String())
			}
		}
	}
	log.Printf("Found ip address(%v) without lo in machine...", ipAddress)
	if len(ipAddress) == 1 {
		//只有一个ip，直接返回
		return ipAddress[0], nil
	} else if len(ipAddress) == 0 {
		return "", errors.New("had not found suitable ip address")
	} else {
		log.Printf("Found multi ip address(%v), prepare to find etcd endPoints...", ipAddress)
		endPoints, err := grpclb.GetEtcdv3Endpoints()
		if err != nil {
			return "", err
		}
		log.Printf("Found multi ip address(%v), prepare to dial etcd endPoints(%v)...", ipAddress, endPoints)
		conn, err := net.DialTimeout("tcp", endPoints[0], time.Second*5)
		if err != nil {
			return "", err
		}
		defer conn.Close()
		return strings.Split(conn.LocalAddr().String(), ":")[0], nil
	}
}

func (s *Server) createGRPCServer(ctx context.Context) (*grpc.Server, error) {
	svr := grpc.NewServer(s.opt.gRPCServerOptions...)

	for _, initializer := range s.opt.gRPCServerInitializer {
		if err := initializer(ctx, svr, s.sc); err != nil {
			return nil, err
		}
	}

	return svr, nil
}

func (s *Server) registerToNameServer(ctx context.Context) {
	if !s.sf.withNameRegistration {
		log.Printf("Skip name registration")
		return
	}

	// Register server to etcd.
	r := grpclb.DefaultEtcdv3Resolver()
	fullName := s.sc.FullName()
	if len(fullName) > 0 {
		log.Printf("Registering `%s` -> `%s` to naming server.", fullName, s.sc.Addr)
		r.Register(ctx, fullName, s.sc.Addr)
	}
	grpclb.GetEtcdProcess().SetResolve(r)
}

func (s *Server) revokeFromNameServer(ctx context.Context) error {
	if !s.sf.withNameRegistration {
		log.Printf("Skip name revocation")
		return nil
	}

	r := grpclb.DefaultEtcdv3Resolver()
	fullName := s.sc.FullName()
	if len(fullName) > 0 {
		log.Printf("Revoking `%s` -> `%s` from naming server.", fullName, s.sc.Addr)
		ctx, cancel := context.WithTimeout(ctx, time.Second*3)
		defer cancel()
		err := r.Unregister(ctx, fullName, s.sc.Addr)
		if err == nil {
			log.Printf("Revoked `%s` -> `%s` from naming server.", fullName, s.sc.Addr)
		} else {
			log.Printf("Revoke `%s` -> `%s` from naming	 failed, err: %v", fullName, s.sc.Addr, err)
		}

		return err
	}
	return nil
}

func (s *Server) signalHandler(shutdown chan interface{}) {
	wait := make(chan os.Signal, 1)
	signal.Notify(wait, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGQUIT)

	select {
	case sig := <-wait:
		log.Printf("Receive signal %v, stop server [%s].", sig, s.sc.Name)
		// remove from name server to avoid future requests
		s.revokeFromNameServer(s.ctx)

		// stop the server after finishing all pending RPCs
		if s.svr != nil {
			log.Printf("GracefulStop server [%s].", s.sc.Name)
			// prevent pending requests
			s.svr.GracefulStop()

			if s.opt.gRPCServerCloser != nil {
				s.opt.gRPCServerCloser(s.ctx, s.svr)
			}

			close(shutdown)
		}
	}
}

func (s *Server) runDebugServer() (svr *http.Server) {
	http.HandleFunc(web.MONITOR_ROUTER, web.MonitorHandle)
	http.HandleFunc(web.WARNING_ROUTER, web.WarningHandle)
	svr = &http.Server{Handler: http.DefaultServeMux}
	addr := ":"
	if len(s.sc.DebugListen) > 0 {
		addr = s.sc.DebugListen
	}

	l, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalln("Failed to start debug server on", addr)
		return
	}
	go func() {
		log.Println("Running debug server on", l.Addr())
		log.Println("Debug server stopped:", svr.Serve(l))
	}()
	return
}
