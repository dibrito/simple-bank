package gapi

import (
	"context"
	"database/sql"
	"time"

	db "github.com/dibrito/simple-bank/db/sqlc"
	"github.com/dibrito/simple-bank/pb"
	"github.com/dibrito/simple-bank/util"
	"github.com/dibrito/simple-bank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	_ "google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) UpdateUser(ctx context.Context, req *pb.UpdateUserRequest) (*pb.UpdateUserResponse, error) {
	payload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if payload.Username != req.Username {
		return nil, status.Errorf(codes.PermissionDenied, "cannot update other user's info")
	}

	violations := validateUpdateUserRequest(req)
	if violations != nil {
		return nil, invalidArgumentsError(violations)
	}

	arg := db.UpdateUserParams{
		Username: req.GetUsername(),
		Email: sql.NullString{
			String: req.GetEmail(),
			Valid:  req.Email != nil,
		},
		FullName: sql.NullString{
			String: req.GetFullName(),
			Valid:  req.FullName != nil,
		},
	}
	if req.Password != nil {
		hashedPassword, err := util.HashPassword(req.GetPassword())
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to hash password:%s", err)
		}
		arg.HashedPassword = sql.NullString{
			Valid:  true,
			String: hashedPassword,
		}
		arg.PasswordChangedAt = sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		}
	}

	u, err := server.store.UpdateUser(ctx, arg)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Errorf(codes.NotFound, "user not found")
		}
		return nil, status.Errorf(codes.Internal, "fail to update user:%s", err)
	}

	resp := &pb.UpdateUserResponse{
		User: convertUser(u),
	}
	return resp, nil
}

func validateUpdateUserRequest(req *pb.UpdateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if req.Password != nil {
		if err := val.ValidatePassword(req.GetPassword()); err != nil {
			violations = append(violations, fieldViolation("password", err))
		}
	}

	if req.FullName != nil {
		if err := val.ValidateFullName(req.GetFullName()); err != nil {
			violations = append(violations, fieldViolation("full_name", err))
		}
	}

	if req.Email != nil {
		if err := val.ValidateEmail(req.GetEmail()); err != nil {
			violations = append(violations, fieldViolation("email", err))
		}
	}

	return violations
}
