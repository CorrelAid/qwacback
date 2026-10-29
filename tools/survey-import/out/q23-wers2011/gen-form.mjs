// Generates tools/survey-import/out/q23-wers2011/form.json from a verbatim
// transcription of the WERS 2011 (WERS6) Management Questionnaire (MQ),
// BIS/11/803, "Published: March 2011", as downloaded from GOV.UK:
//   - Section C "RECRUITMENT, TRAINING AND ORGANISATION OF WORK" (pp. 30-37), complete;
//   - from Section EPQ (p. 4-5) the questions Section C depends on: ZALLEMPS
//     (used by the Section C soft checks), the occupation table SOC_1 to SOC_9
//     (defines the text fill [LOGIDX], the largest non-managerial group that
//     most of Section C asks about) and ZSOCDESC.
// Only literal text of the MQ PDF is used. PDF line wraps inside one sentence
// are joined with a space; separate printed lines stay separate ("\n").
// verify-labels.py checks every text in form.json against the PDF.
//
//   node gen-form.mjs form.json
import { writeFileSync } from 'node:fs';

const out = process.argv[2];
const S = [];
const C = [];
const row = (r) => S.push(r);
const begin = (name, label, extra = {}) => row({ type: 'begin_group', name, label, ...extra });
const end = () => row({ type: 'end_group' });
const L = (...lines) => lines.join('\n');
// items: [code, label, extra?]
const list = (name, items) =>
  items.forEach(([code, label, extra = {}]) => C.push({ list_name: name, name: code, label, ...extra }));

const LOG = '[LOGIDX: TITLE OF THE LARGEST OCCUPATIONAL GROUP]';
const PROBE = "PROBE: Which others? UNTIL 'None'.";
const ALL = 'CODE ALL THAT APPLY.';
const soft = (v) =>
  `SOFT CHECK if ${v}>ZALLEMPS: "Are you sure? The number recorded here is greater than the total number of employees recorded at the start of the interview."`;
const exact97 = 'INTERVIEWER: If respondent gives answer as an exact number you can code “97” here and record the number at the next question.';

// ---------------------------------------------------------------- choices
list('yes_no', [['1', 'Yes'], ['2', 'No']]);
list('cfillvac', [
  ['1', 'Internal applicants are only source, no external recruitment'],
  ['2', 'Internal applicants are given preference, other things being equal, over external applicants'],
  ['3', 'Applications from internal and external applicants are treated equally'],
  ['4', 'External applicants are given preference, other things being equal, over internal applicants'],
  ['5', 'External applicants are only source, no internal recruitment'],
]);
list('cfactors', [
  ['1', 'References'],
  ['2', 'Availability'],
  ['3', 'Recommended by another employee'],
  ['4', 'Skills'],
  ['5', 'Age'],
  ['6', 'Qualifications'],
  ['7', 'Experience'],
  ['8', 'Motivation'],
  ['other', 'Other (please specify CFACTOTH)'],
]);
list('cspecix', [
  ['1', 'Women returning to work after having children'],
  ['2', 'Women in general'],
  ['3', 'Members of minority ethnic groups'],
  ['4', 'Older workers'],
  ['5', 'Disabled people'],
  ['6', 'People who have been unemployed for 12 months or more'],
  ['7', 'Part-time workers or job sharers'],
  ['8', 'Gay, lesbian and transgender communities'],
  ['9', 'None of these', { exclusive: 'yes' }],
]);
list('catestw', [
  ['1', 'Managerial'],
  ['2', 'Non-managerial'],
  ['3', 'No tests conducted', { exclusive: 'yes' }],
]);
list('cptestw', [
  ['1', 'Managerial'],
  ['2', 'Non-managerial'],
  ['3', 'No tests conducted.', { exclusive: 'yes' }],
]);
list('cchanx', [
  ['1', 'Local or regional newspapers'],
  ['2', 'Trade, professional or specialist press'],
  ['3', 'Internal notices'],
  ['4', 'Notice in office or shop window'],
  ['5', 'Jobcentre / Employment Service office'],
  ['6', 'Careers Service (School/College/University)'],
  ['7', 'Fee charging, private employment or recruitment agency'],
  ['8', 'Direct approach to potential recruit(s)'],
  ['9', 'Replying to speculative applications'],
  ['10', 'Recommendation or enquiry by existing employee'],
  ['11', 'Word of mouth'],
  ['12', 'Intranet or own website'],
  ['13', 'Internet – external website(s)'],
  ['other', 'Some other way (please specify COTHCHN)'],
]);
list('cstuckin', [
  ['1', 'One week or less'],
  ['2', 'More than one week, up to one month'],
  ['3', 'More than one month, up to six months'],
  ['4', 'More than six months, up to one year'],
  ['5', 'More than one year'],
]);
const prop = [
  ['1', 'All (100%)'],
  ['2', 'Almost all (80-99%)'],
  ['3', 'Most (60-79%)'],
  ['4', 'Around half (40-59%)'],
  ['5', 'Some (20-39%)'],
];
list('coffjob', [...prop, ['6', 'Just a few (1-19%)'], ['7', 'None (0%)']]);
list('ctrain', [
  ['1', 'No time'],
  ['2', 'Less than one day'],
  ['3', '1 to less than 2 days'],
  ['4', '2 to less than 5 days'],
  ['5', '5 to less than 10 days'],
  ['6', '10 days or more'],
]);
list('chow', [
  ['1', 'Computing skills'],
  ['2', 'Team working'],
  ['3', 'Communication skills'],
  ['4', 'Leadership skills'],
  ['5', 'Operation of new equipment'],
  ['6', 'Customer service/liaison'],
  ['7', 'Health and safety'],
  ['8', 'Problem-solving methods'],
  ['9', 'Equal opportunities and diversity'],
  ['10', 'Reliability and working to deadlines'],
  ['11', 'Quality control procedures'],
  ['12', 'None of these', { exclusive: 'yes' }],
]);
// SHOW CARD C8 (COTHJOB, COTHDO): code 97 as printed.
list('card_c8', [
  ...prop,
  ['6', 'Just a few (1-19%)'],
  ['7', 'None (0%)'],
  ['97', 'INTERVIEWER: Exact number to be entered.'],
]);
// SHOW CARD C10 (CTEAMS), printed with "(1-19%," and no final full stop.
list('cteams', [
  ...prop,
  ['6', 'Just a few (1-19%,'],
  ['7', 'None (0%)'],
  ['97', 'INTERVIEWER: Exact number to be entered'],
]);
list('card_c9', [['1', 'A lot'], ['2', 'Some'], ['3', 'A little'], ['4', 'None']]);

// ---------------------------------------------------------------- Section EPQ
// INTROEPQ, and the SOC_1 to SOC_9 instructions, are put on the rows: a
// group's guidance_hint does not survive the DDI conversion.
begin('section_epq', 'EMPLOYEE PROFILE QUESTIONNAIRE');
row({
  type: 'integer',
  name: 'zallemps',
  label: 'Currently how many employees do you have on the payroll at this workplace?',
  guidance_hint: L(
    'INTERVIEWER: To start the interview, please enter the relevant data from the EPQ.',
    'INTERVIEWER: THIS IS Q1 IN THE PAPER EPQ',
    'PROMPT: Remember to include yourself if you are an employee at the workplace but do NOT include casual workers without a contract of employment, freelance, self-employed or agency workers.',
  ),
  required: 'yes',
  constraint: '. >= 5 and . <= 999997',
  constraint_message: 'THERE MUST BE AT LEAST 5 EMPLOYEES AT THE WORKPLACE. PLEASE CHECK.',
});
begin('soc', 'How many male and female employees are in each of the following occupational groups?', {
  appearance: 'table-list',
});
const socGuidance = L(
  'INTERVIEWER: THIS IS Q2 IN THE PAPER EPQ.',
  'PROMPT: Please refer to the accompanying information sheet for definitions of the occupations or SHOWCARD A1 for some examples.',
);
for (const [code, label] of [
  ['zmgr', 'Managers and senior officials'],
  ['zpro', 'Professional occupations'],
  ['ztec', 'Associate professional and technical occupations'],
  ['zadm', 'Administrative and secretarial occupations'],
  ['zskl', 'Skilled trades occupations'],
  ['zper', 'Caring, leisure and other personal service occupations'],
  ['zsal', 'Sales and customer service occupations'],
  ['zope', 'Process, plant and machine operatives and drivers'],
  ['zrou', 'Routine occupations'],
]) {
  for (const [sfx, col] of [['male', 'Male'], ['fem', 'Female']]) {
    row({
      type: 'integer',
      name: `${code}_${sfx}`,
      label,
      hint: col,
      guidance_hint: socGuidance,
      constraint: '. <= ${zallemps}',
      constraint_message: 'THE NUMBER OF MEN / WOMEN GIVEN FOR THIS GROUP EXCEEDS THE TOTAL NUMBER OF EMPLOYEES.',
    });
  }
}
end();
row({
  type: 'text',
  name: 'zsocdesc',
  label: L(
    'From what you have said, the largest non-managerial group of employees at this workplace is classified as [LOGIDX: TITLE OF LARGEST OCCUPATIONAL GROUP].',
    'How would you describe the main tasks and activities of this largest group?',
  ),
  guidance_hint: 'INTERVIEWER: Probe for details about the most common job in this group, record verbatim: OPEN',
});
end();

// ---------------------------------------------------------------- Section C
begin('section_c', 'RECRUITMENT, TRAINING AND ORGANISATION OF WORK');
row({
  type: 'select_one cfillvac',
  name: 'cfillvac',
  label: 'Which of these statements best describes your approach to filling vacancies at this workplace?',
  guidance_hint: L(
    'SHOW CARD C1',
    'INTERVIEWER: For larger public sector employers, individual departments and agencies should be treated as separate employers',
  ),
});
row({
  type: 'select_multiple cfactors',
  name: 'cfactors',
  label: 'Which of the following factors are important when recruiting new employees?',
  guidance_hint: L('SHOW CARD C2', PROBE, ALL),
});
row({
  type: 'text',
  name: 'cfactors_other',
  label: 'Other (please specify CFACTOTH)',
  relevant: "selected(${cfactors}, 'other')",
});
row({
  type: 'select_multiple cspecix',
  name: 'cspecix',
  label: 'When filling vacancies, do you have any special procedures to encourage applications from the following groups?',
  guidance_hint: L('SHOW CARD C3', PROBE, ALL),
});
row({
  type: 'select_multiple catestw',
  name: 'catestw',
  label: 'When filling vacancies at this workplace, do you ever conduct any type of personality or attitude test for managerial positions, non-managerial positions or not at all?',
  guidance_hint: ALL,
});
row({
  type: 'select_multiple cptestw',
  name: 'cptestw',
  label: 'When filling vacancies at this workplace, do you ever conduct any type of performance or competency test for managerial positions, non-managerial positions or not at all?',
  guidance_hint: ALL,
});
row({
  type: 'select_one yes_no',
  name: 'clogvac',
  label: L(
    `Now I would like to ask you about the largest non-managerial group of employees here, that is, the ${LOG}.`,
    `Have there been any vacancies for ${LOG} in the past 12 months?`,
  ),
});
row({
  type: 'select_multiple cchanx',
  name: 'cchanx',
  label: 'This card shows a list of recruitment channels. What are the three most common channels used when trying to fill these vacancies?',
  guidance_hint: L('SHOW CARD C4', PROBE),
  relevant: "${clogvac} = '1'",
});
row({
  type: 'text',
  name: 'cchanx_other',
  label: 'Some other way (please specify COTHCHN)',
  relevant: "selected(${cchanx}, 'other')",
});
row({
  type: 'select_one yes_no',
  name: 'cinduct',
  label: `Is there a standard induction programme designed to introduce new ${LOG} to this workplace?`,
  guidance_hint: 'INTERVIEWER: Do not include probation periods.',
});
row({
  type: 'select_one cstuckin',
  name: 'cstuckin',
  label: `About how long does it normally take before new ${LOG} are able to do their job as well as more experienced employees already working here?`,
  guidance_hint: 'SHOW CARD C5',
});
row({
  type: 'note',
  name: 'trintro',
  label: 'I am now going to ask about training activities. I am interested in all training for which employees are given a break from their normal work duties, whether the training takes place at their immediate work position or elsewhere.',
});
row({
  type: 'select_one coffjob',
  name: 'coffjob',
  label: `What proportion of experienced ${LOG} have been given time off from their normal daily work duties to undertake training over the past 12 months?`,
  guidance_hint: 'SHOW CARD C6',
});
// Asked if COFFJOB = DK. "Don't know" is a CAPI key, not a printed code; an
// unanswered COFFJOB stands for it here.
row({
  type: 'select_one yes_no',
  name: 'coffchk',
  label: `Did any experienced ${LOG} have time off from their normal daily work duties to undertake training over the past 12 months?`,
  relevant: 'string-length(${coffjob}) = 0',
});
const anyTraining = "(string-length(${coffjob}) > 0 and ${coffjob} != '7') or ${coffchk} = '1'";
row({
  type: 'select_one ctrain',
  name: 'ctrain',
  label: `On average, how many days of training did experienced ${LOG} undertake over the past 12 months?`,
  relevant: anyTraining,
});
row({
  type: 'select_multiple chow',
  name: 'chow',
  label: 'Did this training cover any of the matters listed on this card?:',
  guidance_hint: L('SHOW CARD C7', PROBE, ALL),
  relevant: anyTraining,
});
row({
  type: 'select_one card_c8',
  name: 'cothjob',
  label: `Approximately, what proportion of ${LOG} are formally trained to be able to do jobs other than their own?`,
  guidance_hint: L('SHOW CARD C8', exact97),
});
row({
  type: 'integer',
  name: 'cothjobt',
  label: `(How many ${LOG} are formally trained to be able to do jobs other than their own?)`,
  guidance_hint: L('ENTER NUMBER', soft('COTHJOBT')),
  constraint: '. >= 1 and . <= 999997',
  relevant: "${cothjob} = '97'",
});
row({
  type: 'select_one card_c8',
  name: 'cothdo',
  label: `And approximately, what proportion of ${LOG} actually do jobs other than their own at least once a week?`,
  guidance_hint: L('SHOW CARD C8', exact97),
});
row({
  type: 'integer',
  name: 'cothdot',
  label: `(And approximately, how many ${LOG} actually do jobs other than their own at least once a week?)`,
  guidance_hint: L('ENTER NUMBER', soft('COTHDOT')),
  constraint: '. >= 1 and . <= 999997',
  relevant: "${cothdo} = '97'",
});
begin('card_c9_grid', `Using the scale on this card, to what extent would you say that individual ${LOG} here have...`, {
  appearance: 'table-list',
});
for (const [name, label] of [
  ['cvariety', '...variety in their work?'],
  ['cdiscrete', '...discretion over how they do their work? :'],
  ['control', '...control over the pace at which they work? :'],
  ['cdesign', '...involvement in decisions over how their work is organised?'],
]) {
  row({ type: 'select_one card_c9', name, label, guidance_hint: 'SHOW CARD C9' });
}
end();
row({
  type: 'select_one cteams',
  name: 'cteams',
  label: `What proportion, if any, of ${LOG} at this workplace work in formally designated teams?`,
  guidance_hint: L(
    'SHOW CARD C10',
    'INTERVIEWER: If respondent gives answer as an exact number you can code “97” here and record the number at the next question',
  ),
});
row({
  type: 'integer',
  name: 'cteamst',
  label: `(How many ${LOG} at this workplace work in formally designated teams?)`,
  guidance_hint: L('ENTER NUMBER', soft('CTEAMST')),
  constraint: '. >= 1 and . <= 999997',
  relevant: "${cteams} = '97'",
});
begin('cteamho', `Do each of the following statements apply to the way that teamworking operates among ${LOG} at this workplace?`, {
  appearance: 'table-list',
  relevant: "${cteams} != '7'",
});
for (const [name, label] of [
  ['cteamhoa', "...Team members depend on each other's work to be able to do their job."],
  ['cteamhoc', '...Team members jointly decide how the work is to be done.'],
  ['cteamhod', '...Teams are given responsibility for specific products or services.'],
]) {
  row({ type: 'select_one yes_no', name, label });
}
end();
row({
  type: 'integer',
  name: 'ccomput',
  label: L(
    'I would now like you to think about the whole workforce at this workplace.',
    'In this workplace, what percentage of all employees use computers as part of their normal work duties? Please answer to the nearest 10%.',
  ),
  guidance_hint: L(
    'INTERVIEWER: If respondent gives answer as an exact number you can code “997” here and record the number at the next question.',
    'SOFT CHECK if CCOMPUT>100 and CCOMPUT≠997: "The number entered here is a percentage. It should not be greater than 100."',
  ),
  constraint: '. >= 0 and . <= 997',
});
row({
  type: 'integer',
  name: 'ccomputt',
  label: '(In this workplace, how many employees use computers as part of their normal work duties?)',
  guidance_hint: L('ENTER NUMBER', soft('CCOMPUTT')),
  constraint: '. >= 1 and . <= 999997',
  relevant: '${ccomput} = 997',
});
end();

// ---------------------------------------------------------------- study
const URL = 'https://www.gov.uk/government/publications/the-2011-workplace-employment-relations-study-wers';
const PDF =
  'https://assets.publishing.service.gov.uk/media/5a74f4a0e5274a59fa716449/11-803-wers6-sixth-workplace-study-management-questionnaire.pdf';
// label_short (formtransform 0.9.0, varGrp/labl): the short name the source
// prints for a group, here the designator of the MQ's section headings ("SECTION C").
// Its questions carry printed variable codes (CFACTORS), not numbers; those
// are the names. Only groups (sections, grids,
// select_multiple) carry it: qwacback shows group labels only, so a standalone
// question's number is left out. A number's trailing period is dropped.
const LABEL_SHORT = {
  section_epq: "SECTION EPQ",
  section_c: "SECTION C",
};
for (const r of S) {
  const short = LABEL_SHORT[r.name];
  if (!short || !/^(begin_group|select_multiple)/.test(r.type ?? '')) continue;
  const { type, name, ...rest } = r;
  for (const k of Object.keys(r)) delete r[k];
  Object.assign(r, { type, name, label_short: short, ...rest });
  delete LABEL_SHORT[name];
}
if (Object.keys(LABEL_SHORT).length) throw new Error(`label_short: no group ${Object.keys(LABEL_SHORT)}`);

const study = {
  title: 'Workplace Employment Relations Study 2011 (WERS6): Management Questionnaire, Section C Recruitment, training and organisation of work',
  shortTitle: 'WERS6',
  id: 'q23-wers2011',
  authors: [
    { name: 'Department for Business, Innovation and Skills (BIS)', affiliation: 'UK Government' },
    { name: 'Advisory, Conciliation and Arbitration Service (Acas)', affiliation: 'Acas' },
    { name: 'Economic and Social Research Council (ESRC)', affiliation: 'ESRC' },
    { name: 'UK Commission for Employment and Skills (UKCES)', affiliation: 'UKCES' },
    { name: 'National Institute of Economic and Social Research (NIESR)', affiliation: 'NIESR' },
  ],
  producer: { name: 'Department for Business, Innovation and Skills (BIS)', affiliation: 'UK Government' },
  citation:
    'Department for Business, Innovation and Skills. (2011). Sixth Workplace Employment Relations Study (WERS6): Management questionnaire (MQ) (BIS/11/803). GOV.UK. ' +
    URL,
  url: URL,
  sourceLabel: 'BIS/11/803, WERS6 Management Questionnaire (MQ), Section C and the EPQ items it depends on (GOV.UK)',
  abstract:
    "This study contains Section C (recruitment, training and organisation of work) of the management interview of the sixth Workplace Employment Relations Study (WERS 2011), a CAPI interview with the most senior manager responsible for staff and employment relations at British workplaces, plus the employee-profile items it builds on (number of employees, occupational structure). The manager reports how vacancies are filled and which recruitment criteria, tests and channels are used, and, for the largest non-managerial occupational group, induction, the share given off-the-job training, training days and training content, multi-skilling, job autonomy and teamworking. A last item asks for the share of all employees using computers.",
  keywords: [
    'training',
    'recruitment',
    'skills',
    'job autonomy',
    'teamworking',
    'employment relations',
    'workplaces',
    'WERS',
  ],
  topics: ['Employment', 'Labour relations', 'Education and training'],
  nation: 'United Kingdom (Great Britain)',
  universe:
    'Workplaces in Great Britain with at least 5 employees; respondent: the most senior manager responsible for staff and employment relations at the workplace. Most of Section C refers to the largest non-managerial occupational group of the workplace.',
  anlyUnit: 'Workplaces',
  timePrd: '2011',
  dataKind: 'Questionnaire (CAPI specification)',
  license:
    'Open Government Licence v3.0 (https://www.nationalarchives.gov.uk/doc/open-government-licence/version/3/). GOV.UK publication page (' +
    URL +
    '): "All content is available under the Open Government Licence v3.0, except where otherwise stated" "© Crown copyright". The questionnaire PDF (' +
    PDF +
    ') states no other terms. Contains public sector information licensed under the Open Government Licence v3.0. Changes: format conversion of Section C and of the EPQ items ZALLEMPS, SOC_1 to SOC_9 and ZSOCDESC to DDI; wording unchanged; all other sections omitted. This DDI version was made by the qwac project (CorrelAid), not by the Department for Business, Innovation and Skills or the other WERS sponsors, and does not imply their endorsement.',
  copyright: '© Crown copyright',
};

writeFileSync(
  out,
  JSON.stringify(
    {
      study,
      survey: S,
      choices: C,
      settings: { default_language: 'English (en)', form_title: study.title, form_id: study.id },
    },
    null,
    2,
  ) + '\n',
);
console.log(`${S.length} survey rows, ${C.length} choices -> ${out}`);
