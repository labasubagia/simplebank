package gapi

import (
	"context"
	"errors"

	db "github.com/labasubagia/simplebank/db/sqlc"
	"github.com/labasubagia/simplebank/pb"
	"github.com/labasubagia/simplebank/util"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (server *Server) CreateAccount(ctx context.Context, req *pb.CreateAccountRequest) (*pb.CreateAccountResponse, error) {
	authPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if violations := validateCreateAccountRequest(req); violations != nil {
		return nil, invalidArgumentError(violations)
	}

	arg := db.CreateAccountParams{
		Owner:    authPayload.Username,
		Currency: req.Currency,
		Balance:  0,
	}
	account, err := server.store.CreateAccount(ctx, arg)
	if err != nil {
		errCode := db.ErrorCode(err)
		if errCode == db.ForeignKeyViolation || errCode == db.UniqueViolation {
			return nil, status.Error(codes.AlreadyExists, err.Error())
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &pb.CreateAccountResponse{Account: convertAccount(account)}
	return res, nil
}

func validateCreateAccountRequest(req *pb.CreateAccountRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if ok := util.IsSupportedCurrency(req.GetCurrency()); !ok {
		violations = append(violations, fieldValidation("currency", errors.New("not supported currency")))
	}
	return violations
}
