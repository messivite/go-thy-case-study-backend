#!/usr/bin/env bash
set -eu
cd "$(dirname "$0")/.."

if [ ! -f api.yaml ]; then
  cat > api.yaml <<'EOF'
version: "1"
basePath: /api
output:
  serverDir: server
  handlersDir: handlers
endpoints:
  - method: GET
    path: /health
    handler: Health
    auth: false
  - method: GET
    path: /me
    handler: Me
    auth: true
  - method: GET
    path: /providers
    handler: ListProviders
    auth: true
  - method: GET
    path: /sessions
    handler: ListSessions
    auth: true
  - method: POST
    path: /sessions
    handler: CreateSession
    auth: true
  - method: GET
    path: /sessions/{sessionID}/messages
    handler: ListMessages
    auth: true
  - method: POST
    path: /sessions/{sessionID}/messages
    handler: PostMessage
    auth: true
EOF
  echo "Created api.yaml"
else
  echo "api.yaml already exists"
fi

if [ ! -f .env.example ]; then
  cat > .env.example <<'EOF'
PORT=8080
SUPABASE_URL=
SUPABASE_ANON_KEY=
SUPABASE_SERVICE_ROLE_KEY=
SUPABASE_JWT_SECRET=
SUPABASE_JWT_VALIDATION_MODE=auto
SUPABASE_ROLE_CLAIM_KEY=role
OPENAI_API_KEY=
GEMINI_API_KEY=
EOF
  echo "Created .env.example"
else
  echo ".env.example already exists"
fi

echo "Setup complete. To generate handlers or server scaffolding, run:"
echo "  go run github.com/messivite/gosupabase/cmd/gosupabase init"
echo "  go run github.com/messivite/gosupabase/cmd/gosupabase gen"
