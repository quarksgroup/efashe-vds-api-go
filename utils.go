package efashevdsapigo

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func setHeaders(dest, src http.Header) {

	for k, v := range src {
		dest.Del(k)
		for _, _v := range v {
			dest.Add(k, _v)
		}
	}
}

func addBearerToken(hds http.Header, token string) {
	hds.Set("Authorization", fmt.Sprintf("Bearer %s", token))
}

func httpDo(cl *http.Client, req *http.Request, jsonOut any, expectBody bool) (statusCode int, statusText string, err error) {

	res, err := cl.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer res.Body.Close()

	if !expectBody {
		return res.StatusCode, res.Status, nil
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, "", err
	}

	return res.StatusCode, res.Status, json.Unmarshal(body, jsonOut)
}

func parseTokenTstamp(tokenString string) (time.Time, error) {

	token, _, err := jwt.NewParser().ParseUnverified(tokenString, jwt.MapClaims{})
	if err != nil {
		return time.Time{}, err
	}

	numDate, err := token.Claims.GetExpirationTime()
	if err != nil {
		return time.Time{}, err
	}
	if numDate == nil {
		numDate, err := token.Claims.GetNotBefore()
		if err != nil {
			return time.Time{}, err
		}
		if numDate == nil {
			return time.Time{}, errors.New("invalid token")
		}
	}
	return numDate.Time, nil
}
