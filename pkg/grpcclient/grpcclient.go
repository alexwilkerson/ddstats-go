package grpcclient

import (
	"context"
	"fmt"
	"time"

	pb "github.com/alexwilkerson/ddstats-server/gamesubmission"
	"google.golang.org/grpc"
)

type Client struct {
	gameRecorderClient pb.GameRecorderClient
	conn               *grpc.ClientConn
}

func New(addr string) (*Client, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("New: could not get connection to grpc server: %w", err)
	}
	c := pb.NewGameRecorderClient(conn)

	return &Client{c, conn}, nil
}

func (c *Client) Close() {
	c.conn.Close()
}

func (c *Client) SubmitGame(game *pb.SubmitGameRequest) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	r, err := c.gameRecorderClient.SubmitGame(ctx, game)
	if err != nil {
		return 0, fmt.Errorf("SubmitGame: failed to submit game over grpc: %w", err)
	}
	return int(r.GetGameID()), nil
}

func (c *Client) ClientConnect(version string) (*pb.ClientStartReply, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	r, err := c.gameRecorderClient.ClientStart(ctx, &pb.ClientStartRequest{Version: version})
	if err != nil {
		return nil, fmt.Errorf("SubmitGame: failed to submit game over grpc: %w", err)
	}
	return r, nil
}
