package words

import "encore.dev/cron"

// Executes every night at 4:00AM UTC (12:00AM EDT)
var _ = cron.NewJob("send-messages", cron.JobConfig{
	Title:    "Send Daily SMS Messages",
	Schedule: "0 4 * * *",
	Endpoint: SendMessages,
})
