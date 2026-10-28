// Generates tools/survey-import/out/q01-cvts7/form.json from a verbatim
// transcription of the CVTS7 paper questionnaire (Statistikamt Nord PDF).
import { writeFileSync } from 'node:fs';

const out = process.argv[2];
const S = [];
const C = [];
const row = (r) => S.push(r);
const begin = (name, label, extra = {}) => row({ type: 'begin_group', name, label, ...extra });
const end = () => row({ type: 'end_group' });
const choices = (list, items) => items.forEach(([name, label]) => C.push({ list_name: list, name, label }));
// Stem printed once above a table of rows: a note, which formtransform puts in
// the first row's preQTxt.
const stem = (name, label) => row({ type: 'note', name, label });

choices('ja_nein', [['1', 'Ja'], ['2', 'Nein']]);

// Grid of Ja/Nein rows sharing one stem.
const jnGrid = (name, stem, rows, extra = {}) => {
  begin(name, stem, { appearance: 'table-list', ...extra });
  for (const [code, label] of rows) row({ type: 'select_one ja_nein', name: `${name}_${code}`, label });
  end();
};
const yes = (n) => `\${${n}} = '1'`;

const OHNE = '(ohne Auszubildende, Studierende in Dualen Studiengängen und an Berufsakademien sowie Praktikanten/Trainees mit einem Ausbildungsvertrag)';
const GESCHL_HINT =
  'Bitte machen Sie Angaben zu „Divers“ und „Ohne Angabe (nach Geburtenregister)“ nur, wenn Ihnen diese Angaben aus dem Geburtenregister bekannt sind.';

// ---- Hinweise -------------------------------------------------------------
row({
  type: 'note',
  name: 'hinweis_definition',
  label:
    'Was ist berufliche Weiterbildung im Sinne dieser Erhebung?\n' +
    'Berufliche Weiterbildung umfasst\n' +
    '– Lehrgänge, Kurse und Seminare (Weiterbildung im engeren Sinne) sowie\n' +
    '– andere Formen der betrieblichen Weiterbildung (z. B. arbeitsplatznahe Formen der Qualifizierung, selbstgesteuertes Lernen und Informationsveranstaltungen).\n' +
    'Bei allen Weiterbildungsaktivitäten im Sinne dieser Erhebung handelt es sich um vorausgeplantes, organisiertes Lernen.\n' +
    'Für die „berufliche Weiterbildung in Unternehmen“ kennzeichnend ist, dass sie vollständig oder teilweise von Unternehmen finanziert wird. Die Finanzierung kann sowohl direkt erfolgen (z. B. Kosten für externe Dozentinnen und Dozenten) als auch indirekt (z. B. Kosten für betriebliche Weiterbildung, die während der bezahlten Arbeitszeit stattfindet).\n' +
    'Zu beruflichen Weiterbildungsmaßnahmen in Unternehmen gehören nicht\n' +
    '– Bildungsmaßnahmen für Arbeitslose, die von der Agentur für Arbeit initiiert und vollständig finanziert werden,\n' +
    '– betriebliche Weiterbildungsmaßnahmen, die vollständig von öffentlichen Stellen finanziert werden,\n' +
    '– betriebliche Weiterbildungsmaßnahmen, die ausschließlich von Beschäftigten des Unternehmens finanziert werden,\n' +
    '– die berufliche Erstausbildung von Auszubildenden im Dualen System, Studierenden in Dualen Studiengängen sowie an Berufsakademien mit einem besonderen Ausbildungsvertrag.',
});
row({
  type: 'note',
  name: 'hinweis_beschaeftigte',
  label:
    'Auf welche Beschäftigten beziehen sich die Fragen zur Weiterbildung?\n' +
    'Die Fragen zur beruflichen Weiterbildung betreffen die Beschäftigten Ihres Unternehmens in Deutschland mit einem Arbeitsvertrag.\n' +
    'Ausgeschlossen von diesen Fragen sind die Erstausbildung von Auszubildenden, die Ausbildung von Studierenden in Dualen Studiengängen sowie an Berufsakademien und die Ausbildung der Praktikanten/Trainees mit einem besonderen Ausbildungsvertrag.',
});

// ---- Teil 1 ---------------------------------------------------------------
begin('teil1', 'Teil 1: Fragen an alle Unternehmen – Strukturdaten des Unternehmens');
const q11 = `Wie viele in Deutschland Beschäftigte ${OHNE} hatte Ihr Unternehmen insgesamt?`;
row({ type: 'integer', name: 'f1_1_002', label: q11, hint: 'Beschäftigte am 31.12.2025 insgesamt. Anzahl' });
stem('f1_1_geschlecht', 'davon: Geschlecht (nach Geburtenregister)');
for (const [c, l] of [['002a', 'Männlich'], ['002b', 'Weiblich'], ['002c', 'Divers'], ['002d', 'Ohne Angabe (nach Geburtenregister)']])
  row({ type: 'integer', name: `f1_1_${c}`, label: l, hint: `Anzahl. ${GESCHL_HINT}` });
row({
  type: 'integer',
  name: 'f1_2_006',
  label: `Geben Sie bitte den Gesamtumfang der von den Beschäftigten ${OHNE} in Ihrem Unternehmen im Jahr 2025 geleisteten Arbeitsstunden an.`,
  hint: 'Geleistete Arbeitsstunden im Jahr 2025 insgesamt. Stunden',
});
row({
  type: 'integer',
  name: 'f1_3_007',
  label: 'Bitte tragen Sie die Personalaufwendungen (Bruttolöhne und -gehälter sowie Sozialaufwendungen) für die unter Frage 1.1 angegebenen Beschäftigten Ihres Unternehmens für das Jahr 2025 ein.',
  hint: 'Personalaufwendungen im Jahr 2025. Volle Euro',
});
end();

// ---- Teil 2 ---------------------------------------------------------------
begin('teil2', 'Teil 2: Fragen an alle Unternehmen – Weiterbildungsstrategien des Unternehmens');
row({ type: 'select_one ja_nein', name: 'f2_1_010', label: 'Gibt es in Ihrem Unternehmen eine bestimmte Person oder Organisationseinheit, die für betriebliche Weiterbildung verantwortlich ist?' });
choices('bedarf', [
  ['1', 'Nein'],
  ['2', 'Ja, aber nicht regelmäßig (hauptsächlich bei personellen Veränderungen)'],
  ['3', 'Ja, dies ist Teil des allgemeinen Planungsprozesses im Unternehmen'],
]);
row({ type: 'select_one bedarf', name: 'f2_2_011', label: 'Ermittelt Ihr Unternehmen regelmäßig den zukünftigen Bedarf an Qualifikationen (Fertigkeiten, Kenntnisse und Fähigkeiten) innerhalb des Unternehmens?' });
jnGrid('f2_3', 'Welche Maßnahmen ergreift Ihr Unternehmen normalerweise, um einem zukünftigen Bedarf an Qualifikationen (Fertigkeiten, Kenntnisse und Fähigkeiten) zu begegnen?', [
  ['012', 'Regelmäßige berufliche Weiterbildung der derzeitigen Beschäftigten'],
  ['013', 'Einstellung neuen Personals, das über die erforderlichen Qualifikationen verfügt'],
  ['014', 'Einstellung neuen Personals verbunden mit spezifischer Weiterbildung zur Erlangung der erforderlichen Qualifikationen'],
  ['015', 'Interne Reorganisation, um die bei den Beschäftigten bestehenden Qualifikationen besser zu nutzen'],
]);
const quali = (list, codes, techLabel, withKeine) => {
  const labels = [
    'Allgemeine IT-Kenntnisse (z. B. Computerbedienung, Textverarbeitung, Tabellenkalkulation)',
    'Professionelle IT-Kenntnisse (z. B. Erstellen von Webseiten, Programmierung)',
    'Führungskompetenzen',
    'Teamfähigkeit',
    'Kundenorientierung',
    'Problemlösungskompetenz',
    'Büro- und Verwaltungsfertigkeiten (z. B. Buchhaltung, Büroorganisation und -kommunikation, Sekretariats- und Büroarbeit, Zeitmanagement)',
    'Fremdsprachenkenntnisse',
    techLabel,
    'Mündliche und schriftliche Ausdrucksfähigkeit',
    'Mathematische und/oder Lese-/Schreibkompetenz',
    'Andere, noch nicht genannte Arten von Qualifikationen',
  ];
  if (withKeine) labels.push('Keine Einschätzung möglich');
  choices(list, labels.map((l, i) => [`k${codes[i]}`, l]));
};
quali('quali_wichtig', ['021', '022', '023', '024', '025', '026', '027', '028', '029', '030', '031', '032', '033'], 'Technische, praktische oder arbeitsplatzspezifische Fertigkeiten', true);
row({
  type: 'select_multiple quali_wichtig',
  name: 'f2_4',
  label: 'Welche Qualifikationen (Fertigkeiten, Kenntnisse und Fähigkeiten) der Beschäftigten werden in Ihrem Unternehmen im Allgemeinen für die Unternehmensentwicklung der nächsten Jahre als am wichtigsten erachtet?',
  hint: 'Kreuzen Sie bitte bis zu drei Qualifikationen an.',
  constraint: 'count-selected(.) <= 3',
});
row({ type: 'select_one ja_nein', name: 'f2_5_035', label: 'Erstellt Ihr Unternehmen normalerweise im Rahmen seiner Planung der betrieblichen Weiterbildung einen schriftlichen Weiterbildungsplan oder ein Weiterbildungsprogramm?' });
row({ type: 'select_one ja_nein', name: 'f2_6_036', label: 'Hat Ihr Unternehmen normalerweise ein Jahresbudget für die Aus- und Weiterbildung, das Mittel für die betriebliche Weiterbildung einschließt?' });
row({ type: 'select_one ja_nein', name: 'f2_7_037', label: 'Gelten für Ihr Unternehmen tarifvertragliche oder Betriebsvereinbarungen, die die betriebliche Weiterbildung zum Gegenstand haben?' });
row({ type: 'select_one ja_nein', name: 'f2_8_038', label: 'Gibt es in Ihrem Unternehmen ein Gremium der Arbeitnehmervertretung, wie z. B. einen Betriebsrat?', hint: 'Nein: Weiter mit Frage 3.1.' });
row({ type: 'select_one ja_nein', name: 'f2_8_039', label: 'Falls „Ja“, ist das Gremium der Arbeitnehmervertretung an Planung oder Durchführung der betrieblichen Weiterbildung beteiligt?', hint: 'Nein: Weiter mit Frage 3.1.', relevant: yes('f2_8_038') });
jnGrid('f2_8', 'Falls „Ja“, an welchen Aspekten der Planung oder Durchführung der betrieblichen Weiterbildung ist die Arbeitnehmervertretung normalerweise beteiligt?', [
  ['040', 'Festlegung der Ziele und Prioritäten der betrieblichen Weiterbildung'],
  ['041', 'Festlegung der Kriterien für die Auswahl von Teilnehmenden an Weiterbildungsmaßnahmen'],
  ['042', 'Art der Weiterbildung (z. B. interne oder externe Lehrveranstaltungen, andere Formen wie Lernzirkel)'],
  ['043', 'Themen der Weiterbildung'],
  ['044', 'Festlegung des Weiterbildungsbudgets'],
  ['045', 'Auswahl externer Weiterbildungsanbieter'],
  ['046', 'Evaluierung/Bewertung der Ergebnisse von Weiterbildung'],
], { relevant: yes('f2_8_039') });
end();

// ---- Teil 3 ---------------------------------------------------------------
begin('teil3', 'Teil 3: Fragen an alle Unternehmen – Weiterbildungsaktivitäten des Unternehmens');
row({
  type: 'note',
  name: 'teil3_hinweis',
  label:
    'Dieser Teil des Fragebogens bezieht sich auf die Weiterbildungsaktivitäten Ihres Unternehmens für seine Beschäftigten.\n' +
    'Die im Folgenden aufgeführten Aktivitäten gelten jedoch nur dann als betriebliche Weiterbildung im Sinne dieser Erhebung, wenn\n' +
    '– es sich um geplante Maßnahmen handelt, denen eine Entscheidung über Weiterbildungsbedarf vorausging,\n' +
    '– deren vorrangiges Ziel die berufliche Weiterbildung der Beschäftigten ist und\n' +
    '– sie ganz oder teilweise vom Unternehmen finanziert werden.\n' +
    'Betriebliche Weiterbildungsmaßnahmen umfassen sowohl Lehrveranstaltungen als auch andere Formen der betrieblichen Weiterbildung.\n' +
    'Die beiden folgenden Fragen beziehen sich auf interne und auf externe Lehrveranstaltungen.',
});
row({
  type: 'select_one ja_nein',
  name: 'f3_1_052',
  label: `Haben Beschäftigte Ihres Unternehmens ${OHNE} 2025 an internen Lehrveranstaltungen teilgenommen?`,
  hint: 'Interne Lehrveranstaltungen sind Maßnahmen, bei denen die Verantwortung für Ziele, Inhalte und Organisation beim Unternehmen selbst liegt und an denen ausschließlich oder überwiegend Beschäftigte des Unternehmens teilnehmen.',
});
row({
  type: 'select_one ja_nein',
  name: 'f3_2_053',
  label: `Haben Beschäftigte Ihres Unternehmens ${OHNE} 2025 an externen Lehrveranstaltungen teilgenommen?`,
  hint: 'Externe Lehrveranstaltungen sind Maßnahmen, die von externen Trägern (z. B. Kammern, Bildungswerken, Management-Akademien) auf dem freien Markt angeboten werden. Die Verantwortung für die Konzeption, Organisation und Durchführung liegt im Wesentlichen bei dem externen Träger.',
});
const q33 = `Haben Beschäftigte Ihres Unternehmens ${OHNE} 2025 an den folgenden geplanten Weiterbildungsaktivitäten teilgenommen?`;
const act33 = [
  ['054', 'Geplante Phasen der Weiterbildung am Arbeitsplatz (Unterweisung durch Vorgesetzte, Spezialisten oder Kollegen, Lernen durch die normalen Arbeitsmittel und andere Medien (Einarbeitung))'],
  ['056', 'Geplante Weiterbildung durch Job-Rotation, Austauschprogramme, Abordnungen, Studienbesuche'],
  ['058', 'Geplante Weiterbildung durch Besuch von Informationsveranstaltungen wie Fachvorträgen, Fachtagungen/Kongressen/Symposien/Kolloquien, Workshops, Fachmessen, Erfahrungsaustauschkreisen'],
  ['060', 'Geplante Weiterbildung durch Teilnahme an Lern- und Qualitätszirkeln'],
  ['062', 'Geplante Weiterbildung durch selbstgesteuertes Lernen z. B. durch E-Learning, mit audio-visuellen Hilfen wie Videos und anderen Medien'],
];
jnGrid('f3_3', q33, act33, { hint: 'Die nächste Frage bezieht sich auf andere Formen der betrieblichen Weiterbildung für Beschäftigte Ihres Unternehmens.' });
choices('anteil', [['1', 'unter 10 %'], ['2', '10 % bis unter 50 %'], ['3', '50 % und mehr']]);
const q33a = 'Falls Beschäftigte Ihres Unternehmens an folgenden Weiterbildungsaktivitäten teilgenommen haben, kreuzen Sie bitte zusätzlich die jeweils zutreffende Kategorie zum Anteil der Teilnehmenden an den Beschäftigten an.';
begin('f3_3_anteil', q33a, { appearance: 'table-list' });
for (const [c, l] of act33) row({ type: 'select_one anteil', name: `f3_3_${c}a`, label: l, relevant: yes(`f3_3_${c}`) });
end();
choices('ja_nein_2024', [['1', 'Ja'], ['2', 'Nein'], ['3', 'Unternehmen existierte 2024 nicht']]);
row({ type: 'select_one ja_nein_2024', name: 'f3_4_064', label: 'Hat Ihr Unternehmen 2024 für seine Beschäftigten interne oder externe Lehrveranstaltungen in Form von Lehrgängen, Kursen und Seminaren durchgeführt?' });
row({ type: 'select_one ja_nein_2024', name: 'f3_5_065', label: 'Hat Ihr Unternehmen 2024 eine der Maßnahmen im Bereich der anderen Formen der betrieblichen Weiterbildung (siehe Frage 3.3) für seine Beschäftigten durchgeführt?' });
begin('f3_6', 'Beiträge und Einnahmen für Weiterbildungsaktivitäten');
row({
  type: 'note',
  name: 'f3_6_hinweis',
  label: 'Bitte machen Sie hier Angaben zu Beiträgen und Einnahmen Ihres Unternehmens 2025 für die Weiterbildung seiner Beschäftigten. Beiträge und Einnahmen für die Aus- und Weiterbildung von Auszubildenden, Studierenden in Dualen Studiengängen sowie Praktikanten/Trainees mit einem Ausbildungsvertrag sind hier auszuschließen.',
});
const q36a = 'Musste Ihr Unternehmen 2025 Umlagen oder Beiträge an Fonds für die Finanzierung von Weiterbildungsprogrammen leisten?';
row({ type: 'select_one ja_nein', name: 'f3_6_066', label: q36a });
row({ type: 'integer', name: 'f3_6_067', label: 'Falls „Ja“, geben Sie bitte die Umlagen und Beiträge in vollen Euro an.', hint: 'Volle Euro', relevant: yes('f3_6_066') });
const q36b = 'Hatte Ihr Unternehmen 2025 Einnahmen aus Fonds oder erhielt es Zuschüsse (z. B. von der EU oder vom Staat) oder sonstige finanzielle Unterstützung aus anderen Quellen (z. B. Stiftungen) für Weiterbildungsaktivitäten?';
row({ type: 'select_one ja_nein', name: 'f3_6_068', label: q36b, hint: 'Nein: Weiter mit Teil 4, 5 oder 6 des Fragebogens gemäß der Einteilung auf Seite 11 unten.' });
row({ type: 'integer', name: 'f3_6_069', label: 'Falls „Ja“, geben Sie bitte die Einnahmen in vollen Euro an.', hint: 'Volle Euro', relevant: yes('f3_6_068') });
end();
jnGrid('f3_7', 'Falls „Ja“, welche Fördermaßnahmen nimmt Ihr Unternehmen in Anspruch?', [
  ['070', 'Zahlungen aus Branchenfonds zur Finanzierung der Weiterbildung'],
  ['071', 'Zuschüsse der Europäischen Union (z. B. Mitfinanzierung von Qualifizierungsangeboten für die Bezieher von Kurzarbeitergeld aus dem Europäischen Sozialfonds)'],
  ['072', 'Öffentliche Zuschüsse (z. B. Förderung der Weiterbildung Geringqualifizierter und beschäftigter älterer Arbeitnehmer in Unternehmen (WeGebAU), Qualifizierungsangebote für die Bezieher von Kurzarbeitergeld, Bildungsscheck NRW)'],
  ['073', 'Einnahmen aus sonstigen Quellen (z. B. aus privaten Stiftungen, Zahlungen, die das Unternehmen aufgrund der Durchführung von Weiterbildungskursen für externe Organisationen und Personen erhält)'],
], { relevant: yes('f3_6_068') });
end();

// Routing after Teil 3 (page 11 of the questionnaire).
const lv = `(${yes('f3_1_052')} or ${yes('f3_2_053')})`;
const andere = act33.map(([c]) => yes(`f3_3_${c}`)).join(' or ');
const wb = `(${lv.slice(1, -1)} or ${andere})`;

// ---- Teil 4 ---------------------------------------------------------------
begin('teil4', 'Teil 4: Fragen an Unternehmen, die 2025 Lehrveranstaltungen angeboten haben', { relevant: lv });
row({
  type: 'note',
  name: 'teil4_hinweis',
  label: 'In den Fragen 3.1 und 3.2 wurden Sie gefragt, ob Beschäftigte Ihres Unternehmens im Jahr 2025 an internen oder externen Lehrveranstaltungen teilgenommen haben. Beantworten Sie bitte die nachfolgenden Fragen, wenn Sie mindestens eine der Fragen mit „Ja“ beantwortet haben.',
});
row({
  type: 'note',
  name: 'f4_1_hinweis',
  label:
    'Teilnehmende an internen und externen Lehrveranstaltungen\n' +
    'Zunächst hätten wir gerne Angaben zu den Teilnehmenden an internen und externen Lehrveranstaltungen. Dabei sollten die Beschäftigten jeweils nur einmal gezählt werden, unabhängig davon, an wie vielen Lehrveranstaltungen sie im Laufe des Jahres 2025 teilgenommen haben. Wenn Sie Angaben auf der Basis von Personen machen können, tragen Sie bitte diese Angaben in die erste Spalte der Tabelle ein.\n' +
    'Liegen in Ihrem Unternehmen Angaben zu Teilnehmenden-Fällen vor (Beschäftigte, die im Jahr 2025 an mehreren Lehrveranstaltungen teilgenommen haben, werden mehrfach gezählt), so tragen Sie bitte diese Angaben in die zweite Spalte der Tabelle ein.\n' +
    'Liegen Ihnen beide Angaben vor, so tragen Sie bitte entsprechende Angaben in die Tabelle ein.',
});
for (const [suffix, col] of [['075', 'Personen'], ['078', 'Teilnehmenden-Fälle']]) {
  row({ type: 'integer', name: `f4_1_${suffix}`, label: 'Teilnehmende insgesamt', hint: `Anzahl. ${col}` });
  stem(`f4_1_${suffix}_geschlecht`, 'davon: Geschlecht (nach Geburtenregister)');
  for (const [s, l] of [['a', 'Männlich'], ['b', 'Weiblich'], ['c', 'Divers'], ['d', 'Ohne Angabe (nach Geburtenregister)']])
    row({ type: 'integer', name: `f4_1_${suffix}${s}`, label: l, hint: `Anzahl. ${col}. ${GESCHL_HINT}` });
}
const q42 = 'Geben Sie bitte die Gesamtzahl der Teilnahmestunden im Jahr 2025 an. Unterscheiden Sie dabei auch nach internen und externen Lehrveranstaltungen.';
row({ type: 'integer', name: 'f4_2_081', label: q42, hint: 'Teilnahmestunden insgesamt' });
row({ type: 'integer', name: 'f4_2_082', label: 'davon an: internen Lehrveranstaltungen' });
row({ type: 'integer', name: 'f4_2_083', label: 'externen Lehrveranstaltungen' });
quali('quali_stunden', ['084', '085', '086', '087', '088', '089', '090', '091', '092', '093', '094', '095'], 'Technische, praktische oder arbeitsplatzspezifische Kompetenzen', false);
row({
  type: 'select_multiple quali_stunden',
  name: 'f4_3',
  label: 'Auf welche Qualifikationen (Fertigkeiten, Kenntnisse und Fähigkeiten) entfielen in Ihrem Unternehmen die meisten Stunden interner und externer Lehrveranstaltungen?',
  hint: 'Bitte kreuzen Sie bis zu drei Qualifikationen an.',
  constraint: 'count-selected(.) <= 3',
});
const q44 = 'Wie hoch war der Anteil der gesetzlich oder anderweitig vorgeschriebenen Lehrveranstaltungen aus dem Bereich „Gesundheit und Arbeitsschutz“ (gemessen in Teilnahmestunden an Lehrveranstaltungen) im Jahr 2025?';
const h44 =
  'Beispiele für vorgeschriebene Lehrveranstaltungen im Bereich „Gesundheit und Arbeitsschutz“ sind § 12 Arbeitsschutzgesetz, § 11 Lärm- und Vibrations-Arbeitsschutzverordnung, § 20 Druckluftverordnung, § 14 Gefahrstoffverordnung, §§ 35 und 43 Absatz 4 Infektionsschutzgesetz; Gabelstaplerschein, Schweißerpass. Bitte tragen Sie die entsprechenden prozentualen Anteile ein. = Teilnahmestunden an Lehrveranstaltungen insgesamt: 100 %';
stem('f4_4_frage', q44);
row({ type: 'decimal', name: 'f4_4_097', label: 'Teilnahmestunden an gesetzlich oder anderweitig vorgeschriebenen Lehrveranstaltungen aus dem Bereich Gesundheit und Arbeitsschutz', hint: 'Prozent. ' + h44 });
row({ type: 'decimal', name: 'f4_4_098', label: '+ Sonstige Teilnahmestunden an Lehrveranstaltungen', hint: 'Prozent' });
choices('anbieter', [
  ['k099', 'Fachschulen, Universitäten und Fachhochschulen'],
  ['k100', 'Öffentliche Bildungsanbieter (z. B. Volkshochschulen)'],
  ['k101', 'Private Bildungsanbieter mit Erwerbszweck'],
  ['k102', 'Hersteller/Lieferanten, Muttergesellschaften/Beteiligungsgesellschaften'],
  ['k103', 'Organisationen der Wirtschaft inkl. der berufsständischen Kammern (IHK, HwK, Innungen) und ihre Bildungseinrichtungen'],
  ['k104', 'Gewerkschaften und ihre Bildungseinrichtungen'],
  ['k105', 'Sonstige Bildungsanbieter (z. B. Kirchen)'],
]);
row({
  type: 'select_multiple anbieter',
  name: 'f4_5',
  label: 'Auf welche Anbieter externer Lehrveranstaltungen entfielen im Jahr 2025 die meisten Stunden?',
  hint: 'Falls Ihr Unternehmen externe Lehrveranstaltungen angeboten hat (siehe Frage 3.2), beantworten Sie bitte die folgende Frage 4.5. Bitte kreuzen Sie bis zu drei Anbieter an.',
  constraint: 'count-selected(.) <= 3',
  relevant: yes('f3_2_053'),
});
begin('f4_6', 'Kosten für interne und externe Lehrveranstaltungen');
row({
  type: 'note',
  name: 'f4_6_hinweis',
  label: 'Bitte geben Sie hier die Kosten Ihres Unternehmens im Jahr 2025 für die Durchführung von internen und externen Lehrveranstaltungen an, und zwar aufgeschlüsselt nach den unten angegebenen Kategorien. Kosten für Kurse für Auszubildende, Studierende in Dualen Studiengängen und an Berufsakademien sowie Praktikanten/Trainees mit einem Ausbildungsvertrag sind hier auszuschließen. Falls Ihnen Kosten in einer der folgenden Kategorien entstanden sind, geben Sie diese bitte in vollen Euro an.',
});
for (const [no, amt, l] of [
  ['107', '108', 'Zahlungen und Gebühren an Weiterbildungsanbieter für die Teilnahme der Beschäftigten an externen Lehrveranstaltungen; Kosten für den Einsatz von externem Weiterbildungspersonal in internen Lehrveranstaltungen'],
  ['109', '110', 'Reisekosten, Spesen und Tagegeld für Weiterbildungsteilnehmende für die Zeit der Teilnahme an Lehrgängen, Kursen und Seminaren'],
  ['111', '112', 'Personalaufwendungen für internes Weiterbildungspersonal, das ausschließlich oder teilweise mit der Konzeption, Organisation und Durchführung von Lehrgängen, Kursen und Seminaren beschäftigt war'],
  ['113', '114', 'Kosten für Räume und Ausstattung einschließlich Unterhaltungskosten für Schulungszentren (bei nur zeitweiliger Nutzung für Weiterbildung sind die Kosten anteilig anzugeben)'],
  ['115', '116', 'Kosten für Unterrichtsmaterial, das speziell für Weiterbildungszwecke gekauft wurde'],
]) {
  row({ type: 'select_one ja_nein', name: `f4_6_${no}`, label: l });
  row({ type: 'integer', name: `f4_6_${amt}`, label: l, hint: 'Volle Euro', relevant: yes(`f4_6_${no}`) });
}
row({
  type: 'integer',
  name: 'f4_6_117',
  label: 'Geben Sie bitte hier noch die Gesamtkosten für Lehrveranstaltungen an, auch wenn Sie die Kosten nicht nach den obigen Kategorien aufschlüsseln konnten.',
  hint: 'Volle Euro',
});
end();
end();

// ---- Teil 5 ---------------------------------------------------------------
const gruende = (prefix, codes, first) => [
  [codes[0], first],
  [codes[1], 'Die benötigten Qualifikationen wurden bevorzugt durch Neueinstellung entsprechend qualifizierten Personals gewonnen'],
  [codes[2], 'Schwierigkeiten, den Weiterbildungsbedarf des Unternehmens einzuschätzen'],
  [codes[3], 'Mangelndes Angebot an passenden Lehrveranstaltungen auf dem Weiterbildungsmarkt'],
  [codes[4], 'Die hohen Kosten für Lehrveranstaltungen'],
  [codes[5], 'Das Unternehmen legte den Schwerpunkt auf die betriebliche Erstausbildung statt auf Weiterbildung'],
  [codes[6], 'Umfangreiche Weiterbildungsmaßnahmen in den letzten Jahren'],
  [codes[7], 'Die hohe Arbeitsbelastung und die begrenzt verfügbare Zeit der Beschäftigten für Weiterbildung'],
  [codes[8], 'Sonstige Gründe'],
];
begin('teil5', 'Teil 5: Fragen an Unternehmen, die 2025 Lehrveranstaltungen und/oder andere Formen der Weiterbildung angeboten haben', { relevant: wb });
row({
  type: 'note',
  name: 'teil5_hinweis',
  label: 'In den Fragen 3.1 bis 3.3 wurden Sie gefragt, ob Beschäftigte Ihres Unternehmens im Jahr 2025 an Lehrveranstaltungen oder anderen Formen der betrieblichen Weiterbildung teilgenommen haben. Wenn Sie mindestens eine der in Frage 3.1 bis 3.3 aufgeführten Weiterbildungsaktivitäten angeboten haben, beantworten Sie bitte die Fragen in Teil 5 des Fragebogens.',
});
choices('bewertung', [['1', 'Ja, nach jeder Maßnahme'], ['2', 'Ja, nach einigen Maßnahmen'], ['3', 'Nein, die Feststellung der Teilnahme reicht aus']]);
row({ type: 'select_one bewertung', name: 'f5_1_123', label: 'Bewertet Ihr Unternehmen normalerweise die Ergebnisse von Weiterbildungsaktivitäten?', hint: 'Nein, die Feststellung der Teilnahme reicht aus: Weiter mit Frage 5.2.' });
jnGrid('f5_1', 'Falls „Ja“, welche Methode verwendet Ihr Unternehmen zur Bewertung der Weiterbildungsergebnisse?', [
  ['124', 'Bescheinigung/Zertifizierung nach schriftlichem oder praktischem Test der Teilnehmenden'],
  ['125', 'Messung der Zufriedenheit der Teilnehmenden durch einen Fragebogen'],
  ['126', 'Bewertung des Verhaltens oder der Leistung der Teilnehmenden in Bezug auf die Weiterbildungsziele'],
  ['127', 'Bewertung/Messung der Auswirkungen der Weiterbildung auf die Leistung bestimmter Abteilungen oder des ganzen Unternehmens'],
  ['127a', 'Andere Bewertungsmethoden'],
], { relevant: "${f5_1_123} = '1' or ${f5_1_123} = '2'" });
jnGrid('f5_2', 'Wurde im Jahr 2025 das Weiterbildungsangebot Ihres Unternehmens durch einen oder mehrere der folgenden Gründe eingeschränkt?',
  gruende('f5_2', ['128', '129', '130', '131', '132', '133', '134', '135', '136'], 'Das Weiterbildungsangebot entsprach dem Bedarf des Unternehmens (keine Einschränkungen)'),
  { hint: 'Beantworten Sie bitte als nächstes Teil 7 des Fragebogens.' });
end();

// ---- Teil 6 ---------------------------------------------------------------
begin('teil6', 'Teil 6: Fragen an Unternehmen, die 2025 keine Weiterbildung angeboten haben', { relevant: `not(${wb})` });
row({
  type: 'note',
  name: 'teil6_hinweis',
  label: 'In den Fragen 3.1 bis 3.3 wurden Sie gefragt, ob Beschäftigte Ihres Unternehmens im Jahr 2025 an Lehrveranstaltungen oder anderen Formen der betrieblichen Weiterbildung teilgenommen haben. Wenn Sie alle drei Fragen mit „Nein“ beantwortet haben, beantworten Sie bitte die folgende Frage.',
});
jnGrid('f6', 'Was sind die Gründe dafür, dass Sie im Jahr 2025 keine berufliche Weiterbildung angeboten haben?',
  gruende('f6', ['137', '138', '139', '140', '141', '142', '143', '144', '145'], 'Die vorhandenen Qualifikationen der Beschäftigten entsprachen dem Bedarf des Unternehmens'));
end();

// ---- Teil 7 ---------------------------------------------------------------
begin('teil7', 'Teil 7: Fragen an alle Unternehmen zur betrieblichen Erstausbildung');
row({
  type: 'note',
  name: 'teil7_hinweis',
  label:
    'Dieser Teil des Fragebogens bezieht sich auf die betriebliche Erstausbildung, die Sie im Jahr 2025 in Ihrem Unternehmen angeboten haben.\n' +
    'Als betriebliche Erstausbildung im Sinne dieser Erhebung gilt nur die Ausbildung im Dualen System nach Berufsbildungsgesetz/Handwerksordnung (Lehre).\n' +
    'Nicht zur betrieblichen Erstausbildung zählen Unternehmenspraktika von Schülerinnen und Schülern oder Studierenden, Ausbildung von Studierenden in Dualen Studiengängen sowie an Berufsakademien sowie Traineeprogramme für Hochschulabsolventinnen und -absolventen.',
});
row({ type: 'select_one ja_nein', name: 'f7_1_146a', label: 'Beschäftigt Ihr Unternehmen normalerweise Auszubildende?' });
jnGrid('f7_2', 'Aus welchen Gründen bietet Ihr Unternehmen normalerweise eine Ausbildung im Dualen System an?', [
  ['147', 'Um zukünftige Beschäftigte entsprechend des Bedarfs des Unternehmens zu qualifizieren'],
  ['148', 'Um die besten Auszubildenden nach Abschluss ihrer Ausbildung zu übernehmen'],
  ['149', 'Um das Risiko personeller Fehlentscheidungen bei der Einstellung externer Arbeitskräfte zu vermeiden'],
  ['150', 'Um Auszubildende bereits während der Ausbildung als Arbeitskräfte einzusetzen'],
  ['151', 'Sonstige Gründe (z. B. um das Unternehmen attraktiver für potentielle Beschäftigte zu machen, Sicherung des Fachkräftenachwuchses, Firmentradition)'],
], { relevant: yes('f7_1_146a'), hint: 'Falls Sie normalerweise Auszubildende haben, beantworten Sie bitte noch die nachfolgende Frage 7.2.' });
end();

const study = {
  title: 'Europäische Erhebung über die berufliche Weiterbildung in Unternehmen (CVTS7): Fragebogen',
  id: 'q01-cvts7',
  authors: [{ name: 'Statistische Ämter des Bundes und der Länder', affiliation: 'Statistisches Amt für Hamburg und Schleswig-Holstein (Statistikamt Nord)' }],
  producer: { name: 'Statistisches Amt für Hamburg und Schleswig-Holstein (Statistikamt Nord)' },
  citation:
    'Statistische Ämter des Bundes und der Länder. (2026). Europäische Erhebung über die berufliche Weiterbildung in Unternehmen (CVTS7): Fragebogen [Erhebungsunterlage]. Statistikamt Nord. https://www.statistik-nord.de/fileadmin/Dokumente/Erhebungen/Bildung-Rechtspflege/CVTS7_Fragebogen.pdf',
  url: 'https://www.statistik-nord.de/fileadmin/Dokumente/Erhebungen/Bildung-Rechtspflege/CVTS7_Fragebogen.pdf',
  sourceLabel: 'CVTS7 Fragebogen „Berufliche Weiterbildung in Unternehmen 2025“ (PDF, Statistikamt Nord)',
  abstract:
    'Paper questionnaire of the seventh European Continuing Vocational Training Survey (CVTS7) as used in Germany, reference year 2025. It asks enterprises with at least 10 employees about their structure, training strategies, courses and other forms of continuing vocational training, participants, hours and costs, quality assessment, reasons for not training, and initial vocational training. Routing sends training and non-training enterprises to different parts.',
  keywords: ['berufliche Weiterbildung', 'betriebliche Weiterbildung', 'CVTS', 'Weiterbildungskosten', 'Lehrveranstaltungen', 'Erstausbildung', 'amtliche Statistik'],
  topics: ['Bildung', 'Weiterbildung', 'Unternehmen'],
  nation: 'Deutschland',
  universe: 'Unternehmen mit mindestens 10 Beschäftigten in Deutschland (Bezugsjahr 2025)',
  anlyUnit: 'Unternehmen',
  timePrd: '2025',
  dataKind: 'Erhebungsunterlage (Fragebogen)',
  license:
    '© Statistisches Amt für Hamburg und Schleswig-Holstein, Hamburg 2026. „Sie können Informationen speichern, (auch auszugsweise) mit Quellenangabe weitergeben, vervielfältigen und verbreiten sowie Links zur Homepage des Statistikamtes Nord speichern. Die Informationen dürfen im Übrigen nicht verändert oder verfälscht werden.“ (https://www.statistik-nord.de/impressum). Erhebungsunterlage einer amtlichen Statistik (Statistische Ämter des Bundes und der Länder).',
  copyright: '© Statistisches Amt für Hamburg und Schleswig-Holstein, Hamburg 2026',
};

writeFileSync(
  out,
  JSON.stringify({ study, survey: S, choices: C, settings: { default_language: 'Deutsch (de)', form_title: study.title, form_id: study.id } }, null, 2) + '\n',
);
console.log(`rows: ${S.length}, choices: ${C.length}`);
