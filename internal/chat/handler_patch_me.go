package chat

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/messivite/go-thy-case-study-backend/internal/auth"
	domain "github.com/messivite/go-thy-case-study-backend/internal/domain/chat"
	"github.com/messivite/go-thy-case-study-backend/internal/httpx"
)

const patchMeMaxMultipartMemory = 32 << 20

// PatchMe updates the authenticated user's profile; optional avatar is resized to 300×300 JPEG and stored in Supabase Storage.
func (h *Handler) PatchMe(w http.ResponseWriter, r *http.Request) {
	user, ok := auth.AuthenticatedUserFromContext(r.Context())
	if !ok {
		httpx.Unauthorized(w)
		return
	}

	ct := strings.ToLower(r.Header.Get("Content-Type"))
	var patch domain.ProfilePatch
	var rawAvatar []byte
	var err error

	if strings.Contains(ct, "multipart/form-data") {
		if err := r.ParseMultipartForm(patchMeMaxMultipartMemory); err != nil {
			httpx.BadRequest(w, "invalid multipart body")
			return
		}
		patch, rawAvatar, err = parsePatchMeMultipart(r)
	} else {
		patch, err = parsePatchMeJSON(r.Body)
	}
	if err != nil {
		httpx.BadRequest(w, err.Error())
		return
	}

	prof, err := h.uc.PatchMe(r.Context(), user.UserID, patch, rawAvatar)
	if err != nil {
		writeAppError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toMeResponse(user, prof))
}

type patchMeJSONBody struct {
	DisplayName         *string `json:"displayName"`
	PreferredProvider   *string `json:"preferredProvider"`
	PreferredModel      *string `json:"preferredModel"`
	Locale              *string `json:"locale"`
	Timezone            *string `json:"timezone"`
	OnboardingCompleted *bool   `json:"onboardingCompleted"`
}

func parsePatchMeJSON(r io.ReadCloser) (domain.ProfilePatch, error) {
	defer r.Close()
	var body patchMeJSONBody
	if err := json.NewDecoder(r).Decode(&body); err != nil {
		return domain.ProfilePatch{}, err
	}
	return domain.ProfilePatch{
		DisplayName:         body.DisplayName,
		PreferredProvider:   body.PreferredProvider,
		PreferredModel:      body.PreferredModel,
		Locale:              body.Locale,
		Timezone:            body.Timezone,
		OnboardingCompleted: body.OnboardingCompleted,
	}, nil
}

func parsePatchMeMultipart(r *http.Request) (domain.ProfilePatch, []byte, error) {
	form := r.MultipartForm
	if form == nil {
		return domain.ProfilePatch{}, nil, errors.New("missing multipart form")
	}
	var patch domain.ProfilePatch
	setFormString := func(key string, dest **string) {
		if vals, ok := form.Value[key]; ok {
			s := ""
			if len(vals) > 0 {
				s = vals[0]
			}
			*dest = &s
		}
	}
	setFormString("displayName", &patch.DisplayName)
	setFormString("preferredProvider", &patch.PreferredProvider)
	setFormString("preferredModel", &patch.PreferredModel)
	setFormString("locale", &patch.Locale)
	setFormString("timezone", &patch.Timezone)

	if vals, ok := form.Value["onboardingCompleted"]; ok && len(vals) > 0 && strings.TrimSpace(vals[0]) != "" {
		b, err := strconv.ParseBool(strings.TrimSpace(vals[0]))
		if err != nil {
			return domain.ProfilePatch{}, nil, err
		}
		patch.OnboardingCompleted = &b
	}

	files := form.File["avatar"]
	if len(files) == 0 {
		return patch, nil, nil
	}
	f, err := files[0].Open()
	if err != nil {
		return domain.ProfilePatch{}, nil, err
	}
	defer f.Close()
	raw, err := io.ReadAll(io.LimitReader(f, patchMeMaxMultipartMemory))
	if err != nil {
		return domain.ProfilePatch{}, nil, err
	}
	return patch, raw, nil
}
