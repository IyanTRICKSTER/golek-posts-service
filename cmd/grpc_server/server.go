package grpc_server

import (
	"context"
	"fmt"
	"golek_posts_service/pkg/contracts"
	ps "golek_posts_service/pkg/models/proto_schema"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
)

type GRPCPostServer struct {
	ps.UnimplementedPostServiceServer
	postService contracts.PostServiceContract
}

func (s *GRPCPostServer) Fetch(ctx context.Context, IDs *ps.PostIDs) (*ps.Posts, error) {

	posts := make([]*ps.Post, 0)

	for _, id := range IDs.Id {
		post, err := s.postService.FindById(ctx, id)
		if err != nil {
			log.Printf("RPC SERVER, Fetching Document %v Error: %v", IDs.Id, err.Error())
		} else {
			posts = append(posts, &ps.Post{
				Id:       post.ID.Hex(),
				Name:     post.Title,
				ImageUrl: post.ImageURL,
			})
		}

	}
	return &ps.Posts{List: posts}, nil
}

func (s *GRPCPostServer) Run() {
	srv := grpc.NewServer()
	ps.RegisterPostServiceServer(srv, s)

	l, err := net.Listen("tcp", fmt.Sprintf("%v:%v", os.Getenv("RPC_HOST"), os.Getenv("RPC_PORT")))
	if err != nil {
		log.Fatalf("could not listen to %s: %v", os.Getenv("RPC_PORT"), err)
	}

	log.Println("RPC listening and serving TCP on", os.Getenv("RPC_PORT"))
	log.Fatal(srv.Serve(l))
}

func New(postService *contracts.PostServiceContract) *GRPCPostServer {

	return &GRPCPostServer{postService: *postService}
}
