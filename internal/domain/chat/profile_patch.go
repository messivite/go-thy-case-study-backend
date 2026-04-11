package chat

// ProfilePatch is a partial update for public.profiles. Nil pointer = field unchanged.
type ProfilePatch struct {
	DisplayName         *string
	PreferredProvider   *string
	PreferredModel      *string
	Locale              *string
	Timezone            *string
	AvatarURL           *string
	OnboardingCompleted *bool
}
