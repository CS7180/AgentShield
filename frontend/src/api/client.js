export const API_BASE = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

function authHeaders(accessToken) {
  return {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${accessToken}`,
  };
}

async function request(path, options) {
  const res = await fetch(`${API_BASE}${path}`, options);
  const body = await res.json().catch(() => ({}));
  if (!res.ok) {
    const err = new Error(body.error ?? `HTTP ${res.status}`);
    err.code = body.code;
    err.status = res.status;
    err.isAuthError = res.status === 401 || res.status === 403;
    err.isGatewayError = res.status >= 500;
    throw err;
  }
  return body;
}

export function createScan(payload, accessToken) {
  return request('/api/v1/scans', {
    method: 'POST',
    headers: authHeaders(accessToken),
    body: JSON.stringify(payload),
  });
}

export function listScans(accessToken, { limit = 50, offset = 0 } = {}) {
  const params = new URLSearchParams({
    limit: String(limit),
    offset: String(offset),
  });
  return request(`/api/v1/scans?${params.toString()}`, {
    method: 'GET',
    headers: authHeaders(accessToken),
  });
}

export function getScan(id, accessToken) {
  return request(`/api/v1/scans/${id}`, {
    method: 'GET',
    headers: authHeaders(accessToken),
  });
}

export function startScan(id, accessToken) {
  return request(`/api/v1/scans/${id}/start`, {
    method: 'POST',
    headers: authHeaders(accessToken),
  });
}

export function stopScan(id, accessToken) {
  return request(`/api/v1/scans/${id}/stop`, {
    method: 'POST',
    headers: authHeaders(accessToken),
  });
}

export function listAttackResults(id, accessToken, { limit = 200, offset = 0 } = {}) {
  const params = new URLSearchParams({
    limit: String(limit),
    offset: String(offset),
  });
  return request(`/api/v1/scans/${id}/attack-results?${params.toString()}`, {
    method: 'GET',
    headers: authHeaders(accessToken),
  });
}

export function listScanDeadLetters(id, accessToken, { limit = 20, offset = 0 } = {}) {
  const params = new URLSearchParams({
    limit: String(limit),
    offset: String(offset),
  });
  return request(`/api/v1/scans/${id}/dead-letters?${params.toString()}`, {
    method: 'GET',
    headers: authHeaders(accessToken),
  });
}

export function getScanReport(id, accessToken) {
  return request(`/api/v1/scans/${id}/report`, {
    method: 'GET',
    headers: authHeaders(accessToken),
  });
}

export function generateScanReport(id, accessToken, includePDF = true) {
  return request(`/api/v1/scans/${id}/report/generate`, {
    method: 'POST',
    headers: authHeaders(accessToken),
    body: JSON.stringify({ include_pdf: includePDF }),
  });
}

export function compareScans(baseScanID, otherScanID, accessToken) {
  return request(`/api/v1/scans/${baseScanID}/compare/${otherScanID}`, {
    method: 'GET',
    headers: authHeaders(accessToken),
  });
}

export function calibrateJudge(samples, accessToken) {
  return request('/api/v1/judge/calibrate', {
    method: 'POST',
    headers: authHeaders(accessToken),
    body: JSON.stringify({ samples }),
  });
}

export function getJudgeCalibrationReport(accessToken) {
  return request('/api/v1/judge/calibration-report', {
    method: 'GET',
    headers: authHeaders(accessToken),
  });
}
