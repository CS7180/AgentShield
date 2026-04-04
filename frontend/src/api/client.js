const API_BASE = import.meta.env.VITE_API_URL ?? 'http://localhost:8080';

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

export function listScans(accessToken) {
  return request('/api/v1/scans', {
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

export function getScanReport(id, accessToken) {
  return request(`/api/v1/scans/${id}/report`, {
    method: 'GET',
    headers: authHeaders(accessToken),
  });
}
