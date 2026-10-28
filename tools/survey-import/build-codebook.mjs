#!/usr/bin/env node
// Builds a DDI codebook for seed_data/ from an XLSForm-as-JSON file plus
// study metadata, with the formtransform release qwacback pins
// (ddi-emitter/node_modules; run `npm ci` there first).
//
//   node tools/survey-import/build-codebook.mjs out/q13-attari/form.json [out.xml]
//
// form.json: {"study": {...}, "survey": [...], "choices": [...], "settings": {...}}
// See PLAN.md for the study fields. Writes codebook.xml next to form.json
// unless an output path is given, then converts it back with ddiToXlsform and
// prints the warnings, so lost fields show up before the file is seeded.
import { readFileSync, writeFileSync } from 'node:fs';
import { dirname, join } from 'node:path';
import { fileURLToPath } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));
const ft = await import(
  join(here, '../../ddi-emitter/node_modules/@correlaid/formtransform/dist/index.js')
);

const [input, output = join(dirname(input ?? '.'), 'codebook.xml')] = process.argv.slice(2);
if (!input) {
  console.error('usage: build-codebook.mjs form.json [codebook.xml]');
  process.exit(2);
}
const form = JSON.parse(readFileSync(input, 'utf8'));
const study = form.study ?? {};
const settings = Array.isArray(form.settings) ? (form.settings[0] ?? {}) : (form.settings ?? {});
if (!settings.form_title && study.title) settings.form_title = study.title;
if (!settings.form_id && study.id) settings.form_id = study.id;

let xml;
try {
  xml = ft.xlsformToDdi(
    { surveyData: form.survey ?? [], choicesData: form.choices ?? [] },
    { settings, onWarning: (d) => console.warn(`warning: ${d.code}: ${d.message}`) },
  );
} catch (e) {
  if (e instanceof ft.ConversionError) {
    for (const d of e.details.length ? e.details : [e]) console.error(`error: ${d.message}`);
    process.exit(1);
  }
  throw e;
}

const esc = (s) =>
  String(s).replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');
const el = (name, text, attrs = {}) => {
  if (text == null || text === '') return '';
  const a = Object.entries(attrs)
    .filter(([, v]) => v != null && v !== '')
    .map(([k, v]) => ` ${k}="${esc(v)}"`)
    .join('');
  return `<${name}${a}>${esc(text)}</${name}>`;
};
const insert = (after, content) => {
  if (!content) return;
  const i = xml.indexOf(after);
  if (i < 0) throw new Error(`build-codebook: no ${after} in formtransform output`);
  xml = xml.slice(0, i + after.length) + content + xml.slice(i + after.length);
};

// Inserted in DDI 2.5 sequence order: citation (titlStmt, rspStmt, prodStmt,
// …, biblCit, holdings), then stdyInfo and dataAccs before stdyDscr/notes.
// titlStmt: titl, subTitl, altTitl, then formtransform's parTitl and IDNo.
insert('</titl>', el('subTitl', study.subtitle) + el('altTitl', study.shortTitle));
const authors = (study.authors ?? []).map((a) =>
  typeof a === 'string' ? el('AuthEnty', a) : el('AuthEnty', a.name, { affiliation: a.affiliation }),
);
insert('</titlStmt>', authors.length ? `<rspStmt>${authors.join('')}</rspStmt>` : '');
insert(
  '<prodStmt>',
  el('producer', study.producer?.name ?? study.producer, { affiliation: study.producer?.affiliation }) +
    el('copyright', study.copyright),
);
insert(
  '</prodStmt>',
  el('biblCit', study.citation) + el('holdings', study.sourceLabel ?? study.url, { URI: study.url }),
);
const subject = [
  ...(study.keywords ?? []).map((k) => el('keyword', k)),
  ...(study.topics ?? []).map((t) => el('topcClas', t)),
].join('');
const sumDscr =
  el('timePrd', study.timePrd, { event: 'single' }) +
  el('nation', study.nation) +
  el('anlyUnit', study.anlyUnit) +
  el('universe', study.universe, { clusion: 'I' }) +
  el('dataKind', study.dataKind);
const stdyInfo =
  (subject ? `<subject>${subject}</subject>` : '') +
  el('abstract', study.abstract) +
  (sumDscr ? `<sumDscr>${sumDscr}</sumDscr>` : '');
const dataAccs = study.license
  ? `<dataAccs><useStmt>${el('conditions', study.license)}</useStmt></dataAccs>`
  : '';
insert('</citation>', (stdyInfo ? `<stdyInfo>${stdyInfo}</stdyInfo>` : '') + dataAccs);

writeFileSync(output, xml);
console.log(`wrote ${output}`);

// Round trip: what qwacback's /xlsform export will give back.
const back = ft.ddiToXlsform(xml);
const rows = (f) => (f.survey ?? []).filter((r) => !/^(begin|end)[ _]/.test(r.type ?? '')).length;
console.log(`round trip: ${rows(form)} survey rows in, ${rows(back)} out`);
for (const w of back.warnings ?? []) console.warn(`round-trip warning: ${w.code}: ${w.message}`);
