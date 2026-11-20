# Data Model: 007 - Cockpit UI/UX Redesign & Telemetry Synchronization

---

## 1. Frontend State Models (`web/app.js`)

### 1.1 `CockpitState`
```typescript
interface CockpitState {
  activeTab: 'dashboard' | 'map' | 'probes' | 'subscriptions';
  reports: Map<string, ProxyEvaluationReport>;
  filter: {
    status: 'all' | 'clean' | 'available' | 'geo' | 'error' | 'residential';
    country: string | null;
    residentialOnly: boolean;
    searchQuery: string;
  };
  globeFilter: 'all' | 'clean' | 'residential' | string; // 'all' or ISO country code
  activeProbes: string[]; // e.g. ['gemini', 'aistudio', 'chatgpt', 'claude', 'speedtest']
  customProbes: CustomProbeDef[];
  concurrency: number; // 5 - 100 workers
  stage1Timeout: number; // 1000 - 8000 ms
  stage2Timeout: number; // 3000 - 15000 ms
  speedtestEnabled: boolean;
  speedtestSizeMb: number; // 1 - 5 MB
  scanRunning: boolean;
  selectedNodeId: string | null;
  origin: HostOrigin;
}

interface HostOrigin {
  ip: string;
  country_code: string;
  country_name: string;
  city: string;
  latitude: number;
  longitude: number;
  isp?: string;
  flag?: string;
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

## 2. Geodesic Trajectory Mathematical Model (`GeodesicOrbitCurve`)

```typescript
class GeodesicOrbitCurve extends THREE.Curve<THREE.Vector3> {
  r: number; // Globe radius (7.5)
  v1: THREE.Vector3; // Unit vector for origin (Host Egress)
  v2: THREE.Vector3; // Unit vector for destination (Proxy Exit)
  axis: THREE.Vector3; // Perpendicular rotation axis: normalize(v1 x v2)
  theta: number; // Angular separation: acos(clamp(v1 . v2, -1, 1))
  maxAlt: number; // Peak elevation: 0.55 + theta * 0.65

  getPoint(t: number, target: THREE.Vector3): THREE.Vector3;
}
```
**Formula Guarantee**:  
At any parameter $t \in [0, 1]$:
$$\mathbf{u}(t) = \text{applyAxisAngle}(\mathbf{v}_1, \mathbf{axis}, t \cdot \theta)$$
$$R(t) = r + \sin(t \cdot \pi) \cdot \text{maxAlt}$$
$$\mathbf{P}(t) = \mathbf{u}(t) \cdot R(t)$$
Since $R(t) \ge r$ for all $t$, the curve strictly vaults above the spherical surface of the globe.

---

## 3. Dynamic Country & Flag Registry
- **Universal Emoji Flag Generation**: Derived dynamically via Unicode regional indicator math:
  `String.fromCodePoint(127397 + c0, 127397 + c1)` for all 249 ISO 3166-1 alpha-2 codes.
- **65+ Country Capital/Tech-Hub Geocoordinates**: Comprehensive fallback dictionary mapping ISO codes to exact coordinates (Frankfurt, Amsterdam, Ashburn, Montreal, Tokyo, Sydney, Zurich, etc.).
- **Deterministic Hash Spatial Fallback**: For any uncharted ISO code, coordinates are generated via character hash ensuring no country collides onto Germany.
