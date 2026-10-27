// Generates tools/survey-import/out/q20-eu-ict-ent-ai/form.json from a verbatim
// transcription of Module E "Artificial Intelligence" of the 2025 model
// questionnaire (ICT-ENT 2025 - Model Questionnaire V1.5) as printed in
// Annex 1 of Eurostat's "European businesses statistics compilers' manual for
// ICT usage and e-commerce in enterprises, 2025 survey" (KS-01-25-065,
// 17 December 2025). Only Annex 1's literal text is used: question texts,
// item texts, printed instructions/examples and, for E1, Annex 1's glossary
// definitions of the named techniques. Line-end hyphenation of the PDF is
// undone ("tech-/nologies" -> "technologies").
// verify-labels.py checks every text in form.json against Annex 1 of the PDF.
//
//   node gen-form.mjs form.json
import { writeFileSync } from 'node:fs';

const out = process.argv[2];
const S = [];
const C = [];
const row = (r) => S.push(r);
const begin = (name, label, extra = {}) => row({ type: 'begin_group', name, label, ...extra });
const end = () => row({ type: 'end_group' });
const choices = (list, items) => items.forEach(([name, label]) => C.push({ list_name: list, name, label }));
const L = (...lines) => lines.join('\n');

// Annex 1 prints only "Yes" / "No" tick boxes. Codes 1/0 follow the manual's
// data transmission codes for the AI_* variables ("1 Yes", "0 No").
choices('yes_no', [['1', 'Yes'], ['0', 'No']]);

const E1 = ['a', 'b', 'c', 'd', 'e', 'f', 'g', 'h'].map((x) => `\${e1_${x}}`);
const anyAI = E1.map((v) => `${v} = '1'`).join(' or ');
const noAI = E1.map((v) => `${v} = '0'`).join(' and ');

const grid = (q, letters) =>
  letters.forEach(([x, label, hint]) =>
    row({ type: 'select_one yes_no', name: `${q}_${x}`, label, ...(hint ? { hint } : {}) }),
  );

begin('module_e', 'Module E: Artificial Intelligence', {
  hint: '(Scope: enterprises with access to the internet, i.e. if A1>0)',
});

row({
  type: 'note',
  name: 'e_intro',
  label: L(
    'Artificial intelligence refers to systems that use technologies such as: text mining, computer vision, speech recognition, natural language generation, machine learning, deep learning to gather and/or use data to predict, recommend or decide, with varying levels of autonomy, the best action to achieve specific goals.',
    'Artificial intelligence systems can be purely software based, e.g.:',
    '• chatbots and business virtual assistants based on natural language processing;',
    '• face recognition systems based on computer vision or speech recognition systems;',
    '• machine translation software;',
    '• data analysis based on machine learning;',
    'or embedded in devices, e.g.:',
    '• autonomous robots for warehouse automation or production assembly works;',
    '• autonomous drones for production surveillance or parcel handling.',
  ),
});

// ---- E1 (mandatory)
begin('e1', 'Does your enterprise use any of the following Artificial Intelligence (AI) technologies?', {
  appearance: 'field-list',
  hint: 'E1 a) to h) = “No” then go to E4 (if optional included) else go to F1.',
});
grid('e1', [
  ['a', 'a) AI technologies performing analysis of written language (e.g. text mining)',
    'Text mining refers to the use of advanced techniques for automated detection of patterns in (large) texts.'],
  ['b', 'b) AI Technologies converting spoken language into machine-readable format (speech recognition)',
    'Speech recognition is the ability of a machine or program to identify words and phrases in spoken language and convert them to a machine-readable format.'],
  ['c', 'c) AI Technologies generating written or spoken language (natural language generation, speech synthesis)',
    'Natural language generation is the ability for a computer program to convert data into natural language representation.'],
  ['d', 'd) AI Technologies generating pictures, videos, sound/audio'],
  ['e', 'e) AI Technologies identifying objects or persons based on images or videos (image recognition, image processing)'],
  ['f', 'f) Machine learning (e.g. deep learning) for data analysis',
    'Machine learning (e.g. deep learning) involves ‘training’ a computer model to better perform an automated task, e.g. pattern recognition.'],
  ['g', 'g) AI Technologies automating different workflows or assisting in decision making (e.g. AI based software robotic process automation)',
    'Artificial Intelligence based robotic process automation refers to software that automates business processes (e.g. workflows automation) based on Artificial Intelligence technologies.'],
  ['h', 'h) AI Technologies enabling physical movement of machines via autonomous decisions based on observation of surroundings (autonomous robots, self-driving vehicles, autonomous drones)'],
]);
end();

// ---- E2 (mandatory in 2025; AI users)
const ex = 'some of the examples may be:';
begin('e2', 'Does your enterprise use Artificial Intelligence software or systems for any of the following purposes?', {
  appearance: 'field-list',
  relevant: anyAI,
});
grid('e2', [
  ['a', 'a) Use of AI for marketing or sales', L(ex,
    '• customer profiling, price optimisation, personalised marketing offers, market analysis based on machine learning',
    '• chatbots based on natural language processing for customer support',
    '• autonomous robots for orders processing')],
  ['b', 'b) Use of AI for production or service processes', L(ex,
    '• predictive maintenance or process optimization based on machine learning',
    '• tools to classify products or find defects in products based on computer vision',
    '• autonomous drones for production surveillance, security or inspection tasks',
    '• assembly works performed by autonomous robots')],
  ['c', 'c) Use of AI for organisation of business administration processes or management', L(ex,
    '• business virtual assistants based on machine learning and/or natural language processing, e.g. for document drafting',
    '• data analysis or strategic decision making based on machine learning, e.g. risk assessment based on machine learning',
    '• planning or business forecasting based on machine learning',
    '• human resources management based on machine learning or natural language processing, e.g. candidates pre-selection screening, employee profiling or performance analysis')],
  ['d', 'd) Use of AI for logistics', L(ex,
    '• autonomous robots for pick-and-pack solutions in warehouses for parcel shipping, tracing, distribution or sorting',
    '• route optimization based on machine learning')],
  ['e', 'e) Use of AI for ICT security', L(ex,
    '• face recognition based on computer vision for authentication of ICT users',
    '• detection and prevention of cyber-attacks based on machine learning')],
  ['f', 'f) Use of AI for accounting, controlling or finance management', L(ex,
    '• machine learning to analyse data that helps to make financial decisions',
    '• invoice processing based on machine learning',
    '• machine learning or natural language processing used for bookkeeping tasks')],
  ['g', 'g) Use of AI for research and development (R&D) or innovation activity (excluding research on AI)', L(ex,
    '• analysis of data for conducting research, solving research problems, developing a new or significantly improved product/service based on machine learning')],
]);
end();

// ---- E3 (optional; AI users)
begin('e3', 'How did you enterprise acquire the Artificial Intelligence software or systems that it uses?', {
  appearance: 'field-list',
  relevant: anyAI,
  hint: '– Optional',
});
grid('e3', [
  ['a', 'a) They were developed by own employees (including those employed in parent or affiliate enterprise)'],
  ['b', 'b) Commercial software or systems were modified by own employees (including those employed in parent or affiliate enterprise)'],
  ['c', 'c) Open-source software or systems were modified by own employees (including those employed in parent or affiliate enterprise)'],
  ['d', 'd) Commercial software or systems ready to use were purchased (including examples where it was already incorporated in a purchased item or system)'],
  ['e', 'e) External providers were contracted to develop or modify them'],
]);
end();

// ---- E4, E5 (optional; non-users)
row({
  type: 'select_one yes_no',
  name: 'e4',
  label: 'Has your enterprise ever considered using any of the Artificial Intelligence technologies listed in question E1?',
  hint: L(
    'Questions E4 is presented only to respondents who answered ‘No’ to E1a)-h) i.e. enterprises that did not use any of the Artificial Intelligence technologies listed in question E1.',
    '– Optional',
    '(Filter question)',
  ),
  relevant: noAI,
});
begin('e5', 'What are the reasons for your enterprise not to use any of the Artificial Intelligence technologies listed in question E1?', {
  appearance: 'field-list',
  relevant: `\${e4} = '1'`,
  hint: '– Optional',
});
grid('e5', [
  ['a', 'a) The costs seem too high'],
  ['b', 'b) There is a lack of relevant expertise in the enterprise'],
  ['c', 'c) Incompatibility with existing equipment, software or systems'],
  ['d', 'd) Difficulties with availability or quality of the necessary data'],
  ['e', 'e) Concerns regarding violation of data protection and privacy'],
  ['f', 'f) Lack of clarity about the legal consequences (e.g. liability in case of damage caused by the use of Artificial Intelligence)'],
  ['g', 'g) Ethical considerations'],
  ['h', 'h) Artificial Intelligence technologies are not useful for the enterprise'],
]);
end();

end(); // module_e

const URL_DOC = 'https://ec.europa.eu/eurostat/web/products-manuals-and-guidelines/w/ks-01-25-065';

const study = {
  title: 'EU Survey on ICT Usage and e-Commerce in Enterprises 2025: AI module (Model Questionnaire, Module E)',
  id: 'q20-eu-ict-ent-ai',
  authors: [{ name: 'Eurostat', affiliation: 'European Commission, Eurostat' }],
  producer: { name: 'Eurostat' },
  citation: `Eurostat. (2025, December 17). European businesses statistics compilers' manual for ICT usage and e-commerce in enterprises, 2025 survey (2025 edition; Manuals and guidelines, KS-01-25-065). Publications Office of the European Union. ${URL_DOC}`,
  url: URL_DOC,
  sourceLabel: 'KS-01-25-065, Annex 1 – Model questionnaire (ICT-ENT 2025 V1.5), Module E "Artificial Intelligence" (Eurostat)',
  abstract:
    'This study contains only Module E (Artificial Intelligence) of the Eurostat model questionnaire for the 2025 EU survey on ICT usage and e-commerce in enterprises, as printed in the 2025 compilers\' manual; national statistical institutes adapt it for their own surveys. The enterprise reports which AI technologies it uses, for which business purposes, and how it acquired the AI software or systems. Enterprises not using AI are asked whether they ever considered it and, if so, why they do not use it (for example costs, lack of expertise, data protection concerns or unclear legal consequences).',
  keywords: ['artificial intelligence', 'AI adoption', 'barriers to AI use', 'machine learning', 'data protection', 'enterprises', 'ICT usage', 'official statistics'],
  topics: ['Information society', 'Digital economy', 'Artificial intelligence'],
  nation: 'European Union',
  universe:
    'Enterprises with 10 or more employees and self-employed persons (NACE Rev. 2 sections C–J, L–N and group 95.1; optional: 0–9). Module E: enterprises with access to the internet.',
  anlyUnit: 'Enterprises',
  timePrd: '2025',
  dataKind: 'Model questionnaire',
  license:
    '© European Union, 2025. Manual KS-01-25-065: "Except otherwise noted, the reuse of this document is authorised under a Creative Commons Attribution 4.0 International (CC-BY 4.0) licence https://creativecommons.org/licenses/by/4.0/). This means that reuse is allowed provided appropriate credit is given and any changes are indicated." Eurostat copyright notice (https://ec.europa.eu/eurostat/help/copyright-notice): "The copyright for the editorial content of this website, which is owned by the EU, is licensed under the Creative Commons Attribution 4.0 International licence. This means that you can re-use the content provided you acknowledge the source and indicate any changes you have made." "Reuse of statistical data, metadata, publications, and other dissemination tools published on this website for commercial or non-commercial purposes is authorised provided the source is acknowledged. The reuse policy of the European Commission is implemented by the Decision of 12 December 2011." (Commission Decision 2011/833/EU). "When reuse involves translations of publications or modifications to the data or text, this must be stated clearly to the end user of the information. A disclaimer regarding the non-responsibility of Eurostat shall be included." ' +
    'Changes: format conversion of Module E "Artificial Intelligence" of the 2025 model questionnaire (Annex 1 of manual KS-01-25-065) to DDI; wording unchanged; all other modules, explanatory notes and layout omitted; glossary definitions of the named AI techniques attached to the E1 items as hints. Disclaimer: this DDI version was made by the qwac project (CorrelAid), not by Eurostat; Eurostat bears no responsibility for it.',
  copyright: '© European Union, 2025 (Eurostat)',
};

writeFileSync(
  out,
  JSON.stringify({ study, survey: S, choices: C, settings: { default_language: 'English (en)', form_title: study.title, form_id: study.id } }, null, 2) + '\n',
);
console.log(`rows: ${S.length}, choices: ${C.length}`);
