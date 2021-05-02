package ctxms

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"

	api "github.com/rotationalio/ctxms/proto"
)

func init() {
	// Set the random seed
	rand.Seed(time.Now().UnixNano())

	// Initialize zerolog with GCP logging requirements
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: "15:04:05"})
}

type Server struct {
	api.UnimplementedHopperServer
	srv  *grpc.Server
	errc chan error
	conf *Config
	cc   *grpc.ClientConn
	fwd  api.HopperClient
}

func New(conf *Config) (_ *Server, err error) {
	if err = conf.Validate(); err != nil {
		return nil, err
	}

	s := &Server{
		srv:  grpc.NewServer(),
		errc: make(chan error, 1),
		conf: conf,
	}
	api.RegisterHopperServer(s.srv, s)
	return s, nil
}

func (s *Server) Trace(ctx context.Context, in *api.Packet) (out *api.Packet, err error) {
	// Check terminal condition (we're the source and are getting trace called again)
	if (len(in.Route) > 0 && in.Route[0] == s.conf.Name) || s.fwd == nil {
		log.Info().Int("length", len(in.Route)).Msg("end of chain")
		return in, nil
	}

	// Do hard work or wait for context to be canceled
	select {
	case <-ctx.Done():
		err = ctx.Err()
		log.Warn().Err(err).Msg("context canceled")
		return nil, err
	case <-s.hardWork():
		log.Debug().Msg("sending trace to next hop")
	}

	// Add ourselves to the route and send to the next hop on the route
	in.Route = append(in.Route, s.conf.Name)

	// Send out the request to the next hop
	out, err = s.fwd.Trace(ctx, in)
	if err != nil {
		log.Error().Err(err).Msg("trace failed")
	} else {
		log.Info().Str("source", out.Route[0]).Str("id", out.Id).Msg("trace complete")
	}
	return out, err
}

func (s *Server) hardWork() <-chan bool {
	// Create the done channel
	done := make(chan bool, 1)

	// Kick off the go routine to do the hard work
	go func(done chan<- bool) {
		// Delay for a random amount of time to simulate hard work
		delay := time.Duration(rand.Int63n(int64(s.conf.Delay)))
		log.Debug().Str("eta", delay.String()).Msg("starting hard work")
		time.Sleep(delay)
		log.Debug().Msg("hard work complete")
		done <- true
	}(done)

	// Return the done channel
	return done
}

func (s *Server) Serve() (err error) {
	// Listen for CTRL+C and call shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	go func() {
		<-quit
		s.errc <- s.Shutdown()
	}()

	// Listen on the address (ipaddr:port)
	var sock net.Listener
	if sock, err = net.Listen("tcp", s.conf.Addr()); err != nil {
		return fmt.Errorf("could not listen on %q: %s", s.conf.Addr(), err)
	}
	defer sock.Close()

	// Handle gRPC methods in a go routine
	go func() {
		log.Info().Str("listen", s.conf.Addr()).Msg("server started")
		if err := s.srv.Serve(sock); err != nil {
			s.errc <- err
		}
	}()

	// Dial the next hop in the sequence
	go func() {
		if err := s.Dial(); err != nil {
			s.errc <- err
			return
		}
		log.Info().Str("addr", s.conf.NextHop()).Msg("dialed next hop")
	}()

	// Wait for server error or shutdown
	if err = <-s.errc; err != nil {
		return err
	}
	return nil
}

func (s *Server) Dial() (err error) {
	// Add a bit of delay for dialing to make sure the server is up
	time.Sleep(time.Duration(rand.Int63n(int64(2 * time.Second))))

	if s.cc, err = grpc.Dial(s.conf.NextHop(), grpc.WithInsecure()); err != nil {
		return err
	}
	s.fwd = api.NewHopperClient(s.cc)
	return nil
}

func (s *Server) Shutdown() (err error) {
	// Shutdown the gRPC server
	s.srv.GracefulStop()

	// Shutdown hopper client
	if err = s.cc.Close(); err != nil {
		log.Error().Err(err).Msg("could not close hopper client")
	}

	return nil
}
