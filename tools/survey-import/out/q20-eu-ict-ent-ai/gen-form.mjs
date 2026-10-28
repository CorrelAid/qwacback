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
  appearance: 'table-list',
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
  appearance: 'table-list',
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
  appearance: 'table-list',
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
  appearance: 'table-list',
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

// German: the wording Destatis uses for these questions in its national
// questionnaire, as printed in the quality report "IKT-Erhebung in
// Unternehmen 2025" (questionnaire pp. 6–8, PDF pp. 22–24). Destatis drops
// Eurostat's E3 (how AI was acquired), so E3 has no German text, and its
// E3/E4 are Eurostat's E4/E5; mapped by content. Soft hyphens, line-end
// hyphenation and footnote numbers are removed; item numbers and the
// "weiter mit Frage …" routing lines (Destatis numbering) are left out.
// verify-labels-de.py checks every German text against the PDF.
const DE = {
  module_e: { label: 'Künstliche Intelligenz (KI)' },
  e_intro: {
    label: L(
      'Künstliche Intelligenz (KI) bezieht sich auf Systeme, die Technologien wie Text Mining, Computer Vision, Spracherkennung, Generierung natürlicher Sprache, maschinelles Lernen und Deep Learning nutzen, um Daten zu sammeln, zu verwenden und/oder zu produzieren, um mit einem unterschiedlichen Grad an Autonomie z. B. Vorhersagen, Empfehlungen oder Entscheidungen zu treffen.',
      'Künstliche Intelligenz kann rein softwarebasiert sein, z. B.:',
      '– Systeme,die Inhalte erstellen (generative KI)',
      '– Chatbots und virtuelle Assistenten, die auf der Verarbeitung natürlicher Sprache basieren,',
      '– Gesichtserkennungsysteme, die auf maschinellem Sehen oder Spracherkennungssystemen basieren',
      '– Datenanalyse basierend auf maschinellem Lernen',
      'oder in Geräten eingebettet sein z. B.',
      '– Autonome Roboter für Lagerautomatisierung oder Montagearbeiten',
      '– Autonome Drohnen zur Produktionsüberwachung oder Paketabfertigung',
    ),
  },
  e1: { label: 'Nutzt Ihr Unternehmen eine der folgenden Technologien der künstlichen Intelligenz (KI)?' },
  e1_a: { label: 'Technologien zur Analyse von Schriftsprache (Text Mining)' },
  e1_b: { label: 'Technologien zur Umwandlung gesprochener Sprache in maschinenlesbare Form (Spracherkennung)' },
  e1_c: { label: 'Technologien, die geschriebene oder gesprochene Sprache oder Programmiercodes erzeugen (Erzeugung natürlicher Sprache)' },
  e1_d: { label: 'Technologien zur Erzeugung von Bildern, Videos, Ton/Audio' },
  e1_e: { label: 'Technologien zur Identifizierung von Objekten oder Personen anhand von Bildern oder Videos (Bilderkennung, Bildverarbeitung)' },
  e1_f: { label: 'Maschinelles Lernen (z. B. Tiefenlernen /Deep Learning) zur Datenanalyse' },
  e1_g: { label: 'Technologien, die unterschiedliche Arbeitsabläufe automatisieren oder bei der Entscheidungsfindung helfen (z. B. Prozessautomatisierung durch Roboter auf der Grundlage von KI-basierter Software)' },
  e1_h: { label: 'Technologien, die die physische Bewegung von Maschinen anhand autonomer Entscheidungen ermöglichen, die auf der Beobachtung der Umgebung basieren (autonome Roboter, selbstfahrende Fahrzeuge, autonome Drohnen)' },
  e2: { label: 'Nutzt Ihr Unternehmen Software oder Systeme für künstliche Intelligenz für einen der folgenden Zwecke?' },
  e2_a: {
    label: 'Für Marketing oder Vertrieb',
    hint: L(
      'Einige der Beispiele können sein:',
      '– Kundenprofiling, Preisoptimierung, personalisierte Marketingangebote, Marktanalyse basierend auf maschinellem Lernen',
      '– Chatbots basierend auf der Verarbeitung natürlicher Sprache für den Kundensupport',
      '– autonome Roboter für die Auftragsabwicklung',
    ),
  },
  e2_b: {
    label: 'Für Produktions- oder Dienstleistungsprozesse',
    hint: L(
      'Einige der Beispiele können sein:',
      '– vorausschauende Wartung basierend auf maschinellem Lernen',
      '– Tools zum Klassifizieren von Produkten oder zur Feststellung von Mängeln in Produkten basierend auf Computer Vision',
      '– autonome Drohnen für Produktionsüberwachung, Sicherheits- oder Inspektionsaufgaben',
      '– Montagearbeiten durch autonome Roboter',
    ),
  },
  e2_c: {
    label: 'Zur Organisation von Unternehmensverwaltungsprozessen oder das Management',
    hint: L(
      'Einige der Beispiele können sein:',
      '– virtuelle Geschäftsassistenten, die auf maschinellem Lernen und/oder Verarbeitung natürlicher Sprache basieren, z. B. zur Dokumentenerstellung',
      '– Datenanalyse oder strategische Entscheidungsfindung, z. B. Risikobewertung, basierend auf maschinellem Lernen',
      '– Planung oder Geschäftsprognosen basierend auf maschinellem Lernen',
      '– Personalmanagement basierend auf maschinellem Lernen oder der Verarbeitung natürlicher Sprache, z. B. Vorauswahl von Bewerbern, Erstellung von Mitarbeiterprofilen oder Leistungsanalysen',
    ),
  },
  e2_d: {
    label: 'Für die Logistik',
    hint: L(
      'Einige der Beispiele können sein:',
      '– autonome Roboter für Pick-and Pack-Lösungen in Lagern für Paketversand, Rückverfolgung, Verteilung oder Sortierung',
      '– Routenoptimierung basierend auf maschinellem Lernen',
    ),
  },
  e2_e: {
    label: 'Für die IT-Sicherheit',
    hint: L(
      'Einige der Beispiele können sein:',
      '– Gesichtserkennung basierend auf Computer Vision zur Authentifizierung von IT-Benutzern',
      '– Erkennung und Prävention von Cyber-Angriffen auf der Grundlage von maschinellem Lernen',
    ),
  },
  e2_f: {
    label: 'Für Buchführung, Controlling oder Finanzverwaltung',
    hint: L(
      'Einige der Beispiele können sein:',
      '– Maschinelles Lernen zur Analyse von Daten, die bei finanziellen Entscheidungen helfen',
      '– Rechnungsverarbeitung basierend auf maschinellem Lernen',
      '– Maschinelles Lernen oder Verarbeitung natürlicher Sprache für Buchführungsunterlagen',
    ),
  },
  e2_g: {
    label: 'Für Forschung und Entwicklung (F&E) oder Innovationstätigkeit (ohne KI-Forschung)',
    hint: L(
      'Einige der Beispiele können sein:',
      '– Analyse von Daten zur Durchführung von Forschungsarbeiten, Lösung von Forschungsproblemen, Entwicklung eines neuen oder deutlich verbesserten Produkts/Dienstes auf der Grundlage von maschinellem Lernen',
    ),
  },
  e4: { label: 'Hat Ihr Unternehmen jemals in Betracht gezogen, eine der in Frage E1 aufgeführten Technologien der künstlichen Intelligenz einzusetzen?' },
  e5: { label: 'Aus welchen Gründen verwendet Ihr Unternehmen keine der in Frage E1 aufgeführten Technologien der künstlichen Intelligenz?' },
  e5_a: { label: 'Die Kosten erscheinen zu hoch' },
  e5_b: { label: 'Es fehlt an entsprechendem Wissen im Unternehmen' },
  e5_c: { label: 'Inkompatibillität mit vorhandenem Bestand an Geräten, Software oder Systemen' },
  e5_d: { label: 'Schwierigkeiten mit der Verfügbarkeit oder Qualität der erforderlichen Daten' },
  e5_e: { label: 'Bedenken hinsichtlich der Wahrung des Datenschutzes und der Privatsphäre' },
  e5_f: { label: 'Unklarheit über die rechtlichen Folgen (z. B. Haftung bei Schäden durch den Einsatz von künstlicher Intelligenz)' },
  e5_g: { label: 'Ethische Überlegungen' },
  e5_h: { label: 'Technologien der künstlichen Intelligenz sind für das Unternehmen nicht sinnvoll.' },
};
const DE_LANG = 'Deutsch (de)';
for (const r of S) {
  const de = DE[r.name];
  if (!de) continue;
  r[`label::${DE_LANG}`] = de.label;
  if (de.hint) r[`hint::${DE_LANG}`] = de.hint;
}
const DE_CHOICES = { 1: 'Ja', 0: 'Nein' };
for (const c of C) if (c.list_name === 'yes_no') c[`label::${DE_LANG}`] = DE_CHOICES[c.name];

const study = {
  title: 'EU Survey on ICT Usage and e-Commerce in Enterprises 2025',
  subtitle: 'Module E: Artificial Intelligence',
  shortTitle: 'ICT-ENT 2025',
  id: 'q20-eu-ict-ent-ai',
  authors: [{ name: 'Eurostat', affiliation: 'European Commission, Eurostat' }],
  producer: { name: 'Eurostat' },
  citation: "Eurostat. (2025). European business statistics compiler's manual for ICT usage and e-commerce in enterprises, 2025 survey (2025 ed.; Manuals and guidelines, KS-01-25-065). Publications Office of the European Union. https://doi.org/10.2785/2641073 — German wording: Statistisches Bundesamt (Destatis). (2025). Qualitätsbericht: IKT-Erhebung in Unternehmen 2025 [Fragebogen, Abschnitt E]. https://www.destatis.de/DE/Methoden/Qualitaet/Qualitaetsberichte/Unternehmen/ikt-unternehmen-2025.pdf",
  url: URL_DOC,
  sourceLabel: 'KS-01-25-065, Annex 1 – Model questionnaire (ICT-ENT 2025 V1.5), Module E "Artificial Intelligence" (Eurostat)',
  abstract:
    'This study contains only Module E (Artificial Intelligence) of the Eurostat model questionnaire for the 2025 EU survey on ICT usage and e-commerce in enterprises, as printed in the 2025 compilers\' manual; national statistical institutes adapt it for their own surveys. The enterprise reports which AI technologies it uses, for which business purposes, and how it acquired the AI software or systems. Enterprises not using AI are asked whether they ever considered it and, if so, why they do not use it (for example costs, lack of expertise, data protection concerns or unclear legal consequences). The German texts are Destatis\u2019 wording of these questions in its national questionnaire for 2025, which does not ask E3 (how AI was acquired).',
  keywords: ['artificial intelligence', 'AI adoption', 'barriers to AI use', 'machine learning', 'data protection', 'enterprises', 'ICT usage', 'official statistics'],
  topics: ['Information society', 'Digital economy', 'Artificial intelligence'],
  nation: 'European Union',
  universe:
    'Enterprises with 10 or more employees and self-employed persons (NACE Rev. 2 sections C–J, L–N and group 95.1; optional: 0–9). Module E: enterprises with access to the internet.',
  anlyUnit: 'Enterprises',
  timePrd: '2025',
  dataKind: 'Model questionnaire',
  license:
    'CC BY 4.0 (https://creativecommons.org/licenses/by/4.0/). © European Union, 2025. Manual KS-01-25-065: "Except otherwise noted, the reuse of this document is authorised under a Creative Commons Attribution 4.0 International (CC-BY 4.0) licence https://creativecommons.org/licenses/by/4.0/). This means that reuse is allowed provided appropriate credit is given and any changes are indicated." Eurostat copyright notice (https://ec.europa.eu/eurostat/help/copyright-notice): "The copyright for the editorial content of this website, which is owned by the EU, is licensed under the Creative Commons Attribution 4.0 International licence. This means that you can re-use the content provided you acknowledge the source and indicate any changes you have made." "Reuse of statistical data, metadata, publications, and other dissemination tools published on this website for commercial or non-commercial purposes is authorised provided the source is acknowledged. The reuse policy of the European Commission is implemented by the Decision of 12 December 2011." (Commission Decision 2011/833/EU). "When reuse involves translations of publications or modifications to the data or text, this must be stated clearly to the end user of the information. A disclaimer regarding the non-responsibility of Eurostat shall be included." ' +
    'Changes: format conversion of Module E "Artificial Intelligence" of the 2025 model questionnaire (Annex 1 of manual KS-01-25-065) to DDI; wording unchanged; all other modules, explanatory notes and layout omitted; glossary definitions of the named AI techniques attached to the E1 items as hints. Disclaimer: this DDI version was made by the qwac project (CorrelAid), not by Eurostat; Eurostat bears no responsibility for it. German texts: Destatis, Qualitätsbericht "IKT-Erhebung in Unternehmen 2025" (questionnaire, section E): "© Statistisches Bundesamt (Destatis), 2025. Vervielfältigung und Verbreitung, auch auszugsweise, mit Quellenangabe gestattet." They are attached as the German version of the matching Eurostat questions (Destatis E3/E4 = Eurostat E4/E5); item numbers and routing lines omitted.',
  copyright: '© European Union, 2025 (Eurostat); German texts © Statistisches Bundesamt (Destatis), 2025',
};

writeFileSync(
  out,
  JSON.stringify({ study, survey: S, choices: C, settings: { default_language: 'English (en)', form_title: study.title, form_id: study.id } }, null, 2) + '\n',
);
console.log(`rows: ${S.length}, choices: ${C.length}`);
