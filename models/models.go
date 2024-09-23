package models

type jsonb = interface{}

// synced with social_net
const SOCIAL_NET_REDDIT string = "Reddit"

type SocialNet struct {
	ID   int16
	Name string
}

type PersonaWithActivities struct {
	Persona        *Persona
	WeeklyActivity *WeeklyActivity
	YearlyActivity *YearlyActivity

	RedisScore float64
}

func NewPersonaWithActivities() *PersonaWithActivities {
	return &PersonaWithActivities{
		Persona:        NewPersona(),
		WeeklyActivity: NewWeeklyActivity(),
		YearlyActivity: NewYearlyActivity(),
	}
}
