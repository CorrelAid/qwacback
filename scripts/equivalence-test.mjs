#!/usr/bin/env node
// Formtransform equivalence test against qwacback (formtransform#14, qwacback#3).
//
// Compares formtransform's buildDdiXml output with qwacback's
// POST /api/convert/xlsform-to-ddi output. For each supported answer type, the
// same XLSForm goes through both, and the DDI shapes must match.

// Path to a built formtransform dist/. Either clone formtransform and run
// `npm ci && npm run build`, or install the release tarball into a temp
// directory and point this at its dist/index.js.
const FORMTRANSFORM_DIST = process.env.FORMTRANSFORM_DIST
  || '/tmp/opencode/formtransform/package/dist/index.js';

const { buildDdiXml } = await import(FORMTRANSFORM_DIST);

const QWACBACK_URL = process.env.QWACBACK_URL || 'http://127.0.0.1:8090';
const QWACBACK_EMAIL = process.env.QWACBACK_EMAIL || 'admin@example.com';
const QWACBACK_PASSWORD = process.env.QWACBACK_PASSWORD || 'yourpassword123';

const EQUIVALENT_TYPES = [
  {
    id: 'single_choice',
    survey: [{ type: 'select_one bildung', name: 'bildung', label: 'Bildungsgrad', required: 'false', appearance: null }],
    choices: {
      bildung: [
        { name: '1', label: 'Kein Abschluss' },
        { name: '2', label: 'Abitur' },
        { name: '3', label: 'Hochschulabschluss' },
      ],
    },
  },
  {
    id: 'multiple_choice',
    survey: [{ type: 'select_multiple tage', name: 'wochenende', label: 'Wochenendtage', required: 'false', appearance: null }],
    choices: {
      tage: [
        { name: 'sa', label: 'Samstag' },
        { name: 'so', label: 'Sonntag' },
      ],
    },
  },
  {
    id: 'single_choice_other',
    survey: [
      { type: 'select_one quelle', name: 'src', label: 'Source', required: 'false', appearance: null },
      { type: 'text', name: 'src_other', label: 'Other', required: 'false', appearance: null },
    ],
    choices: {
      quelle: [
        { name: 'a', label: 'A' },
        { name: 'other', label: 'Other' },
      ],
    },
  },
  {
    id: 'multiple_choice_other',
    survey: [
      { type: 'select_multiple dev', name: 'own', label: 'Own', required: 'false', appearance: null },
      { type: 'text', name: 'own_other', label: 'Other', required: 'false', appearance: null },
    ],
    choices: {
      dev: [
        { name: 'a', label: 'A' },
        { name: 'other', label: 'Other' },
      ],
    },
  },
  {
    id: 'grid',
    survey: [
      { type: 'begin_group', name: 'trust', label: 'Trust', required: 'false', appearance: 'table-list' },
      { type: 'select_one s5', name: 'trust_a', label: 'A', required: 'false', appearance: null },
      { type: 'end_group', name: null, label: null, required: null, appearance: null },
    ],
    choices: {
      s5: [
        { name: '1', label: 'One' },
        { name: '2', label: 'Two' },
      ],
    },
  },
  {
    id: 'integer',
    survey: [{ type: 'integer', name: 'alter', label: 'Alter', required: 'false', appearance: null }],
    choices: {},
  },
  {
    id: 'decimal',
    survey: [{ type: 'decimal', name: 'rating', label: 'Rating', required: 'false', appearance: null }],
    choices: {},
  },
  {
    id: 'range',
    survey: [{ type: 'range', name: 'score', label: 'Score', required: 'false', appearance: null }],
    choices: {},
  },
  {
    id: 'date',
    survey: [{ type: 'date', name: 'besuch', label: 'Besuchsdatum', required: 'false', appearance: null }],
    choices: {},
  },
  {
    id: 'text',
    survey: [{ type: 'text', name: 'anmerkung', label: 'Anmerkungen', required: 'false', appearance: null }],
    choices: {},
  },
  {
    id: 'note',
    survey: [{ type: 'note', name: 'thanks', label: 'Thank you', required: 'false', appearance: null }],
    choices: {},
    // After the swap, both sides run through formtransform, so a note row
    // emits no <var> on either side (it folds into <notes>). The Go converter
    // that used to emit a <var> is gone — see qwacback#3.
  },
  {
    id: 'single_choice_long_list',
    survey: [{ type: 'select_one_from_file iso_3166_1.csv', name: 'country', label: 'Country', required: 'false', appearance: null }],
    choices: {},
  },
  {
    id: 'multiple_choice_long_list',
    survey: [{ type: 'select_multiple_from_file iso_3166_1.csv', name: 'visited', label: 'Visited', required: 'false', appearance: null }],
    choices: {},
  },
  {
    id: 'section',
    survey: [
      { type: 'begin_group', name: 'section1', label: 'Section 1', required: 'false', appearance: null },
      { type: 'text', name: 'q1', label: 'Q1', required: 'false', appearance: null },
      { type: 'end_group', name: null, label: null, required: null, appearance: null },
    ],
    choices: {},
  },
];

function findDataDscr(xmlStr) {
  // Same logic as formtransform's Python test: find <dataDscr> in codeBook,
  // or treat a bare <var>/<varGrp> root as a single-element container.
  const codeBookMatch = xmlStr.match(/<dataDscr[^>]*>([\s\S]*?)<\/dataDscr>/);
  if (codeBookMatch) return codeBookMatch[1];
  const varMatch = xmlStr.match(/<var[\s>][\s\S]*?<\/var>/);
  if (varMatch) return varMatch[0];
  const grpMatch = xmlStr.match(/<varGrp[\s>][\s\S]*?<\/varGrp>/);
  if (grpMatch) return grpMatch[0];
  return null;
}

function shapeVars(xmlStr) {
  const inner = findDataDscr(xmlStr) ?? '';
  const vars = [];
  for (const m of inner.matchAll(/<var[\s>][\s\S]*?<\/var>/g)) {
    const id = (m[0].match(/ID="([^"]+)"/) || [])[1];
    const name = (m[0].match(/\sname="([^"]+)"/) || [])[1];
    const intrvl = (m[0].match(/intrvl="([^"]+)"/) || [])[1];
    const rdt = (m[0].match(/responseDomainType="([^"]+)"/) || [])[1];
    vars.push({ id, name, intrvl, responseDomainType: rdt });
  }
  return vars;
}

function shapeGroups(xmlStr) {
  const inner = findDataDscr(xmlStr) ?? '';
  const groups = [];
  for (const m of inner.matchAll(/<varGrp[\s>][\s\S]*?<\/varGrp>/g)) {
    const id = (m[0].match(/ID="([^"]+)"/) || [])[1];
    const name = (m[0].match(/\sname="([^"]+)"/) || [])[1];
    const type = (m[0].match(/type="([^"]+)"/) || [])[1];
    const v = (m[0].match(/\svar="([^"]+)"/) || [])[1];
    groups.push({ id, name, type, var: v });
  }
  return groups;
}

function payloadFor(test) {
  // buildDdiXml wants survey/choices rows in flat form; qwacback wants the
  // same shape (it forwards JSON to the ddi-emitter).
  const choices = [];
  for (const [listName, list] of Object.entries(test.choices)) {
    for (const c of list) {
      choices.push({ list_name: listName, ...c });
    }
  }
  return { survey: test.survey, choices, settings: {} };
}

async function login() {
  const resp = await fetch(`${QWACBACK_URL}/api/collections/_superusers/auth-with-password`, {
    method: 'POST',
    headers: { 'content-type': 'application/json' },
    body: JSON.stringify({ identity: QWACBACK_EMAIL, password: QWACBACK_PASSWORD }),
  });
  if (!resp.ok) {
    throw new Error(`login failed: HTTP ${resp.status}: ${await resp.text()}`);
  }
  const data = await resp.json();
  if (!data.token) {
    throw new Error('login response missing token');
  }
  return data.token;
}

let qbToken = null;
async function ensureLoggedIn() {
  if (!qbToken) qbToken = await login();
  return qbToken;
}

async function callQwacback(payload) {
  const token = await ensureLoggedIn();
  const resp = await fetch(`${QWACBACK_URL}/api/convert/xlsform-to-ddi`, {
    method: 'POST',
    headers: {
      'content-type': 'application/json',
      'authorization': token,
    },
    body: JSON.stringify(payload),
  });
  if (!resp.ok) {
    throw new Error(`HTTP ${resp.status}: ${await resp.text()}`);
  }
  return resp.text();
}

async function callFormtransform(payload) {
  return buildDdiXml(payload.survey, payload.choices);
}

let passed = 0;
let xfailed = 0;
let failed = 0;
const failures = [];

// Authenticated as superuser, qwacback allows 30 conversions/minute; 3s
// spacing keeps the run under 1min total.
const QWACBACK_DELAY_MS = 3000;

for (const test of EQUIVALENT_TYPES) {
  const payload = payloadFor(test);
  try {
    const ftDdi = await callFormtransform(payload);
    const qbDdi = await callQwacback(payload);

    const ftVars = shapeVars(ftDdi);
    const qbVars = shapeVars(qbDdi);
    const ftGroups = shapeGroups(ftDdi);
    const qbGroups = shapeGroups(qbDdi);

    const sameVars = JSON.stringify(ftVars) === JSON.stringify(qbVars);
    const sameGroups = JSON.stringify(ftGroups) === JSON.stringify(qbGroups);

    if (sameVars && sameGroups) {
      console.log(`✓ ${test.id}: ${qbVars.length} vars, ${qbGroups.length} groups (match)`);
      passed++;
    } else {
      console.log(`✗ ${test.id}: MISMATCH`);
      console.log(`  formtransform vars: ${JSON.stringify(ftVars)}`);
      console.log(`  qwacback vars:      ${JSON.stringify(qbVars)}`);
      console.log(`  formtransform grps: ${JSON.stringify(ftGroups)}`);
      console.log(`  qwacback grps:      ${JSON.stringify(qbGroups)}`);
      failed++;
      failures.push(test.id);
    }
  } catch (e) {
    console.log(`✗ ${test.id}: ERROR: ${e.message}`);
    failed++;
    failures.push(test.id);
  }
  // Pace requests to stay under the rate limit.
  await new Promise((r) => setTimeout(r, QWACBACK_DELAY_MS));
}

console.log('');
console.log(`Total: ${EQUIVALENT_TYPES.length}, passed: ${passed}, xfailed: ${xfailed}, failed: ${failed}`);
if (failed > 0) {
  console.log(`Failures: ${failures.join(', ')}`);
  process.exit(1);
}
