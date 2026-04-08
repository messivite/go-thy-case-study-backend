import { createClient } from "https://esm.sh/@supabase/supabase-js@2";

const PUSH_TOKEN_RE = /^ExponentPushToken\[.+\]$/;
const LANGUAGES = new Set(["tr", "en", "fr", "de", "es", "ar"]);

const corsHeaders: Record<string, string> = {
  "Access-Control-Allow-Origin": "*",
  "Access-Control-Allow-Headers":
    "authorization, x-client-info, apikey, content-type",
  "Access-Control-Allow-Methods": "POST, OPTIONS",
};

function jsonResponse(
  status: number,
  body: Record<string, unknown>,
): Response {
  return new Response(JSON.stringify(body), {
    status,
    headers: { ...corsHeaders, "Content-Type": "application/json" },
  });
}

Deno.serve(async (req) => {
  if (req.method === "OPTIONS") {
    return new Response(null, { status: 204, headers: corsHeaders });
  }

  if (req.method !== "POST") {
    return jsonResponse(405, { error: "method not allowed" });
  }

  const supabaseUrl = Deno.env.get("SUPABASE_URL");
  const anonKey = Deno.env.get("SUPABASE_ANON_KEY");
  const serviceKey = Deno.env.get("SUPABASE_SERVICE_ROLE_KEY");

  if (!supabaseUrl || !anonKey || !serviceKey) {
    return jsonResponse(500, { error: "server misconfiguration" });
  }

  const authHeader = req.headers.get("Authorization");
  if (!authHeader?.startsWith("Bearer ")) {
    return jsonResponse(401, { error: "unauthorized" });
  }
  const token = authHeader.slice("Bearer ".length).trim();
  if (!token) {
    return jsonResponse(401, { error: "unauthorized" });
  }

  const supabaseAuth = createClient(supabaseUrl, anonKey);
  const { data: userData, error: userErr } = await supabaseAuth.auth.getUser(
    token,
  );

  if (userErr || !userData?.user?.id) {
    return jsonResponse(401, { error: "unauthorized" });
  }

  const userId = userData.user.id;

  let payload: unknown;
  try {
    payload = await req.json();
  } catch {
    return jsonResponse(400, { error: "invalid json body" });
  }

  if (
    typeof payload !== "object" ||
    payload === null ||
    Array.isArray(payload)
  ) {
    return jsonResponse(400, { error: "invalid json body" });
  }

  const body = payload as Record<string, unknown>;
  const pushToken = body.push_token;
  const language = body.language;

  if (typeof pushToken !== "string" || !PUSH_TOKEN_RE.test(pushToken)) {
    return jsonResponse(400, { error: "invalid push_token format" });
  }

  if (typeof language !== "string" || !LANGUAGES.has(language)) {
    return jsonResponse(400, { error: "invalid language" });
  }

  const supabaseAdmin = createClient(supabaseUrl, serviceKey);
  // Upsert: satır yoksa oluşturur; varsa günceller. Sadece UPDATE bazen 0 satır
  // döndürür (RLS/yanlış API key/postgrest); upsert daha tutarlıdır.
  const { data, error } = await supabaseAdmin
    .from("users")
    .upsert(
      {
        id: userId,
        push_token: pushToken,
        language,
        push_token_updated_at: new Date().toISOString(),
      },
      { onConflict: "id" },
    )
    .select("id");

  if (error) {
    console.error("[register-push-token]", error);
    return jsonResponse(500, {
      error: "internal error",
      detail: error.message,
      code: error.code,
    });
  }

  if (!data?.length) {
    return jsonResponse(404, {
      error: "user not found",
      detail: "upsert returned no row; check RLS and Edge secret SUPABASE_SERVICE_ROLE_KEY (must be service role / sb_secret, not anon)",
    });
  }

  return jsonResponse(200, { success: true });
});
