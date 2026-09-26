// Node sidecar that turns XLSForm JSON into DDI Codebook XML via
// @correlaid/formtransform. Started by qwacback (Dockerfile) and reached over
// HTTP from internal/converter/ddi_client.go.
import { createServer } from 'node:http';
import { ConversionError, xlsformToDdi } from '@correlaid/formtransform';

const PORT = Number(process.env.DDI_EMITTER_PORT ?? 8091);

const HEADERS = {
  'content-type': 'application/xml; charset=utf-8',
  'x-content-type-options': 'nosniff',
  'x-frame-options': 'DENY',
};

function readJson(req) {
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
    req.on('end', () => {
      try {
        resolve(body ? JSON.parse(body) : {});
      } catch (e) {
        reject(new Error(`invalid JSON: ${e.message}`));
      }
    });
    req.on('error', reject);
  });
}

function pickSettings(input) {
  const s = input?.settings ?? {};
  return {
    id_string: s.form_id ?? s.id_string,
    form_title: s.form_title,
    version: s.version,
  };
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

server.listen(PORT, '0.0.0.0', () => {
  console.log(`ddi-emitter listening on http://0.0.0.0:${PORT}`);
});
