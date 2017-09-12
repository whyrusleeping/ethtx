package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/btcsuite/btcd/btcec"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/urfave/cli"
)

func main() {
	app := cli.NewApp()

	app.Commands = []cli.Command{
		mktx,
	}

	app.RunAndExitOnError()
}

var mktx = cli.Command{
	Name: "new",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name: "privkey",
		},
		cli.StringFlag{
			Name: "to",
		},
		cli.StringFlag{
			Name: "data",
		},
		cli.StringFlag{
			Name: "value",
		},
		cli.StringFlag{
			Name:  "gasPrice",
			Value: "4000000000",
		},
		cli.StringFlag{
			Name:  "gasLimit",
			Value: "100000",
		},
		cli.Int64Flag{
			Name: "nonce",
		},
	},
	Action: func(c *cli.Context) error {
		nonce := c.Int64("nonce")
		gasprice := c.String("gasPrice")
		gaslimit := c.String("gasLimit")
		val := c.String("value")
		data := c.String("data")
		to := c.String("to")
		privkey := c.String("privkey")

		var toset bool
		var toaddr common.Address
		if to != "" {
			toaddr = common.HexToAddress(to)
			toset = true
		}
		ethval, err := Parse(val)
		if err != nil {
			return err
		}

		gaspr, ok := big.NewInt(0).SetString(gasprice, 10)
		if !ok {
			return fmt.Errorf("invalid value for gas price")
		}

		gaslim, ok := big.NewInt(0).SetString(gaslimit, 10)
		if !ok {
			return fmt.Errorf("invalid value for gas limit")
		}

		datab := common.FromHex(data)
		if datab == nil && data != "" {
			return fmt.Errorf("bad hex data: %q", data)
		}

		var tx *types.Transaction
		if toset {
			tx = types.NewTransaction(uint64(nonce), toaddr, ethval, gaslim, gaspr, datab)
		} else {
			tx = types.NewContractCreation(uint64(nonce), ethval, gaslim, gaspr, datab)
		}

		privk, err := hex.DecodeString(privkey)
		if err != nil {
			return fmt.Errorf("error decoding private key")
		}

		ecpriv, _ := btcec.PrivKeyFromBytes(btcec.S256(), privk)

		signer := types.NewEIP155Signer(big.NewInt(1))
		signed, err := types.SignTx(tx, signer, ecpriv.ToECDSA())
		if err != nil {
			return err
		}

		fmt.Println(signed.String())

		return nil
	},
}

func Parse(val string) (*big.Int, error) {
	denom, ok := big.NewInt(0).SetString("1000000000000000000", 10)
	if !ok {
		panic("not okay")
	}
	// Alg:
	// Split on decimal.
	// Let P be the number of characters to the right of the decimal
	// Concat the left and right sides
	// Parse that number as an integer
	// Multiply that number by the denominator value
	// Divide that number by 10^P
	parts := strings.Split(val, ".")

	dec := big.NewInt(1)
	fval := parts[0]
	if len(parts) > 2 {
		return nil, fmt.Errorf("invalid currency amount, expected at most one decimal: %s", val)
	}
	if len(parts) == 2 {
		decstr := parts[1]
		places := int64(len(decstr))
		dec.Exp(big.NewInt(10), big.NewInt(places), nil)
		fval += decstr
	}

	valint, ok := big.NewInt(0).SetString(fval, 10)
	if !ok {
		return nil, fmt.Errorf("error parsing value as currency: %s", val)
	}

	value := valint.Mul(valint, denom)
	value = value.Div(value, dec)

	return value, nil
}
