#!/usr/bin/env node
// Formtransform equivalence test against qwacback (formtransform#14, qwacback#3).
//
// Part 1 compares formtransform's xlsformToDdi output with qwacback's
// POST /api/convert/xlsform-to-ddi output. For each case, the same XLSForm goes
// through both, and every <var>/<varGrp> must match after normalization
// (whitespace, self-closing tags, entity spelling, the `files` IDREF and the
// namespace declaration) — not only the var/group shape. qwacback passes the
// sidecar's elements through, so any difference is a qwacback bug (#14).
//
// Part 2 (#37) imports codebooks (the seed files and one formtransform wrote)
// and checks that qwacback gives them back: /api/studies/{id}/export is the
// file byte for byte, and /api/studies/{id}/xlsform and
// /api/convert/ddi-to-xlsform equal formtransform's ddiToXlsform of the file.

// Path to a built formtransform dist/. Defaults to the version ddi-emitter
// pins (run `npm ci` in ddi-emitter/ first), so both sides run the same code.
const FORMTRANSFORM_DIST = process.env.FORMTRANSFORM_DIST
  || new URL('../ddi-emitter/node_modules/@correlaid/formtransform/dist/index.js', import.meta.url).pathname;

const { xlsformToDdi, ddiToXlsform } = await import(FORMTRANSFORM_DIST);
const { readFileSync, readdirSync } = await import('node:fs');

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
    // A note emits no <var>; formtransform folds its text into the next
    // question's <preQTxt>. (A form of only notes is a 400 in qwacback.)
    survey: [
      { type: 'note', name: 'intro', label: 'Please answer honestly', required: 'false', appearance: null },
      { type: 'integer', name: 'alter', label: 'Alter', required: 'false', appearance: null },
    ],
    choices: {},
  },
  {
    id: 'hint_and_guidance',
    // hint → <postQTxt> (v0.7.0), guidance_hint → <ivuInstr> (v0.2.0)
    // (qwacback#12); both must reach the client unchanged.
    survey: [{ type: 'integer', name: 'alter', label: 'Alter', hint: 'In Jahren', parameters: 'guidance_hint=Nachfragen', required: 'false', appearance: null }],
    choices: {},
  },
  {
    id: 'relevant_and_required',
    survey: [
      { type: 'select_one yn', name: 'hund', label: 'Hund?', required: 'yes', appearance: null },
      { type: 'text', name: 'hundname', label: 'Name des Hundes', relevant: "${hund} = 'ja'", required: 'false', appearance: null },
    ],
    choices: { yn: [{ name: 'ja', label: 'Ja' }, { name: 'nein', label: 'Nein' }] },
  },
  {
    id: 'special_characters',
    survey: [{ type: 'text', name: 'zitat', label: 'Was heißt „<b>“ & \'x\' "y"?', required: 'false', appearance: null }],
    choices: {},
  },
  {
    id: 'multilingual',
    // formtransform#135: base language untagged and first, then one
    // xml:lang sibling per other language; qwacback forwards label::<lang>
    // columns and default_language unchanged.
    survey: [
      { type: 'select_one yn', name: 'hund', 'label::Deutsch (de)': 'Haben Sie einen Hund?', 'label::English (en)': 'Do you have a dog?', 'hint::Deutsch (de)': 'Auch Mitbewohner', 'hint::English (en)': 'Flatmates count', required: 'false', appearance: null },
    ],
    choices: {
      yn: [
        { name: 'ja', 'label::Deutsch (de)': 'Ja', 'label::English (en)': 'Yes' },
        { name: 'nein', 'label::Deutsch (de)': 'Nein', 'label::English (en)': 'No' },
      ],
    },
    settings: { default_language: 'Deutsch (de)' },
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

// Every <var>/<varGrp> in the fragment, normalized so that serializer
// differences between formtransform (JS) and qwacback (Go encoding/xml) vanish
// but content differences don't.
function normalizedElements(xmlStr) {
  const inner = findDataDscr(xmlStr) ?? '';
  const out = [];
  for (const m of inner.matchAll(/<(var|varGrp)[\s>][\s\S]*?<\/\1>/g)) {
    out.push(
      m[0]
        // qwacback puts codeBook/@xml:lang on the fragment root; formtransform
        // keeps it on the <codeBook>. Compare the elements without it.
        .replace(/^(<(?:var|varGrp)\b[^>]*?)\s+xml:lang="[^"]*"/, '$1')
        .replace(/\s+files="[^"]*"/g, '')
        .replace(/\s+xmlns(:\w+)?="[^"]*"/g, '')
        .replace(/<([\w:]+)((?:\s+[\w:]+="[^"]*")*)\s*\/>/g, '<$1$2></$1>')
        .replace(/>\s+</g, '><')
        .replace(/&#34;|&quot;/g, '"')
        .replace(/&#39;|&apos;/g, "'")
        .replace(/&#xA;/g, '\n')
        .replace(/&#x9;/g, '\t')
        .trim(),
    );
  }
  return out;
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
  // xlsformToDdi wants survey/choices rows in flat form; qwacback wants the
  // same shape (it forwards JSON to the ddi-emitter).
  const choices = [];
  for (const [listName, list] of Object.entries(test.choices)) {
    for (const c of list) {
      choices.push({ list_name: listName, ...c });
    }
  }
  return { survey: test.survey, choices, settings: test.settings ?? {} };
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
  return xlsformToDdi({ surveyData: payload.survey, choicesData: payload.choices }, { settings: payload.settings, onWarning: () => {} });
}

let passed = 0;
let xfailed = 0;
let failed = 0;
const failures = [];

// Authenticated as superuser, qwacback allows 30 conversions/minute; 3s
// spacing keeps the run under 1min total.
const QWACBACK_DELAY_MS = Number(process.env.QWACBACK_DELAY_MS ?? 3000);

for (const test of EQUIVALENT_TYPES) {
  const payload = payloadFor(test);
  try {
    const ftDdi = await callFormtransform(payload);
    const qbDdi = await callQwacback(payload);

    const ftVars = shapeVars(ftDdi);
    const qbVars = shapeVars(qbDdi);
    const ftGroups = shapeGroups(ftDdi);
    const qbGroups = shapeGroups(qbDdi);

    const ftElems = normalizedElements(ftDdi);
    const qbElems = normalizedElements(qbDdi);
    const same = ftElems.length > 0 && JSON.stringify(ftElems) === JSON.stringify(qbElems);

    if (same) {
      console.log(`✓ ${test.id}: ${qbVars.length} vars, ${qbGroups.length} groups (identical content)`);
      passed++;
    } else {
      console.log(`✗ ${test.id}: MISMATCH`);
      console.log(`  formtransform vars: ${JSON.stringify(ftVars)}`);
      console.log(`  qwacback vars:      ${JSON.stringify(qbVars)}`);
      console.log(`  formtransform grps: ${JSON.stringify(ftGroups)}`);
      console.log(`  qwacback grps:      ${JSON.stringify(qbGroups)}`);
      for (let i = 0; i < Math.max(ftElems.length, qbElems.length); i++) {
        if (ftElems[i] !== qbElems[i]) {
          console.log(`  first differing element #${i}:`);
          console.log(`    formtransform: ${ftElems[i]}`);
          console.log(`    qwacback:      ${qbElems[i]}`);
          break;
        }
      }
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

// ---- Part 2: codebooks back out (#37) ----

// A formtransform codebook with what the records don't model: a section,
// skip logic, a constraint, a hint, question order.
const FT_CODEBOOK = xlsformToDdi({
  surveyData: [
    { type: 'begin_group', name: 'teil1', label: 'Teil 1' },
    { type: 'integer', name: 'alter', label: 'Wie alt sind Sie?', hint: 'In Jahren', constraint: '. < 120', required: 'yes' },
    { type: 'select_one ja_nein', name: 'rente', label: 'Sind Sie in Rente?', relevant: '${alter} > 60' },
    { type: 'end_group' },
  ],
  choicesData: [
    { list_name: 'ja_nein', name: '1', label: 'Ja' },
    { list_name: 'ja_nein', name: '2', label: 'Nein' },
  ],
}, { settings: { form_title: 'Equivalence round trip', form_id: 'equivalence_rt' }, onWarning: () => {} });

const seedDir = new URL('../seed_data/', import.meta.url);
const CODEBOOKS = [
  ...readdirSync(seedDir).filter((f) => f.endsWith('.xml')).sort()
    .map((f) => ({ id: `seed ${f}`, xml: readFileSync(new URL(f, seedDir), 'utf8') })),
  { id: 'formtransform codebook', xml: FT_CODEBOOK },
];

// The sheets only: warnings are compared by code, not message wording.
function sheets(form) {
  return JSON.stringify({ survey: form.survey, choices: form.choices, settings: form.settings });
}

async function qwacbackJSON(path, init = {}) {
  const resp = await fetch(`${QWACBACK_URL}${path}`, init);
  if (!resp.ok) throw new Error(`${path}: HTTP ${resp.status}: ${await resp.text()}`);
  return resp.json();
}

for (const cb of CODEBOOKS) {
  try {
    const token = await ensureLoggedIn();
    const body = new FormData();
    body.append('file', new Blob([cb.xml], { type: 'application/xml' }), 'codebook.xml');
    const imported = await qwacbackJSON('/api/import', { method: 'POST', headers: { authorization: token }, body });
    if (!imported.imported) throw new Error(`import: ${JSON.stringify(imported)}`);
    const id = imported.study_id;

    const want = ddiToXlsform(cb.xml, { onWarning: () => {} });
    const problems = [];

    const exported = await (await fetch(`${QWACBACK_URL}/api/studies/${id}/export`)).text();
    if (exported !== cb.xml) problems.push('export differs from the imported file');

    const studyForm = await qwacbackJSON(`/api/studies/${id}/xlsform`);
    if (sheets(studyForm) !== sheets(want)) problems.push('/api/studies/{id}/xlsform differs from ddiToXlsform');

    const converted = await qwacbackJSON('/api/convert/ddi-to-xlsform', {
      method: 'POST', headers: { 'content-type': 'application/xml', authorization: token }, body: cb.xml,
    });
    if (sheets(converted) !== sheets(want)) problems.push('/api/convert/ddi-to-xlsform differs from ddiToXlsform');

    if (problems.length === 0) {
      console.log(`✓ ${cb.id}: export identical, ${want.survey.length} survey rows identical`);
      passed++;
    } else {
      console.log(`✗ ${cb.id}: ${problems.join('; ')}`);
      failed++;
      failures.push(cb.id);
    }
  } catch (e) {
    console.log(`✗ ${cb.id}: ERROR: ${e.message}`);
    failed++;
    failures.push(cb.id);
  }
  await new Promise((r) => setTimeout(r, QWACBACK_DELAY_MS));
}

console.log('');
console.log(`Total: ${EQUIVALENT_TYPES.length + CODEBOOKS.length}, passed: ${passed}, xfailed: ${xfailed}, failed: ${failed}`);
if (failed > 0) {
  console.log(`Failures: ${failures.join(', ')}`);
  process.exit(1);
}
