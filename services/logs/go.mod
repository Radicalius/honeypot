module log_server

go 1.15

require (
	github.com/aws/aws-sdk-go v1.44.189
	logging v0.0.0-00010101000000-000000000000
)

replace logging => ../../packages/logging
