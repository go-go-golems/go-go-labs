import { useState, useEffect, useRef, useCallback, useMemo } from "react";

// ============================================================
// CORE RUNTIME — EventQueue, DB, VM Pool
// ============================================================

function createEventQueue() {
  const listeners = {};
  const log = [];

  return {
    on(type, handler) {
      if (!listeners[type]) listeners[type] = [];
      listeners[type].push(handler);
      return () => {
        listeners[type] = listeners[type].filter((h) => h !== handler);
      };
    },
    emit(type, payload) {
      const event = { type, payload, timestamp: Date.now(), id: `evt_${log.length}` };
      log.push(event);
      (listeners[type] || []).forEach((h) => {
        try { h(event); } catch (e) { console.error(`[Event:${type}]`, e); }
      });
      (listeners["*"] || []).forEach((h) => {
        try { h(event); } catch (e) { console.error(`[Event:*]`, e); }
      });
      return event;
    },
    getLog: () => [...log],
  };
}

function createDB(onChange) {
  const store = {};

  function notify() {
    onChange?.({ ...store });
  }

  return {
    put(collection, id, doc) {
      if (!store[collection]) store[collection] = {};
      store[collection][id] = { ...doc, _id: id, _updated: Date.now() };
      notify();
    },
    get(collection, id) {
      return store[collection]?.[id] || null;
    },
    query(collection, filter = {}) {
      const col = store[collection] || {};
      let results = Object.values(col);
      for (const [key, val] of Object.entries(filter)) {
        results = results.filter((doc) => doc[key] === val);
      }
      return results;
    },
    delete(collection, id) {
      if (store[collection]) {
        delete store[collection][id];
        notify();
      }
    },
    deleteAll(collection) {
      store[collection] = {};
      notify();
    },
    dump: () => JSON.parse(JSON.stringify(store)),
    _raw: store,
  };
}

function createVMPool(events, db) {
  let vmCounter = 0;
  const vms = {};

  function spawn(name, fn, env = {}) {
    const started = Date.now();
    const id = `vm_${vmCounter++}_${name}`;
    vms[id] = {
      id,
      name,
      status: "running",
      started,
      finished: null,
      durationMs: null,
      error: null,
    };
    events.emit("system.vm.started", { vm_id: id, name });

    setTimeout(() => {
      try {
        const result = fn({ env, db, events, spawn: (n, f, e) => spawn(n, f, e) });
        if (vms[id]) {
          vms[id].status = "done";
          vms[id].finished = Date.now();
          vms[id].durationMs = vms[id].finished - vms[id].started;
        }
        events.emit("system.vm.stopped", { vm_id: id, reason: "complete" });
        return result;
      } catch (err) {
        if (vms[id]) {
          vms[id].status = "error";
          vms[id].finished = Date.now();
          vms[id].durationMs = vms[id].finished - vms[id].started;
          vms[id].error = err.message;
        }
        events.emit("system.vm.stopped", { vm_id: id, reason: err.message });
      }
    }, 80 + Math.random() * 220);

    return id;
  }

  return { spawn, list: () => ({ ...vms }) };
}

// ============================================================
// PLUGIN PRESETS — each is a function(runtime) that sets up
// event handlers, DB seeds, and UI surface renderers
// ============================================================

const PLUGIN_PRESETS = {

  // ── Claims Ingestion ──────────────────────────────────
  "claims-ingestion": {
    id: "claims-ingestion",
    name: "Claims Ingestion",
    icon: "📝",
    description: "Accepts raw text, extracts atomic claims via VM workers",
    install({ events, db, vms }) {
      // Seed some sample submissions
      const sampleClaims = [
        { text: "Emergency rooms should use AI cameras for triage to expedite patient sorting.", author: "u_1", topic: "medical-ai", votes: 14, stance: "support" },
        { text: "AI-driven diagnostics could reduce misdiagnosis rates by 30% in rural clinics.", author: "u_2", topic: "medical-ai", votes: 22, stance: "support" },
        { text: "Patient privacy must be prioritized over efficiency gains from medical AI.", author: "u_3", topic: "medical-ai", votes: 31, stance: "concern" },
        { text: "Taiwan's AI model outperforms GPT-4 in certain medical contexts.", author: "u_4", topic: "medical-ai", votes: 8, stance: "support" },
        { text: "The quality of AI solutions released by the government is poor.", author: "u_5", topic: "ai-policy", votes: 19, stance: "concern" },
        { text: "AI should be deeply integrated into K-12 education within 5 years.", author: "u_6", topic: "ai-education", votes: 27, stance: "support" },
        { text: "Teachers need comprehensive AI literacy training before classroom deployment.", author: "u_7", topic: "ai-education", votes: 45, stance: "support" },
        { text: "Over-reliance on AI tutoring may weaken critical thinking skills.", author: "u_8", topic: "ai-education", votes: 33, stance: "concern" },
        { text: "AI-powered personalized learning has shown 40% improvement in test scores.", author: "u_9", topic: "ai-education", votes: 15, stance: "support" },
        { text: "Cultural identity preservation should be a requirement for all AI communication tools.", author: "u_10", topic: "society-culture", votes: 38, stance: "support" },
        { text: "Social media AI algorithms are eroding traditional community bonds.", author: "u_11", topic: "society-culture", votes: 29, stance: "concern" },
        { text: "AI translation tools are making minority languages more accessible.", author: "u_12", topic: "society-culture", votes: 12, stance: "support" },
        { text: "Japan's collaborative approach to AI development is a model Taiwan should follow.", author: "u_13", topic: "ai-policy", votes: 16, stance: "support" },
        { text: "AI career displacement will disproportionately affect workers over 50.", author: "u_14", topic: "ai-careers", votes: 41, stance: "concern" },
        { text: "New AI industries will create more jobs than they eliminate by 2030.", author: "u_15", topic: "ai-careers", votes: 20, stance: "support" },
        { text: "Telemedicine AI should be available in all rural townships.", author: "u_16", topic: "medical-ai", votes: 36, stance: "support" },
        { text: "AI grading systems remove human bias from educational assessment.", author: "u_17", topic: "ai-education", votes: 11, stance: "support" },
        { text: "Digital literacy gaps will worsen social inequality without intervention.", author: "u_18", topic: "society-culture", votes: 44, stance: "concern" },
      ];

      sampleClaims.forEach((c, i) => {
        db.put("claims", `c_${i}`, { ...c, status: "approved", created: Date.now() - (sampleClaims.length - i) * 60000 });
      });

      // Handle new submissions
      events.on("raw-text.submit", (e) => {
        const text = e.payload.value;
        if (!text?.trim()) return;
        const id = `c_${Date.now()}`;
        vms.spawn("claim-extractor", ({ db, events }) => {
          db.put("claims", id, {
            text: text.trim(),
            author: "u_you",
            topic: "uncategorized",
            votes: 0,
            stance: "support",
            status: "approved",
            created: Date.now(),
          });
          events.emit("claim.extracted", { claim_id: id, text: text.trim() });
        });
      });
    },
  },

  // ── Topic Clustering ──────────────────────────────────
  "clustering": {
    id: "clustering",
    name: "Topic Clustering",
    icon: "🔬",
    description: "Groups claims into topics using similarity analysis",
    install({ events, db, vms }) {
      const topicDefs = [
        { id: "medical-ai", label: "Medical AI", summary: "The development of medical AI in clinical diagnosis, patient care, telemedicine and emergency response. Citizens see both promise and privacy concerns in deploying AI across healthcare systems.", color: "#4F46E5", icon: "🏥" },
        { id: "ai-education", label: "AI and Education", summary: "In the next five years, AI technology in education is expected to reach a new milestone, with predictions that it will be deeply integrated into the learning environment, providing personalized learning pathways.", color: "#0891B2", icon: "🎓" },
        { id: "society-culture", label: "Society and Culture", summary: "In exploring the diverse contexts of contemporary Taiwanese society and culture, we find that interpersonal relationships, communication methods, social participation, and cultural identity are undergoing profound changes.", color: "#9333EA", icon: "🏛️" },
        { id: "ai-policy", label: "AI Policy & Governance", summary: "How should Taiwan approach AI regulation, international collaboration, and government-led AI initiatives? Citizens debate between following established models and forging an independent path.", color: "#DC2626", icon: "⚖️" },
        { id: "ai-careers", label: "AI Impact on Careers", summary: "The workforce transformation driven by AI adoption raises questions about displacement, retraining, and whether new industries will compensate for disrupted ones.", color: "#D97706", icon: "💼" },
      ];

      topicDefs.forEach((t) => {
        const claims = db.query("claims", { topic: t.id });
        db.put("topics", t.id, {
          ...t,
          claimCount: claims.length,
          peopleCount: new Set(claims.map((c) => c.author)).size + Math.floor(Math.random() * 80 + 20),
          subtopics: [],
        });
      });

      events.on("claim.extracted", () => {
        vms.spawn("re-cluster", ({ db }) => {
          const topicIds = ["medical-ai", "ai-education", "society-culture", "ai-policy", "ai-careers"];
          topicIds.forEach((tid) => {
            const claims = db.query("claims", { topic: tid });
            const topic = db.get("topics", tid);
            if (topic) {
              db.put("topics", tid, {
                ...topic,
                claimCount: claims.length,
                peopleCount: new Set(claims.map((c) => c.author)).size + Math.floor(Math.random() * 80 + 20),
              });
            }
          });
          // put uncategorized into a misc topic
          const uncategorized = db.query("claims", { topic: "uncategorized" });
          if (uncategorized.length > 0) {
            uncategorized.forEach((c) => {
              db.put("claims", c._id, { ...c, topic: topicIds[Math.floor(Math.random() * topicIds.length)] });
            });
          }
        });
      });
    },
  },

  // ── Voting & Consensus ────────────────────────────────
  "voting": {
    id: "voting",
    name: "Voting & Consensus",
    icon: "🗳️",
    description: "Tracks agree/disagree/nuance votes per claim",
    install({ events, db }) {
      events.on("vote.cast", (e) => {
        const { claim_id, stance } = e.payload;
        const claim = db.get("claims", claim_id);
        if (claim) {
          const delta = stance === "agree" ? 1 : stance === "disagree" ? -1 : 0;
          db.put("claims", claim_id, { ...claim, votes: (claim.votes || 0) + delta });
          events.emit("claim.votes_updated", { claim_id, newVotes: claim.votes + delta });
        }
      });
    },
  },

  // ── Moderation Agent ──────────────────────────────────
  "moderation": {
    id: "moderation",
    name: "Moderation Agent",
    icon: "🛡️",
    description: "Auto-classifies new claims against community guidelines",
    install({ events, db, vms }) {
      events.on("claim.extracted", (e) => {
        vms.spawn("mod-check", ({ db, events }) => {
          const claim = db.get("claims", e.payload.claim_id);
          if (!claim) return;
          // simulate classification — all pass in demo
          const label = claim.text.length < 10 ? "flagged" : "approved";
          db.put("moderation-log", `mod_${Date.now()}`, {
            claim_id: e.payload.claim_id,
            label,
            checked: Date.now(),
          });
          events.emit(`moderation.${label}`, { claim_id: e.payload.claim_id });
        });
      });
    },
  },

  // ── Summarizer Agent ──────────────────────────────────
  "summarizer": {
    id: "summarizer",
    name: "AI Summarizer",
    icon: "🤖",
    description: "Generates topic summaries by recursively spawning LLM agents",
    install({ events, db, vms }) {
      events.on("summary.request", (e) => {
        const topicId = e.payload.topic_id;
        vms.spawn("summarize-topic", ({ db, events, spawn }) => {
          const topic = db.get("topics", topicId);
          if (!topic) return;
          const claims = db.query("claims", { topic: topicId });
          // Simulate LLM summarization
          const summary = `Based on ${claims.length} citizen contributions, the community has expressed diverse perspectives on ${topic.label.toLowerCase()}. Key themes include both optimism about AI's transformative potential and caution regarding implementation risks, privacy, and equity.`;
          db.put("summaries", topicId, { text: summary, generated: Date.now(), claimCount: claims.length });
          events.emit("summary.generated", { topic_id: topicId });

          // Recursive: spawn a meta-summarizer if we have enough topic summaries
          const allSummaries = db.query("summaries", {});
          if (allSummaries.length >= 3) {
            spawn("meta-summarize", ({ db }) => {
              db.put("summaries", "meta", {
                text: `Across ${allSummaries.length} major topics, Taiwan's citizens show a consistent pattern: enthusiastic adoption of AI capabilities paired with deep concern for social equity and cultural preservation.`,
                generated: Date.now(),
                level: "meta",
              });
            });
          }
        });
      });
    },
  },
};

// ============================================================
// UI COMPONENTS
// ============================================================

const FONT_DISPLAY = "'Instrument Serif', Georgia, serif";
const FONT_BODY = "'DM Sans', system-ui, sans-serif";

const COLORS = {
  bg: "#0D0D0F",
  surface: "#17171B",
  surfaceHover: "#1E1E24",
  surfaceRaised: "#222228",
  border: "#2A2A32",
  borderLight: "#35353F",
  textPrimary: "#E8E6E3",
  textSecondary: "#9A9890",
  textMuted: "#6B6960",
  accent: "#E8A83E",
  accentDim: "#E8A83E33",
  accentHover: "#F0B84E",
  support: "#4ADE80",
  concern: "#F87171",
  nuance: "#A78BFA",
  topicColors: ["#4F46E5", "#0891B2", "#9333EA", "#DC2626", "#D97706"],
};

const SIMULATION_TEXTS = [
  "AI in public hospitals should be audited quarterly for bias.",
  "Community centers need AI literacy nights for older residents.",
  "Taiwan should publish transparent benchmark reports for state AI tools.",
  "Every school should keep a human review process for AI grading decisions.",
  "City governments should use AI to predict emergency response bottlenecks.",
];

function randomOf(items) {
  return items[Math.floor(Math.random() * items.length)];
}

// ── Claim Grid (the little squares) ─────────────────────
function ClaimGrid({ claims, onClaimClick, maxShow = 60 }) {
  const shown = claims.slice(0, maxShow);
  return (
    <div style={{ display: "flex", flexWrap: "wrap", gap: 3, margin: "12px 0" }}>
      {shown.map((c, i) => (
        <div
          key={c._id || i}
          onClick={() => onClaimClick?.(c)}
          title={c.text}
          style={{
            width: 14,
            height: 14,
            borderRadius: 3,
            cursor: "pointer",
            background: c.stance === "support" ? COLORS.support + "55" : c.stance === "concern" ? COLORS.concern + "55" : COLORS.nuance + "55",
            border: `1px solid ${c.stance === "support" ? COLORS.support + "88" : c.stance === "concern" ? COLORS.concern + "88" : COLORS.nuance + "88"}`,
            transition: "transform 0.15s, box-shadow 0.15s",
          }}
          onMouseEnter={(e) => {
            e.target.style.transform = "scale(1.6)";
            e.target.style.boxShadow = `0 0 8px ${c.stance === "support" ? COLORS.support : COLORS.concern}66`;
            e.target.style.zIndex = 10;
          }}
          onMouseLeave={(e) => {
            e.target.style.transform = "scale(1)";
            e.target.style.boxShadow = "none";
            e.target.style.zIndex = 1;
          }}
        />
      ))}
    </div>
  );
}

// ── Topic Card ──────────────────────────────────────────
function TopicCard({ topic, claims, summary, onExpand, onClaimClick }) {
  const [hovered, setHovered] = useState(false);
  return (
    <div
      onMouseEnter={() => setHovered(true)}
      onMouseLeave={() => setHovered(false)}
      style={{
        background: hovered ? COLORS.surfaceHover : COLORS.surface,
        border: `1px solid ${hovered ? COLORS.borderLight : COLORS.border}`,
        borderRadius: 12,
        padding: 24,
        marginBottom: 16,
        transition: "all 0.2s",
        borderLeft: `3px solid ${topic.color || COLORS.accent}`,
      }}
    >
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "flex-start", marginBottom: 8 }}>
        <h3 style={{ fontFamily: FONT_DISPLAY, fontSize: 24, color: COLORS.textPrimary, margin: 0, fontWeight: 400, fontStyle: "italic" }}>
          {topic.icon} {topic.label}
        </h3>
        <div style={{ display: "flex", gap: 12, fontSize: 13, color: COLORS.textMuted, fontFamily: FONT_BODY }}>
          <span>📋 {topic.claimCount} claims</span>
          <span>👥 {topic.peopleCount} people</span>
        </div>
      </div>

      <ClaimGrid claims={claims} onClaimClick={onClaimClick} />

      <p style={{ fontFamily: FONT_BODY, fontSize: 14, lineHeight: 1.65, color: COLORS.textSecondary, margin: "12px 0" }}>
        {summary?.text || topic.summary}
      </p>

      <div style={{ display: "flex", gap: 8, flexWrap: "wrap", marginTop: 12 }}>
        <button
          onClick={() => onExpand?.(topic.id || topic._id)}
          style={{
            fontFamily: FONT_BODY,
            fontSize: 13,
            fontWeight: 500,
            padding: "6px 16px",
            borderRadius: 20,
            border: `1px solid ${topic.color || COLORS.accent}`,
            color: topic.color || COLORS.accent,
            background: "transparent",
            cursor: "pointer",
            transition: "all 0.15s",
          }}
          onMouseEnter={(e) => {
            e.target.style.background = (topic.color || COLORS.accent) + "22";
          }}
          onMouseLeave={(e) => {
            e.target.style.background = "transparent";
          }}
        >
          Expand topic →
        </button>
      </div>
    </div>
  );
}

// ── Claim Detail Modal ──────────────────────────────────
function ClaimModal({ claim, onClose, onVote }) {
  if (!claim) return null;
  return (
    <div
      onClick={onClose}
      style={{
        position: "fixed", inset: 0, background: "#000000AA", zIndex: 1000,
        display: "flex", alignItems: "center", justifyContent: "center",
        backdropFilter: "blur(4px)",
      }}
    >
      <div
        onClick={(e) => e.stopPropagation()}
        style={{
          background: COLORS.surfaceRaised, borderRadius: 16, padding: 32,
          maxWidth: 520, width: "90%", border: `1px solid ${COLORS.border}`,
          boxShadow: "0 24px 80px #000000AA",
        }}
      >
        <div style={{ fontSize: 12, color: COLORS.textMuted, fontFamily: FONT_BODY, marginBottom: 8, textTransform: "uppercase", letterSpacing: 1 }}>
          Claim #{claim._id}
        </div>
        <p style={{ fontFamily: FONT_DISPLAY, fontSize: 22, color: COLORS.textPrimary, lineHeight: 1.5, margin: "0 0 8px", fontStyle: "italic", fontWeight: 400 }}>
          "{claim.text}"
        </p>
        <div style={{ fontSize: 13, color: COLORS.textMuted, fontFamily: FONT_BODY, marginBottom: 20 }}>
          by {claim.author} · {claim.votes} votes · stance: {claim.stance}
        </div>
        <div style={{ display: "flex", gap: 10 }}>
          {[
            { label: "👍 Agree", stance: "agree", color: COLORS.support },
            { label: "👎 Disagree", stance: "disagree", color: COLORS.concern },
            { label: "🤔 Nuanced", stance: "nuance", color: COLORS.nuance },
          ].map((btn) => (
            <button
              key={btn.stance}
              onClick={() => onVote(claim._id, btn.stance)}
              style={{
                fontFamily: FONT_BODY, fontSize: 14, fontWeight: 500,
                padding: "10px 20px", borderRadius: 10, border: `1px solid ${btn.color}55`,
                color: btn.color, background: btn.color + "15",
                cursor: "pointer", transition: "all 0.15s", flex: 1,
              }}
              onMouseEnter={(e) => e.target.style.background = btn.color + "30"}
              onMouseLeave={(e) => e.target.style.background = btn.color + "15"}
            >
              {btn.label}
            </button>
          ))}
        </div>
        <button
          onClick={onClose}
          style={{
            marginTop: 16, width: "100%", padding: "10px",
            fontFamily: FONT_BODY, fontSize: 13, color: COLORS.textMuted,
            background: "transparent", border: `1px solid ${COLORS.border}`,
            borderRadius: 8, cursor: "pointer",
          }}
        >
          Close
        </button>
      </div>
    </div>
  );
}

// ── Sidebar: Topic Outline ──────────────────────────────
function Sidebar({ topics, activeTopic, onSelectTopic, installedPlugins, compact = false }) {
  return (
    <div style={{
      width: compact ? "100%" : 260,
      minWidth: compact ? 0 : 260,
      background: COLORS.surface,
      borderRight: compact ? "none" : `1px solid ${COLORS.border}`,
      borderBottom: compact ? `1px solid ${COLORS.border}` : "none",
      padding: "20px 0",
      display: "flex",
      flexDirection: "column",
      height: compact ? "auto" : "100vh",
      overflowY: "auto",
    }}>
      <div style={{ padding: "0 20px 20px", borderBottom: `1px solid ${COLORS.border}` }}>
        <div style={{ fontFamily: FONT_DISPLAY, fontSize: 22, color: COLORS.textPrimary, fontStyle: "italic" }}>
          🏛️ Talk to the City
        </div>
        <div style={{ fontFamily: FONT_BODY, fontSize: 11, color: COLORS.textMuted, marginTop: 4, textTransform: "uppercase", letterSpacing: 1.5 }}>
          Plugin-Driven Deliberation
        </div>
      </div>

      <div style={{ padding: "16px 20px 8px", fontSize: 10, fontFamily: FONT_BODY, color: COLORS.textMuted, textTransform: "uppercase", letterSpacing: 2 }}>
        Outline
      </div>
      {topics.map((t) => (
        <div
          key={t._id}
          onClick={() => onSelectTopic(t._id)}
          style={{
            padding: "10px 20px",
            fontFamily: FONT_BODY, fontSize: 14,
            color: activeTopic === t._id ? COLORS.accent : COLORS.textSecondary,
            cursor: "pointer",
            borderLeft: activeTopic === t._id ? `2px solid ${COLORS.accent}` : "2px solid transparent",
            background: activeTopic === t._id ? COLORS.accentDim + "22" : "transparent",
            transition: "all 0.15s",
          }}
        >
          {t.icon} {t.label}
          <span style={{ float: "right", fontSize: 11, color: COLORS.textMuted }}>{t.peopleCount}</span>
        </div>
      ))}

      <div style={{ padding: "24px 20px 8px", fontSize: 10, fontFamily: FONT_BODY, color: COLORS.textMuted, textTransform: "uppercase", letterSpacing: 2 }}>
        Active Plugins
      </div>
      {installedPlugins.map((p) => (
        <div key={p.id} style={{ padding: "6px 20px", fontFamily: FONT_BODY, fontSize: 12, color: COLORS.textMuted }}>
          {p.icon} {p.name}
        </div>
      ))}
    </div>
  );
}

// ── Submission Bar ──────────────────────────────────────
function SubmissionBar({ onSubmit }) {
  const [text, setText] = useState("");
  const handleSubmit = () => {
    if (!text.trim()) return;
    onSubmit(text);
    setText("");
  };
  return (
    <div style={{
      background: COLORS.surface, border: `1px solid ${COLORS.border}`,
      borderRadius: 12, padding: 20, marginBottom: 20,
    }}>
      <div style={{ fontFamily: FONT_BODY, fontSize: 13, color: COLORS.textMuted, marginBottom: 10 }}>
        📣 Share your perspective with the community
      </div>
      <div style={{ display: "flex", gap: 10, flexWrap: "wrap" }}>
        <input
          value={text}
          onChange={(e) => setText(e.target.value)}
          onKeyDown={(e) => e.key === "Enter" && handleSubmit()}
          placeholder="What do you think about AI in Taiwan?"
          style={{
            flex: 1, padding: "12px 16px", borderRadius: 8,
            border: `1px solid ${COLORS.border}`, background: COLORS.bg,
            color: COLORS.textPrimary, fontFamily: FONT_BODY, fontSize: 14,
            outline: "none", minWidth: 220,
          }}
          onFocus={(e) => e.target.style.borderColor = COLORS.accent}
          onBlur={(e) => e.target.style.borderColor = COLORS.border}
        />
        <button
          onClick={handleSubmit}
          style={{
            padding: "12px 24px", borderRadius: 8, border: "none",
            background: COLORS.accent, color: COLORS.bg,
            fontFamily: FONT_BODY, fontSize: 14, fontWeight: 600,
            cursor: "pointer", transition: "background 0.15s",
          }}
          onMouseEnter={(e) => e.target.style.background = COLORS.accentHover}
          onMouseLeave={(e) => e.target.style.background = COLORS.accent}
        >
          Submit
        </button>
      </div>
    </div>
  );
}

// ── Event Log (debug panel) ─────────────────────────────
function EventLog({ events, vms, compact = false }) {
  const log = events.getLog().slice(-30).reverse();
  const vmList = Object.values(vms.list());
  return (
    <div style={{
      width: compact ? "100%" : 300,
      minWidth: compact ? 0 : 300,
      background: COLORS.surface,
      borderLeft: compact ? "none" : `1px solid ${COLORS.border}`,
      borderTop: compact ? `1px solid ${COLORS.border}` : "none",
      padding: 20,
      height: compact ? 300 : "100vh",
      overflowY: "auto",
    }}>
      <div style={{ fontSize: 10, fontFamily: FONT_BODY, color: COLORS.textMuted, textTransform: "uppercase", letterSpacing: 2, marginBottom: 12 }}>
        ⚡ Event Stream
      </div>
      {vmList.length > 0 && (
        <div style={{ marginBottom: 16 }}>
          <div style={{ fontSize: 10, fontFamily: FONT_BODY, color: COLORS.textMuted, textTransform: "uppercase", letterSpacing: 2, marginBottom: 6 }}>
            Active VMs: {vmList.filter(v => v.status === "running").length}
          </div>
          {vmList.slice(-5).reverse().map((vm) => (
            <div key={vm.id} style={{
              fontSize: 11, fontFamily: "monospace", color: vm.status === "running" ? COLORS.support : COLORS.textMuted,
              padding: "2px 0",
            }}>
              {vm.status === "running" ? "●" : "○"} {vm.name}
            </div>
          ))}
        </div>
      )}
      {log.map((e) => (
        <div key={e.id} style={{
          fontSize: 11, fontFamily: "monospace", padding: "4px 0",
          borderBottom: `1px solid ${COLORS.border}22`,
          color: e.type.startsWith("system.") ? COLORS.textMuted :
                 e.type.startsWith("claim.") ? COLORS.support :
                 e.type.startsWith("vote.") ? COLORS.nuance :
                 e.type.startsWith("moderation.") ? COLORS.concern :
                 COLORS.textSecondary,
        }}>
          <span style={{ color: COLORS.textMuted }}>
            {new Date(e.timestamp).toLocaleTimeString().slice(0, 8)}
          </span>{" "}
          {e.type}
        </div>
      ))}
      {log.length === 0 && (
        <div style={{ fontSize: 12, color: COLORS.textMuted, fontFamily: FONT_BODY, fontStyle: "italic" }}>
          Waiting for events...
        </div>
      )}
    </div>
  );
}

// ── Stats Bar ───────────────────────────────────────────
function StatsBar({ db }) {
  const claims = db.query("claims", {});
  const topics = db.query("topics", {});
  const totalPeople = topics.reduce((s, t) => s + (t.peopleCount || 0), 0);
  const stats = [
    { label: "Topics", value: topics.length, icon: "📊" },
    { label: "Claims", value: claims.length, icon: "📋" },
    { label: "Participants", value: totalPeople, icon: "👥" },
  ];
  return (
    <div style={{ display: "flex", gap: 16, marginBottom: 20 }}>
      {stats.map((s) => (
        <div key={s.label} style={{
          flex: 1, background: COLORS.surface, border: `1px solid ${COLORS.border}`,
          borderRadius: 10, padding: "14px 18px",
          display: "flex", alignItems: "center", gap: 12,
        }}>
          <span style={{ fontSize: 22 }}>{s.icon}</span>
          <div>
            <div style={{ fontFamily: FONT_DISPLAY, fontSize: 24, color: COLORS.textPrimary, fontWeight: 400 }}>{s.value}</div>
            <div style={{ fontFamily: FONT_BODY, fontSize: 11, color: COLORS.textMuted, textTransform: "uppercase", letterSpacing: 1 }}>{s.label}</div>
          </div>
        </div>
      ))}
    </div>
  );
}

// ── Runtime Panel ───────────────────────────────────────
function RuntimePanel({ events, vms, claims, autoSimulation, onStep, onBurst, onToggleAuto }) {
  const vmList = Object.values(vms.list()).sort((a, b) => b.started - a.started);
  const runningCount = vmList.filter((vm) => vm.status === "running").length;
  const doneCount = vmList.filter((vm) => vm.status === "done").length;
  const errorCount = vmList.filter((vm) => vm.status === "error").length;
  const lastEvent = events.getLog().at(-1);

  return (
    <div style={{
      marginBottom: 20, background: COLORS.surface,
      border: `1px solid ${COLORS.border}`, borderRadius: 12, padding: 20,
    }}>
      <div style={{ display: "flex", justifyContent: "space-between", gap: 12, flexWrap: "wrap", marginBottom: 14 }}>
        <div>
          <div style={{ fontFamily: FONT_BODY, fontSize: 11, color: COLORS.textMuted, textTransform: "uppercase", letterSpacing: 1.4 }}>
            Runtime
          </div>
          <div style={{ fontFamily: FONT_DISPLAY, fontSize: 28, color: COLORS.textPrimary, lineHeight: 1.2 }}>
            VM Simulation State
          </div>
          <div style={{ fontFamily: FONT_BODY, fontSize: 12, color: COLORS.textMuted, marginTop: 3 }}>
            Live status from the runtime, not mocked UI values.
          </div>
        </div>
        <div style={{ display: "flex", gap: 8, alignItems: "flex-start", flexWrap: "wrap" }}>
          <button
            onClick={onStep}
            style={{
              padding: "8px 14px", borderRadius: 8, border: `1px solid ${COLORS.border}`,
              color: COLORS.textSecondary, background: COLORS.bg, cursor: "pointer",
              fontFamily: FONT_BODY, fontSize: 12,
            }}
          >
            Simulate Step
          </button>
          <button
            onClick={onBurst}
            style={{
              padding: "8px 14px", borderRadius: 8, border: `1px solid ${COLORS.border}`,
              color: COLORS.textSecondary, background: COLORS.bg, cursor: "pointer",
              fontFamily: FONT_BODY, fontSize: 12,
            }}
          >
            Burst x5
          </button>
          <button
            onClick={onToggleAuto}
            style={{
              padding: "8px 14px", borderRadius: 8, border: `1px solid ${autoSimulation ? COLORS.accent : COLORS.border}`,
              color: autoSimulation ? COLORS.accent : COLORS.textSecondary, background: COLORS.bg, cursor: "pointer",
              fontFamily: FONT_BODY, fontSize: 12,
            }}
          >
            {autoSimulation ? "Stop Auto-Run" : "Start Auto-Run"}
          </button>
        </div>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(120px, 1fr))", gap: 10, marginBottom: 14 }}>
        {[
          { label: "Running", value: runningCount, color: COLORS.support },
          { label: "Completed", value: doneCount, color: COLORS.textPrimary },
          { label: "Errors", value: errorCount, color: COLORS.concern },
          { label: "Claims in DB", value: claims.length, color: COLORS.accent },
        ].map((item) => (
          <div key={item.label} style={{
            background: COLORS.bg, borderRadius: 10, border: `1px solid ${COLORS.border}`,
            padding: "10px 12px",
          }}>
            <div style={{ fontFamily: FONT_BODY, fontSize: 10, color: COLORS.textMuted, textTransform: "uppercase", letterSpacing: 1 }}>
              {item.label}
            </div>
            <div style={{ fontFamily: FONT_DISPLAY, fontSize: 24, color: item.color, lineHeight: 1.1 }}>
              {item.value}
            </div>
          </div>
        ))}
      </div>

      <div style={{ fontFamily: FONT_BODY, fontSize: 12, color: COLORS.textMuted, marginBottom: 10 }}>
        Last event: {lastEvent ? `${lastEvent.type} @ ${new Date(lastEvent.timestamp).toLocaleTimeString()}` : "none yet"}.
      </div>

      <div style={{ fontFamily: FONT_BODY, fontSize: 11, color: COLORS.textMuted, textTransform: "uppercase", letterSpacing: 1.4, marginBottom: 6 }}>
        Recent VMs
      </div>
      <div style={{ border: `1px solid ${COLORS.border}`, borderRadius: 8, overflow: "hidden" }}>
        {vmList.slice(0, 8).map((vm) => {
          const vmColor = vm.status === "running" ? COLORS.support : vm.status === "error" ? COLORS.concern : COLORS.textSecondary;
          const duration = vm.durationMs == null ? "running" : `${vm.durationMs}ms`;
          return (
            <div key={vm.id} style={{
              display: "grid", gridTemplateColumns: "1.5fr 1fr 1fr", gap: 8,
              padding: "8px 10px", background: COLORS.bg, borderBottom: `1px solid ${COLORS.border}`,
              fontFamily: "monospace", fontSize: 11, color: COLORS.textSecondary,
            }}>
              <span>{vm.id}</span>
              <span style={{ color: vmColor }}>{vm.name} ({vm.status})</span>
              <span style={{ textAlign: "right" }}>{duration}</span>
            </div>
          );
        })}
        {vmList.length === 0 && (
          <div style={{ padding: 10, fontFamily: FONT_BODY, fontSize: 12, color: COLORS.textMuted }}>
            No VM jobs started yet. Run a simulation step or submit a claim.
          </div>
        )}
      </div>
    </div>
  );
}

// ── Installed Plugins ───────────────────────────────────
function PluginInstaller({ installed }) {
  return (
    <div style={{ marginBottom: 20 }}>
      <div style={{
        marginTop: 10, background: COLORS.surface,
        border: `1px solid ${COLORS.border}`, borderRadius: 12,
        padding: 16, display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(220px, 1fr))", gap: 10,
      }}>
        {installed.map((plugin) => (
          <div
            key={plugin.id}
            style={{
              padding: 14, borderRadius: 10,
              border: `1px solid ${COLORS.accent + "44"}`,
              background: COLORS.accentDim + "11",
            }}
          >
            <div style={{ fontFamily: FONT_BODY, fontSize: 14, color: COLORS.textPrimary, marginBottom: 4 }}>
              {plugin.icon} {plugin.name}
              <span style={{ float: "right", fontSize: 11, color: COLORS.support }}>Live</span>
            </div>
            <div style={{ fontFamily: FONT_BODY, fontSize: 12, color: COLORS.textMuted, lineHeight: 1.5 }}>
              {plugin.description}
            </div>
          </div>
        ))}
      </div>
    </div>
  );
}


// ============================================================
// MAIN APP
// ============================================================

export default function App() {
  const [tick, setTick] = useState(0);
  const [activeTopic, setActiveTopic] = useState(null);
  const [selectedClaim, setSelectedClaim] = useState(null);
  const [showEventLog, setShowEventLog] = useState(true);
  const [autoSimulation, setAutoSimulation] = useState(false);
  const [isMobile, setIsMobile] = useState(() => {
    if (typeof window === "undefined") return false;
    return window.innerWidth < 1120;
  });

  // Stable runtime refs
  const runtimeRef = useRef(null);
  const installedRef = useRef([]);

  // Initialize runtime once
  if (!runtimeRef.current) {
    const events = createEventQueue();
    const db = createDB(() => setTick((t) => t + 1));
    const vms = createVMPool(events, db);
    runtimeRef.current = { events, db, vms };

    // Install all plugins by default
    const allPlugins = Object.values(PLUGIN_PRESETS);
    allPlugins.forEach((p) => p.install({ events, db, vms }));
    installedRef.current = allPlugins;
  }

  const { events, db, vms } = runtimeRef.current;

  // Re-render on events
  useEffect(() => {
    const unsub = events.on("*", () => setTick((t) => t + 1));
    return unsub;
  }, [events]);

  useEffect(() => {
    const onResize = () => setIsMobile(window.innerWidth < 1120);
    window.addEventListener("resize", onResize);
    return () => window.removeEventListener("resize", onResize);
  }, []);

  const topics = useMemo(() => db.query("topics", {}), [tick]);
  const allClaims = useMemo(() => db.query("claims", {}), [tick]);

  const handleSubmit = useCallback((text) => {
    events.emit("raw-text.submit", { value: text });
  }, [events]);

  const handleVote = useCallback((claimId, stance) => {
    events.emit("vote.cast", { claim_id: claimId, stance });
    setSelectedClaim(null);
  }, [events]);

  const handleExpandTopic = useCallback((topicId) => {
    events.emit("summary.request", { topic_id: topicId });
    setActiveTopic(topicId);
  }, [events]);

  const simulateStep = useCallback(() => {
    const actions = ["submit", "vote", "summary"];
    const action = randomOf(actions);

    if (action === "submit") {
      events.emit("raw-text.submit", { value: randomOf(SIMULATION_TEXTS) });
      return;
    }

    if (action === "vote" && allClaims.length > 0) {
      const claim = randomOf(allClaims);
      const stance = randomOf(["agree", "disagree", "nuance"]);
      events.emit("vote.cast", { claim_id: claim._id, stance });
      return;
    }

    if (action === "summary" && topics.length > 0) {
      const topic = randomOf(topics);
      events.emit("summary.request", { topic_id: topic._id });
      return;
    }

    events.emit("raw-text.submit", { value: randomOf(SIMULATION_TEXTS) });
  }, [events, allClaims, topics]);

  const simulateBurst = useCallback(() => {
    for (let i = 0; i < 5; i++) {
      setTimeout(() => simulateStep(), i * 140);
    }
  }, [simulateStep]);

  useEffect(() => {
    if (!autoSimulation) return undefined;
    const id = setInterval(() => simulateStep(), 1200);
    return () => clearInterval(id);
  }, [autoSimulation, simulateStep]);

  const filteredTopics = activeTopic ? topics.filter((t) => t._id === activeTopic) : topics;

  return (
    <>
      <style>{`
        *, *::before, *::after { box-sizing: border-box; margin: 0; padding: 0; }
        body { background: ${COLORS.bg}; margin: 0; }
        ::-webkit-scrollbar { width: 6px; }
        ::-webkit-scrollbar-track { background: transparent; }
        ::-webkit-scrollbar-thumb { background: ${COLORS.border}; border-radius: 3px; }
        ::selection { background: ${COLORS.accent}44; color: ${COLORS.textPrimary}; }
      `}</style>

      <div style={{ display: "flex", flexDirection: isMobile ? "column" : "row", minHeight: "100vh", fontFamily: FONT_BODY }}>
        <Sidebar
          topics={topics}
          activeTopic={activeTopic}
          onSelectTopic={(id) => setActiveTopic(activeTopic === id ? null : id)}
          installedPlugins={installedRef.current}
          compact={isMobile}
        />

        <div style={{ flex: 1, padding: isMobile ? 16 : 32, overflowY: "auto", height: isMobile ? "auto" : "100vh" }}>
          {/* Header */}
          <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 24 }}>
            <div>
              <h1 style={{ fontFamily: FONT_DISPLAY, fontSize: 36, color: COLORS.textPrimary, fontWeight: 400, fontStyle: "italic" }}>
                {activeTopic ? topics.find(t => t._id === activeTopic)?.label || "Topic" : "Community Deliberation"}
              </h1>
              <p style={{ fontFamily: FONT_BODY, fontSize: 14, color: COLORS.textMuted, marginTop: 4 }}>
                {activeTopic
                  ? `Viewing claims and perspectives for this topic`
                  : `All topics · powered by ${installedRef.current.length} plugins running in sandboxed VMs`
                }
              </p>
            </div>
            <div style={{ display: "flex", gap: 8 }}>
              {activeTopic && (
                <button
                  onClick={() => setActiveTopic(null)}
                  style={{
                    fontFamily: FONT_BODY, fontSize: 13, padding: "8px 16px",
                    borderRadius: 8, border: `1px solid ${COLORS.border}`,
                    color: COLORS.textSecondary, background: COLORS.surface,
                    cursor: "pointer",
                  }}
                >
                  ← All Topics
                </button>
              )}
              <button
                onClick={() => setShowEventLog(!showEventLog)}
                style={{
                  fontFamily: FONT_BODY, fontSize: 13, padding: "8px 16px",
                  borderRadius: 8, border: `1px solid ${COLORS.border}`,
                  color: showEventLog ? COLORS.accent : COLORS.textSecondary,
                  background: COLORS.surface,
                  cursor: "pointer",
                }}
              >
                ⚡ {showEventLog ? "Hide" : "Show"} Events
              </button>
            </div>
          </div>

          <StatsBar db={db} />
          <SubmissionBar onSubmit={handleSubmit} />
          <RuntimePanel
            events={events}
            vms={vms}
            claims={allClaims}
            autoSimulation={autoSimulation}
            onStep={simulateStep}
            onBurst={simulateBurst}
            onToggleAuto={() => setAutoSimulation((prev) => !prev)}
          />
          <PluginInstaller
            installed={installedRef.current}
          />

          {/* Topic Cards — rendered from DB state */}
          {filteredTopics.map((topic) => {
            const claims = allClaims.filter((c) => c.topic === topic._id);
            const summary = db.get("summaries", topic._id);
            return (
              <TopicCard
                key={topic._id}
                topic={topic}
                claims={claims}
                summary={summary}
                onExpand={handleExpandTopic}
                onClaimClick={setSelectedClaim}
              />
            );
          })}

          {/* Expanded topic: show individual claims */}
          {activeTopic && (
            <div style={{ marginTop: 8 }}>
              <div style={{ fontSize: 10, fontFamily: FONT_BODY, color: COLORS.textMuted, textTransform: "uppercase", letterSpacing: 2, marginBottom: 12 }}>
                Individual Claims
              </div>
              {allClaims
                .filter((c) => c.topic === activeTopic)
                .sort((a, b) => (b.votes || 0) - (a.votes || 0))
                .map((claim) => (
                  <div
                    key={claim._id}
                    onClick={() => setSelectedClaim(claim)}
                    style={{
                      padding: "14px 18px", marginBottom: 8,
                      background: COLORS.surface, border: `1px solid ${COLORS.border}`,
                      borderRadius: 10, cursor: "pointer",
                      display: "flex", justifyContent: "space-between", alignItems: "center",
                      transition: "border-color 0.15s",
                    }}
                    onMouseEnter={(e) => e.currentTarget.style.borderColor = COLORS.borderLight}
                    onMouseLeave={(e) => e.currentTarget.style.borderColor = COLORS.border}
                  >
                    <div style={{ flex: 1 }}>
                      <span style={{
                        display: "inline-block", width: 8, height: 8, borderRadius: 4, marginRight: 10,
                        background: claim.stance === "support" ? COLORS.support : claim.stance === "concern" ? COLORS.concern : COLORS.nuance,
                      }} />
                      <span style={{ fontFamily: FONT_BODY, fontSize: 14, color: COLORS.textPrimary }}>
                        {claim.text}
                      </span>
                    </div>
                    <div style={{ fontSize: 12, color: COLORS.textMuted, fontFamily: FONT_BODY, whiteSpace: "nowrap", marginLeft: 16 }}>
                      {claim.votes > 0 ? "👍" : "👎"} {claim.votes}
                    </div>
                  </div>
                ))}
            </div>
          )}
        </div>

        {/* Event log sidebar */}
        {showEventLog && <EventLog events={events} vms={vms} compact={isMobile} />}
      </div>

      {/* Claim detail modal */}
      <ClaimModal claim={selectedClaim} onClose={() => setSelectedClaim(null)} onVote={handleVote} />
    </>
  );
}
