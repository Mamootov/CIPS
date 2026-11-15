// CIPS Web Application Client Logic
document.addEventListener('DOMContentLoaded', () => {
    // State
    const state = {
        configs: new Map(), // id -> ProxyEvaluationReport
        filter: 'all',
        searchQuery: '',
        isScanning: false,
        stats: { total: 0, alive: 0, clean: 0, dead: 0 }
    };

    // Elements
    const dropZone = document.getElementById('drop-zone');
    const fileInput = document.getElementById('file-input');
    const subUrlInput = document.getElementById('sub-url');
    const btnFetchSub = document.getElementById('btn-fetch-sub');
    const rawTextArea = document.getElementById('raw-text');
    const btnParseText = document.getElementById('btn-parse-text');
    const btnStart = document.getElementById('btn-start');
    const btnCancel = document.getElementById('btn-cancel');
    const btnCopySub = document.getElementById('btn-copy-sub');
    const btnDownloadSub = document.getElementById('btn-download-sub');

    const concurrencySlider = document.getElementById('concurrency');
    const concurrencyVal = document.getElementById('concurrency-val');
    const timeoutP1Slider = document.getElementById('timeout-p1');
    const timeoutP1Val = document.getElementById('timeout-p1-val');
    const timeoutP2Slider = document.getElementById('timeout-p2');
    const timeoutP2Val = document.getElementById('timeout-p2-val');

    const statTotal = document.getElementById('stat-total');
    const statAlive = document.getElementById('stat-alive');
    const statClean = document.getElementById('stat-clean');
    const statDead = document.getElementById('stat-dead');

    const countAll = document.getElementById('count-all');
    const countClean = document.getElementById('count-clean');
    const countAlive = document.getElementById('count-alive');
    const countDead = document.getElementById('count-dead');

    const matrixTbody = document.getElementById('matrix-tbody');
    const filterTabs = document.querySelectorAll('.tab');
    const filterSearch = document.getElementById('filter-search');
    const toast = document.getElementById('toast');

    // Sliders
    concurrencySlider.addEventListener('input', (e) => concurrencyVal.textContent = e.target.value);
    timeoutP1Slider.addEventListener('input', (e) => timeoutP1Val.textContent = e.target.value);
    timeoutP2Slider.addEventListener('input', (e) => timeoutP2Val.textContent = e.target.value);

    // Toast helper
    function showToast(msg, duration = 3000) {
        toast.textContent = msg;
        toast.classList.add('show');
        setTimeout(() => toast.classList.remove('show'), duration);
    }

    // Drag and Drop
    ['dragenter', 'dragover'].forEach(name => {
        dropZone.addEventListener(name, (e) => {
            e.preventDefault();
            dropZone.classList.add('dragover');
        });
    });

    ['dragleave', 'drop'].forEach(name => {
        dropZone.addEventListener(name, (e) => {
            e.preventDefault();
            dropZone.classList.remove('dragover');
        });
    });

    dropZone.addEventListener('drop', (e) => {
        const files = e.dataTransfer.files;
        if (files.length > 0) handleFileUpload(files[0]);
    });

    fileInput.addEventListener('change', (e) => {
        if (e.target.files.length > 0) handleFileUpload(e.target.files[0]);
    });

    function handleFileUpload(file) {
        const formData = new FormData();
        formData.append('file', file);

        showToast(`Uploading ${file.name}...`);
        fetch('/api/ingest', { method: 'POST', body: formData })
            .then(res => res.json())
            .then(data => handleIngestSuccess(data))
            .catch(err => showToast(`Upload failed: ${err.message}`));
    }

    // Subscription Fetch
    btnFetchSub.addEventListener('click', () => {
        const url = subUrlInput.value.trim();
        if (!url) return showToast('Please enter a subscription URL');

        showToast('Fetching subscription feed...');
        fetch('/api/ingest', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ subscription_url: url })
        })
        .then(res => res.json())
        .then(data => handleIngestSuccess(data))
        .catch(err => showToast(`Fetch error: ${err.message}`));
    });

    // Parse Text
    btnParseText.addEventListener('click', () => {
        const text = rawTextArea.value.trim();
        if (!text) return showToast('Please paste links in text box');

        fetch('/api/ingest', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ text: text })
        })
        .then(res => res.json())
        .then(data => {
            handleIngestSuccess(data);
            rawTextArea.value = '';
        })
        .catch(err => showToast(`Parse error: ${err.message}`));
    });

    function handleIngestSuccess(data) {
        if (data.error) {
            showToast(`Error: ${data.error}`);
            return;
        }

        if (Array.isArray(data.configs)) {
            data.configs.forEach(cfg => {
                if (!state.configs.has(cfg.id)) {
                    state.configs.set(cfg.id, {
                        config: cfg,
                        state: 'QUEUED',
                        stage1_latency_ms: 0,
                        stage1_passed: false,
                        probe_results: {},
                        ip_info: {},
                        is_clean: false
                    });
                }
            });
        }

        showToast(`Ingested ${data.total_parsed} links (${data.unique_candidates} unique)`);
        updateStats();
        renderTable();
    }

    // Start Scan
    btnStart.addEventListener('click', () => {
        if (state.configs.size === 0) {
            showToast('Please ingest proxy configurations first');
            return;
        }

        btnStart.disabled = true;
        btnCancel.disabled = false;
        state.isScanning = true;

        fetch('/api/scan/start', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                concurrency: parseInt(concurrencySlider.value),
                stage1_timeout_ms: parseInt(timeoutP1Slider.value),
                stage2_timeout_ms: parseInt(timeoutP2Slider.value)
            })
        })
        .then(res => res.json())
        .then(data => {
            showToast('Parallel scan started');
        })
        .catch(err => {
            showToast(`Failed to start: ${err.message}`);
            btnStart.disabled = false;
            btnCancel.disabled = true;
            state.isScanning = false;
        });
    });

    // Cancel Scan
    btnCancel.addEventListener('click', () => {
        fetch('/api/scan/cancel', { method: 'POST' })
            .then(res => res.json())
            .then(() => {
                showToast('Scan cancellation requested');
                btnStart.disabled = false;
                btnCancel.disabled = true;
                state.isScanning = false;
            });
    });

    // Server-Sent Events (SSE) Live Feed
    function setupSSE() {
        const evtSource = new EventSource('/api/stream');

        evtSource.addEventListener('item_update', (e) => {
            const report = JSON.parse(e.data);
            if (report && report.config && report.config.id) {
                state.configs.set(report.config.id, report);
                updateRow(report);
                updateStats();
            }
        });

        evtSource.addEventListener('stats', (e) => {
            const stats = JSON.parse(e.data);
            if (stats) {
                state.stats = stats;
                if (!stats.is_running && state.isScanning) {
                    state.isScanning = false;
                    btnStart.disabled = false;
                    btnCancel.disabled = true;
                    showToast('Scan session completed!');
                }
                updateCounters();
            }
        });

        evtSource.onerror = () => {
            // Automatic reconnection handled by browser EventSource
        };
    }
    setupSSE();

    // Table Rendering & Updating
    function renderTable() {
        const filtered = getFilteredConfigs();
        if (filtered.length === 0) {
            matrixTbody.innerHTML = `
                <tr class="empty-state">
                    <td colspan="8">${state.configs.size === 0 ? 'No proxies loaded. Ingest files on the left to start.' : 'No proxies match current filter.'}</td>
                </tr>
            `;
            return;
        }

        matrixTbody.innerHTML = '';
        filtered.forEach((report, index) => {
            const tr = createRowElement(report, index + 1);
            matrixTbody.appendChild(tr);
        });
    }

    function createRowElement(report, index) {
        const tr = document.createElement('tr');
        tr.id = `row-${report.config.id}`;
        populateRowContent(tr, report, index);
        return tr;
    }

    function updateRow(report) {
        const tr = document.getElementById(`row-${report.config.id}`);
        if (tr) {
            populateRowContent(tr, report, tr.dataset.index);
        } else {
            renderTable();
        }
    }

    function populateRowContent(tr, report, index) {
        tr.dataset.index = index;
        const cfg = report.config;

        // Ping badge
        let pingBadge = '<span class="badge badge-pending">-</span>';
        if (report.stage1_latency_ms > 0) {
            pingBadge = `<span class="badge badge-pass">${report.stage1_latency_ms}ms</span>`;
        } else if (report.state === 'STAGE1_DEAD') {
            pingBadge = '<span class="badge badge-fail">Dead</span>';
        } else if (report.state === 'STAGE1_HANDSHAKE') {
            pingBadge = '<span class="badge badge-running">Ping...</span>';
        }

        // Service checkpoints
        const geminiCheck = renderCheckpoint(report.probe_results['gemini'], cfg.id, 'gemini');
        const studioCheck = renderCheckpoint(report.probe_results['aistudio'], cfg.id, 'aistudio');
        const ipleakCheck = renderIPLeakCell(report.ip_info);

        // State pill
        let stateHtml = '';
        switch(report.state) {
            case 'QUEUED':
                stateHtml = '<span class="state-pill" style="color:var(--text-muted)">Queued</span>'; break;
            case 'STAGE1_HANDSHAKE':
                stateHtml = '<span class="state-pill handshake">Pinging...</span>'; break;
            case 'STAGE1_DEAD':
                stateHtml = '<span class="state-pill stage1_dead">Stage 1 Dead</span>'; break;
            case 'STAGE2_PROBING':
                stateHtml = '<span class="state-pill probing">Probing AI...</span>'; break;
            case 'COMPLETE':
                stateHtml = report.is_clean ? 
                    '<span class="state-pill complete clean">🌟 AI Clean</span>' : 
                    '<span class="state-pill complete nonclean">Probed (Failed)</span>';
                break;
            default:
                stateHtml = `<span class="state-pill">${report.state}</span>`;
        }

        tr.innerHTML = `
            <td class="col-idx">${index}</td>
            <td class="col-proto"><span class="badge">${cfg.protocol.toUpperCase()}</span></td>
            <td class="col-server" title="${cfg.raw_uri}">
                <strong>${cfg.server}:${cfg.port}</strong>
                ${cfg.remark ? `<div style="font-size:0.75rem;color:var(--text-muted)">${escapeHtml(cfg.remark)}</div>` : ''}
            </td>
            <td class="col-ping">${pingBadge}</td>
            <td class="col-service">${geminiCheck}</td>
            <td class="col-service">${studioCheck}</td>
            <td class="col-service">${ipleakCheck}</td>
            <td class="col-status">${stateHtml}</td>
        `;
    }

    function renderCheckpoint(res, proxyId, probeId) {
        if (!res) return '<span class="badge badge-pending">○</span>';
        if (res.status === 'PASS') {
            return `<span class="badge badge-pass" title="Latency: ${res.latency_ms}ms">✓ PASS (${res.latency_ms}ms)</span>`;
        }
        if (res.status === 'FAIL') {
            let label = '✗ FAIL';
            let title = res.error_message || ('HTTP ' + res.http_code);
            if (res.diagnostic) {
                if (res.diagnostic.interception_type === 'CLOUDFLARE_CHALLENGE') {
                    label = '🛡️ Cloudflare';
                    title = `Cloudflare Challenge (Ray-ID: ${res.diagnostic.ray_id || 'N/A'}) - Click to inspect`;
                } else if (res.diagnostic.interception_type === 'GOOGLE_UNUSUAL_TRAFFIC') {
                    label = '🤖 Captcha';
                    title = 'Google Unusual Traffic Captcha (/sorry/index) - Click to inspect';
                } else if (res.diagnostic.interception_type === 'GEO_RESTRICTED') {
                    label = '🌍 Geoblock';
                    title = 'Region Restricted - Click to inspect';
                } else if (res.diagnostic.interception_type === 'LOCAL_ENGINE_ERROR') {
                    label = '⚠️ Port Refused';
                    title = 'Local Engine Port Refused - Click to inspect';
                }
            }
            return `<span class="badge badge-fail" onclick="openDiagModal('${proxyId}', '${probeId}')" title="${escapeHtml(title)}">${label}</span>`;
        }
        if (res.status === 'RUNNING') {
            return '<span class="badge badge-running">Testing...</span>';
        }
        return '<span class="badge badge-pending">○</span>';
    }

    function renderIPLeakCell(ipInfo) {
        if (!ipInfo || !ipInfo.exit_ip) return '<span class="badge badge-pending">○</span>';
        return `
            <div style="font-size:0.75rem">
                <strong>${ipInfo.exit_ip}</strong>
                <div style="color:var(--text-secondary)">[${ipInfo.country_code || '??'}] ${ipInfo.isp || 'Unknown'}</div>
            </div>
        `;
    }

    function getFilteredConfigs() {
        const list = Array.from(state.configs.values());
        return list.filter(r => {
            // Tab filter
            if (state.filter === 'clean' && !r.is_clean) return false;
            if (state.filter === 'alive' && (!r.stage1_passed || r.state === 'STAGE1_DEAD')) return false;
            if (state.filter === 'dead' && r.state !== 'STAGE1_DEAD' && (r.is_clean || r.state !== 'COMPLETE')) return false;

            // Search filter
            if (state.searchQuery) {
                const q = state.searchQuery.toLowerCase();
                const srv = (r.config.server || '').toLowerCase();
                const rmk = (r.config.remark || '').toLowerCase();
                const isp = (r.ip_info && r.ip_info.isp ? r.ip_info.isp : '').toLowerCase();
                if (!srv.includes(q) && !rmk.includes(q) && !isp.includes(q)) return false;
            }

            return true;
        });
    }

    function updateStats() {
        let total = state.configs.size;
        let alive = 0;
        let clean = 0;
        let dead = 0;

        state.configs.forEach(r => {
            if (r.stage1_passed) alive++;
            if (r.is_clean) clean++;
            if (r.state === 'STAGE1_DEAD') dead++;
        });

        countAll.textContent = total;
        countAlive.textContent = alive;
        countClean.textContent = clean;
        countDead.textContent = dead;

        statTotal.textContent = total;
        statAlive.textContent = alive;
        statClean.textContent = clean;
        statDead.textContent = dead;
    }

    function updateCounters() {
        if (state.stats) {
            statTotal.textContent = state.stats.total_ingested;
            statAlive.textContent = state.stats.stage1_alive;
            statClean.textContent = state.stats.stage2_clean;
            statDead.textContent = state.stats.stage1_dead;
        }
    }

    // Tabs & Search
    filterTabs.forEach(tab => {
        tab.addEventListener('click', (e) => {
            filterTabs.forEach(t => t.classList.remove('active'));
            tab.classList.add('active');
            state.filter = tab.dataset.filter;
            renderTable();
        });
    });

    filterSearch.addEventListener('input', (e) => {
        state.searchQuery = e.target.value.trim();
        renderTable();
    });

    // Copy Clean Subscription
    btnCopySub.addEventListener('click', () => {
        fetch('/api/export?format=sub')
            .then(res => res.text())
            .then(text => {
                if (!text.trim()) {
                    showToast('No clean verified configurations available to export');
                    return;
                }
                navigator.clipboard.writeText(text).then(() => {
                    showToast('Clean subscription copied to clipboard!');
                });
            })
            .catch(err => showToast(`Export failed: ${err.message}`));
    });

    // Download Clean Subscription
    btnDownloadSub.addEventListener('click', () => {
        window.open('/api/export?format=sub', '_blank');
    });

    // Diagnostic Inspection Modal Handlers
    window.openDiagModal = function(proxyId, probeId) {
        const report = state.configs.get(proxyId);
        if (!report) return;
        const res = report.probe_results[probeId];
        if (!res || !res.diagnostic) {
            showToast('No detailed diagnostic available for this probe');
            return;
        }

        const d = res.diagnostic;
        document.getElementById('modal-probe-name').textContent = res.probe_name || probeId;
        document.getElementById('modal-error-layer').textContent = d.layer || 'UNKNOWN';
        document.getElementById('modal-interception').textContent = d.interception_type || 'NONE';
        document.getElementById('modal-http-status').textContent = d.http_status ? d.http_status : (res.http_code || 'N/A');
        document.getElementById('modal-server').textContent = d.server_header || 'N/A';
        document.getElementById('modal-ray-id').textContent = d.ray_id || 'N/A';
        document.getElementById('modal-final-url').textContent = d.final_url || 'N/A';
        document.getElementById('modal-summary').textContent = d.summary || res.error_message || 'N/A';

        document.getElementById('diag-modal').classList.add('open');
    };

    document.getElementById('modal-close').addEventListener('click', () => {
        document.getElementById('diag-modal').classList.remove('open');
    });

    document.getElementById('diag-modal').addEventListener('click', (e) => {
        if (e.target.id === 'diag-modal') {
            document.getElementById('diag-modal').classList.remove('open');
        }
    });

    document.addEventListener('keydown', (e) => {
        if (e.key === 'Escape') {
            document.getElementById('diag-modal').classList.remove('open');
        }
    });

    function escapeHtml(str) {
        return (str || '').replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
    }
});
