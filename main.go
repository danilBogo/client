package main

import (
	"context"
	"fmt"
	chatv1 "github.com/danilBogo/protos/gen/go/chat"
	"github.com/danilBogo/server/pkg/config"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"io"
	"log"
	"math/rand"
	"net"
	"strconv"
)

const (
	grpcHost = "localhost"
	username = "John Doe"
)

func main() {
	cfg := config.MustLoad()

	cc, err := grpc.DialContext(context.Background(),
		grpcAddress(cfg),
		grpc.WithTransportCredentials(insecure.NewCredentials()))

	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer cc.Close()

	client := chatv1.NewChatClient(cc)

	chatId := join(cfg, client)

	send(cfg, client, chatId)

	getMessages(client, chatId)

	leave(cfg, client, chatId)
}

func join(cfg *config.Config, client chatv1.ChatClient) string {
	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)
	joinResp, err := client.Join(ctx, &chatv1.JoinRequest{Username: username})
	cancelCtx()
	if err != nil {
		log.Fatalf("Error joining to chat: %v", err)
	}

	return joinResp.ChatId
}

func send(cfg *config.Config, client chatv1.ChatClient, chatId string) {
	count := rand.Int()%10 + 2
	for i := 1; i < count; i++ {
		ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)
		_, err := client.Send(ctx, &chatv1.SendRequest{
			Username: username,
			ChatId:   chatId,
			Text:     "Hello! My name is John. This is my " + strconv.Itoa(i) + " message",
		})
		cancelCtx()
		if err != nil {
			log.Fatalf("Error sending message: %v", err)
		}
	}
}

func getMessages(client chatv1.ChatClient, chatId string) {
	getMessagesResp, err := client.GetMessages(context.Background(), &chatv1.GetMessagesRequest{ChatId: chatId})
	if err != nil {
		log.Fatalf("Error receiving messages: %v", err)
	}

	fmt.Println("Messages:")
	for {
		message, err := getMessagesResp.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Fatalf("Error receiving message: %v", err)
		}

		fmt.Printf("'%v' '%v'\n", message.Username, message.Text)
	}
}

func leave(cfg *config.Config, client chatv1.ChatClient, chatId string) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), cfg.GRPC.Timeout)
	_, err := client.Leave(ctx, &chatv1.LeaveRequest{Username: username, ChatId: chatId})
	cancelCtx()
	if err != nil {
		log.Fatalf("Error leaving from chat: %v", err)
	}
}

func grpcAddress(cfg *config.Config) string {
	return net.JoinHostPort(grpcHost, strconv.Itoa(cfg.GRPC.Port))
}
