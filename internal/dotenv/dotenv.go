package dotenv

import (
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// LoadLocalEnv, çalışma dizininden başlayarak üst klasörlere doğru ilk bulduğu `.env` dosyasını yükler.
// godotenv.Overload kullanılır: dosyadaki değerler, shell'de kalmış eski export'ların üzerine yazar
// (ör. daha önce `export CHAT_PERSISTENCE=memory` kaldıysa .env içindeki supabase geçerli olur).
// Üretimde genelde .env olmaz; platform env yine set edilmişse Overload sonrası o dosyadaki anahtarlar geçerli olur — imajda .env koymayın.
func LoadLocalEnv() {
	dir, err := os.Getwd()
	if err != nil {
		return
	}
	for range 32 {
		p := filepath.Join(dir, ".env")
		if st, statErr := os.Stat(p); statErr == nil && !st.IsDir() {
			_ = godotenv.Overload(p)
			return
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
}
