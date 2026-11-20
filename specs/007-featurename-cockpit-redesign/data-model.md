# Data Model: 007 - Cockpit UI/UX Redesign & Telemetry Synchronization

---

## 1. Frontend State Models (`web/app.js`)

### 1.1 `CockpitState`
```typescript
interface CockpitState {
  currentTab: 'dashboard' | 'map' | 'probes' | 'subscriptions';
  reports: Map<string, EvaluationReport>;
  filter: {
    status: 'all' | 'clean' | 'geo' | 'fail';
    country: string | null;
    residentialOnly: boolean;
    searchQuery: string;
  };
  metrics: {
    total: number;
    cleanCount: number;
    geoCount: number;
    failCount: number;
    avgSpeedMbps: number;
    p50LatencyMs: number;
    p95LatencyMs: number;
  };
  globe: {
    selectedNodeId: string | null;
    selectedCity: string | null;
    cityClusters: Map<string, ClusterNode[]>;
  };
}
```

### 1.2 `ClusterNode`
```typescript
interface ClusterNode {
  id: string;
  remark: string;
  protocol: string;
  city: string;
  country: string;
  lat: number;
  lon: number;
  exitIP: string;
  isp: string;
  isClean: boolean;
  speedMbps: number;
  latencyMs: number;
  probeResults: Record<string, ProbeResult>;
}
```

---

## 2. City Coordinate Registry
```javascript
const CITY_GEO_REGISTRY = {
  "Frankfurt": { lat: 50.1109, lon: 8.6821, country: "DE" },
  "Nuremberg": { lat: 49.4521, lon: 11.0767, country: "DE" },
  "Amsterdam": { lat: 52.3676, lon: 4.9041, country: "NL" },
  "Helsinki":  { lat: 60.1699, lon: 24.9384, country: "FI" },
  "New York":  { lat: 40.7128, lon: -74.0060, country: "US" },
  "San Jose":  { lat: 37.3382, lon: -121.8863, country: "US" },
  "London":    { lat: 51.5074, lon: -0.1278, country: "GB" },
  "Paris":     { lat: 48.8566, lon: 2.3522, country: "FR" },
  "Moscow":    { lat: 55.7558, lon: 37.6173, country: "RU" },
  "Singapore": { lat: 1.3521, lon: 103.8198, country: "SG" },
  "Tokyo":     { lat: 35.6762, lon: 139.6503, country: "JP" },
  "Tehran":    { lat: 35.6892, lon: 51.3890, country: "IR" } // Source Hub
};
```
