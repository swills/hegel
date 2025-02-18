package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log"

	"github.com/tinkerbell/hegel/grpc/protos/hegel"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

func main() {
	config := &tls.Config{
		// TODO: Investigate whether it is safe to remove this dangerous default
		InsecureSkipVerify: true, //nolint:gosec // G402: TLS InsecureSkipVerify set true
	}

	// TODO: Remove this hard-coded hostname
	conn, err := grpc.Dial("metadata.packet.net:50060", grpc.WithTransportCredentials(credentials.NewTLS(config)))
	if err != nil {
		log.Panic(err)
	}
	client := hegel.NewHegelClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	err = subscribe(ctx, client, func(str string) {
		fmt.Println(str)
	})
	if err != nil {
		log.Panic(err)
	}
}

func subscribe(ctx context.Context, client hegel.HegelClient, onJSON func(string)) error {
	res, err := client.Get(ctx, &hegel.GetRequest{})
	if err != nil {
		return err
	}
	str := res.GetJSON()
	onJSON(str)

	watcher, err := client.Subscribe(ctx, &hegel.SubscribeRequest{})
	if err != nil {
		return err
	}

	for {
		hw, err := watcher.Recv()
		if errors.Is(err, io.EOF) {
			return errors.New("hegel closed the subscription")
		}

		onJSON(hw.GetJSON())
	}
}
