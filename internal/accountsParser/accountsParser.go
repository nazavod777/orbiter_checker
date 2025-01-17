package accountsParser

import (
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	log "github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"main/pkg/types"
	"main/pkg/util"
	"strconv"
)

func getBalances(
	accountData types.AccountData,
	authToken string,
) float64 {
	var err error

	type responseStruct struct {
		Code   int `json:"code"`
		Result struct {
			Amount string `json:"amount"`
		} `json:"result"`
	}

	for {
		client := util.GetClient(util.GetProxy())

		req := fasthttp.AcquireRequest()

		req.SetRequestURI("https://airdrop-api.orbiter.finance/airdrop/snapshot")
		req.Header.SetMethod("POST")
		req.Header.SetContentType("application/json")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("accept", "*/*")
		req.Header.Set("accept-language", "ru,en;q=0.9,vi;q=0.8,es;q=0.7,cy;q=0.6")
		req.Header.Set("origin", "https://www.orbiter.finance")
		req.Header.Set("referrer", "https://www.orbiter.finance")
		req.Header.SetReferer("https://www.orbiter.finance")
		req.Header.Set("token", authToken)

		resp := fasthttp.AcquireResponse()

		if err = client.Do(req, resp); err != nil {
			log.Printf("%s | Error When Doing Request When Parsing Balance: %s", accountData.AccountKeyHex, err)

			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		var response responseStruct

		if err = json.Unmarshal(resp.Body(), &response); err != nil {
			log.Printf("%s | Failed To Parse JSON Response When Parsing Balance: %s, response: %s",
				accountData.AccountKeyHex, err, string(resp.Body()))
			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		if resp.StatusCode() != 201 || response.Code != 0 {
			log.Printf("%s | Wrong Response When Parsing Balance: %s",
				accountData.AccountKeyHex, string(resp.Body()))
			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		if response.Result.Amount == "" {
			return 0
		}

		formattedAmount, err := strconv.ParseFloat(response.Result.Amount, 64)

		if err != nil {
			log.Printf("%s | Failed To Converting Amount When Parsing Balance: %s, response: %s",
				accountData.AccountKeyHex, err, string(resp.Body()))
			fasthttp.ReleaseRequest(req)
			fasthttp.ReleaseResponse(resp)
			continue
		}

		fasthttp.ReleaseRequest(req)
		fasthttp.ReleaseResponse(resp)

		return formattedAmount
	}
}

func ParseAccount(accountData types.AccountData) {
	signHash, err := crypto.Sign(accounts.TextHash([]byte("Orbiter Airdrop")), accountData.AccountKey)

	if err != nil {
		log.Printf("%s | Failed to sign auth message: %v", accountData.AccountAddress.String(), err)
		return
	}

	signHash[64] += 27
	signHashHex := hexutil.Encode(signHash)

	tokensBalance := getBalances(accountData, signHashHex)

	log.Printf("%s | Balance: %g $OBT",
		accountData.AccountKeyHex, tokensBalance)

	if tokensBalance > 0 {
		util.AppendFile("with_tokens.txt",
			fmt.Sprintf("%s | Balance: %g $OBT\n",
				accountData.AccountKeyHex, tokensBalance))
	}
}
