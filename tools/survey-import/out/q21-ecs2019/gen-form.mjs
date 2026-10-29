// Generates tools/survey-import/out/q21-ecs2019/form.json from a verbatim
// transcription of "European Company Survey 2019 – Online questionnaire for
// Management respondents" (Eurofound, version for publication, clean,
// 14 July 2020; ecs_2019_-_mm_questionnaire_for_publication_-_clean_14_july_2020.pdf).
// Only the PDF's literal text is used: question texts, item texts, answer
// options, the definitions printed after the (i) symbol, the italic
// question-type lines and the CAPITAL routing instructions. Bracketed variable
// labels ([MANAGER]) are not text: they become the `name`s. Printed lines
// wrapped by the layout are joined with a single space; nothing is joined
// across separate printed paragraphs (those become separate lines).
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
const choices = (list, items) => items.forEach((label, i) => C.push({ list_name: list, name: String(i + 1), label }));
const choicesCoded = (list, items) => items.forEach(([name, label]) => C.push({ list_name: list, name, label }));
const L = (...lines) => lines.filter(Boolean).join('\n');
const opt = (o) => Object.fromEntries(Object.entries(o).filter(([, v]) => v !== undefined && v !== ''));

const ONE = 'Tick one box only.';
const EACH_S = 'For each statement, Tick one box only.';
const BEST = 'Your best estimate is good enough.';

// ---- choice lists (codes 1..n in printed order; the PDF prints no codes) ----
choices('yes_no', ['Yes', 'No']);
choicesCoded('dk', [['1', 'Don’t know']]);
choicesCoded('mainact', [['nace', '[LIST OF NACE 1-DIGIT CODES]'], ['dk', 'Don’t know'], ['na', 'No answer']]);
choicesCoded('mainact2d', [
  ['nace', '[LIST OF NACE 2-DIGIT CODES (FOR RELEVANT 1-DIFIT CODE SELECTED IN MAINACT)]'],
  ['dk', 'Don’t know'],
  ['na', 'No answer'],
]);
choices('activity', [
  'Yes, this is mainly carried out internally at this establishment',
  'Yes, this is mainly carried out in collaboration with one or more other establishments within our company',
  'Yes, this is mainly carried out in collaboration with one or more other companies',
  'Yes, this is mainly contracted out',
  'No',
]);
choices('salesint', [
  'We do not engage in export (0%)',
  '1% to 24%',
  '25% to 49%',
  '50% or more',
  'Not applicable - our establishment does not engage in sales',
]);
choices('chowner', ['Yes, and it involved a change of management', 'Yes, but management remained the same', 'No']);
choices('inno', ['Yes, new to the market', 'Yes, new to the establishment, but not new to the market', 'No']);
choices('chemp', [
  'Increased by more than 10%',
  'Increased by up to 10%',
  'Stayed about the same',
  'Decreased by up to 10%',
  'Decreased by more than 10%',
]);
choices('pct', ['None at all', 'Less than 20%', '20% to 39%', '40% to 59%', '60% to 79%', '80% to 99%', 'All']);
choices('teasin', ['Most of them work in a single team', 'Most of them work in more than one team']);
choices('tauton', ['Team members decide among themselves', 'Tasks are distributed by a superior']);
choices('itchange', [
  'It has increased',
  'It has stayed about the same',
  'It has decreased',
  'This establishment does not use data analytics',
]);
choices('supcheck', [
  'Managers control whether employees follow the tasks assigned to them',
  'Managers create an environment in which employees can autonomously carry out their tasks',
]);
choices('important', ['Not at all important', 'Not very important', 'Fairly important', 'Very important']);
choices('often', ['Very often', 'Fairly often', 'Not very often', 'Never']);
choices('skillch', ['No change at all', 'Not very quickly', 'Fairly quickly', 'Very quickly']);
choices('wpsupp', [
  'Participation in training and professional development activities is only possible if workload and work schedules allow for it',
  'Workload and work schedules are adjusted to allow employees to participate in training and professional development activities',
]);
choices('hirinter', ['Always', 'Most of the time', 'Sometimes', 'Rarely', 'Never']);
choices('ca', [
  'A collective agreement negotiated at the national or cross-sectoral level',
  'A collective agreement negotiated at the sectoral level',
  'A collective agreement at the regional level',
  'A collective agreement negotiated at the establishment or company level',
  'A collective agreement negotiated on behalf of employees with a specific occupation',
  'Another type of collective agreement',
]);
choices('mmerconfirm', [
  'Trade union representation',
  'Trade union representative/shop steward',
  'Works council',
  'Public sector equivalent of works council',
  'Non-union staff representation',
  'Non-union staff representative',
  'Other country-specific bodies',
  'Other country-specific individuals',
  'There is no official employee representation',
]);
choices('constructive', ['Not at all constructive', 'Not very constructive', 'Fairly constructive', 'Very constructive']);
choices('indir', [
  'Management prefers to consult with the employee representation',
  'Management prefers to consult with employees directly',
  'Management prefers to consult with the employee representation and with employees directly',
  'Management prefers not to consult with employees or their representatives',
]);
choices('extent', ['To a great extent', 'To a moderate extent', 'To a small extent', 'Not at all']);
choices('regular', ['Yes, on a regular basis', 'Yes, on an irregular basis', 'No']);
choices('extent_dec', [
  'To a great extent',
  'To a moderate extent',
  'To a small extent',
  'Not at all',
  'No decisions were made in this area',
]);
choices('noer', ['There is no employee representation at this establishment']);
choices('motivated', ['Not at all motivated', 'Not very motivated', 'Fairly motivated', 'Very motivated']);
choices('difficult', ['Not at all difficult', 'Not very difficult', 'Fairly difficult', 'Very difficult']);
choices('relations', ['Very bad', 'Bad', 'Neither good nor bad', 'Good', 'Very good']);
choices('competitive', ['Not at all competitive', 'Not very competitive', 'Fairly competitive', 'Very competitive']);
choices('predictable', ['Not at all predictable', 'Not very predictable', 'Fairly predictable', 'Very predictable']);
choices('prodvol', ['It has increased', 'It has stayed about the same', 'It has decreased']);
choices('profit', [
  'Yes, we made a profit',
  'No, we made loss',
  'We broke even',
  'Not applicable, our company is a not-for-profit organisation',
]);
choices('chempfut', ['It will increase', 'It will stay about the same', 'It will decrease']);
choices('sexres', ['Woman', 'Man', 'Other']);
choices('posres', [
  'General manager',
  'Owner-manager',
  'Human Resource Manager (Personnel Manager)',
  'Training Manager',
  'Finance/Accounting Manager',
  'Other',
]);

// ---- helpers ---------------------------------------------------------------
const one = (name, list, label, hint, extra = {}) =>
  row(opt({ type: `select_one ${list}`, name, label, hint, ...extra }));

// Grid: one printed question above several items sharing one scale.
const grid = (name, list, label, hint, items, extra = {}) => {
  begin(name, label, opt({ appearance: 'table-list', hint, ...extra }));
  for (const [code, text] of items) row({ type: `select_one ${list}`, name: `${name}_${code}`, label: text });
  end();
};

// Table of number fields under one printed question (ranking questions,
// SKILL): a table-list group of integer rows.
const numTable = (name, label, hint, items, rowExtra = {}, extra = {}) => {
  begin(name, label, opt({ appearance: 'table-list', hint, ...extra }));
  for (const [code, text, rowHint] of items)
    row(opt({ type: 'integer', name: `${name}_${code}`, label: text, hint: rowHint, ...rowExtra }));
  end();
};

// "Enter the exact number of employees or tick one of the percentage
// categories." The question is printed once above two alternative answer
// fields, "Number" and "Percentage". A grid holds rows of one kind only, so
// the question becomes two one-row grids that share the printed question as
// their label: <code>_number (integer row "Number") and <code>_percentage
// (select_one pct row "Percentage"). The printed hint goes on both.
const numPct = (code, label, hint, { relevant, numGuide, pctGuide } = {}) => {
  begin(`${code}_number`, label, opt({ appearance: 'table-list', hint, relevant }));
  row(opt({ type: 'integer', name: `${code}_n`, label: 'Number', guidance_hint: numGuide, constraint: '. >= 0' }));
  end();
  begin(`${code}_percentage`, label, opt({ appearance: 'table-list', hint, relevant }));
  row(opt({ type: 'select_one pct', name: `${code}_pct`, label: 'Percentage', guidance_hint: pctGuide }));
  end();
};
const ENTER_EMP = 'Enter the exact number of employees or tick one of the percentage categories.';

// ============================================================================
// Questions 1–74
// ============================================================================

// 1. [WPSIZE_MM_N]
row({
  type: 'integer',
  name: 'wpsize_mm_n',
  label: 'Approximately how many people work in this establishment?',
  hint: L(
    'Please include all people that are employed by the establishment, regardless of the type of contract and regardless of whether they are physically present or carry out their work outside of the premises. Each employee is counted as one person, regardless whether they are working full-time or part-time. Your best estimate is good enough.',
    'When the term \'establishment\' is used in this survey, it refers the local site at which you work.',
    'Indicate a number or tick the ‘Don’t know’ box.',
  ),
  constraint: '. >= 0',
});
row({
  type: 'select_multiple dk',
  name: 'wpsize_mm_n_dk',
  label: 'Indicate a number or tick the ‘Don’t know’ box.',
});

// 2. [MAINACT]
one('mainact', 'mainact', 'Please select this establishment\'s main activity category from the list below.', ONE);

// 3. [YEARSOP]
row({
  type: 'integer',
  name: 'yearsop',
  label: 'Since what year has this establishment been carrying out this activity?',
  hint: L('Please disregard any changes in ownership structure or location.', 'Indicate the year [format – YYYY].'),
  constraint: '. >= 1000 and . <= 9999',
});

// 4.–6.
one('actprod', 'activity', 'Is this establishment engaged in the production of goods, assembly of parts or delivery of services?', ONE);
one('actdede', 'activity', 'Is this establishment engaged in the design or development of new products or services?', ONE);
one('ecommerce', 'yes_no', 'Does this establishment buy or sell goods or services on the internet? For instance, by using business-to-business portals, e-commerce etc.', ONE);

// 7. [SALESINT]
one('salesint', 'salesint', 'Since the beginning of 2016, what percentage of this establishment\'s sales were to customers in other countries?', L(
  'If this establishment was set up after 2016, please answer the question for the period since the establishment was set up.',
  'If this establishment is not engaged in sales, please select ‘Not applicable’. If this establishment does sell its products or services, but does not sell them internationally, please select \'We do not engage in export (0%)\'.',
  ONE,
));

const MGMT_DEF =
  '‘Management’ refers to the people that direct the establishment or a subset of it, e.g. units or departments focusing on different geographical markets or on specific functions (sales, human resources, logistic, purchases, production or lines of production). Managers have employees reporting to them, and are responsible for pay increases, bonuses or promotion of these employees.';

// 8. [CHOWNER]
one('chowner', 'chowner', 'Since the beginning of 2016, has there been any change in the ownership of the company to which this establishment belongs?', L(MGMT_DEF, ONE));

// 9.–11.
one('innoprod', 'inno', 'Since the beginning of 2016, has this establishment introduced any new or significantly changed products or services?', ONE);
one('innoproc', 'inno', 'Since the beginning of 2016, has this establishment introduced any new or significantly changed processes either for producing goods or supplying services?', ONE);
one('innomark', 'inno', 'Since the beginning of 2016, has this establishment introduced any new or significantly changed marketing methods?', ONE);

// 12. [CHEMP]
one('chemp', 'chemp', 'How has the total number of employees in this establishment changed since the beginning of 2016?', ONE);

// 13.–15.
numPct('manager', 'How many people that work in this establishment are managers? Your best estimate is good enough.', L(
  MGMT_DEF,
  'Enter the exact number of managers or tick one of the percentage categories.',
));
numPct('empperm', 'How many employees in this establishment have an open-ended contract? Your best estimate is good enough.', L(
  'When the term \'employees\' is used in this survey, it refers to the people that work in this establishment that are not managers.',
  ENTER_EMP,
));
numPct('emppart', 'How many employees in this establishment work part-time? Your best estimate is good enough.', L(
  '‘Part-time’ refers to working less than 35 hours per week.',
  ENTER_EMP,
));

// 16.–18. Teams
one('teamex', 'yes_no', 'A team is a group of people working together with a shared responsibility for the execution of allocated tasks. Team members can come from the same unit or from different units across the establishment. Do you have any teams fitting this definition in this establishment?', ONE, {
  guidance_hint: L('Yes ➔ GO TO QUESTION 17', 'No ➔ GO TO QUESTION 19'),
});
const team = "${teamex} = '1'";
one('teasin', 'teasin', 'With regard to the employees doing teamwork, do most of them work in a single team or do most of them work in more than one team?', ONE, { relevant: team });
one('tauton', 'tauton', 'Please think about the tasks to be performed by these teams. Who usually decides how the tasks are distributed within the team?', ONE, { relevant: team });

// 19. [ICTCOMP]
numPct('ictcomp', 'How many employees in this establishment use personal computers or laptops to carry out their daily tasks? Your best estimate is good enough.', L(
  'The term \'employees\' refers to the people that work in this establishment that are not managers.',
  ENTER_EMP,
), {
  numGuide: L('➔ IF 0, GO TO QUESTION 23', '➔ IF AT LEAST 1, GO TO QUESTION 20'),
  pctGuide: L('None at all ➔ GO TO QUESTION 23', 'GO TO QUESTION 20'),
});
const ict = "not(${ictcomp_n} = 0 or ${ictcomp_pct} = '1')";

// 20.–24. Digitalisation
one('ictapp', 'yes_no', 'Since the beginning of 2016, did this establishment purchase any software that was specifically developed or customised to meet the needs of the establishment?', ONE, { relevant: ict });
one('ictrob', 'yes_no', 'Robots are programmable machines that are capable of carrying out a complex series of actions automatically, which may include the interaction with people. Does this establishment use robots?', ONE, { relevant: ict });
one('itprodimp', 'yes_no', 'Does this establishment use data analytics to improve the processes of production or service delivery?', L(
  'Data analytics refers to the use of digital tools for analysing data collected at this establishment or from other sources.',
  ONE,
), { relevant: ict });
one('itpermon', 'yes_no', 'Does this establishment use data analytics to monitor employee performance?', ONE);
one('itperfmonuse', 'itchange', 'Since the beginning of 2016, how would you say the use of data analytics in this establishment has changed?', ONE, {
  guidance_hint: '(ASK IF QUESTION 22 = YES OR QUESTION 23 = YES)',
  relevant: "${itprodimp} = '1' or ${itpermon} = '1'",
});

// 25. [HIERA]
row({
  type: 'integer',
  name: 'hiera',
  label: 'Counting in the same way as is done in the examples below, how many hierarchical levels do you have in this establishment?',
  hint: 'Indicate the number of hierarchical levels.',
  constraint: '. >= 1',
});

// 26. [SUPCHECK]
one('supcheck', 'supcheck', 'Different establishments use different approaches to manage the way employees carry out their tasks. Which of these two statements best describes the general approach to management at this establishment? Please think about the approach that is used the most by managers.', ONE);

// 27. [DISC]
grid('disc', 'important', 'To be evaluated positively, how important is it that employees at this establishment show the following behaviour?', 'For each question, Tick one box only.', [
  ['help', 'Helping colleagues without being asked?'],
  ['hours', 'Staying longer when the work requires it?'],
  ['sugg', 'Making suggestions for improving the way things are done in the company?'],
]);

// 28. [MOTI]
grid('moti', 'often', 'How often are the following practices used to motivate and retain employees at this establishment?', EACH_S, [
  ['mon', 'Offering monetary rewards'],
  ['mis', 'Communicating a strong mission and vision, providing meaning to our work'],
  ['chal', 'Providing interesting and stimulating work'],
  ['learn', 'Providing opportunities for training and development'],
]);

// 29.–31.
numPct('compprobs', 'For how many employees in this establishment does their job include finding solutions to unfamiliar problems they are confronted with? Your best estimate is good enough.', ENTER_EMP);
numPct('comorg', 'For how many employees in this establishment does their job include independently organising their own time and scheduling their own tasks? Your best estimate is good enough.', ENTER_EMP);
numPct('pcwkmach', 'For how many employees at this establishment is the pace of work determined by machines or computers? Your best estimate is good enough.', ENTER_EMP);

// 32. [SKILL] Two alternative tables (percentages / exact numbers).
const SKILL_Q = L(
  'Some employees have about the right skills required to do their job, while others have a lower level of skills than needed and would need additional training. There can also be employees who have a higher level of skills than is needed in their job and can take on more demanding duties.',
  'Please think about the employees in this establishment, their current skills and the skills they need to do their job.',
);
numTable('skill_pct', SKILL_Q, L(
  'You can answer again with an exact number of employees, or by using the percentage of employees.',
  'QUESTION USING PERCENTAGES',
  'The answers to these three questions should add up to 100%. The term \'employees\' refers to the people that work in this establishment that are not managers. Your best estimate is good enough.',
), [
  ['skillsmatch', 'What percentage of employees have the skills that are about right to do the job?', '%'],
  ['overskill', 'What percentage of employees have a higher level of skills than is needed in their job?', '%'],
  ['underskill', 'What percentage of employees have a lower level of skills than is needed in their job?', '%'],
], { constraint: '. >= 0 and . <= 100' });
numTable('skill_n', SKILL_Q, L(
  'QUESTION USING AN EXACT NUMBER OF EMPLOYEES',
  'There are [WPSIZE_MM_N - MANAGER] people that work in this establishment that are not managers. The answers to these three questions should add up to this number. Your best estimate is good enough.',
), [
  ['skillsmatch', 'How many employees have the skills that are about right to do the job?'],
  ['overskill', 'How many employees have a higher level of skills than is needed in their job?'],
  ['underskill', 'How many employees have a lower level of skills than is needed in their job?'],
], { constraint: '. >= 0' });

// 33. [SKILLCH]
one('skillch', 'skillch', 'How quickly do the knowledge and skills needed from the employees in this establishment change?', L(
  'If this differs a lot between different groups of employees, please think of the largest group of employees in this establishment.',
  ONE,
));

// 34.–35.
numPct('contr', 'How many employees in this establishment are in jobs that require continuous training? Your best estimate is good enough.', L(
  'Continuous training: training that is received at a frequent, regular basis, and that is required to keep up with changes in the equipment that is being used, or changes in the requirements that come along with the certification that is needed to carry out the job.',
  ENTER_EMP,
));
numPct('learnnoneed', 'How many employees in this establishment are in jobs that offer limited opportunities to learn new things? Your best estimate is good enough.', ENTER_EMP);

// 36. [TRAINLEARN] ranking 1–3 (items carry no printed codes)
numTable('trainlearn', 'What are the most important ways through which employees in this establishment can become more skilled at their jobs? Please order them from most to least important, entering 1 for the most important down to 3 for the least important.', undefined, [
  ['1', 'Participating in training (on-the-job training as well as training courses followed at work, on location or on the internet)'],
  ['2', 'Learning from more experienced colleagues and supervisors, for example, by asking for help, through conversations, or by observing'],
  ['3', 'Learning by doing (trying out different ways to carry out their tasks; trial and error)'],
], { constraint: '. >= 1 and . <= 3' });

// 37.–38.
numPct('paidtrain', 'In 2018, how many employees in this establishment participated in training sessions on the establishment premises or at other locations during paid working time? Your best estimate is good enough.', ENTER_EMP);
numPct('onjob', 'In 2018, how many employees in this establishment have received on-the-job training or other forms of direct instruction in the workplace from more experienced colleagues? Your best estimate is good enough.', ENTER_EMP);

// 39. [TRAINATALL]
one('trainatall', 'yes_no', 'Since the beginning of 2016, has this establishment provided any training to any of its employees?', L(
  'Training includes on-the-job training as well as training courses followed at work, on location or on the internet.',
  ONE,
), { guidance_hint: L('Yes ➔ GO TO QUESTION 40', 'No ➔ GO TO QUESTION 42') });
const trained = "${trainatall} = '1'";

// 40. [WPSUPP]
one('wpsupp', 'wpsupp', 'Workload and work schedules can prevent the participation of employees in training activities. Which of the following statements best describes what happens in practice at this establishment?', ONE, { relevant: trained });

// 41. [TR]
grid('tr', 'important', 'How important are the following reasons for providing training to employees in this establishment?', EACH_S, [
  ['ski', 'Ensuring that employees have the skills they need to do their current job'],
  ['flex', 'Allowing employees to acquire skills they need to do other jobs than their current job. For instance, to allow for job rotation or career advancement.'],
  ['inn', 'Increasing the capacity of employees to articulate ideas about improvements to the establishment'],
  ['mot', 'Improving employee morale'],
], { relevant: trained });

// 42. [NEWLYHIR]
row({
  type: 'integer',
  name: 'newlyhir',
  label: 'Since the beginning of 2016, how many new employees have been hired at this establishment?',
  hint: 'Indicate a number.',
  guidance_hint: L('➔ IF 0, GO TO QUESTION 44', '➔ IF AT LEAST 1, GO TO QUESTION 43'),
  constraint: '. >= 0',
});

// 43. [HIRREADY]
numPct('hirready', 'How many of these newly recruited employees did not yet have the skills needed to do their job to the required level? Your best estimate is good enough.', ENTER_EMP, {
  relevant: '${newlyhir} >= 1',
});

// 44. [HIRINTER]
one('hirinter', 'hirinter', 'When recruiting, how often does management start by looking whether there are any suitable internal candidates?', ONE);

// 45. [HIR] ranking 1–4
numTable('hir', 'When recruiting new employees, how important are the following four characteristics? Please order them from most to least important, entering 1 for the most important down to 4 for the least important.', undefined, [
  ['att', 'The candidate has a personality that fits the company'],
  ['qual', 'The candidate has all the educational and vocational qualifications that are required'],
  ['exp', 'The candidate has professional experience in similar positions'],
  ['ecando', 'The candidate has the skills required to do the job and does not need any additional training'],
], { constraint: '. >= 1 and . <= 4' });

// 46. Variable pay (no question label printed): four items, each with
// "Number" or "Percentage" → a number grid and a percentage grid.
const VP_Q = 'In 2018, how many employees at this establishment received the following types of variable pay? Your best estimate is good enough.';
const VP = [
  ['bres', 'Payment by results, for example piece rates, provisions, brokerages or commissions'],
  ['inper', 'Variable extra pay linked to individual performance following management appraisal'],
  ['grpe', 'Variable extra pay linked to the performance of the team, working group or department'],
  ['prsh', 'Variable extra pay linked to the results of the company or establishment (profit sharing scheme)'],
];
begin('vp_number', VP_Q, { appearance: 'table-list', hint: L(ENTER_EMP, 'Number') });
for (const [c, t] of VP) row({ type: 'integer', name: `vp${c}_n`, label: t, constraint: '. >= 0' });
end();
begin('vp_percentage', VP_Q, { appearance: 'table-list', hint: L(ENTER_EMP, 'Percentage') });
for (const [c, t] of VP) row({ type: 'select_one pct', name: `vp${c}_pct`, label: t });
end();

// 47.–49. Collective agreements, employee representation
one('comemporg', 'yes_no', 'Is the company to which this establishment belongs a member of any employers\' organisation which participates in the negotiation of collective agreements?', ONE);
row({
  type: 'select_multiple ca',
  name: 'ca',
  label: 'Are the wages of any employees in this establishment set by any of the following types of collective agreements?',
  hint: 'Tick all that apply.',
  guidance_hint: '[COUNTRY SPECIFIC LIST]',
});
row({
  type: 'select_multiple mmerconfirm',
  name: 'mmerconfirm',
  label: 'Which of the following forms of official employee representation currently exist in this establishment?',
  hint: 'Tick all that apply.',
  guidance_hint: L('[COUNTRY SPECIFIC LIST]', 'GO TO QUESTION 50', 'There is no official employee representation ➔ GO TO QUESTION 53'),
  constraint: "not(selected(., '9') and count-selected(.) > 1)",
});
const er = "not(selected(${mmerconfirm}, '9'))";

// 50.–52.
one('eratt', 'constructive', 'How would you describe the general attitude of the employee representation at this establishment?', L(
  'The term establishment\' refers to the local site at which you work.',
  ONE,
), { relevant: er });
one('indir', 'indir', 'Would you say the management at this establishment prefers to consult with the employee representation or would management rather consult directly with employees?', ONE, { relevant: er });
one('ertrus', 'extent', 'In your opinion, to what extent does management at this establishment trust the employee representation?', ONE, { relevant: er });

// 53. (no question label printed)
grid('invol', 'regular', 'Which of the following practices are used to involve employees in this establishment in how work is organised?', EACH_S, [
  ['regmee', 'Meetings between employees and their immediate manager'],
  ['staffme', 'Meetings open to all employees at the establishment'],
  ['dissinf', 'Dissemination of information through newsletters, website, notice boards, email etc.'],
  ['somedi', 'Discussions with employees through social media or in online discussion boards'],
]);

// 54. [SUGGS]
one('suggs', 'yes_no', 'Does this establishment make use of suggestion schemes?', L(
  'Suggestion scheme: The collection of ideas and suggestions from the employees, voluntary and at any time, often by means of a physical or virtual ‘suggestion box’.',
  ONE,
));

row({
  type: 'note',
  name: 'intro_ei',
  label: 'The following questions are about your experiences with the direct involvement of employees in decision-making in this establishment.',
});

// 55.–56.
one('eidelay', 'extent', 'In your opinion, to what extent does involving employees cause delays in the implementation of changes?', ONE);
one('eicomp', 'extent', 'In your opinion, to what extent does involving employees in work organisation changes give the establishment a competitive advantage?', ONE);

// 57. [MMEPIN]
const PIN = [
  ['org', 'The organisation and efficiency of work processes'],
  ['dism', 'Dismissals'],
  ['train', 'Training and skill development'],
  ['time', 'Working time arrangements'],
  ['pay', 'Payment schemes'],
];
grid('mmepin', 'extent_dec', 'Please think of the period since the beginning of 2016. In your opinion, to what extent have employees directly influenced management decisions in the following areas?', EACH_S, PIN);

// 58. [MMERIN] The PDF reuses the item labels [MMEPINORG] … [MMEPINPAY] of 57;
// names here are mmerin_<item>.
row({
  type: 'select_multiple noer',
  name: 'mmerin_noer',
  label: 'If there is no employee representation at your establishment, tick the box below and continue with the next question.',
  guidance_hint: '➔ GO TO QUESTION 59',
});
grid('mmerin', 'extent_dec', 'Please think of the period since the beginning of 2016. In your opinion, to what extent has the employee representation influenced management decisions in the following areas?', EACH_S, PIN, {
  relevant: "not(selected(${mmerin_noer}, '1'))",
});

row({
  type: 'note',
  name: 'intro_hrm',
  label: 'The following questions are about difficulties with human resource management that establishments might face.',
});

// 59.–64.
one('sickleave', 'yes_no', 'Do you think the level of sickness leave in this establishment is too high?', ONE);
one('lowmot', 'motivated', 'Overall, how motivated do you think employees in this establishment are?', ONE);
one('findskill', 'difficult', 'How difficult is it for this establishment to find employees with the required skills?', ONE);
one('retainemp', 'difficult', 'How difficult is it for this establishment to retain employees?', ONE);
one('qwprel', 'relations', 'How would you describe the relations between management and employees in this establishment in general?', ONE);
one('indact', 'yes_no', 'Since the beginning of 2016, did any industrial action (such as a strike, work-to-rule, or manifestation) take place that was directly related to an issue specific to this establishment?', ONE);

// 65. [PMSTRATLP] ranking 1–4 (the PDF labels the question and its first item
// [PMSTRATLP]; the group is named pmstrat).
numTable('pmstrat', 'How important are the following four factors for the competitive success of this establishment? Please order them from most to least important, entering 1 for the most important down to 4 for the least important.', undefined, [
  ['lp', 'Offering products or services at lower prices than the competition'],
  ['bq', 'Offering products or services that are of better quality than those offered by the competition'],
  ['cust', 'Customising products or services to meet specific customer requirements'],
  ['nps', 'Regularly developing products, services or processes that are new to the market'],
], { constraint: '. >= 1 and . <= 4' });

// 66.–71.
one('competmark', 'competitive', 'How competitive would you say the market for the main products or services provided by this establishment is?', ONE);
one('pdemstab', 'predictable', 'How predictable would you say that the demand for the main products or services of this establishment is?', ONE);
one('prodvol', 'prodvol', 'Since the beginning of 2016, how has the amount of goods or services produced by this establishment changed?', ONE);
one('profit', 'profit', 'In 2018, did this establishment make a profit?', ONE, {
  guidance_hint: L('GO TO QUESTION 70', '➔ GO TO QUESTION 71'),
});
one('profplan', 'yes_no', 'Did this establishment expect to make a profit in 2018?', ONE, { relevant: "${profit} != '4'" });
one('chempfut', 'chempfut', 'In the next three years, how do you expect the total number of employees in this establishment to change?', ONE);

// 72. [MAINACT2D]
one('mainact2d', 'mainact2d', 'Please select this establishment\'s main activity category from the list below.', ONE, {
  relevant: "${mainact} = 'nace'",
});

// 73.–74. About the respondent
row({
  type: 'note',
  name: 'intro_res',
  label: 'Thank you very much for completing this questionnaire. Could you finally answer a few questions about yourself?',
});
one('sexres', 'sexres', 'Which of the following categories best captures how you think of yourself?', ONE);
one('posres', 'posres', 'What position do you hold?', ONE);

// ============================================================================
const URL_PDF =
  'https://assets.eurofound.europa.eu/f/279033/9287637f70/ecs_2019_-_mm_questionnaire_for_publication_-_clean_14_july_2020.pdf';
const URL_PAGE = 'https://www.eurofound.europa.eu/en/surveys-and-data/surveys/european-company-survey/ecs-2019';

// label_short (formtransform 0.9.0, varGrp/labl): the short name the source
// prints for a group, here the number printed before each question ("13. [MANAGER]"); the
// number and percentage tables of one question share it. Only groups (sections, grids,
// select_multiple) carry it: qwacback shows group labels only, so a standalone
// question's number is left out. A number's trailing period is dropped.
const LABEL_SHORT = {
  wpsize_mm_n_dk: "1",
  manager_number: "13",
  manager_percentage: "13",
  empperm_number: "14",
  empperm_percentage: "14",
  emppart_number: "15",
  emppart_percentage: "15",
  ictcomp_number: "19",
  ictcomp_percentage: "19",
  disc: "27",
  moti: "28",
  compprobs_number: "29",
  compprobs_percentage: "29",
  comorg_number: "30",
  comorg_percentage: "30",
  pcwkmach_number: "31",
  pcwkmach_percentage: "31",
  skill_pct: "32",
  skill_n: "32",
  contr_number: "34",
  contr_percentage: "34",
  learnnoneed_number: "35",
  learnnoneed_percentage: "35",
  trainlearn: "36",
  paidtrain_number: "37",
  paidtrain_percentage: "37",
  onjob_number: "38",
  onjob_percentage: "38",
  tr: "41",
  hirready_number: "43",
  hirready_percentage: "43",
  hir: "45",
  vp_number: "46",
  vp_percentage: "46",
  ca: "48",
  mmerconfirm: "49",
  invol: "53",
  mmepin: "57",
  mmerin_noer: "58",
  mmerin: "58",
  pmstrat: "65",
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
  title: 'European Company Survey 2019',
  subtitle: 'Online questionnaire for Management respondents',
  shortTitle: 'ECS 2019',
  id: 'q21-ecs2019',
  authors: [
    { name: 'Eurofound', affiliation: 'European Foundation for the Improvement of Living and Working Conditions' },
    { name: 'Cedefop', affiliation: 'European Centre for the Development of Vocational Training' },
  ],
  producer: { name: 'European Foundation for the Improvement of Living and Working Conditions (Eurofound)' },
  citation:
    'Eurofound & Cedefop. (2020). European Company Survey 2019: Online questionnaire for management respondents [Questionnaire, version for publication, 14 July 2020]. European Foundation for the Improvement of Living and Working Conditions. ' +
    URL_PDF,
  url: URL_PAGE,
  sourceLabel: 'European Company Survey 2019 – Online questionnaire for Management respondents (Eurofound, 14 July 2020)',
  abstract:
    'Management questionnaire of the fourth European Company Survey (2019), run jointly by Eurofound and Cedefop as a push-to-web survey of establishments in the EU27 and the United Kingdom. A senior manager reports on the establishment\'s work organisation, digitalisation (computer use, custom software, robots, data analytics), skills match and skill requirements, training and learning, recruitment, variable pay, employee voice and representation, human resource difficulties and business strategy and performance. The published version shows the wording for subsidiary sites of multi-site companies.',
  keywords: ['establishment survey', 'work organisation', 'skills', 'training', 'digitalisation', 'human resource management', 'employee participation', 'social dialogue', 'variable pay', 'innovation'],
  topics: ['Workplace practices', 'Skills and training', 'Digitalisation', 'Social dialogue'],
  nation: 'EU27 and United Kingdom',
  universe:
    'Establishments with 10 or more employees in economic sectors engaged in market activities in the 27 EU Member States and the United Kingdom; respondent: the most senior manager in charge of personnel (management questionnaire).',
  anlyUnit: 'Establishments',
  timePrd: '2019',
  dataKind: 'Questionnaire',
  license:
    'Eurofound copyright, reuse with attribution (not a Creative Commons licence). ECS 2019 overview report (Eurofound and Cedefop, 2020, EF20001): "© European Foundation for the Improvement of Living and Working Conditions, 2020 and European Centre for the Development of Vocational Training, 2020. Reproduction is authorised provided the source is acknowledged. For any use or reproduction of photos or other material that is not under the Eurofound copyright, permission must be sought directly from the copyright holders." Eurofound legal information (https://www.eurofound.europa.eu/en/legal-information), Copyright: “The reuse of any information of this website is authorised for commercial and non-commercial purposes, under the following conditions: The re-user is obliged to acknowledge the source of the document. The original meaning or the message of the documents should not be distorted. Eurofound cannot be held liable for any consequence stemming from the reuse.” The questionnaire PDF itself carries no copyright notice. ' +
    'Changes: format conversion of the management questionnaire to DDI, wording unchanged; welcome and closing texts, the employee-representative contact block and the report/follow-up consent questions omitted; layout omitted. This DDI version was made by the qwac project (CorrelAid), not by Eurofound or Cedefop.',
  copyright:
    '© European Foundation for the Improvement of Living and Working Conditions (Eurofound) and European Centre for the Development of Vocational Training (Cedefop), 2020',
};

writeFileSync(
  out,
  JSON.stringify({ study, survey: S, choices: C, settings: { default_language: 'English (en)', form_title: study.title, form_id: study.id } }, null, 2) + '\n',
);
console.log(`rows: ${S.length}, choices: ${C.length}`);
