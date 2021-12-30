package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
)

type Lambda struct {
	cognitoPoolID string
	region        string
}

const (
	ALLOW = "ALLOW"
	DENY  = "DENY"
)

var (
	accountID string
	apiID     string
	stage     string
)

func (l *Lambda) handler(r events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	token, err := l.parseToken(r.AuthorizationToken)
	if err != nil {
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("unauthorized")
	}

	split := strings.Split(strings.Split(r.MethodArn, ":")[5], "/")
	apiID = split[0]
	stage = split[1]
	accountID = split[4]

	principalID := l.getUsername(*token)
	resp := l.buildPolicy(principalID)

	groups := l.getGroups(*token)
	if len(groups) > 1 {
		log.Printf("[ERROR] too many groups (len %d)", len(groups))
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("forbidden")
	}

	l.deviceAccess(&resp, groups[0])
	l.userAccess(&resp)

	return resp, nil
}

func (l *Lambda) deviceAccess(r *events.APIGatewayCustomAuthorizerResponse, deviceID string) {
	// devices are only allowed to fetch information about themselves
	l.applyMethod(r, ALLOW, http.MethodGet, "/devices/"+deviceID)

	// devices are not allowed to access capture api points
	l.applyMethod(r, DENY, "*", "/captures*")
}

func (l *Lambda) userAccess(r *events.APIGatewayCustomAuthorizerResponse) {
	// users have access to all data
	l.applyMethod(r, ALLOW, "*", "/devices*")
	l.applyMethod(r, ALLOW, "*", "/captures*")
}

func (l *Lambda) applyMethod(r *events.APIGatewayCustomAuthorizerResponse, effect, method, resource string) {
	arn := fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/%s/%s/%s", l.region, accountID, apiID, stage, method, resource)
	policy := events.IAMPolicyStatement{
		Action:   []string{"execute-api:Invoke"},
		Effect:   effect,
		Resource: []string{arn},
	}

	r.PolicyDocument.Statement = append(r.PolicyDocument.Statement, policy)
}

func (l *Lambda) parseToken(t string) (*jwt.Token, error) {
	keys, err := jwk.Fetch(context.Background(), "https://cognito-idp."+l.region+".amazonaws.com/"+l.cognitoPoolID+"/.well-known/jwks.json")
	if err != nil {
		return nil, err
	}

	token, err := jwt.Parse([]byte(t), jwt.WithKeySet(keys), jwt.WithValidate(false))
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (l *Lambda) getUsername(t jwt.Token) string {
	username, ok := t.Get("cognito:username")
	if !ok {
		return ""
	}

	return username.(string)
}

func (l *Lambda) getGroups(t jwt.Token) []string {
	var result []string

	claims, _ := t.Get("cognito:groups")
	for _, s := range claims.([]interface{}) {
		m, _ := regexp.Match(`[0-9]+`, []byte(s.(string)))
		if m {
			result = append(result, s.(string))
		}
	}

	return result
}

func (l *Lambda) buildPolicy(principalID string) events.APIGatewayCustomAuthorizerResponse {
	var resp = events.APIGatewayCustomAuthorizerResponse{
		PrincipalID: principalID,
		PolicyDocument: events.APIGatewayCustomAuthorizerPolicy{
			Version: "2012-10-17",
		},
	}

	return resp
}

func main() {
	l := &Lambda{
		cognitoPoolID: os.Getenv("COGNITO_POOL_ID"),
		region:        os.Getenv("AWS_REGION"),
	}

	lambda.Start(l.handler)
}
