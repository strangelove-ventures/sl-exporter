package cosmos

import (
	"context"
	"fmt"
	"net/url"
	"path"
	"strconv"
)

type AccountBalance struct {
	Account string
	Denom   string
	Amount  float64
}

// AccountBalance returns the balance of an account. The account is a bech32 address with prefix.
// The denom is likely lowercase, e.g. "uatom".
// To query ibc denoms use the prefix "ibc/hash", e.g. "ibc/E7D5E9D0E9BF8B7354929A817DD28D4D017E745F638954764AA88522A7A409EC"
// If the denom is not found, API returns a balance of 0 instead of an error.
func (c RestClient) AccountBalance(ctx context.Context, account, denom string) (AccountBalance, error) {
	var u url.URL
	u.Path = path.Join("/cosmos/bank/v1beta1/balances", account, "by_denom")
	q := u.Query()
	q.Set("denom", denom)
	u.RawQuery = q.Encode()

	var resp struct {
		Balance struct {
			Denom  string `json:"denom"`
			Amount string `json:"amount"`
		}
	}

	err := c.get(ctx, u, &resp)
	if err != nil {
		return AccountBalance{}, err
	}

	amount, err := strconv.ParseFloat(resp.Balance.Amount, 64)
	if err != nil {
		return AccountBalance{}, fmt.Errorf("malformed amount: %w", err)
	}

	return AccountBalance{
		Account: account,
		Denom:   resp.Balance.Denom,
		Amount:  amount,
	}, nil
}
