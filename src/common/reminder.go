package common

type ReminderType int

const (
	Callback ReminderType = iota	
)

type Reminder struct {
	Id string
	Type ReminderType
	Payload string
}


