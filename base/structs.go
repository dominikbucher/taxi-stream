package base

// Configure this program using the following parameters.
type Configuration struct {
	Mode      string
	TaxiData  []string
	NumTaxis  int32
	MaxRoutes int32

	MaxClients           int
	ClientRequestsPerSec float64

	TargetSpeedPerSecond     float64
	TrackpointPrepWindowSize float64
	TimeWarp                 float64

	WebSocketPort int

	Log bool

	DbUser     string
	DbPassword string
	DbName     string
	DbHost     string
	DbPort     string
	DbSSLMode  string
}
