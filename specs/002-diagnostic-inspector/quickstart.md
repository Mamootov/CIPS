# Quickstart & Verification: Diagnostic Logging & Interception Inspector

## Verification Scenarios

### 1. Test Cloudflare Challenge Detection
- Target: A URL behind Cloudflare Under-Attack mode.
- Expected Result:
  - `status`: `FAIL`
  - `diagnostic.interception_type`: `CLOUDFLARE_CHALLENGE`
  - `diagnostic.ray_id`: populated
  - `diagnostic.page_title`: `"Just a moment..."`

### 2. Test Google Unusual Traffic (/sorry/index) Detection
- Target: Google Gemini through a flagged/abused datacenter IP.
- Expected Result:
  - `status`: `FAIL`
  - `diagnostic.interception_type`: `GOOGLE_UNUSUAL_TRAFFIC`
  - `diagnostic.final_url`: `https://www.google.com/sorry/index?...`

### 3. Test Local Socket Poller
- Simulate startup delay on local inbound port.
- Verify: Poller polls at 20ms intervals until port accepts TCP handshake; probes do NOT receive `connectex: actively refused it`.

### 4. Test Clean OAuth Sign-in Pass
- Target: Clean Google sign-in response containing standard JavaScript with `access_denied` script tokens.
- Verify: Evaluator strips `<script>` tags, identifies valid sign-in, and flags `PASS`.
