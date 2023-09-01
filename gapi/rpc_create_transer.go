package gapi

import (
	"context"
	"errors"

	db "github.com/labasubagia/simplebank/db/sqlc"
	"github.com/labasubagia/simplebank/pb"
	"github.com/labasubagia/simplebank/util"
	"github.com/labasubagia/simplebank/val"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var ErrCurrencyInvalid = errors.New("currency invalid")

func (server *Server) CreateTransfer(ctx context.Context, req *pb.CreateTransferRequest) (*pb.CreateTransferResponse, error) {

	authPayload, err := server.authorizeUser(ctx)
	if err != nil {
		return nil, unauthenticatedError(err)
	}

	if violations := validateTransferRequest(req); violations != nil {
		return nil, invalidArgumentError(violations)
	}

	fromAccount, err := server.validAccount(ctx, req.GetFromAccountId(), req.GetCurrency())
	if err != nil {
		if errors.Is(err, ErrCurrencyInvalid) {
			return nil, status.Errorf(codes.InvalidArgument, "mismatch currency %s and %s", fromAccount.Currency, req.GetCurrency())
		}
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "from account not found")
		}
	}

	if fromAccount.Owner != authPayload.Username {
		return nil, status.Errorf(codes.PermissionDenied, "this is not your account")
	}

	toAccount, err := server.validAccount(ctx, req.GetToAccountId(), req.GetCurrency())
	if err != nil {
		if errors.Is(err, ErrCurrencyInvalid) {
			return nil, status.Errorf(codes.InvalidArgument, "mismatch currency %s and %s", fromAccount.Currency, req.GetCurrency())
		}
		if errors.Is(err, db.ErrRecordNotFound) {
			return nil, status.Errorf(codes.NotFound, "from account not found")
		}
	}

	arg := db.TransferTxParams{
		FromAccountID: fromAccount.ID,
		ToAccountID:   toAccount.ID,
		Amount:        req.GetAmount(),
	}
	result, err := server.store.TransferTx(ctx, arg)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	res := &pb.CreateTransferResponse{
		Transfer:    convertTransfer(result.Transfer),
		FromAccount: convertAccount(result.FromAccount),
		ToAccount:   convertAccount(result.ToAccount),
		FromEntry:   convertEntry(result.FromEntry),
		ToEntry:     convertEntry(result.ToEntry),
	}

	return res, nil
}

func (server *Server) validAccount(ctx context.Context, accountID int64, currency string) (db.Account, error) {
	account, err := server.store.GetAccount(ctx, accountID)
	if err != nil {
		return account, db.ErrRecordNotFound
	}

	if account.Currency != currency {
		return account, ErrCurrencyInvalid
	}

	return account, nil
}

func validateTransferRequest(req *pb.CreateTransferRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := val.ValidateID(req.GetFromAccountId()); err != nil {
		violations = append(violations, fieldValidation("from_account_id", err))
	}
	if err := val.ValidateID(req.GetToAccountId()); err != nil {
		violations = append(violations, fieldValidation("to_account_id", err))
	}
	if req.GetAmount() < 1 {
		violations = append(violations, fieldValidation("amount", errors.New("amount minimal 1")))
	}
	if ok := util.IsSupportedCurrency(req.GetCurrency()); !ok {
		violations = append(violations, fieldValidation("currency", errors.New("not supported currency")))
	}
	return violations
}
