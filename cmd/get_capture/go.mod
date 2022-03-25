module codeltin.io/safeguard/control/get-capture

go 1.18

replace (
	model => ../../pkg/model
	utils => ../../pkg/utils
)

require (
	github.com/aws/aws-lambda-go v1.28.0
	github.com/aws/aws-sdk-go v1.43.25
	github.com/sirupsen/logrus v1.8.1
	model v0.0.0-00010101000000-000000000000
	utils v0.0.0-00010101000000-000000000000
)

require (
	github.com/jmespath/go-jmespath v0.4.0 // indirect
	golang.org/x/sys v0.0.0-20211216021012-1d35b9e2eb4e // indirect
)
