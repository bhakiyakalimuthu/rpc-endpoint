package server

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/ethereum/go-ethereum/core/types"
	"github.com/pkg/errors"
)

func GetIP(r *http.Request) string {
	forwarded := r.Header.Get("X-Forwarded-For")
	if forwarded != "" {
		return forwarded
	}
	return r.RemoteAddr
}

// CHROME_ID: nkbihfbeogaeaoehlefnkodbefgpgknn
func IsMetamask(r *http.Request) bool {
	return r.Header.Get("Origin") == "chrome-extension://nkbihfbeogaeaoehlefnkodbefgpgknn"
}

// FIREFOX_ID: webextension@metamask.io
func IsMetamaskMoz(r *http.Request) bool {
	return r.Header.Get("Origin") == "moz-extension://57f9aaf6-270a-154f-9a8a-632d0db4128c"
}

func ProxyRequest(proxyUrl string, body []byte) (*http.Response, error) {
	// Create new request:
	req, err := http.NewRequest("POST", proxyUrl, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Length", strconv.Itoa(len(body)))

	client := &http.Client{
		Timeout: time.Duration(10 * time.Second),
	}
	return client.Do(req)
}

func GetTx(rawTxHex string) (*types.Transaction, error) {
	if len(rawTxHex) < 2 {
		return nil, errors.New("invalid raw transaction")
	}

	rawTxBytes, err := hex.DecodeString(rawTxHex[2:])
	if err != nil {
		return nil, errors.New("invalid raw transaction")
	}

	tx := new(types.Transaction)
	if err := tx.UnmarshalBinary(rawTxBytes); err != nil {
		return nil, errors.New("error unmarshalling")
	}

	return tx, nil
}

func GetSenderFromTx(tx *types.Transaction) (string, error) {
	signer := types.LatestSignerForChainID(tx.ChainId())
	sender, err := types.Sender(signer, tx)
	if err != nil {
		return "", err
	}
	return sender.Hex(), nil
}

func GetSenderFromRawTx(tx *types.Transaction) (string, error) {
	from, err := GetSenderFromTx(tx)
	if err != nil {
		return "", errors.New("error getting from")
	}

	return from, nil
}

func TruncateText(s string, max int) string {
	if len(s) > max {
		r := 0
		for i := range s {
			r++
			if r > max {
				return s[:i]
			}
		}
	}
	return s
}

func SendRpcAndParseResponseTo(url string, req *JsonRpcRequest) (*JsonRpcResponse, error) {
	jsonData, err := json.Marshal(req)
	if err != nil {
		return nil, errors.Wrap(err, "marshal")
	}

	// fmt.Printf("%s\n", jsonData)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, errors.Wrap(err, "post")
	}

	respData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "read")
	}

	// Unmarshall JSON-RPC response and check for error inside
	jsonRpcResp := new(JsonRpcResponse)
	if err := json.Unmarshal(respData, jsonRpcResp); err != nil {
		return nil, errors.Wrap(err, "unmarshal")
	}

	return jsonRpcResp, nil
}
