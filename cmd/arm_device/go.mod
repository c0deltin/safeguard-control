module codeltin.io/safeguard/control/arm-device

go 1.17

replace (
	model => ../../pkg/model
	utils => ../../pkg/utils
)

require (
	github.com/aws/aws-lambda-go v1.27.1
	github.com/aws/aws-sdk-go v1.42.25
	model v0.0.0-00010101000000-000000000000
	utils v0.0.0-00010101000000-000000000000
)

require github.com/jmespath/go-jmespath v0.4.0 // indirect
