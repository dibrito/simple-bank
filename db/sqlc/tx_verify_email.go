package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/rs/zerolog/log"
)

type VerifyEmailTxParams struct {
	EmailID    int64
	SecretCode string
}

type VerifyEmailTxResult struct {
	User        User
	VerifyEmail VerifyEmail
}

func (store *SQLStore) VerifyEmailTx(ctx context.Context, arg VerifyEmailTxParams) (VerifyEmailTxResult, error) {
	var result VerifyEmailTxResult
	err := store.execTx(ctx, func(q *Queries) error {
		var err error

		verifyEmail, err := q.UpdateVerifyEmail(ctx, UpdateVerifyEmailParams{
			ID:         arg.EmailID,
			SecretCode: arg.SecretCode,
		})
		if err != nil {
			log.Info().Msg(fmt.Sprintf("args:%v", arg))
			log.Err(err).Err(err).Msg("cannot update verify email:%v")
			return err
		}

		result.User, err = q.UpdateUser(ctx, UpdateUserParams{
			Username: verifyEmail.Username,
			IsEmailVerified: sql.NullBool{
				Valid: true,
				Bool:  true,
			},
		})
		if err != nil {
			log.Info().Msg(fmt.Sprintf("args:%v", arg))
			log.Err(err).Err(err).Msg("cannot update user:%v")
		}

		return err
	})

	return result, err
}
