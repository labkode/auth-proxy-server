package main

import (
	"context"
	"flag"
	"log"
	"time"

	"github.com/vkuznet/auth-proxy-server/grpc/cms"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

var defaultRequestTimeout = time.Second * 10

// Service defines the interface exposed by this package.
type GRPCService interface {
	GetData(request *cms.Request) (*cms.Response, error)
}

type grpcService struct {
	grpcClient cms.DataServiceClient
}

// NewGRPCService creates a new gRPC user service connection using the specified connection string.
func NewGRPCService(connString, cert string) (GRPCService, error) {
	var err error
	var conn *grpc.ClientConn
	if cert == "" {
		// insecure gRPC connection
		conn, err = grpc.Dial(connString, grpc.WithInsecure())
	} else {

		// secure (TLS) gRPC connection
		// for details see
		// https://github.com/grpc/grpc-go/blob/master/Documentation/grpc-auth-support.md
		// https://pkg.go.dev/google.golang.org/grpc/credentials#NewClientTLSFromCert
		conn, err = grpc.Dial(connString, grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(nil, "")))
	}

	if err != nil {
		return nil, err
	}
	return &grpcService{grpcClient: cms.NewDataServiceClient(conn)}, nil
}

// GetData implements grpcServer GetData API
func (s *grpcService) GetData(req *cms.Request) (*cms.Response, error) {
	ctx, cancelFunc := context.WithTimeout(context.Background(), defaultRequestTimeout)
	defer cancelFunc()
	// pass incoming gRPC request to backend gRPC server
	resp, err := s.grpcClient.GetData(ctx, req)
	return resp, err
}

func main() {
	var address string
	flag.StringVar(&address, "address", "", "gRPC address")
	var token string
	flag.StringVar(&token, "token", "", "gRPC authorization token")
	var cert string
	flag.StringVar(&cert, "cert", "", "client certificate")
	flag.Parse()

	backendGRPC, err := NewGRPCService(address, cert)
	if err != nil {
		log.Fatal(err)
	}
	data := &cms.Data{Id: 1, Token: token}
	req := &cms.Request{Data: data}
	resp, err := backendGRPC.GetData(req)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("gRPC response", resp.String())
}