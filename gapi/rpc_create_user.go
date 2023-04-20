package gapi

import (
	"context"

	db "github.com/dibrito/simple-bank/db/sqlc"
	"github.com/dibrito/simple-bank/pb"
	"github.com/dibrito/simple-bank/util"
	"github.com/lib/pq"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.CreateUserResponse, error) {
	hashedPassword, err := util.HashPassword(req.GetPassword())
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to hash password:%s", err)
	}
	u, err := server.store.CreateUser(ctx, db.CreateUserParams{
		Username:       req.GetUsername(),
		HashedPassword: hashedPassword,
		Email:          req.GetEmail(),
		FullName:       req.GetFullName(),
	})
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			switch pqErr.Code.Name() {
			// user table does not have FKs
			case "unique_violation":
				return nil, status.Errorf(codes.AlreadyExists, "username already exists:%s", err)
			}
		}
		return nil, status.Errorf(codes.Internal, "fail to create user:%s", err)
	}

	resp := &pb.CreateUserResponse{
		User: convertUser(u),
	}
	return resp, nil
}

func convertUser(user db.User) *pb.User {
	return &pb.User{
		Username:          user.Username,
		FullName:          user.FullName,
		Email:             user.Email,
		PasswordChangedAt: timestamppb.New(user.PasswordChangedAt),
		CreatedAt:         timestamppb.New(user.CreatedAt),
	}
}
