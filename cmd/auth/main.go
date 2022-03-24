package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/lestrrat-go/jwx/jwk"
	"github.com/lestrrat-go/jwx/jwt"
	"github.com/sirupsen/logrus"
)

type Lambda struct {
	logger        *logrus.Logger
	cognitoPoolID string
	region        string
}

const (
	ALLOW = "Allow"
	DENY  = "Deny"
)

var (
	accountID string
	apiID     string
	stage     string
	region    string
)

func (l *Lambda) handler(r events.APIGatewayCustomAuthorizerRequest) (events.APIGatewayCustomAuthorizerResponse, error) {
	token, err := l.parseToken(r.AuthorizationToken)
	if err != nil {
		l.logger.Errorf("failed to parse token, %v", err)
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("unauthorized")
	}

	methodArn := strings.Split(r.MethodArn, ":")
	region = methodArn[3]
	accountID = methodArn[4]

	gatewayArn := strings.Split(methodArn[5], "/")
	apiID = gatewayArn[0]
	stage = gatewayArn[1]

	groups := l.getGroups(*token)
	if len(groups) < 1 {
		l.logger.Error("no groups found")
	} else if len(groups) > 1 {
		l.logger.Errorf("too many groups: %d", len(groups))
		return events.APIGatewayCustomAuthorizerResponse{}, errors.New("forbidden")
	}

	principalID := l.getUsername(*token)
	resp := l.buildPolicy(principalID)
	if l.isAdmin(groups) {
		l.logger.Printf("applying user access to user %s", principalID)
		l.userAccess(&resp)
	} else {
		l.logger.Printf("applying device access to user %s", principalID)
		l.deviceAccess(&resp, groups[0])
	}

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

func (l *Lambda) isAdmin(groups []string) bool {
	for _, x := range groups {
		if x == "admin" {
			return true
		}
	}

	return false
}

func (l *Lambda) applyMethod(r *events.APIGatewayCustomAuthorizerResponse, effect, method, resource string) {
	resource = strings.TrimPrefix(resource, "/")
	arn := fmt.Sprintf("arn:aws:execute-api:%s:%s:%s/%s/%s/%s", region, accountID, apiID, stage, method, resource)

	r.PolicyDocument.Statement = append(r.PolicyDocument.Statement, events.IAMPolicyStatement{
		Action:   []string{"execute-api:Invoke"},
		Effect:   effect,
		Resource: []string{arn},
	})
}

func (l *Lambda) parseToken(t string) (*jwt.Token, error) {
	keys, err := jwk.Fetch(context.Background(), "https://cognito-idp."+l.region+".amazonaws.com/"+l.cognitoPoolID+"/.well-known/jwks.json")
	if err != nil {
		return nil, err
	}

	split := strings.Split(t, "Bearer ")
	if len(split) != 2 {
		return nil, errors.New("invalid token format")
	}

	token, err := jwt.Parse([]byte(split[1]), jwt.WithKeySet(keys), jwt.WithValidate(false))
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (l *Lambda) getUsername(t jwt.Token) string {
	username, ok := t.Get("username")
	if !ok {
		return ""
	}

	return username.(string)
}

func (l *Lambda) getGroups(t jwt.Token) []string {
	var result []string

	claims, ok := t.Get("cognito:groups")
	if ok {
		for _, s := range claims.([]interface{}) {
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
		logger:        logrus.New(),
	}

	l.logger.SetFormatter(&logrus.JSONFormatter{})

	lambda.Start(l.handler)
}
