module hosting-provisioning-service

go 1.24.4

require (
	github.com/joho/godotenv v1.5.1 // indirect
	github.com/wagslane/go-rabbitmq v0.15.0 // indirect
)

require hosting-kit v0.0.0-00010101000000-000000000000

require (
	github.com/ardanlabs/conf/v3 v3.9.0
	github.com/google/uuid v1.6.0
	github.com/rabbitmq/amqp091-go v1.10.0 // indirect
	hosting-events-contract v0.0.0-00010101000000-000000000000
)

replace hosting-events-contract => ../hosting-events-contract

replace hosting-kit => ../hosting-kit
