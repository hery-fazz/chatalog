import type { APIGatewayProxyResultV2, APIGatewayProxyEventV2 } from 'aws-lambda';

// Build CSV export URL dari Google Sheets (PUBLIC)
function buildCsvUrl(spreadsheetId: string, gidOrSheet?: string, sheetQueryName?: string) {
  const base = `https://docs.google.com/spreadsheets/d/${encodeURIComponent(spreadsheetId)}/export?format=csv`;
  if (gidOrSheet && /^\d+$/.test(gidOrSheet)) {
    return `${base}&gid=${gidOrSheet}`;
  }
  // Jika bukan numeric gid, coba by sheet name (param path) atau query ?sheet=
  const name = gidOrSheet || sheetQueryName || 'Sheet1';
  return `${base}&sheet=${encodeURIComponent(name)}`;
}

async function fetchCsv(spreadsheetId: string, gidOrSheet?: string, sheetQueryName?: string): Promise<string> {
  const url = buildCsvUrl(spreadsheetId, gidOrSheet, sheetQueryName);
  const res = await fetch(url);
  if (!res.ok) {
    const text = await res.text().catch(() => '');
    throw new Error(`Fetch CSV gagal: HTTP ${res.status} ${res.statusText} ${text ? `- ${text.slice(0, 200)}` : ''}`);
  }
  return await res.text();
}

// CSV -> JSON (assume row1 = header)
function csvToJson(csv: string) {
  const lines = csv.replace(/\r\n?/g, '\n').split('\n');
  const rows = lines.filter(l => l.length > 0).map(l => parseCsvLine(l));
  if (rows.length === 0) return [];
  const [header, ...data] = rows;
  return data.map(r => Object.fromEntries(header.map((h, i) => [h, r[i] ?? ''])));
}

// Minimal CSV parser (handles quotes & commas)
function parseCsvLine(line: string): string[] {
  const out: string[] = [];
  let cur = '';
  let inQ = false;
  for (let i = 0; i < line.length; i++) {
    const ch = line[i];
    if (inQ) {
      if (ch === '"') {
        if (line[i + 1] === '"') { cur += '"'; i++; } else { inQ = false; }
      } else { cur += ch; }
    } else {
      if (ch === ',') { out.push(cur); cur = ''; }
      else if (ch === '"') { inQ = true; }
      else { cur += ch; }
    }
  }
  out.push(cur);
  return out;
}

function ok(body: string | object, contentType = 'application/json'): APIGatewayProxyResultV2 {
  return {
    statusCode: 200,
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Access-Control-Allow-Headers': 'Content-Type,Authorization',
      'Access-Control-Allow-Methods': 'GET,OPTIONS',
      'Content-Type': contentType
    },
    body: typeof body === 'string' ? body : JSON.stringify(body)
  };
}

function err(e: unknown, code = 500): APIGatewayProxyResultV2 {
  const message = e instanceof Error ? e.message : String(e);
  return {
    statusCode: code,
    headers: {
      'Access-Control-Allow-Origin': '*',
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({ error: message })
  };
}

export async function publicCsv(event: APIGatewayProxyEventV2): Promise<APIGatewayProxyResultV2> {
  try {
    const spreadsheetId = event.pathParameters?.spreadsheetId;
    const gidOrSheet = event.pathParameters?.gidOrSheet;
    const sheetQuery = event.queryStringParameters?.sheet; // alternatif nama sheet via query
    if (!spreadsheetId) return err('Missing spreadsheetId', 400);
    const csv = await fetchCsv(spreadsheetId, gidOrSheet, sheetQuery);
    return ok(csv, 'text/csv; charset=utf-8');
  } catch (e) {
    return err(e);
  }
}

export async function publicJson(event: APIGatewayProxyEventV2): Promise<APIGatewayProxyResultV2> {
  try {
    const spreadsheetId = event.pathParameters?.spreadsheetId;
    const gidOrSheet = event.pathParameters?.gidOrSheet;
    const sheetQuery = event.queryStringParameters?.sheet;
    if (!spreadsheetId) return err('Missing spreadsheetId', 400);
    const csv = await fetchCsv(spreadsheetId, gidOrSheet, sheetQuery);
    const json = csvToJson(csv);
    return ok({ values: json });
  } catch (e) {
    return err(e);
  }
}
