package devices

type HostInterface interface {
	Log(msg string, args ...any)
	LogErr(msg string, args ...any)
	LogWarn(msg string, args ...any)
	SerialCable() SerialCable
}
