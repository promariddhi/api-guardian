import { useState, useEffect, useRef } from "react";
import {
  AreaChart,
  Area,
  LineChart,
  Line,
  BarChart,
  Bar,
  PieChart,
  Pie,
  Cell,
  XAxis,
  YAxis,
  CartesianGrid,
  Tooltip,
  ResponsiveContainer,
  Legend,
} from "recharts";

// ─── DESIGN TOKENS ───────────────────────────────────────────────────────────
const C = {
  bg: "#0a0b0d",
  surface: "#111318",
  card: "#161920",
  border: "#252932",
  border2: "#1c2030",
  text: "#e2e8f0",
  muted: "#64748b",
  dim: "#3a4155",
  accent: "#3b82f6",
  accentDim: "#1d3f6f",
  green: "#10b981",
  greenDim: "#0d2e22",
  yellow: "#f59e0b",
  yellowDim: "#2d2006",
  red: "#ef4444",
  redDim: "#2d0f0f",
  purple: "#8b5cf6",
  purpleDim: "#2d1f4a",
};

// ─── MOCK DATA ────────────────────────────────────────────────────────────────
const generateTrafficHistory = () =>
  Array.from({ length: 40 }, (_, i) => ({
    t: i,
    rps: Math.floor(800 + Math.random() * 600 + Math.sin(i / 5) * 200),
    errors: Math.floor(Math.random() * 30 + 5),
    p99: Math.floor(110 + Math.random() * 80 + Math.sin(i / 3) * 20),
  }));

// Services that back-end routing dispatches to; each has its own instance pool
const SERVICES_DATA = [
  {
    id: "user-svc",
    name: "user-service",
    shortName: "user",
    color: "#3b82f6",
    routes: "/api/v1/users · /api/v1/auth",
    redis: true,
  },
  {
    id: "order-svc",
    name: "order-service",
    shortName: "order",
    color: "#8b5cf6",
    routes: "/api/v1/orders",
    redis: true,
  },
  {
    id: "product-svc",
    name: "product-service",
    shortName: "product",
    color: "#10b981",
    routes: "/api/v1/products · /api/v1/search",
    redis: false,
  },
];

// 3 failures → instance marked DEAD and skipped; background prober retries until /health returns 200
const FAIL_THRESHOLD = 3;

const BACKENDS_INIT = [
  // user-service instances
  {
    id: "us1",
    service: "user-svc",
    name: "user-1",
    host: "10.0.1.10:8080",
    alive: true,
    dead: false,
    probing: false,
    draining: false,
    rps: 312,
    p50: 28,
    p99: 89,
    failures: 0,
    weight: 100,
  },
  {
    id: "us2",
    service: "user-svc",
    name: "user-2",
    host: "10.0.1.11:8080",
    alive: false,
    dead: true,
    probing: true,
    draining: false,
    rps: 0,
    p50: null,
    p99: null,
    failures: 3,
    weight: 0,
  },
  // order-service instances
  {
    id: "os1",
    service: "order-svc",
    name: "order-1",
    host: "10.0.2.10:8080",
    alive: true,
    dead: false,
    probing: false,
    draining: false,
    rps: 142,
    p50: 61,
    p99: 198,
    failures: 1,
    weight: 55,
  },
  {
    id: "os2",
    service: "order-svc",
    name: "order-2",
    host: "10.0.2.11:8080",
    alive: true,
    dead: false,
    probing: false,
    draining: true,
    rps: 98,
    p50: 65,
    p99: 210,
    failures: 0,
    weight: 45,
  },
  // product-service instances
  {
    id: "ps1",
    service: "product-svc",
    name: "product-1",
    host: "10.0.3.10:8080",
    alive: true,
    dead: false,
    probing: false,
    draining: false,
    rps: 224,
    p50: 19,
    p99: 62,
    failures: 2,
    weight: 50,
  },
  {
    id: "ps2",
    service: "product-svc",
    name: "product-2",
    host: "10.0.3.11:8080",
    alive: true,
    dead: false,
    probing: false,
    draining: false,
    rps: 198,
    p50: 17,
    p99: 58,
    failures: 0,
    weight: 50,
  },
];

const ROUTES_DATA = [
  {
    route: "/api/v1/users",
    method: "GET",
    rps: 234,
    p50: 22,
    p99: 78,
    errors: 2,
  },
  {
    route: "/api/v1/auth/token",
    method: "POST",
    rps: 198,
    p50: 41,
    p99: 142,
    errors: 18,
  },
  {
    route: "/api/v1/products",
    method: "GET",
    rps: 187,
    p50: 18,
    p99: 55,
    errors: 0,
  },
  {
    route: "/api/v1/orders",
    method: "POST",
    rps: 142,
    p50: 68,
    p99: 231,
    errors: 4,
  },
  {
    route: "/api/v1/webhooks",
    method: "POST",
    rps: 98,
    p50: 12,
    p99: 44,
    errors: 1,
  },
  {
    route: "/api/v1/search",
    method: "GET",
    rps: 87,
    p50: 89,
    p99: 312,
    errors: 3,
  },
  { route: "/health", method: "GET", rps: 60, p50: 2, p99: 8, errors: 0 },
  {
    route: "/api/v1/files/upload",
    method: "POST",
    rps: 24,
    p50: 421,
    p99: 1840,
    errors: 6,
  },
];

const EVENTS_INIT = [
  {
    id: 1,
    ts: Date.now() - 3000,
    level: "error",
    msg: "user-2 failure 3/3 → circuit OPEN, marked DEAD",
    tag: "CIRCUIT",
  },
  {
    id: 2,
    ts: Date.now() - 9000,
    level: "info",
    msg: "PROBE user-2 (10.0.1.11) GET /health → timeout, still DEAD",
    tag: "PROBE",
  },
  {
    id: 3,
    ts: Date.now() - 15000,
    level: "warn",
    msg: "GET /api/v1/users → user-2 DEAD, retried on user-1 → 200 OK",
    tag: "RETRY",
  },
  {
    id: 4,
    ts: Date.now() - 21000,
    level: "warn",
    msg: "product-1 failure 2/3 — 1 more failure = DEAD",
    tag: "CIRCUIT",
  },
  {
    id: 5,
    ts: Date.now() - 28000,
    level: "error",
    msg: "user-2 timeout 5012ms — GET /api/v1/users/882",
    tag: "TIMEOUT",
  },
  {
    id: 6,
    ts: Date.now() - 37000,
    level: "warn",
    msg: "Rate limit exceeded — 192.168.4.22",
    tag: "RATE_LIMIT",
  },
  {
    id: 7,
    ts: Date.now() - 48000,
    level: "info",
    msg: "order-2 entering drain mode — graceful shutdown",
    tag: "BACKEND",
  },
  {
    id: 8,
    ts: Date.now() - 63000,
    level: "info",
    msg: "Gateway reload — config v2.4.1 applied",
    tag: "GATEWAY",
  },
];

const REQUESTS_INIT = [
  {
    id: "r1",
    ts: Date.now() - 400,
    method: "POST",
    route: "/api/v1/auth/token",
    status: 401,
    backend: "—",
    latency: 12,
    error: "JWT invalid",
  },
  {
    id: "r2",
    ts: Date.now() - 800,
    method: "GET",
    route: "/api/v1/users/882",
    status: 200,
    backend: "user-1",
    latency: 24,
    error: null,
  },
  {
    id: "r3",
    ts: Date.now() - 1200,
    method: "POST",
    route: "/api/v1/orders",
    status: 201,
    backend: "order-1",
    latency: 78,
    error: null,
  },
  {
    id: "r4",
    ts: Date.now() - 1600,
    method: "GET",
    route: "/api/v1/products",
    status: 200,
    backend: "product-1",
    latency: 19,
    error: null,
  },
  {
    id: "r5",
    ts: Date.now() - 2100,
    method: "GET",
    route: "/api/v1/users/441",
    status: 503,
    backend: "user-2",
    latency: 5012,
    error: "Instance dead",
  },
  {
    id: "r6",
    ts: Date.now() - 2600,
    method: "GET",
    route: "/api/v1/users/441",
    status: 200,
    backend: "user-1",
    latency: 27,
    error: null,
    retried: true,
  },
  {
    id: "r7",
    ts: Date.now() - 3100,
    method: "GET",
    route: "/api/v1/search",
    status: 200,
    backend: "product-2",
    latency: 58,
    error: null,
  },
  {
    id: "r8",
    ts: Date.now() - 3600,
    method: "POST",
    route: "/api/v1/orders",
    status: 201,
    backend: "order-2",
    latency: 91,
    error: null,
  },
];

// Per-service load distribution; instances breakdown within each service
const LB_DIST = [
  {
    name: "user-svc",
    label: "user",
    value: 36,
    color: "#3b82f6",
    instances: [
      { name: "user-1", pct: 100, dead: false, draining: false },
      { name: "user-2", pct: 0, dead: true, draining: false },
    ],
  },
  {
    name: "order-svc",
    label: "order",
    value: 27,
    color: "#8b5cf6",
    instances: [
      { name: "order-1", pct: 59, dead: false, draining: false },
      { name: "order-2", pct: 41, dead: false, draining: true },
    ],
  },
  {
    name: "product-svc",
    label: "product",
    value: 37,
    color: "#10b981",
    instances: [
      { name: "product-1", pct: 53, dead: false, draining: false },
      { name: "product-2", pct: 47, dead: false, draining: false },
    ],
  },
];

const TOP_OFFENDERS_IP = [
  { key: "192.168.4.22", hits: 482, blocked: 201 },
  { key: "10.3.2.101", hits: 318, blocked: 97 },
  { key: "203.0.113.55", hits: 201, blocked: 44 },
  { key: "198.51.100.8", hits: 98, blocked: 12 },
];

const TOP_OFFENDERS_USER = [
  { key: "usr_a3f92b", hits: 610, blocked: 284 },
  { key: "usr_7dc481", hits: 374, blocked: 112 },
  { key: "usr_e1290f", hits: 229, blocked: 58 },
  { key: "usr_b84c23", hits: 87, blocked: 9 },
];

// ─── HELPERS ──────────────────────────────────────────────────────────────────
const fmt = {
  ts: (ms) => {
    const s = Math.floor((Date.now() - ms) / 1000);
    if (s < 60) return `${s}s ago`;
    if (s < 3600) return `${Math.floor(s / 60)}m ago`;
    return `${Math.floor(s / 3600)}h ago`;
  },
  latency: (ms) => (ms >= 1000 ? `${(ms / 1000).toFixed(1)}s` : `${ms}ms`),
};

const statusColor = (code) => {
  if (code < 300) return C.green;
  if (code < 400) return C.accent;
  if (code < 500) return C.yellow;
  return C.red;
};

const methodColor = (m) =>
  ({
    GET: "#3b82f6",
    POST: "#8b5cf6",
    PUT: "#f59e0b",
    DELETE: "#ef4444",
    PATCH: "#06b6d4",
  })[m] || C.muted;

// ─── MICRO COMPONENTS ────────────────────────────────────────────────────────

const Badge = ({ children, color = C.muted }) => (
  <span
    style={{
      fontSize: 10,
      fontFamily: "JetBrains Mono, monospace",
      fontWeight: 600,
      color,
      border: `1px solid ${color}44`,
      borderRadius: 3,
      padding: "2px 5px",
      letterSpacing: "0.04em",
      textTransform: "uppercase",
      whiteSpace: "nowrap",
      flexShrink: 0,
      lineHeight: 1,
    }}
  >
    {children}
  </span>
);

const StatusDot = ({ healthy, size = 8, pulse = false }) => (
  <span
    style={{
      position: "relative",
      display: "inline-block",
      width: size,
      height: size,
    }}
  >
    {pulse && healthy && (
      <span
        style={{
          position: "absolute",
          inset: 0,
          borderRadius: "50%",
          background: C.green,
          opacity: 0.4,
          animation: "ping 1.5s cubic-bezier(0,0,0.2,1) infinite",
        }}
      />
    )}
    <span
      style={{
        display: "block",
        width: size,
        height: size,
        borderRadius: "50%",
        background: healthy ? C.green : C.red,
      }}
    />
  </span>
);

const Mono = ({ children, size = 13, color = C.text, style = {} }) => (
  <span
    style={{
      fontFamily: "JetBrains Mono, monospace",
      fontSize: size,
      color,
      ...style,
    }}
  >
    {children}
  </span>
);

const Card = ({ children, style = {}, className = "" }) => (
  <div
    style={{
      background: C.card,
      border: `1px solid ${C.border}`,
      borderRadius: 8,
      ...style,
    }}
  >
    {children}
  </div>
);

const SectionLabel = ({ children }) => (
  <div
    style={{
      fontSize: 10,
      fontWeight: 700,
      color: C.muted,
      letterSpacing: "0.12em",
      textTransform: "uppercase",
      marginBottom: 12,
      fontFamily: "JetBrains Mono, monospace",
    }}
  >
    {children}
  </div>
);

const LevelTag = ({ level }) => {
  const map = {
    error: [C.red, "ERROR"],
    warn: [C.yellow, "WARN"],
    info: [C.accent, "INFO"],
  };
  const [color, label] = map[level] || [C.muted, level];
  return <Badge color={color}>{label}</Badge>;
};

const CustomTooltip = ({ active, payload, label }) => {
  if (!active || !payload?.length) return null;
  return (
    <div
      style={{
        background: "#1a1d26",
        border: `1px solid ${C.border}`,
        borderRadius: 6,
        padding: "8px 12px",
        fontSize: 11,
        fontFamily: "JetBrains Mono, monospace",
      }}
    >
      {payload.map((p, i) => (
        <div key={i} style={{ color: p.color, marginBottom: 2 }}>
          {p.name}: <strong>{p.value}</strong>
        </div>
      ))}
    </div>
  );
};

// ─── TOP KPI BAR ─────────────────────────────────────────────────────────────
const KPIBar = ({ stats }) => {
  const kpis = [
    {
      label: "Requests / sec",
      value: stats.rps.toLocaleString(),
      unit: "req/s",
      color: C.accent,
      delta: "+4.2%",
    },
    {
      label: "Avg Latency",
      value: stats.latency,
      unit: "ms",
      color: C.green,
      delta: "-1.1ms",
    },
    {
      label: "Error Rate",
      value: `${stats.errorRate}%`,
      unit: "",
      color: stats.errorRate > 2 ? C.yellow : C.green,
      delta: "+0.1%",
    },
    {
      label: "Live Instances",
      value: `${stats.healthy}/${stats.total}`,
      unit: "",
      color: stats.healthy === stats.total ? C.green : C.yellow,
      delta: null,
    },
  ];

  return (
    <div
      style={{
        display: "grid",
        gridTemplateColumns: "repeat(4,1fr)",
        gap: 1,
        borderRadius: 8,
        overflow: "hidden",
        border: `1px solid ${C.border}`,
      }}
    >
      {kpis.map((k, i) => (
        <div
          key={i}
          style={{
            background: C.card,
            padding: "16px 20px",
            borderRight: i < 3 ? `1px solid ${C.border}` : "none",
            display: "flex",
            flexDirection: "column",
            gap: 6,
          }}
        >
          <div
            style={{
              fontSize: 10,
              color: C.muted,
              fontWeight: 600,
              letterSpacing: "0.1em",
              textTransform: "uppercase",
              fontFamily: "JetBrains Mono, monospace",
            }}
          >
            {k.label}
          </div>
          <div style={{ display: "flex", alignItems: "baseline", gap: 6 }}>
            <span
              style={{
                fontSize: 26,
                fontWeight: 700,
                fontFamily: "JetBrains Mono, monospace",
                color: k.color,
                lineHeight: 1,
              }}
            >
              {k.value}
            </span>
            {k.unit && (
              <span style={{ fontSize: 11, color: C.muted }}>{k.unit}</span>
            )}
          </div>
          {k.delta && (
            <div
              style={{
                fontSize: 10,
                color: C.muted,
                fontFamily: "JetBrains Mono, monospace",
              }}
            >
              <span
                style={{ color: k.delta.startsWith("+") ? C.yellow : C.green }}
              >
                {k.delta}
              </span>{" "}
              vs prev 5m
            </div>
          )}
        </div>
      ))}
    </div>
  );
};

// ─── TRAFFIC CHART ───────────────────────────────────────────────────────────
const TrafficChart = ({ data }) => (
  <Card style={{ padding: "16px 16px 8px" }}>
    <div
      style={{
        display: "flex",
        justifyContent: "space-between",
        alignItems: "center",
        marginBottom: 12,
      }}
    >
      <SectionLabel style={{ marginBottom: 0 }}>Live Traffic</SectionLabel>
      <div
        style={{
          display: "flex",
          gap: 16,
          fontSize: 10,
          fontFamily: "JetBrains Mono, monospace",
        }}
      >
        <span style={{ color: C.accent }}>● RPS</span>
        <span style={{ color: C.red }}>● Errors</span>
        <span style={{ color: C.yellow }}>● P99 (ms)</span>
      </div>
    </div>
    <ResponsiveContainer width="100%" height={140}>
      <AreaChart data={data} margin={{ top: 4, right: 0, bottom: 0, left: 4 }}>
        <defs>
          <linearGradient id="gRps" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor={C.accent} stopOpacity={0.25} />
            <stop offset="95%" stopColor={C.accent} stopOpacity={0} />
          </linearGradient>
          <linearGradient id="gErr" x1="0" y1="0" x2="0" y2="1">
            <stop offset="5%" stopColor={C.red} stopOpacity={0.3} />
            <stop offset="95%" stopColor={C.red} stopOpacity={0} />
          </linearGradient>
        </defs>
        <CartesianGrid
          strokeDasharray="3 3"
          stroke={C.border2}
          vertical={false}
        />
        <XAxis dataKey="t" hide />
        <YAxis tick={{ fontSize: 9, fill: C.muted }} width={36} />
        <Tooltip content={<CustomTooltip />} />
        <Area
          type="monotone"
          dataKey="rps"
          name="RPS"
          stroke={C.accent}
          fill="url(#gRps)"
          strokeWidth={1.5}
          dot={false}
        />
        <Area
          type="monotone"
          dataKey="errors"
          name="Errors"
          stroke={C.red}
          fill="url(#gErr)"
          strokeWidth={1.5}
          dot={false}
        />
        <Line
          type="monotone"
          dataKey="p99"
          name="P99 ms"
          stroke={C.yellow}
          strokeWidth={1.5}
          dot={false}
        />
      </AreaChart>
    </ResponsiveContainer>
  </Card>
);

// ─── SERVICE TOPOLOGY (arch + pool + load distribution combined) ─────────────
const ServiceTopology = ({ backends, services, onSelect, selected }) => {
  const byService = services.map((svc) => {
    const dist = LB_DIST.find((d) => d.name === svc.id);
    return {
      ...svc,
      loadPct: dist?.value ?? 0,
      instances: backends
        .filter((b) => b.service === svc.id)
        .map((b) => ({
          ...b,
          instPct: dist?.instances.find((i) => i.name === b.name)?.pct ?? 0,
        })),
    };
  });

  return (
    <Card style={{ padding: 20 }}>
      {/* Header */}
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: 16,
        }}
      >
        <SectionLabel style={{ marginBottom: 0 }}>
          Service Topology
        </SectionLabel>
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 6,
            fontSize: 10,
            color: C.muted,
            fontFamily: "JetBrains Mono, monospace",
          }}
        >
          <StatusDot healthy={true} pulse size={6} />
          {backends.filter((b) => b.alive && !b.dead).length}/{backends.length}{" "}
          instances alive
        </div>
      </div>

      {/* Gateway strip */}
      <div
        style={{
          display: "flex",
          alignItems: "center",
          gap: 12,
          padding: "10px 16px",
          background: C.surface,
          borderRadius: 8,
          border: `1px solid ${C.border}`,
          marginBottom: 0,
        }}
      >
        <div
          style={{
            fontSize: 11,
            color: C.muted,
            fontFamily: "JetBrains Mono, monospace",
            fontWeight: 600,
            padding: "4px 12px",
            border: `1px solid ${C.border}`,
            borderRadius: 5,
            whiteSpace: "nowrap",
          }}
        >
          INTERNET
        </div>
        <svg width="36" height="2" style={{ flexShrink: 0 }}>
          <line
            x1="0"
            y1="1"
            x2="36"
            y2="1"
            stroke={C.accent}
            strokeWidth="1.5"
            strokeDasharray="4 3"
            style={{ animation: "march 1s linear infinite" }}
          />
        </svg>
        <div
          style={{
            background: C.accentDim,
            border: `1px solid ${C.accent}55`,
            borderRadius: 6,
            padding: "7px 16px",
            fontSize: 11,
            color: C.accent,
            fontFamily: "JetBrains Mono, monospace",
            fontWeight: 700,
            display: "flex",
            alignItems: "center",
            gap: 8,
            whiteSpace: "nowrap",
          }}
        >
          <span>⬡</span> API GATEWAY <StatusDot healthy={true} pulse size={6} />
        </div>
        {/* Redis health — single shared instance, sits beside gateway box */}
        <div
          style={{
            display: "flex",
            alignItems: "center",
            gap: 6,
            padding: "5px 10px",
            background: C.card,
            border: `1px solid ${C.border}`,
            borderRadius: 5,
          }}
          title="Redis health"
        >
          <div
            style={{
              width: 7,
              height: 7,
              borderRadius: "50%",
              background: C.green,
              boxShadow: `0 0 5px ${C.green}88`,
              flexShrink: 0,
            }}
          />
          <span
            style={{
              fontSize: 10,
              color: C.muted,
              fontFamily: "JetBrains Mono, monospace",
              letterSpacing: "0.04em",
            }}
          >
            redis
          </span>
        </div>
        <div
          style={{
            flex: 1,
            height: 1,
            background: `linear-gradient(90deg, ${C.accent}22, transparent)`,
          }}
        />
      </div>

      {/* Fan lines — point to top-center of each service panel */}
      <svg
        width="100%"
        height="32"
        style={{ display: "block", overflow: "visible" }}
      >
        {byService.map((svc, i) => {
          // Each column is 1/3 of width; centers at 1/6, 3/6, 5/6
          const targets = ["16.67%", "50%", "83.33%"];
          return (
            <line
              key={svc.id}
              x1="50%"
              y1="0"
              x2={targets[i]}
              y2="32"
              stroke={`${svc.color}66`}
              strokeWidth="1.5"
              strokeDasharray="4 4"
              style={{ animation: `march ${1.1 + i * 0.2}s linear infinite` }}
            />
          );
        })}
      </svg>

      {/* Service columns */}
      <div
        style={{
          display: "grid",
          gridTemplateColumns: "repeat(3, 1fr)",
          gap: 12,
        }}
      >
        {byService.map((svc) => (
          <div
            key={svc.id}
            style={{
              background: C.surface,
              border: `1px solid ${svc.color}28`,
              borderRadius: 10,
              overflow: "hidden",
            }}
          >
            {/* Service header */}
            <div
              style={{
                padding: "10px 14px 8px",
                borderBottom: `1px solid ${C.border}`,
                display: "flex",
                alignItems: "center",
                justifyContent: "space-between",
                gap: 8,
              }}
            >
              <div
                style={{
                  display: "flex",
                  alignItems: "center",
                  gap: 8,
                  minWidth: 0,
                }}
              >
                <div
                  style={{
                    width: 8,
                    height: 8,
                    borderRadius: "50%",
                    background: svc.color,
                    flexShrink: 0,
                  }}
                />
                <span
                  style={{
                    fontSize: 12,
                    fontWeight: 600,
                    color: C.text,
                    letterSpacing: "-0.01em",
                    overflow: "hidden",
                    textOverflow: "ellipsis",
                    whiteSpace: "nowrap",
                  }}
                >
                  {svc.name}
                </span>
              </div>
              <span
                style={{
                  fontSize: 14,
                  color: svc.color,
                  fontFamily: "JetBrains Mono, monospace",
                  fontWeight: 700,
                  flexShrink: 0,
                }}
              >
                {svc.loadPct}%
              </span>
            </div>
            {/* Service-level load bar */}
            <div style={{ height: 3, background: C.border2 }}>
              <div
                style={{
                  width: `${svc.loadPct}%`,
                  height: "100%",
                  background: `${svc.color}66`,
                }}
              />
            </div>

            {/* Instances */}
            <div
              style={{
                padding: "10px 10px 8px",
                display: "flex",
                flexDirection: "column",
                gap: 8,
              }}
            >
              {svc.instances.map((b) => (
                <button
                  key={b.id}
                  onClick={() => onSelect(selected?.id === b.id ? null : b)}
                  style={{
                    background:
                      selected?.id === b.id
                        ? `${svc.color}12`
                        : b.dead
                          ? `${C.red}08`
                          : `${C.bg}`,
                    border: `1px solid ${selected?.id === b.id ? `${svc.color}66` : b.dead ? `${C.red}33` : C.border}`,
                    borderRadius: 8,
                    padding: "12px 14px",
                    cursor: "pointer",
                    textAlign: "left",
                    transition: "border-color 0.15s, background 0.15s",
                    width: "100%",
                    opacity: b.dead ? 0.78 : 1,
                  }}
                >
                  {/* Name + badges */}
                  <div
                    style={{
                      display: "flex",
                      alignItems: "center",
                      gap: 8,
                      marginBottom: 10,
                    }}
                  >
                    <StatusDot
                      healthy={b.alive && !b.dead}
                      pulse={b.alive && !b.dead && !b.draining}
                      size={8}
                    />
                    <span
                      style={{
                        fontSize: 13,
                        color: b.dead ? C.red : C.text,
                        fontFamily: "JetBrains Mono, monospace",
                        fontWeight: 600,
                        flex: 1,
                      }}
                    >
                      {b.name}
                    </span>
                    {b.dead && (
                      <Badge color={b.probing ? C.yellow : C.red}>
                        {b.probing ? "probing" : "dead"}
                      </Badge>
                    )}
                    {b.draining && !b.dead && (
                      <Badge color={C.yellow}>drain</Badge>
                    )}
                  </div>

                  {/* Stats */}
                  <div
                    style={{
                      display: "grid",
                      gridTemplateColumns: "1fr 1fr 1fr",
                      gap: 6,
                      marginBottom: 10,
                    }}
                  >
                    {[
                      ["RPS", b.alive && !b.dead ? String(b.rps) : "—", C.text],
                      [
                        "P50",
                        b.p50 != null ? `${b.p50}ms` : "—",
                        b.p50 > 150 ? C.yellow : C.text,
                      ],
                      [
                        "P99",
                        b.p99 != null ? `${b.p99}ms` : "—",
                        b.p99 > 400 ? C.red : b.p99 > 250 ? C.yellow : C.text,
                      ],
                    ].map(([label, val, color]) => (
                      <div key={label}>
                        <div
                          style={{
                            fontSize: 9,
                            color: C.dim,
                            fontFamily: "JetBrains Mono, monospace",
                            textTransform: "uppercase",
                            letterSpacing: "0.06em",
                            marginBottom: 3,
                          }}
                        >
                          {label}
                        </div>
                        <div
                          style={{
                            fontSize: 14,
                            color: b.dead ? C.dim : color,
                            fontFamily: "JetBrains Mono, monospace",
                            fontWeight: 600,
                            lineHeight: 1,
                          }}
                        >
                          {val}
                        </div>
                      </div>
                    ))}
                  </div>

                  {/* Strikes + instance load share */}
                  <div
                    style={{ display: "flex", alignItems: "center", gap: 10 }}
                  >
                    <div
                      style={{ display: "flex", alignItems: "center", gap: 4 }}
                    >
                      {Array.from({ length: FAIL_THRESHOLD }, (_, i) => (
                        <div
                          key={i}
                          style={{
                            width: 10,
                            height: 10,
                            borderRadius: 3,
                            background:
                              i < b.failures
                                ? b.dead
                                  ? C.red
                                  : C.yellow
                                : C.dim,
                          }}
                        />
                      ))}
                      <span
                        style={{
                          fontSize: 10,
                          color: b.dead
                            ? C.red
                            : b.failures > 0
                              ? C.yellow
                              : C.muted,
                          fontFamily: "JetBrains Mono, monospace",
                          marginLeft: 2,
                        }}
                      >
                        {b.failures}/{FAIL_THRESHOLD}
                      </span>
                    </div>
                    {!b.dead && b.instPct > 0 && (
                      <div
                        style={{
                          flex: 1,
                          display: "flex",
                          alignItems: "center",
                          gap: 6,
                        }}
                      >
                        <div
                          style={{
                            flex: 1,
                            height: 4,
                            background: C.border2,
                            borderRadius: 2,
                            overflow: "hidden",
                          }}
                        >
                          <div
                            style={{
                              width: `${b.instPct}%`,
                              height: "100%",
                              background: b.draining
                                ? `${C.yellow}66`
                                : `${svc.color}77`,
                              borderRadius: 2,
                            }}
                          />
                        </div>
                        <span
                          style={{
                            fontSize: 10,
                            color: C.muted,
                            fontFamily: "JetBrains Mono, monospace",
                            flexShrink: 0,
                          }}
                        >
                          {b.instPct}%
                        </span>
                      </div>
                    )}
                  </div>
                </button>
              ))}
            </div>
          </div>
        ))}
      </div>
    </Card>
  );
};

// ─── EVENT + ERROR FEED ───────────────────────────────────────────────────────
const EventFeed = ({ events, filter }) => {
  const visible = filter
    ? events.filter((e) =>
        e.msg.toLowerCase().includes(filter.name.toLowerCase()),
      )
    : events;
  return (
    <Card style={{ padding: 16, height: "100%" }}>
      <div
        style={{
          display: "flex",
          justifyContent: "space-between",
          alignItems: "center",
          marginBottom: 12,
        }}
      >
        <SectionLabel style={{ marginBottom: 0 }}>Live Events</SectionLabel>
        <div style={{ display: "flex", alignItems: "center", gap: 6 }}>
          {filter && (
            <span
              style={{
                fontSize: 9,
                color: C.accent,
                fontFamily: "JetBrains Mono, monospace",
                background: `${C.accent}18`,
                border: `1px solid ${C.accent}33`,
                borderRadius: 4,
                padding: "2px 6px",
              }}
            >
              {filter.name}
            </span>
          )}
          <div style={{ display: "flex", gap: 1 }}>
            <span
              style={{
                width: 6,
                height: 6,
                borderRadius: "50%",
                background: C.red,
                display: "inline-block",
              }}
            />
            <span
              style={{
                width: 6,
                height: 6,
                borderRadius: "50%",
                background: C.red,
                display: "inline-block",
                opacity: 0.5,
                animation: "ping 1s infinite",
              }}
            />
          </div>
        </div>
      </div>
      <div
        style={{
          display: "flex",
          flexDirection: "column",
          gap: 2,
          overflowY: "auto",
          maxHeight: 420,
        }}
      >
        {visible.length === 0 ? (
          <div
            style={{
              padding: "20px 0",
              textAlign: "center",
              fontSize: 11,
              color: C.dim,
              fontFamily: "JetBrains Mono, monospace",
            }}
          >
            no events for {filter?.name}
          </div>
        ) : (
          visible.map((e) => (
            <div
              key={e.id}
              style={{
                padding: "7px 8px",
                borderRadius: 5,
                background:
                  e.level === "error"
                    ? `${C.red}0a`
                    : e.level === "warn"
                      ? `${C.yellow}08`
                      : `${C.accent}08`,
                borderLeft: `2px solid ${e.level === "error" ? C.red : e.level === "warn" ? C.yellow : C.accent}`,
                display: "flex",
                gap: 8,
                alignItems: "flex-start",
              }}
            >
              <LevelTag level={e.level} />
              <div style={{ flex: 1, minWidth: 0 }}>
                <div
                  style={{
                    fontSize: 11,
                    color: C.text,
                    lineHeight: 1.4,
                    wordBreak: "break-word",
                  }}
                >
                  {e.msg}
                </div>
                <div
                  style={{
                    fontSize: 9,
                    color: C.muted,
                    marginTop: 2,
                    fontFamily: "JetBrains Mono, monospace",
                  }}
                >
                  {fmt.ts(e.ts)} · {e.tag}
                </div>
              </div>
            </div>
          ))
        )}
      </div>
    </Card>
  );
};

// ─── MAIN DASHBOARD ───────────────────────────────────────────────────────────
export default function Dashboard() {
  const [traffic, setTraffic] = useState(generateTrafficHistory);
  const [backends] = useState(BACKENDS_INIT);
  const [selectedBackend, setSelectedBackend] = useState(null);
  const [events, setEvents] = useState(EVENTS_INIT);
  const [requests] = useState(REQUESTS_INIT);
  const [tick, setTick] = useState(0);

  // Simulate live traffic updates
  useEffect(() => {
    const id = setInterval(() => {
      setTraffic((prev) => {
        const last = prev[prev.length - 1];
        return [
          ...prev.slice(1),
          {
            t: last.t + 1,
            rps: Math.max(200, last.rps + (Math.random() - 0.5) * 120),
            errors: Math.max(0, last.errors + (Math.random() - 0.5) * 10),
            p99: Math.max(20, last.p99 + (Math.random() - 0.5) * 30),
          },
        ];
      });
      setTick((t) => t + 1);
    }, 1500);
    return () => clearInterval(id);
  }, []);

  // Simulate new events arriving
  useEffect(() => {
    const newEventMsgs = [
      {
        level: "info",
        msg: "PROBE user-2 GET /health → timeout (still DEAD)",
        tag: "PROBE",
      },
      {
        level: "info",
        msg: "PROBE user-2 GET /health → 200 OK → marked ALIVE",
        tag: "PROBE",
      },
      {
        level: "warn",
        msg: "GET /api/v1/users → user-2 DEAD, retried on user-1 → 200",
        tag: "RETRY",
      },
      {
        level: "warn",
        msg: "product-1 failure 2/3 — 1 more failure = DEAD",
        tag: "CIRCUIT",
      },
      {
        level: "error",
        msg: "product-1 failure 3/3 → circuit OPEN, marked DEAD",
        tag: "CIRCUIT",
      },
      {
        level: "info",
        msg: "PROBE product-1 (10.0.3.10) → probing... attempt 1",
        tag: "PROBE",
      },
      {
        level: "warn",
        msg: "Rate limit exceeded — 172.16.4.55",
        tag: "RATE_LIMIT",
      },
      {
        level: "error",
        msg: "JWT validation failed — /api/v1/admin (401)",
        tag: "AUTH",
      },
      {
        level: "info",
        msg: "order-1 health check OK — 0 strikes",
        tag: "BACKEND",
      },
      {
        level: "info",
        msg: "Config reload — no changes detected",
        tag: "GATEWAY",
      },
    ];
    const id = setInterval(() => {
      const e = newEventMsgs[Math.floor(Math.random() * newEventMsgs.length)];
      setEvents((prev) => [
        {
          id: Date.now(),
          ts: Date.now(),
          ...e,
        },
        ...prev.slice(0, 19),
      ]);
    }, 4000);
    return () => clearInterval(id);
  }, []);

  const stats = {
    rps: Math.floor(traffic[traffic.length - 1]?.rps ?? 974),
    latency: 44,
    errorRate: 1.8,
    healthy: backends.filter((b) => b.alive && !b.dead).length,
    total: backends.length,
  };

  return (
    <>
      <style>{`
        @import url('https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700&family=JetBrains+Mono:wght@400;600;700&display=swap');
        * { box-sizing: border-box; margin: 0; padding: 0; }
        body { background: ${C.bg}; color: ${C.text}; font-family: Inter, sans-serif; }
        button { background: none; border: none; color: inherit; font-family: inherit; }
        ::-webkit-scrollbar { width: 4px; }
        ::-webkit-scrollbar-track { background: ${C.surface}; }
        ::-webkit-scrollbar-thumb { background: ${C.border}; border-radius: 2px; }
        @keyframes ping {
          75%, 100% { transform: scale(2); opacity: 0; }
        }
        @keyframes march {
          to { stroke-dashoffset: -12; }
        }
      `}</style>

      <div style={{ minHeight: "100vh", background: C.bg, padding: "0" }}>
        {/* ── TOP NAV ── */}
        <div
          style={{
            height: 48,
            background: C.surface,
            borderBottom: `1px solid ${C.border}`,
            display: "flex",
            alignItems: "center",
            padding: "0 20px",
            justifyContent: "space-between",
          }}
        >
          <div style={{ display: "flex", alignItems: "center", gap: 10 }}>
            <div
              style={{
                width: 24,
                height: 24,
                background: C.accent,
                borderRadius: 5,
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
                fontSize: 11,
                fontWeight: 800,
                color: "#fff",
              }}
            >
              ⬡
            </div>
            <span
              style={{
                fontWeight: 700,
                fontSize: 14,
                letterSpacing: "-0.02em",
              }}
            >
              GateWatch
            </span>
            <span style={{ color: C.muted, fontSize: 11, margin: "0 4px" }}>
              |
            </span>
            <span style={{ fontSize: 11, color: C.muted }}>
              Production · us-east-1
            </span>
          </div>
          <div style={{ display: "flex", alignItems: "center", gap: 16 }}>
            <div
              style={{
                display: "flex",
                alignItems: "center",
                gap: 6,
                fontSize: 11,
                color: C.green,
                fontFamily: "JetBrains Mono, monospace",
              }}
            >
              <StatusDot healthy={true} pulse size={7} />
              All systems operational
            </div>
            <div
              style={{
                fontSize: 10,
                color: C.muted,
                fontFamily: "JetBrains Mono, monospace",
              }}
            >
              {new Date().toLocaleTimeString()} UTC
            </div>
          </div>
        </div>

        {/* ── MAIN LAYOUT ── */}
        <div
          style={{
            padding: "16px",
            display: "flex",
            flexDirection: "column",
            gap: 12,
          }}
        >
          {/* Row 1: KPIs */}
          <KPIBar stats={stats} />

          {/* Row 2: Traffic chart (full width) */}
          <TrafficChart data={traffic} />

          {/* Row 3: Service Topology + Events */}
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 300px",
              gap: 12,
            }}
          >
            <ServiceTopology
              backends={backends}
              services={SERVICES_DATA}
              onSelect={setSelectedBackend}
              selected={selectedBackend}
            />
            <EventFeed events={events} filter={selectedBackend} />
          </div>

          {/* Row 4: Routes + Request Inspector + Rate Limiting */}
          <div
            style={{
              display: "grid",
              gridTemplateColumns: "1fr 1fr 1fr",
              gap: 12,
            }}
          ></div>
        </div>
      </div>
    </>
  );
}
