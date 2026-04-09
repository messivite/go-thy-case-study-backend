package swagger

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/messivite/go-thy-case-study-backend/docs"
	"gopkg.in/yaml.v3"
)

var (
	jsonSpec     []byte
	jsonSpecOnce sync.Once
)

func specJSON() []byte {
	jsonSpecOnce.Do(func() {
		var raw any
		if err := yaml.Unmarshal(docs.OpenAPIYAML, &raw); err != nil {
			jsonSpec = []byte(`{"error":"invalid openapi spec"}`)
			return
		}
		b, err := json.MarshalIndent(raw, "", "  ")
		if err != nil {
			jsonSpec = []byte(`{"error":"json marshal failed"}`)
			return
		}
		jsonSpec = b
	})
	return jsonSpec
}

func Handler(basePath string) http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		_, _ = w.Write(docs.OpenAPIYAML)
	})

	mux.HandleFunc("/openapi.json", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		_, _ = w.Write(specJSON())
	})

	html := swaggerHTML(basePath)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(html))
	})

	return mux
}

func swaggerHTML(basePath string) string {
	specURL := basePath + "/openapi.json"
	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>THY Case Study API</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
  <style>
    body { margin: 0; background: #fafafa; }
    .swagger-ui .topbar { display: none; }
  </style>
</head>
<body>
  <div id="swagger-ui"></div>
  <script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js"></script>
  <script>
    SwaggerUIBundle({
      url: %q,
      dom_id: '#swagger-ui',
      presets: [
        SwaggerUIBundle.presets.apis,
        SwaggerUIBundle.SwaggerUIStandalonePreset
      ],
      layout: 'BaseLayout',
      deepLinking: true,
      defaultModelsExpandDepth: 1,
      defaultModelExpandDepth: 2
    });
  </script>
</body>
</html>`, specURL)
}
