package gapi

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"
	"testing"

	mockdb "github.com/dibrito/simple-bank/db/mocks"
	db "github.com/dibrito/simple-bank/db/sqlc"
	"github.com/dibrito/simple-bank/pb"
	"github.com/dibrito/simple-bank/util"
	"github.com/dibrito/simple-bank/worker"
	mockwk "github.com/dibrito/simple-bank/worker/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type eqCreateUserTxParamsMatcher struct {
	arg      db.CreateUserTxParams
	password string
	user     db.User
}

func (expected eqCreateUserTxParamsMatcher) Matches(x interface{}) bool {
	actual, ok := x.(db.CreateUserTxParams)
	if !ok {
		return false
	}
	err := util.CheckPassword(expected.password, actual.HashedPassword)
	if err != nil {
		return false
	}

	expected.arg.HashedPassword = actual.HashedPassword

	// since we cannot compare two funcs in Go,
	// we should only compare the created user params instead of the whole arg
	if !reflect.DeepEqual(expected.arg.CreateUserParams, actual.CreateUserParams) {
		return false
	}
	err = actual.AfterCreate(expected.user)

	return err == nil
}

func (e eqCreateUserTxParamsMatcher) String() string {
	return fmt.Sprintf("matches arg %v and password %v", e.arg, e.password)
}

func EqCreateUserTxParams(arg db.CreateUserTxParams, password string, user db.User) gomock.Matcher {
	return eqCreateUserTxParamsMatcher{arg: arg, password: password, user: user}
}

func randomUser(t *testing.T) (user db.User, password string) {
	password = util.RandomString(6)
	hashedPassword, err := util.HashPassword(password)
	require.NoError(t, err)

	user = db.User{
		Username:       util.RandomOwner(),
		HashedPassword: hashedPassword,
		FullName:       util.RandomOwner(),
		Email:          util.RandomEmail(),
	}
	return
}

func TestCreateUserApi(t *testing.T) {
	user, password := randomUser(t)
	tcs := []struct {
		name          string
		req           *pb.CreateUserRequest
		accountID     int64
		buildStubs    func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor)
		checkResponse func(t *testing.T, res *pb.CreateUserResponse, err error)
	}{
		{
			name: "OK",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				arg := db.CreateUserTxParams{
					CreateUserParams: db.CreateUserParams{
						Username: user.Username,
						FullName: user.FullName,
						Email:    user.Email,
					},
				}
				store.EXPECT().
					CreateUserTx(gomock.Any(), EqCreateUserTxParams(arg, password, user)).
					Times(1).
					Return(db.CreateUserTxResult{User: user}, nil)

				taskPayload := &worker.PayloadSendVerifyEmail{
					Username: user.Username,
				}
				taskDistributor.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), taskPayload, gomock.Any()).
					Times(1).
					Return(nil)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.NoError(t, err)
				require.NotNil(t, res)
				createdUser := res.GetUser()
				require.Equal(t, user.Username, createdUser.Username)
				require.Equal(t, user.FullName, createdUser.FullName)
				require.Equal(t, user.Email, createdUser.Email)
			},
		},
		{
			name: "internal error",
			req: &pb.CreateUserRequest{
				Username: user.Username,
				Password: password,
				FullName: user.FullName,
				Email:    user.Email,
			},
			buildStubs: func(store *mockdb.MockStore, taskDistributor *mockwk.MockTaskDistributor) {
				store.EXPECT().
					CreateUserTx(gomock.Any(), gomock.Any()).
					Times(1).
					Return(db.CreateUserTxResult{}, sql.ErrConnDone)

				taskDistributor.EXPECT().
					DistributeTaskSendVerifyEmail(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			checkResponse: func(t *testing.T, res *pb.CreateUserResponse, err error) {
				require.Error(t, err)
				st, ok := status.FromError(err)
				require.True(t, ok)
				require.Equal(t, codes.Internal, st.Code())
				require.Nil(t, res)
			},
		},
	}

	for _, tc := range tcs {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// In fact, the problem comes from&nbsp; the way we use the same controller
			// for both the mock store and&nbsp; the mock task distributor.
			// There’s a locking mechanism in the controller
			// every time it checks for a matching function call.
			// So when the CreateUserTx function is&nbsp; being checked for matching arguments,
			// the mock controller will be locked,
			// That’s why when we call the&nbsp; AfterCreate() callback function,
			// It can no longer acquire the lock to record&nbsp; the call to the mock task distributor.
			// To fix this, we can simply&nbsp; use 2 different controllers,
			ctrlStore := gomock.NewController(t)
			defer ctrlStore.Finish()
			store := mockdb.NewMockStore(ctrlStore)

			ctrlTaskDistributor := gomock.NewController(t)
			defer ctrlTaskDistributor.Finish()
			taskDistributor := mockwk.NewMockTaskDistributor(ctrlTaskDistributor)

			tc.buildStubs(store, taskDistributor)
			server := newTestServer(t, store, taskDistributor)

			res, err := server.CreateUser(context.Background(), tc.req)
			tc.checkResponse(t, res, err)
		})
	}
}
