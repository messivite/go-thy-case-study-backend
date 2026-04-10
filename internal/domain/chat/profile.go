package chat

import "time"

// UserProfile is a row from public.profiles (1:1 with auth.users).
type UserProfile struct {
	ID                  string
	DisplayName         string
	AvatarURL           string
	Role                string // profiles.role (user | admin | moderator)
	IsActive            bool
	PreferredProvider   string
	PreferredModel      string
	Locale              string
	Timezone            string
	Metadata            map[string]any
	LastSeenAt          *time.Time
	OnboardingCompleted bool
	CreatedAt           time.Time
	UpdatedAt           time.Time
	IsAnonymous         bool
}
