// Node sidecar that converts between XLSForm JSON and DDI Codebook XML via
// @correlaid/formtransform, both ways, and serves its question-type catalogue. Started by qwacback (Dockerfile) and
// reached over HTTP from internal/converter/ddi_client.go.
import { createServer } from 'node:http';
import { ConversionError, QUESTION_TYPES, ddiToXlsform, xlsformToDdi } from '@correlaid/formtransform';

const PORT = Number(process.env.DDI_EMITTER_PORT ?? 8091);

const HEADERS = {
  'content-type': 'application/xml; charset=utf-8',
  'x-content-type-options': 'nosniff',
  'x-frame-options': 'DENY',
};

function readBody(req) {
  return new Promise((resolve, reject) => {
    let body = '';
    req.setEncoding('utf8');
    req.on('data', (c) => {
      body += c;
      if (body.length > 5 * 1024 * 1024) {
        reject(new Error('payload too large'));
        req.destroy();
      }
    });
    req.on('end', () => resolve(body));
    req.on('error', reject);
  });
}

async function readJson(req) {
  const body = await readBody(req);
  try {
    return body ? JSON.parse(body) : {};
  } catch (e) {
    throw new Error(`invalid JSON: ${e.message}`);
  }
}

// The settings sheet as sent: an object, or a one-row array as XLSForm has
// it. Passed through whole, so default_language (the base language of a
// multilingual form, codeBook/@xml:lang) reaches formtransform along with
// form_id, form_title and version.
function pickSettings(input) {
  const s = input?.settings;
  if (Array.isArray(s)) return s[0] ?? {};
  return s && typeof s === 'object' ? s : {};
}

function send(res, status, body, extraHeaders = {}) {
  res.writeHead(status, { ...HEADERS, ...extraHeaders });
  res.end(body);
}

const server = createServer(async (req, res) => {
  if (req.method === 'GET' && req.url === '/healthz') {
    res.writeHead(200, { 'content-type': 'text/plain' });
    res.end('ok');
    return;
  }
  if (req.method === 'GET' && req.url === '/question-types') {
    // The registry's question-type catalogue, as the pinned release has it.
    send(res, 200, JSON.stringify(QUESTION_TYPES), { 'content-type': 'application/json' });
    return;
  }
  if (req.method === 'POST' && req.url === '/ddi-to-xlsform') {
    await ddiToXlsformRoute(req, res);
    return;
  }
  if (req.method !== 'POST' || req.url !== '/xlsform-to-ddi') {
    send(res, 404, 'not found', { 'content-type': 'text/plain' });
    return;
  }

  let payload;
  try {
    payload = await readJson(req);
  } catch (e) {
    send(res, 400, JSON.stringify({ error: e.message }), {
      'content-type': 'application/json',
    });
    return;
  }

  const survey = Array.isArray(payload?.survey) ? payload.survey : [];
  const choices = Array.isArray(payload?.choices) ? payload.choices : [];

  if (survey.length === 0) {
    send(res, 400, JSON.stringify({ error: 'survey sheet is empty' }), {
      'content-type': 'application/json',
    });
    return;
  }

  try {
    // xlsformToDdi runs formtransform's subset check (target 'ddi') first and
    // throws ConversionError('xlsform-outside-subset') listing every finding:
    // unregistered types (rank, geopoint, ...), selects without choices,
    // dangling ${references}.
    const xml = xlsformToDdi(
      { surveyData: survey, choicesData: choices },
      {
        settings: pickSettings(payload),
        onWarning: (d) => console.warn(`ddi-emitter: ${d.code}: ${d.message}`),
      },
    );
    send(res, 200, xml);
  } catch (e) {
    if (e instanceof ConversionError) {
      // The input is at fault: 400 with every finding; each names the question.
      const findings = e.details.length > 0 ? e.details : [e];
      const errors = findings.filter((d) => d.severity === 'error').map((d) => d.message);
      send(res, 400, JSON.stringify({ error: errors.join('; ') || e.message, errors }), {
        'content-type': 'application/json',
      });
      return;
    }
    // Anything else is a bug in the sidecar or the library, not in the input;
    // qwacback answers 503 for it.
    console.error('ddi-emitter: conversion failed', e);
    send(res, 500, JSON.stringify({ error: 'internal error' }), {
      'content-type': 'application/json',
    });
  }
});

// DDI (a codebook, a <dataDscr>, or bare <var>/<varGrp> elements) →
// { survey, choices, settings, warnings }. ddiToXlsform never refuses DDI it
// can parse; each field it can't supply is a warning, returned with the form.
async function ddiToXlsformRoute(req, res) {
  const json = { 'content-type': 'application/json' };
  let xml;
  try {
    xml = await readBody(req);
  } catch (e) {
    send(res, 400, JSON.stringify({ error: e.message }), json);
    return;
  }
  const warnings = [];
  try {
    const form = ddiToXlsform(xml, {
      onWarning: (d) => warnings.push({ code: d.code, message: d.message }),
    });
    send(res, 200, JSON.stringify({ ...form, warnings }), json);
  } catch (e) {
    if (e instanceof ConversionError) {
      // ddi-invalid: not well-formed, or neither a codeBook nor a var.
      send(res, 400, JSON.stringify({ error: e.message }), json);
      return;
    }
    console.error('ddi-emitter: ddi-to-xlsform failed', e);
    send(res, 500, JSON.stringify({ error: 'internal error' }), json);
  }
}

server.listen(PORT, '0.0.0.0', () => {
  console.log(`ddi-emitter listening on http://0.0.0.0:${PORT}`);
});
