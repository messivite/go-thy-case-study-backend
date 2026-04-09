package landing

import "net/http"

func Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(page))
	}
}

const page = `<!DOCTYPE html>
<html lang="tr">
<head>
<meta charset="utf-8"/>
<meta name="viewport" content="width=device-width,initial-scale=1"/>
<title>Mustafa AKSOY - THY Case Study Çalışması</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
@import url('https://fonts.googleapis.com/css2?family=Inter:wght@300;400;600;700&display=swap');
body{
  font-family:'Inter',system-ui,sans-serif;
  min-height:100vh;
  display:flex;
  align-items:center;
  justify-content:center;
  background:#0a0a0a;
  overflow:hidden;
  position:relative;
}

.bg-grid{
  position:fixed;inset:0;
  background-image:
    linear-gradient(rgba(255,255,255,.03) 1px,transparent 1px),
    linear-gradient(90deg,rgba(255,255,255,.03) 1px,transparent 1px);
  background-size:60px 60px;
  animation:gridMove 20s linear infinite;
}
@keyframes gridMove{
  0%{transform:translate(0,0)}
  100%{transform:translate(60px,60px)}
}

.glow{
  position:fixed;
  width:600px;height:600px;
  border-radius:50%;
  filter:blur(120px);
  opacity:.15;
  animation:float 8s ease-in-out infinite;
}
.glow-1{background:#c41e3a;top:-200px;left:-100px;animation-delay:0s}
.glow-2{background:#1e3a5f;bottom:-200px;right:-100px;animation-delay:4s}
@keyframes float{
  0%,100%{transform:translate(0,0) scale(1)}
  50%{transform:translate(30px,20px) scale(1.1)}
}

.container{
  position:relative;z-index:1;
  text-align:center;
  padding:2rem;
}

.flag-badge{
  position:fixed;
  top:22px;
  right:22px;
  z-index:2;
  width:84px;
  height:56px;
  border-radius:10px;
  overflow:hidden;
  box-shadow:0 10px 30px rgba(0,0,0,.35), 0 0 0 1px rgba(255,255,255,.12) inset;
  animation:flagFloat 4s ease-in-out infinite;
}

.flag{
  width:100%;
  height:100%;
  display:block;
}

@keyframes flagFloat{
  0%,100%{transform:translateY(0)}
  50%{transform:translateY(-6px)}
}

.thy-badge{
  display:inline-flex;
  align-items:center;
  gap:.5rem;
  padding:.5rem 1.25rem;
  background:rgba(196,30,58,.1);
  border:1px solid rgba(196,30,58,.3);
  border-radius:100px;
  color:#e85d75;
  font-size:.75rem;
  font-weight:600;
  letter-spacing:.1em;
  text-transform:uppercase;
  opacity:0;
  animation:fadeUp .8s ease forwards;
  animation-delay:.2s;
}
.thy-badge svg{width:16px;height:16px}

.thy-logo-wrap{
  margin:0 auto 1rem;
  width:250px;
  height:110px;
  border-radius:14px;
  display:grid;
  place-items:center;
  background:radial-gradient(circle at 30% 30%,rgba(255,255,255,.15),rgba(255,255,255,.03));
  border:1px solid rgba(255,255,255,.15);
  box-shadow:0 10px 28px rgba(0,0,0,.35),0 0 0 1px rgba(196,30,58,.15) inset;
  opacity:0;
  transform:translateY(12px) scale(.95);
  animation:fadeUp .8s ease forwards,logoFloat 4.2s ease-in-out 1.2s infinite;
}
.thy-logo{
  width:210px;
  height:auto;
  filter:drop-shadow(0 4px 10px rgba(0,0,0,.25));
}
@keyframes logoFloat{
  0%,100%{transform:translateY(0) scale(1)}
  50%{transform:translateY(-6px) scale(1.03)}
}

h1{
  font-size:clamp(2.5rem,8vw,5rem);
  font-weight:700;
  color:#fff;
  margin-top:1.5rem;
  line-height:1.1;
  opacity:0;
  animation:fadeUp .8s ease forwards;
  animation-delay:.5s;
}
h1 span{
  background:linear-gradient(135deg,#c41e3a 0%,#e85d75 50%,#c41e3a 100%);
  background-size:200% auto;
  -webkit-background-clip:text;
  -webkit-text-fill-color:transparent;
  background-clip:text;
  animation:shimmer 3s linear infinite;
}
@keyframes shimmer{
  0%{background-position:0% center}
  100%{background-position:200% center}
}

.subtitle{
  font-size:1.1rem;
  color:rgba(255,255,255,.5);
  margin-top:1rem;
  font-weight:300;
  letter-spacing:.02em;
  opacity:0;
  animation:fadeUp .8s ease forwards;
  animation-delay:.8s;
}

.divider{
  width:60px;height:2px;
  background:linear-gradient(90deg,transparent,#c41e3a,transparent);
  margin:2rem auto;
  opacity:0;
  animation:fadeUp .8s ease forwards;
  animation-delay:1s;
}

.author{
  font-size:1.4rem;
  font-weight:600;
  color:#fff;
  opacity:0;
  animation:fadeUp .8s ease forwards;
  animation-delay:1.2s;
}

.tech-stack{
  display:flex;
  gap:.5rem;
  justify-content:center;
  flex-wrap:wrap;
  margin-top:1.5rem;
  opacity:0;
  animation:fadeUp .8s ease forwards;
  animation-delay:1.5s;
}
.tech-stack .chip{
  padding:.35rem .75rem;
  background:rgba(255,255,255,.05);
  border:1px solid rgba(255,255,255,.08);
  border-radius:6px;
  color:rgba(255,255,255,.6);
  font-size:.75rem;
  font-weight:400;
  transition:all .3s ease;
}
.chip:hover{
  background:rgba(196,30,58,.1);
  border-color:rgba(196,30,58,.3);
  color:#e85d75;
  transform:translateY(-2px);
}

.status{
  margin-top:2.5rem;
  display:flex;
  align-items:center;
  justify-content:center;
  gap:.5rem;
  color:rgba(255,255,255,.4);
  font-size:.8rem;
  opacity:0;
  animation:fadeUp .8s ease forwards;
  animation-delay:1.8s;
}
.status-dot{
  width:8px;height:8px;
  border-radius:50%;
  background:#22c55e;
  box-shadow:0 0 12px rgba(34,197,94,.5);
  animation:pulse 2s ease-in-out infinite;
}
@keyframes pulse{
  0%,100%{opacity:1;transform:scale(1)}
  50%{opacity:.5;transform:scale(.8)}
}

@keyframes fadeUp{
  from{opacity:0;transform:translateY(20px)}
  to{opacity:1;transform:translateY(0)}
}

.particles{position:fixed;inset:0;pointer-events:none;z-index:0}
.particle{
  position:absolute;
  width:2px;height:2px;
  background:rgba(196,30,58,.4);
  border-radius:50%;
  animation:rise linear infinite;
}
@keyframes rise{
  0%{opacity:0;transform:translateY(100vh) scale(0)}
  10%{opacity:1}
  90%{opacity:1}
  100%{opacity:0;transform:translateY(-10vh) scale(1)}
}
</style>
</head>
<body>
<div class="bg-grid"></div>
<div class="glow glow-1"></div>
<div class="glow glow-2"></div>
<div class="particles" id="particles"></div>

<div class="flag-badge" aria-label="Turkiye Cumhuriyeti Bayragi">
  <svg class="flag" viewBox="0 0 300 200" xmlns="http://www.w3.org/2000/svg" role="img">
    <rect width="300" height="200" fill="#E30A17"/>
    <circle cx="112" cy="100" r="46" fill="#fff"/>
    <circle cx="126" cy="100" r="36" fill="#E30A17"/>
    <polygon fill="#fff" points="188,72 194,90 213,90 198,101 203,120 188,109 173,120 178,101 163,90 182,90"/>
  </svg>
</div>

<div class="container">
  <div class="thy-logo-wrap" aria-hidden="true">
    <img class="thy-logo" src="https://upload.wikimedia.org/wikipedia/commons/0/00/Turkish_Airlines_logo_2019_compact.svg" alt="Turkish Airlines Logo" loading="eager" decoding="async"/>
  </div>

  <div class="thy-badge">
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2">
      <path d="M12 2L2 7l10 5 10-5-10-5z"/>
      <path d="M2 17l10 5 10-5"/>
      <path d="M2 12l10 5 10-5"/>
    </svg>
    CASE STUDY
  </div>

  <h1><span>THY Case Study</span></h1>
  <p class="subtitle">LLM Chat Backend — Go + Supabase</p>

  <div class="divider"></div>

  <p class="author">Mustafa Aksoy</p>

  <div class="tech-stack">
    <span class="chip">Go 1.25</span>
    <span class="chip">Supabase</span>
    <span class="chip">OpenAI</span>
    <span class="chip">Gemini</span>
    <span class="chip">JWT / RBAC</span>
    <span class="chip">SSE Stream</span>
    <span class="chip">Token Quota</span>
    <span class="chip">OpenTelemetry</span>
  </div>

  <div class="status">
    <span class="status-dot"></span>
    API is running
  </div>
</div>

<script>
(function(){
  var c=document.getElementById('particles');
  for(var i=0;i<20;i++){
    var p=document.createElement('div');
    p.className='particle';
    p.style.left=Math.random()*100+'%';
    p.style.animationDuration=(8+Math.random()*12)+'s';
    p.style.animationDelay=Math.random()*8+'s';
    p.style.width=p.style.height=(1+Math.random()*2)+'px';
    c.appendChild(p);
  }
})();
</script>
</body>
</html>
`
