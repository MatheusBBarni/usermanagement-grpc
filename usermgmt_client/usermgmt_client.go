package main

import (
	"context"
	"fmt"
	"log"
	"time"

	pb "github.com/MatheusBBarni/usermgmt-grpc/usermgmt"
	"google.golang.org/grpc"
)

const (
	address = "localhost:50051"
)

func main() {
	conn, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		log.Fatalf("Did not connec: %v", err)
	}

	defer conn.Close()

	c := pb.NewUserManagementClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)

	defer cancel()

	var new_users = make(map[string]int32)
	new_users["Alice"] = 42
	new_users["Bob"] = 30

	for name, age := range new_users {
		r, err := c.CreateNewUser(ctx, &pb.NewUser{
			Name: name,
			Age:  age,
		})

		if err != nil {
			log.Fatalf("Failed to create user: %v", err)
		}

		log.Printf(`User details:
NAME: %s
AGE: %d
ID: %d
		`, r.GetName(), r.GetAge(), r.GetId())
	}

	params := &pb.GetUsersParams{}

	r, err := c.GetUsers(ctx, params)

	if err != nil {
		log.Fatalf("Could not retrieve users: %v", err)
	}

	log.Print("\nUser List:\n")
	fmt.Printf("Users: %v\n", r.GetUsers())
}
