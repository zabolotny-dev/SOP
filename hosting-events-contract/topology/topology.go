package topology

const (
	CommandsExchange    = "server_commands_exchange"
	EventsExchange      = "server_events_exchange"
	DLXExchange         = "hosting.dlx"
	DLQRoutingKeyPrefix = "dlq."
	DLQQueueSuffix      = ".dlq"
)

func GetDLQKey(queue string) string {
	return DLQRoutingKeyPrefix + queue
}

func GetDLQQueueName(originalQueue string) string {
	return originalQueue + DLQQueueSuffix
}
