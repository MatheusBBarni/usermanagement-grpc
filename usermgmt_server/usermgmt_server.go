package main

import (
	"context"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"os"

	pb "github.com/MatheusBBarni/usermgmt-grpc/usermgmt"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	port      = ":50051"
	file_name = "users.json"
)

type UserManagementServer struct {
	pb.UnimplementedUserManagementServer
}

func NewUserManagementServer() *UserManagementServer {
	return &UserManagementServer{}
}

func (server *UserManagementServer) Run() error {
	list, err := net.Listen("tcp", port)

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	s := grpc.NewServer()

	pb.RegisterUserManagementServer(s, server)

	log.Printf("Listening at %v", list.Addr())

	return s.Serve(list)
}

func (s *UserManagementServer) CreateNewUser(ctx context.Context, in *pb.NewUser) (*pb.User, error) {
	log.Printf("Received: %v", in.GetName())
	readBytes, err := ioutil.ReadFile(file_name)
	var users_list *pb.UserList = &pb.UserList{}

	user_id := int32(rand.Intn(10000))

	created_user := &pb.User{
		Name: in.GetName(),
		Age:  in.GetAge(),
		Id:   user_id,
	}

	if err != nil {
		if os.IsNotExist(err) {
			log.Print("File not found, creating...")
			users_list.Users = append(users_list.Users, created_user)
			jsonBytes, err := protojson.Marshal(users_list)

			if err != nil {
				log.Fatalf("JSON Marshaling failed: %v", err)
			}
			if err := ioutil.WriteFile(file_name, jsonBytes, 0664); err != nil {
				log.Fatalf("Failed to write to file: %v", err)
			}

			return created_user, nil
		} else {
			log.Fatalln("Error reading file: ", err)
		}
	}

	if err := protojson.Unmarshal(readBytes, users_list); err != nil {
		log.Fatalf("Failed to parse user list: %v", err)
	}

	users_list.Users = append(users_list.Users, created_user)

	jsonBytes, err := protojson.Marshal(users_list)

	if err != nil {
		log.Fatalf("JSON Marshaling failed: %v", err)
	}
	if err := ioutil.WriteFile(file_name, jsonBytes, 0664); err != nil {
		log.Fatalf("Failed to write to file: %v", err)
	}

	return created_user, nil
}

func (s *UserManagementServer) GetUsers(ctx context.Context, in *pb.GetUsersParams) (*pb.UserList, error) {
	jsonBytes, err := ioutil.ReadFile(file_name)

	if err != nil {
		log.Fatalf("Failed to read from file: %v", err)
	}

	var users_list *pb.UserList = &pb.UserList{}

	if err := protojson.Unmarshal(jsonBytes, users_list); err != nil {
		log.Fatalf("Unmarshling failed: %v", err)
	}

	return users_list, nil
}

func main() {
	var user_mgmt_server *UserManagementServer = NewUserManagementServer()

	if err := user_mgmt_server.Run(); err != nil {
		log.Fatalf("Failed to run: %v", err)
	}
}
