package trace

import (
	"github.com/brocaar/loraserver/api/ns"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Config struct {
	NetworkServer string
	Debug         log.Level
}

type NSClient struct {
	cl          ns.NetworkServerServiceClient
}

func NewNSClient(conf *Config) *NSClient {
	log.SetLevel(conf.Debug)
	grpcDialOpts := []grpc.DialOption{
		grpc.WithInsecure(),
	}
	conn, err := grpc.Dial(conf.NetworkServer, grpcDialOpts...)
	if err != nil {
		log.Error(err)
	}

	return &NSClient{
		cl:          ns.NewNetworkServerServiceClient(conn),
	}
}
