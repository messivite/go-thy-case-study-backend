//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type APIConfig struct {
	BasePath  string     `yaml:"basePath"`
	Endpoints []Endpoint `yaml:"endpoints"`
}

type Endpoint struct {
	Method  string `yaml:"method"`
	Path    string `yaml:"path"`
	Handler string `yaml:"handler"`
	Auth    bool   `yaml:"auth"`
}

func main() {
	cfg, err := readAPIConfig("api.yaml")
	if err != nil {
		fail("read api.yaml: %v", err)
	}

	pathBlocks := buildPathBlocks(cfg.Endpoints)
	doc := renderOpenAPI(pathBlocks)

	if err := os.WriteFile("docs/openapi.yaml", []byte(doc), 0644); err != nil {
		fail("write docs/openapi.yaml: %v", err)
	}

	var raw any
	if err := yaml.Unmarshal([]byte(doc), &raw); err != nil {
		fail("yaml parse generated doc: %v", err)
	}
	b, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		fail("marshal docs/openapi.json: %v", err)
	}
	if err := os.WriteFile("docs/openapi.json", append(b, '\n'), 0644); err != nil {
		fail("write docs/openapi.json: %v", err)
	}

	fmt.Println("docs/openapi.yaml updated from api.yaml")
	fmt.Println("docs/openapi.json updated from generated YAML")
}

func readAPIConfig(path string) (APIConfig, error) {
	var cfg APIConfig
	b, err := os.ReadFile(path)
	if err != nil {
		return cfg, err
	}
	err = yaml.Unmarshal(b, &cfg)
	return cfg, err
}

func buildPathBlocks(endpoints []Endpoint) string {
	byPath := map[string][]Endpoint{}
	for _, e := range endpoints {
		byPath[e.Path] = append(byPath[e.Path], e)
	}
	paths := make([]string, 0, len(byPath))
	for p := range byPath {
		paths = append(paths, p)
	}
	sort.Strings(paths)

	var out strings.Builder
	for _, p := range paths {
		out.WriteString("  " + p + ":\n")
		eps := byPath[p]
		sort.Slice(eps, func(i, j int) bool {
			return strings.ToUpper(eps[i].Method) < strings.ToUpper(eps[j].Method)
		})
		for _, e := range eps {
			out.WriteString(operationFor(e))
		}
	}
	return out.String()
}

func operationFor(e Endpoint) string {
	m := strings.ToLower(strings.TrimSpace(e.Method))
	sec := ""
	if !e.Auth {
		sec = "      security: []\n"
	}
	pathParam := ""
	if strings.Contains(e.Path, "{chatID}") {
		pathParam = "      parameters:\n        - $ref: \"#/components/parameters/chatID\"\n"
	}

	switch e.Handler {
	case "Health":
		return fmt.Sprintf(`    %s:
      tags: [health]
      summary: Health check
      operationId: health
%s      responses:
        "200":
          description: Servis aktif
          content:
            text/plain:
              schema:
                type: string
                example: OK
`, m, sec)
	case "Me":
		return fmt.Sprintf(`    %s:
      tags: [auth]
      summary: Mevcut kullanıcı bilgisi
      operationId: getMe
%s      responses:
        "200":
          description: Kimlik doğrulanmış kullanıcı bilgisi
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Me"
        "401":
          $ref: "#/components/responses/Unauthorized"
`, m, sec)
	case "PatchMe":
		return fmt.Sprintf(`    %s:
      tags: [auth]
      summary: Profil güncelle (JSON veya multipart + avatar)
      operationId: patchMe
%s      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/PatchMeRequest"
          multipart/form-data:
            schema:
              type: object
              properties:
                displayName: { type: string }
                preferredProvider: { type: string }
                preferredModel: { type: string }
                locale: { type: string }
                timezone: { type: string }
                onboardingCompleted: { type: string, description: "true veya false" }
                avatar:
                  type: string
                  format: binary
                  description: 300x300 JPEG'e indirgenir; Supabase Storage avatars bucket
      responses:
        "200":
          description: Güncel kimlik + profil (GET /me ile aynı şema)
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/Me"
        "400":
          $ref: "#/components/responses/BadRequest"
        "401":
          $ref: "#/components/responses/Unauthorized"
`, m, sec)
	case "ListProviders":
		return fmt.Sprintf(`    %s:
      tags: [providers]
      summary: Aktif LLM provider listesi
      operationId: listProviders
%s      responses:
        "200":
          description: Provider listesi
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ListProvidersResponse"
        "401":
          $ref: "#/components/responses/Unauthorized"
`, m, sec)
	case "ListModels":
		return fmt.Sprintf(`    %s:
      tags: [models]
      summary: Kullanılabilir LLM model listesi
      operationId: listModels
%s      responses:
        "200":
          description: Aktif model kataloğu (provider + model)
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ListModelsResponse"
        "401":
          $ref: "#/components/responses/Unauthorized"
`, m, sec)
	case "CreateSession":
		return fmt.Sprintf(`    %s:
      tags: [chats]
      summary: Yeni sohbet oturumu oluştur
      operationId: createSession
%s      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/CreateSessionRequest"
      responses:
        "201":
          description: Oturum oluşturuldu
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/CreateSessionResponse"
        "400":
          $ref: "#/components/responses/BadRequest"
        "401":
          $ref: "#/components/responses/Unauthorized"
`, m, sec)
	case "ListSessions":
		return fmt.Sprintf(`    %s:
      tags: [chats]
      summary: Sohbet listesi
      operationId: listSessions
%s      responses:
        "200":
          description: Sohbet listesi
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: "#/components/schemas/ChatListItem"
        "401":
          $ref: "#/components/responses/Unauthorized"
`, m, sec)
	case "GetChat":
		return fmt.Sprintf(`    %s:
      tags: [chats]
      summary: Sohbet detayı
      operationId: getChat
%s%s      responses:
        "200":
          description: Sohbet detayı
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ChatDetail"
        "401":
          $ref: "#/components/responses/Unauthorized"
        "404":
          $ref: "#/components/responses/NotFound"
`, m, sec, pathParam)
	case "PostMessage":
		return fmt.Sprintf(`    %s:
      tags: [messages]
      summary: Mesaj gönder (non-stream)
      operationId: postMessage
%s%s      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/PostMessageRequest"
      responses:
        "201":
          description: Asistan yanıtı
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/AssistantResponse"
        "400":
          $ref: "#/components/responses/BadRequest"
        "401":
          $ref: "#/components/responses/Unauthorized"
        "404":
          $ref: "#/components/responses/NotFound"
        "429":
          $ref: "#/components/responses/QuotaExceeded"
        "502":
          $ref: "#/components/responses/ProviderError"
        "504":
          $ref: "#/components/responses/ProviderTimeout"
`, m, sec, pathParam)
	case "StreamMessage":
		return fmt.Sprintf(`    %s:
      tags: [messages]
      summary: Mesaj gönder (SSE stream)
      operationId: streamMessage
%s%s      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: "#/components/schemas/PostMessageRequest"
      responses:
        "200":
          description: |
            Server-Sent Events: data satırları JSON taşır. İlk olay genelde type meta — meta.userMessageId
            (DB'ye yazılan kullanıcı mesajı), provider, model. done öncesi meta'da ayrıca assistantMessageId.
            Delta için type delta; bitiş için type done veya hata/iptal olayları.
          content:
            text/event-stream:
              schema:
                type: string
        "400":
          $ref: "#/components/responses/BadRequest"
        "401":
          $ref: "#/components/responses/Unauthorized"
        "404":
          $ref: "#/components/responses/NotFound"
        "429":
          $ref: "#/components/responses/QuotaExceeded"
        "502":
          $ref: "#/components/responses/ProviderError"
        "504":
          $ref: "#/components/responses/ProviderTimeout"
`, m, sec, pathParam)
	default:
		return fmt.Sprintf(`    %s:
      tags: [misc]
      summary: %s
      operationId: %s
%s%s      responses:
        "200":
          description: OK
        "401":
          $ref: "#/components/responses/Unauthorized"
`, m, e.Handler, toOperationID(e.Handler), sec, pathParam)
	}
}

func toOperationID(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return "operation"
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func renderOpenAPI(pathBlocks string) string {
	return `openapi: "3.1.0"
info:
  title: THY Case Study Backend API
  description: |
    LLM sohbet backend servisi.
    Supabase tabanlı kimlik doğrulama, çoklu LLM provider desteği,
    token kota yönetimi ve etkileşim audit loglama içerir.
  version: "1.0.0"
servers:
  - url: /api
    description: Default API base path
security:
  - bearerAuth: []
tags:
  - name: health
    description: Sistem sağlık kontrolü
  - name: auth
    description: Kimlik doğrulama
  - name: providers
    description: LLM provider yönetimi
  - name: models
    description: Desteklenen LLM modelleri
  - name: chats
    description: Sohbet oturumları
  - name: messages
    description: Mesaj gönderme ve stream
paths:
` + pathBlocks + `components:
  securitySchemes:
    bearerAuth:
      type: http
      scheme: bearer
      bearerFormat: JWT
      description: Supabase access token
  parameters:
    chatID:
      name: chatID
      in: path
      required: true
      schema:
        type: string
        format: uuid
  schemas:
    MeUser:
      type: object
      description: JWT / access token özeti (gosupabase doğrulaması sonrası).
      properties:
        id: { type: string, format: uuid }
        email: { type: string, format: email }
        role: { type: string, description: RBAC birincil rol (SUPABASE_ROLE_CLAIM_KEY) }
        roles: { type: array, items: { type: string } }
        phone: { type: string }
        sessionId: { type: string }
        iss: { type: string }
        aud: { type: string }
        iat: { type: integer, format: int64 }
        exp: { type: integer, format: int64 }
        issuedAt: { type: string, format: date-time }
        expiresAt: { type: string, format: date-time }
        appMetadata: { type: object, additionalProperties: true }
        userMetadata: { type: object, additionalProperties: true }
      required: [id]
    MeProfile:
      type: object
      description: public.profiles satırı (display_name, avatar_url, tercihler, …).
      properties:
        id: { type: string, format: uuid }
        displayName: { type: string }
        avatarUrl: { type: string, format: uri }
        role: { type: string, description: profiles.role (user | admin | moderator) }
        isActive: { type: boolean }
        preferredProvider: { type: string }
        preferredModel: { type: string }
        locale: { type: string }
        timezone: { type: string }
        metadata: { type: object, additionalProperties: true }
        lastSeenAt: { type: string, format: date-time }
        onboardingCompleted: { type: boolean }
        createdAt: { type: string, format: date-time }
        updatedAt: { type: string, format: date-time }
        isAnonymous: { type: boolean }
      required: [id, isActive, onboardingCompleted, isAnonymous]
    Me:
      type: object
      description: Kimlik (user) + uygulama profili (profile).
      properties:
        user: { $ref: "#/components/schemas/MeUser" }
        profile: { $ref: "#/components/schemas/MeProfile" }
      required: [user, profile]
    PatchMeRequest:
      type: object
      description: Tüm alanlar opsiyonel; en az bir alan veya multipart istekte avatar dosyası gerekir.
      properties:
        displayName: { type: string }
        preferredProvider: { type: string }
        preferredModel: { type: string }
        locale: { type: string }
        timezone: { type: string }
        onboardingCompleted: { type: boolean }
    ProviderInfo:
      type: object
      properties:
        name: { type: string }
        model: { type: string }
        enabled: { type: boolean }
      required: [name, model, enabled]
    ListProvidersResponse:
      type: object
      properties:
        default: { type: string }
        providers:
          type: array
          items: { $ref: "#/components/schemas/ProviderInfo" }
      required: [default, providers]
    SupportedModelItem:
      type: object
      properties:
        provider: { type: string }
        model: { type: string }
        displayName: { type: string }
        supportsStream: { type: boolean }
      required: [provider, model, displayName, supportsStream]
    ListModelsResponse:
      type: object
      properties:
        models:
          type: array
          items: { $ref: "#/components/schemas/SupportedModelItem" }
      required: [models]
    CreateSessionRequest:
      type: object
      properties:
        title: { type: string }
        provider: { type: string }
        model: { type: string }
        content:
          type: string
          description: |
            Opsiyonel ilk kullanıcı mesajı. Doluysa oturum açılır açılmaz LLM yanıtı üretilir;
            cevapta assistantMessage alanı döner (ayrıca POST /chats/{id}/messages çağrısı gerekmez).
    CreateSessionResponse:
      type: object
      properties:
        id: { type: string, format: uuid }
        provider: { type: string }
        model: { type: string }
        assistantMessage: { $ref: "#/components/schemas/ChatMessage" }
        usage: { type: object, additionalProperties: true }
      required: [id, provider, model]
    ChatMessage:
      type: object
      properties:
        id: { type: string, format: uuid }
        createdAt: { type: string, format: date-time }
        role:
          type: string
          enum: [user, assistant, system]
        content: { type: string }
        provider: { type: string }
        model: { type: string }
      required: [role, content]
    ChatListItem:
      type: object
      properties:
        id: { type: string, format: uuid }
        title: { type: string }
        provider: { type: string }
        model: { type: string }
        createdAt: { type: string, format: date-time }
        updatedAt: { type: string, format: date-time }
        lastMessagePreview: { type: string }
      required: [id, title, provider, model, createdAt, updatedAt]
    ChatDetail:
      type: object
      properties:
        id: { type: string, format: uuid }
        title: { type: string }
        provider: { type: string }
        model: { type: string }
        messages:
          type: array
          items: { $ref: "#/components/schemas/ChatMessage" }
      required: [id, title, provider, model, messages]
    PostMessageRequest:
      type: object
      properties:
        provider: { type: string }
        model: { type: string }
        content: { type: string }
        messages:
          type: array
          items: { $ref: "#/components/schemas/ChatMessage" }
    AssistantResponse:
      type: object
      properties:
        assistantMessage: { $ref: "#/components/schemas/ChatMessage" }
        usage:
          type: object
          additionalProperties: true
      required: [assistantMessage]
    ErrorBody:
      type: object
      properties:
        code: { type: string }
        message: { type: string }
      required: [code, message]
    ErrorEnvelope:
      type: object
      properties:
        error: { $ref: "#/components/schemas/ErrorBody" }
      required: [error]
  responses:
    Unauthorized:
      description: Kimlik doğrulama gerekli
      content:
        application/json:
          schema: { $ref: "#/components/schemas/ErrorEnvelope" }
    BadRequest:
      description: Geçersiz istek
      content:
        application/json:
          schema: { $ref: "#/components/schemas/ErrorEnvelope" }
    NotFound:
      description: Kaynak bulunamadı
      content:
        application/json:
          schema: { $ref: "#/components/schemas/ErrorEnvelope" }
    QuotaExceeded:
      description: Token kotası aşıldı
      content:
        application/json:
          schema: { $ref: "#/components/schemas/ErrorEnvelope" }
    ProviderError:
      description: LLM provider hatası
      content:
        application/json:
          schema: { $ref: "#/components/schemas/ErrorEnvelope" }
    ProviderTimeout:
      description: LLM provider zaman aşımı
      content:
        application/json:
          schema: { $ref: "#/components/schemas/ErrorEnvelope" }
`
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}

