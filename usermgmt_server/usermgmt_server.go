package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"

	pb "github.com/MatheusBBarni/usermgmt-grpc/usermgmt"
	"github.com/jackc/pgx/v4"
	"google.golang.org/grpc"
)

const (
	port      = ":50051"
	file_name = "users.json"
)

type UserManagementServer struct {
	pb.UnimplementedUserManagementServer
	conn *pgx.Conn
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

	create_sql := `
		CREATE TABLE IF NOTE EXISTS users(
			id SERIAL PRIMARY KEY,
			name TEXT,
			age int
		);
	`

	_, err := s.conn.Exec(context.Background(), create_sql)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Table creation failed: %v", err)

		os.Exit(1)
	}

	created_user := &pb.User{
		Name: in.GetName(),
		Age:  in.GetAge(),
	}

	tx, err := s.conn.Begin(context.Background())
	if err != nil {
		log.Fatalf("Connection failed: %v", err)
	}

	defer tx.Conn().Close(ctx)

	_, err = tx.Exec(context.Background(), "insert into users(name, age) values($1, $2)", created_user.Name, created_user.Age)
	if err != nil {
		log.Fatalf("Failed to insert user: %v", err)
	}

	tx.Commit(context.Background())

	return created_user, nil
}

func (s *UserManagementServer) GetUsers(ctx context.Context, in *pb.GetUsersParams) (*pb.UserList, error) {
	rows, err := s.conn.Query(context.Background(), "select * from users")
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var users_list *pb.UserList = &pb.UserList{}

	for rows.Next() {
		user := pb.User{}
		err = rows.Scan(&user.Id, &user.Name, &user.Age)

		if err != nil {
			return nil, err
		}

		users_list.Users = append(users_list.Users, &user)
	}

	return users_list, nil
}

func main() {
	database_url := "postgres://postgres:root@localhost:5432/usermgmt"
	conn, err := pgx.Connect(context.Background(), database_url)
	if err != nil {
		log.Fatalf("Failed to connect to the database: %v", err)
	}

	defer conn.Close(context.Background())

	var user_mgmt_server *UserManagementServer = NewUserManagementServer()
	user_mgmt_server.conn = conn

	if err := user_mgmt_server.Run(); err != nil {
		log.Fatalf("Failed to run: %v", err)
	}
}
