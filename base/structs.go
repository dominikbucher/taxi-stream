package base

// Configure this program using the following parameters.
type Configuration struct {
	Mode                 string
	TaxiData             []string
	NumTaxis             int32
	MaxRoutes            int32
	TargetSpeedPerSecond float32

	DbUser     string
	DbPassword string
	DbName     string
	DbHost     string
	DbPort     string
	DbSSLMode  string
}
