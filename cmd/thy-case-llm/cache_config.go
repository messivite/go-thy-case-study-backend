package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

const (
	cacheEnvMarkerBegin = "# --- thy-case-llm: HTTP response cache ---"
	cacheEnvMarkerEnd   = "# --- end thy-case-llm cache ---"
)

func handleCache(args []string) {
	if len(args) < 1 || args[0] != "config" {
		printCacheUsage()
		os.Exit(1)
	}
	cmdCacheConfig()
}

func printCacheUsage() {
	fmt.Println(`Kullanım:
  thy-case-llm cache config    Etkileşimli sorularla .env içine CACHE_* ayarlarını yazar`)
}

func cmdCacheConfig() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Print("HTTP API yanıt önbelleği açılsın mı? [y/N]: ")
	line := strings.ToLower(strings.TrimSpace(readLine(reader)))
	enabled := line == "y" || line == "yes" || line == "e" || line == "evet"

	if !enabled {
		if err := mergeCacheEnvBlock(buildCacheEnvLines(false, "memory", "", "", "", 0, 0)); err != nil {
			fmt.Fprintf(os.Stderr, "Hata: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✓ CACHE_ENABLED=false olarak kaydedildi:", envFilePath())
		return
	}

	fmt.Print("Backend [memory/redis] [memory]: ")
	backend := strings.ToLower(strings.TrimSpace(readLine(reader)))
	if backend == "" {
		backend = "memory"
	}
	if backend != "redis" {
		backend = "memory"
	}

	fmt.Print("Sohbet listesi önbellek süresi (saniye) [20]: ")
	ttlList := readIntOrDefault(reader, 20)
	fmt.Print("Mesaj listesi önbellek süresi (saniye) [15]: ")
	ttlMsgs := readIntOrDefault(reader, 15)

	redisAddr, redisPw, redisDb := "127.0.0.1:6379", "", "0"
	if backend == "redis" {
		fmt.Printf("Redis adresi [%s]: ", redisAddr)
		if s := strings.TrimSpace(readLine(reader)); s != "" {
			redisAddr = s
		}
		fmt.Print("Redis şifresi (boş bırakılabilir): ")
		redisPw = readLine(reader)
		fmt.Printf("Redis DB numarası [%s]: ", redisDb)
		if s := strings.TrimSpace(readLine(reader)); s != "" {
			redisDb = s
		}
	}

	lines := buildCacheEnvLines(true, backend, redisAddr, redisPw, redisDb, ttlList, ttlMsgs)
	if err := mergeCacheEnvBlock(lines); err != nil {
		fmt.Fprintf(os.Stderr, "Hata: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("✓ Önbellek ayarları yazıldı:", envFilePath())
	fmt.Println("  Sunucuyu yeniden başlatın; Redis kullanıyorsanız REDIS_ADDR erişilebilir olmalı.")
}

func readIntOrDefault(reader *bufio.Reader, def int) int {
	s := strings.TrimSpace(readLine(reader))
	if s == "" {
		return def
	}
	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return def
	}
	return n
}

func buildCacheEnvLines(enabled bool, backend, redisAddr, redisPw, redisDb string, ttlList, ttlMsgs int) []string {
	if !enabled {
		return []string{
			cacheEnvMarkerBegin,
			"CACHE_ENABLED=false",
			cacheEnvMarkerEnd,
		}
	}
	lines := []string{
		cacheEnvMarkerBegin,
		"CACHE_ENABLED=true",
		fmt.Sprintf("CACHE_BACKEND=%s", backend),
		fmt.Sprintf("CACHE_TTL_CHAT_LIST_SEC=%d", ttlList),
		fmt.Sprintf("CACHE_TTL_CHAT_MESSAGES_SEC=%d", ttlMsgs),
	}
	if backend == "redis" {
		lines = append(lines,
			fmt.Sprintf("REDIS_ADDR=%s", redisAddr),
			fmt.Sprintf("REDIS_DB=%s", redisDb),
		)
		if strings.TrimSpace(redisPw) != "" {
			lines = append(lines, fmt.Sprintf("REDIS_PASSWORD=%s", redisPw))
		}
	}
	lines = append(lines, cacheEnvMarkerEnd)
	return lines
}

func envFilePath() string {
	if p := os.Getenv("ENV_FILE"); p != "" {
		return p
	}
	return ".env"
}

func mergeCacheEnvBlock(blockLines []string) error {
	path := envFilePath()
	data, err := os.ReadFile(path)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	content := string(data)
	newBlock := strings.Join(blockLines, "\n") + "\n"

	start := strings.Index(content, cacheEnvMarkerBegin)
	if start >= 0 {
		rel := strings.Index(content[start:], cacheEnvMarkerEnd)
		if rel >= 0 {
			end := start + rel + len(cacheEnvMarkerEnd)
			before := content[:start]
			after := content[end:]
			content = strings.TrimRight(before, "\n") + "\n\n" + newBlock + strings.TrimLeft(after, "\n")
			return os.WriteFile(path, []byte(content), 0o600)
		}
	}

	if strings.TrimSpace(content) != "" && !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	if strings.TrimSpace(content) != "" {
		content += "\n"
	}
	content += newBlock
	return os.WriteFile(path, []byte(content), 0o600)
}
