// Generates tools/survey-import/out/q19-eu-aes2022/form.json from a verbatim
// transcription of the Eurostat 2022 AES manual v1.3 (30 March 2023),
// Section 3 "Explanatory notes": "Model question" and "Codes or values" of each
// variable. Only the manual's literal text is used (labels, hints, choices);
// group labels are the manual's "Detailed topic" / "Topic" entries.
// The Excel questionnaire v1.2 is used only for item order and for routing
// where the manual's "Filter" field is silent. See SOURCE.md.
// verify-labels.py checks every text in form.json against the manual.
import { writeFileSync } from 'node:fs';

const out = process.argv[2];
const S = [];
const C = [];
const row = (r) => S.push(r);
const begin = (name, label, extra = {}) => row({ type: 'begin_group', name, label, ...extra });
const end = () => row({ type: 'end_group' });
const choices = (list, items) => items.forEach(([name, label]) => C.push({ list_name: list, name, label }));
const eq = (n, v) => `\${${n}} = '${v}'`;
const any = (...xs) => xs.join(' or ');

// ---- choice lists ("Codes or values"; "Not stated" / plain "Not applicable" omitted)
choices('yes_no', [['1', 'Yes'], ['2', 'No']]);
// GUIDE_1–3: coded as in the manual ("1 if selected, 2 if not selected").
choices('selected', [['1', 'selected'], ['2', 'not selected']]);
choices('hhtype', [
  ['1', 'One-person household'],
  ['2', 'Lone parent with at least one child aged less than 25'],
  ['3', 'Lone parent with all children aged 25 or more'],
  ['4', 'Couple without any child or children'],
  ['5', 'Couple with at least one child aged less than 25'],
  ['6', 'Couple with all children aged 25 or more'],
  ['7', 'Other type of household'],
]);
choices('sex', [['1', 'Male'], ['2', 'Female']]);
choices('citizen', [
  ['geo', 'Country of main citizenship (SCL GEO alpha-2 code); see annex 6'],
  ['STLS', 'Stateless'],
  ['FOR', 'Foreign citizenship but country unknown'],
  ['RNC', 'Recognised non-citizens'],
]);
choices('birthplace', [
  ['geo', 'Country of birth (SCL GEO alpha-2 code); see annex 6'],
  ['FOR', 'Foreign-born but country of birth unknown'],
]);
choices('restime', [
  ['99', 'Born in this country and never lived abroad for a period of at least 1 year'],
  ['00', 'Less than 1 year in the country but intention to stay at least 1 year in total (residence definition)'],
  ['01_69', 'Number of years in this country (since last establishing the place of usual residence in this country)'],
]);
choices('marstadefacto', [
  ['1', 'Person living with a legal or de facto partner'],
  ['2', 'Person not living with a legal or de facto partner'],
]);
choices('dropeduc', [['1', 'Yes, one'], ['2', 'Yes, several'], ['3', 'No']]);
choices('isced_prog', [
  ['10', 'ISCED 1 Primary education'],
  ['20', 'ISCED 2 Lower secondary education'],
  ['34', 'ISCED 3 Upper secondary education – general'],
  ['35', 'ISCED 3 Upper secondary education – vocational'],
  ['39', 'ISCED 3 Upper secondary education – orientation unknown'],
  ['44', 'ISCED 4 Post-secondary non-tertiary education – general'],
  ['45', 'ISCED 4 Post-secondary non-tertiary education – vocational'],
  ['49', 'ISCED 4 Post-secondary non-tertiary education – orientation unknown'],
  ['54', 'ISCED 5 Short-cycle tertiary education – general'],
  ['55', 'ISCED 5 Short-cycle tertiary education – vocational'],
  ['59', 'ISCED 5 Short-cycle tertiary education – orientation unknown'],
  ['60', 'ISCED 6 Bachelor’s or equivalent level'],
  ['70', 'ISCED 7 Master’s or equivalent level'],
  ['80', 'ISCED 8 Doctoral or equivalent level'],
]);
choices('mainstat', [
  ['10', 'Employed'],
  ['20', 'Unemployed'],
  ['32', 'Retired'],
  ['33', 'Unable to work due to long-standing health problems'],
  ['31', 'Student, pupil'],
  ['35', 'Fulfilling domestic tasks'],
  ['34', 'Compulsory military or civilian service'],
  ['36', 'Other'],
]);
choices('ftpt', [['1', 'Full-time job'], ['2', 'Part-time job']]);
choices('jobstat', [
  ['11', 'Self-employed person with employees'],
  ['12', 'Self-employed person without employees'],
  ['20', 'Employee'],
  ['30', 'Family worker (unpaid)'],
]);
choices('permjob', [['1', 'Fixed-term contract'], ['2', 'Permanent job']]);
choices('locsizefirm', [
  ['1_9', 'Exact number of persons, if between 1 and 9'],
  ['10', '10 to 19 persons'],
  ['11', '20 to 49 persons'],
  ['12', '50 to 249 persons'],
  ['13', '250 persons or more'],
  ['14', 'Do not know but less than 10 persons'],
  ['15', 'Do not know but 10 persons or more'],
]);
const parentLevels = [['1', 'At most lower secondary'], ['2', 'Upper secondary'], ['3', 'Tertiary']];
choices('hatfather', [...parentLevels, ['-2', 'Not applicable (father unknown)']]);
choices('hatmother', [...parentLevels, ['-2', 'Not applicable (mother unknown)']]);
choices('birthfather', [
  ['geo', 'Country of birth of the father (SCL GEO alpha-2 code); see annex 6'],
  ['FOR', 'Father foreign-born but country of birth of the father unknown'],
  ['-2', 'Not applicable (father unknown)'],
]);
choices('birthmother', [
  ['geo', 'Country of birth of the mother (SCL GEO alpha-2 code); see annex 6'],
  ['FOR', 'Mother foreign-born but country of birth of the mother unknown'],
  ['-2', 'Not applicable (mother unknown)'],
]);
choices('guidesource', [
  ['1', 'Education or training institutions (school, college, university, vocational education and training (VET) centre, institution for adult learning, validation centre, etc.)'],
  ['2', 'Public employment services'],
  ['3', 'Employer or prospective employer'],
  ['4', 'Other institutions/organisations'],
]);
choices('guideinter', [
  ['1', 'Interaction with a person: face to face, skype, phone, e-mail, dedicated websites etc.'],
  ['2', 'Interaction without a person: bots/web robots, online automatic applications as long as they provide for interaction'],
]);
choices('mainstat_start', [['1', 'Employed'], ['2', 'Unemployed'], ['3', 'Outside the labour market']]);
choices('fedcomp', [
  ['1', 'No, dropped out before the expected end'],
  ['2', 'No, it is still ongoing'],
  ['3', 'Yes, completed'],
]);
const fields20 = [
  ['2', 'Education'],
  ['3', 'Arts'],
  ['4', 'Humanities'],
  ['5', 'Language acquisition'],
  ['6', 'Literature and linguistics'],
  ['7', 'Social and behavioural sciences'],
  ['8', 'Journalism and information'],
  ['9', 'Business and administration'],
  ['10', 'Law'],
  ['11', 'Natural sciences, mathematics and statistics'],
  ['12', 'Computer use'],
  ['13', 'Database and network design and administration; software and applications development and analysis'],
  ['14', 'Engineering and engineering trades, manufacturing and processing'],
  ['15', 'Architecture and construction'],
  ['16', 'Agriculture, forestry, fisheries and veterinary'],
  ['17', 'Health'],
  ['18', 'Welfare'],
  ['19', 'Personal services, hygiene and occupational health services'],
  ['20', 'Security services, transport services'],
];
choices('fedfield', [['1', 'Generic programmes and qualifications'], ...fields20]);
choices('nfefield', [
  ['1a', 'Basic programmes and qualifications'],
  ['1b', 'Literacy and numeracy'],
  ['1c', 'Personal skills and development'],
  ...fields20,
]);
choices('place', [['1', 'Completely online'], ['2', 'Mostly online'], ['3', 'Mostly on-site'], ['4', 'Completely on-site']]);
choices('freq', [['1', 'Frequently'], ['2', 'Sometimes'], ['3', 'Never']]);
choices('freq_part', [['1', 'Frequently'], ['2', 'Sometimes'], ['3', 'Never'], ['4', 'No other participants']]);
choices('fedreason', [
  ['01a', 'To do my job better'],
  ['01b', 'To improve my career prospects'],
  ['02', 'To be less likely to lose my job'],
  ['03', 'To increase my possibilities of getting a job or changing a job/profession'],
  ['04', 'To start my own business'],
  ['05', 'I was obliged to participate'],
  ['06_7', 'To increase my knowledge/skills for my own general interest and curiosity'],
  ['08', 'To obtain a certificate'],
  ['09', 'To meet new people/for fun'],
]);
choices('worktime', [
  ['1', 'Only during paid working hours'],
  ['2', 'Mostly during paid working hours'],
  ['3', 'Mostly outside paid working hours'],
  ['4', 'Only outside paid working hours'],
  ['5', 'Not working at that time'],
]);
choices('paid', [
  ['1', 'Fully paid by yourself'],
  ['2', 'Partly paid by yourself and partly by somebody else'],
  ['3', 'Fully paid by somebody else'],
  ['4', 'Free activity'],
]);
choices('paidby', [
  ['1', 'Employer or prospective employer'],
  ['2', 'Public employment services'],
  ['3', 'Other public or private institutions'],
  ['4', 'A household member or a relative'],
]);
choices('usea', [['1', 'A lot'], ['2', 'A fair amount'], ['3', 'Very little'], ['4', 'Not at all']]);
choices('outcome', [
  ['1', 'Getting a (new) job'],
  ['3', 'Higher salary/wages'],
  ['2', 'Promotion in the job'],
  ['4', 'New tasks'],
  ['5', 'Better performance in present job'],
  ['6', 'Personal benefits (meet other people, refresh or acquire skills on general or specific subjects, etc.)'],
  ['7', 'No outcome yet'],
]);
choices('nfetype', [
  ['1', 'Courses'],
  ['2', 'Workshops and seminars'],
  ['3', 'Guided on-the-job training'],
  ['4', 'Private lessons'],
]);
choices('purp', [['1', 'Mainly job-related'], ['2', 'Mainly non-job-related']]);
choices('skills', [
  ['1', 'IT skills (general and professional)'],
  ['2', 'Management skills'],
  ['3', 'Team working skills'],
  ['4', 'Customer/clients/patients/students handling skills'],
  ['5', 'Problem solving skills'],
  ['6', 'Office administration skills'],
  ['7', 'Foreign language skills'],
  ['8', 'Technical and/or practical skills'],
  ['9', 'Communication skills'],
  ['10', 'Numeracy and/or literacy skills'],
  ['11', 'Health and safety skills'],
  ['12', 'Creative and musical skills, handicraft, cooking, gardening'],
  ['13', 'Mental strength, intra-personal or self-knowledge skills'],
  ['14', 'Physical skills'],
  ['15', 'Other'],
]);
choices('initia', [
  ['1', 'Suggested by employer or prospective employer'],
  ['2', 'Obliged by employer or prospective employer'],
  ['3', 'Suggested by public employment services'],
  ['4', 'Obliged by public employment services'],
  ['5', 'You yourself'],
]);
choices('nfereason', [
  ['01a', 'To do my job better'],
  ['01b', 'To improve my career prospects'],
  ['02', 'To be less likely to lose my job'],
  ['03', 'To increase my possibilities of getting a job or changing a job/profession'],
  ['04', 'To start my own business'],
  ['13', 'Because of organisational and/or technological changes at work'],
  ['11', 'Required by the employer or prospective employer or by law'],
  ['06_7', 'To increase my knowledge/skills for my own general interest and curiosity'],
  ['08', 'To obtain a certificate'],
  ['09', 'To meet new people/for fun'],
  ['10', 'For health reasons'],
  ['12', 'To do voluntary work better'],
]);
choices('provider', [
  ['1', 'Formal education institution'],
  ['2', 'Non-formal education and training institution'],
  ['4', 'Employer or prospective employer'],
  ['11', 'Other public or private institution'],
  ['8', 'Individual'],
]);
choices('cert', [
  ['1', 'Yes, required by the employer or prospective employer or a professional body or by law'],
  ['2', 'Yes, not required by the employer or prospective employer or a professional body or by law'],
  ['3', 'No (acknowledgement of attendance)'],
]);
// WANT_A / WANT_B answers; codes = the WANT code the answer maps to (manual).
choices('want_a', [['2', 'Yes'], ['1', 'No']]);
choices('want_b', [['4', 'Yes'], ['3', 'No']]);
choices('difftype', [
  ['01', 'Prerequisites'],
  ['02', 'Cost'],
  ['03a', 'Lack of employer support'],
  ['03b', 'Lack of public services support'],
  ['04', 'Schedule'],
  ['05', 'Distance'],
  ['07', 'Family responsibilities'],
  ['08a', 'Health'],
  ['08b', 'Age'],
  ['09', 'Other personal reasons'],
  ['10', 'No suitable education or training activity (offer)'],
  ['12', 'Negative previous learning experience'],
  ['13', 'Course was booked out'],
  ['14', 'Too few registrations'],
]);
choices('inf', [
  ['inffamily', 'Learning from a family member, a friend or a colleague'],
  ['infmaterial', 'Learning by using printed material (books, professional magazines, etc.)'],
  ['infdevice', 'Learning by using electronic devices (online or offline)'],
  ['infmuseum', 'Learning by guided tours in museums, historical or natural or industrial sites'],
  ['inflibraries', 'Learning by visiting learning centres (including libraries)'],
]);
choices('infpurp', [['1', 'At least one informal learning activity job-related'], ['2', 'Non-job-related']]);
choices('langlevel', [
  ['0', 'Hardly any language skills'],
  ['1', 'Beginner (basic user)'],
  ['2', 'Intermediate (independent user)'],
  ['3', 'Advanced (proficient user)'],
]);
choices('genhealth', [
  ['1', 'Very good'],
  ['2', 'Good'],
  ['3', 'Fair (neither good nor bad)'],
  ['4', 'Bad'],
  ['5', 'Very bad'],
]);
choices('gali', [['1', 'Severely limited'], ['2', 'Limited but not severely'], ['3', 'Not limited at all']]);
const ord = ['1st', '2nd', '3rd', '4th', '5th', '6th', '7th', '8th', '9th', '10th', '11th', '12th', '13th', '14th'];
choices('hhincome_band', [
  ['1', `0 to under [${ord[0]} cut-off value]`],
  ...ord.slice(0, 13).map((o, i) => [String(i + 2), `[${o} cut-off value] to under [${ord[i + 1]} cut-off value]`]),
  ['15', `[${ord[13]} cut-off value] or above`],
]);

// ---- Household composition -------------------------------------------------
begin('household', 'Household composition');
begin('hhnbpers', 'How many people usually live in your household? Please include yourself.', { appearance: 'field-list' });
row({ type: 'integer', name: 'hhnbpers_tot', label: 'Total number of members of the household (including the respondent)', constraint: '. >= 1 and . <= 98' });
row({ type: 'integer', name: 'hhnbpers_0_13', label: '0-13 years old members of the household', constraint: '. >= 0 and . <= 98' });
row({ type: 'integer', name: 'hhnbpers_14_24', label: '14-24 years old members of the household (including the respondent)', constraint: '. >= 0 and . <= 98' });
row({ type: 'integer', name: 'hhnbpers_25plus', label: 'Members of the household 25 years old and older (including the respondent)', constraint: '. >= 0 and . <= 98' });
end();
row({ type: 'select_one hhtype', name: 'hhtype', label: 'What is the type of your household?' });
end();

// ---- Person and household characteristics -----------------------------------
begin('person', 'Person and household characteristics');
row({
  type: 'select_one sex',
  name: 'sex',
  label: 'What is your sex?',
  hint: 'Depending on the data collection mode or information being available from administrative sources, it might usually not be necessary to ask the respondents directly.',
});
row({ type: 'integer', name: 'birthyear', label: 'What is your year of birth?', constraint: '. >= 1000 and . <= 9999' });
row({ type: 'select_one yes_no', name: 'birthpass', label: 'Have you already had your birthday this year?' });
row({
  type: 'select_one citizen',
  name: 'citizen',
  label: 'What is your citizenship?',
  hint: 'In case the information is available from administrative sources it might not be necessary to ask the respondents directly.',
});
row({ type: 'text', name: 'citizen_country', label: 'Country of main citizenship (SCL GEO alpha-2 code); see annex 6', relevant: eq('citizen', 'geo') });
row({
  type: 'select_one birthplace',
  name: 'birthplace',
  label: 'In which country were you born?',
  hint:
    "In situations (e.g. specific countries or regions) where this question may not capture appropriately the information on the place of usual residence of the individual's mother at the time of delivery, the following question should be asked: \"Which was the country of usual residence of your mother at the time when you were born?\"",
});
row({ type: 'text', name: 'birthplace_country', label: 'Country of birth (SCL GEO alpha-2 code); see annex 6', relevant: eq('birthplace', 'geo') });
row({
  type: 'select_one restime',
  name: 'restime',
  label:
    'For native-born persons the following questions are recommended:\n' +
    'Did you ever live [have your usual residence] abroad for a period of at least 1 year? y/n\n' +
    'If yes: Considering the date you last arrived in this country (established your usual residence in this country) – for how many years have you lived in this country since then? [please consider whole/completed years only]\n' +
    'For foreign-born persons the following question is recommended:\n' +
    'Considering the date you last arrived in this country (established your usual residence in this country) – for how many years have you lived in this country since then? [please consider whole/completed years only]',
  hint: 'Different questions need to be asked depending on whether a person was born in the country or not.',
});
row({
  type: 'integer',
  name: 'restime_years',
  label: 'Number of years in this country (since last establishing the place of usual residence in this country)',
  constraint: '. >= 1 and . <= 69',
  relevant: eq('restime', '01_69'),
});
row({
  type: 'select_one marstadefacto',
  name: 'marstadefacto',
  label: 'Are you living with a partner in the same household?',
  hint: 'Husband/wife/civil partner/partner/cohabitee',
});
end();

// ---- Educational attainment: education interrupted or abandoned -------------
begin('dropeduc_g', 'Educational attainment - details, including education interrupted or abandoned');
row({
  type: 'select_one dropeduc',
  name: 'dropeduc',
  label: 'Are there any formal education programmes that you started, but then stopped without completing?',
  hint: 'Please consider all formal education programmes that you ever started but then did not successfully complete. Please also include changes of field. If you are currently participating in a formal programme, this programme should not be taken into account. However, if you are currently taking a gap year, please include it.',
});
row({
  type: 'select_one isced_prog',
  name: 'dropeduclevel',
  label: 'What was the level of this formal education programme?',
  hint:
    'In case of several programmes not completed, please report the programme with the highest level; in case of several programmes not completed at the same level, please report the orientation of the most recent one.\n' +
    'Please use the national ISCED mapping to identify and classify the programme.',
  relevant: any(eq('dropeduc', '1'), eq('dropeduc', '2')),
});
end();

// ---- Labour market participation --------------------------------------------
begin('labour', 'Labour market participation');
row({
  type: 'select_one mainstat',
  name: 'mainstat',
  label: 'What is your current main labour status?\nor\nThinking about your current situation, to which of the following groups would you assign yourself?',
});
row({
  type: 'select_one yes_no',
  name: 'emp12m',
  label: 'Have you done any work for pay or profit, at any time during the last 12 months, including paid apprenticeship or paid traineeship, or unpaid work for a family business?',
  relevant: any(...['20', '32', '33', '31', '35', '34', '36'].map((c) => eq('mainstat', c))),
});
begin('job', 'Elementary job characteristics', { relevant: eq('mainstat', '10') });
row({ type: 'select_one ftpt', name: 'ftpt', label: 'Do you work full-time or part-time?' });
row({ type: 'select_one jobstat', name: 'jobstat', label: 'What is your professional status (based on your main job)?' });
row({
  type: 'select_one permjob',
  name: 'permjob',
  label: 'Do you have a fixed-term contract or a permanent contract, i.e. a contract without fixed end?',
  relevant: eq('jobstat', '20'),
});
row({ type: 'text', name: 'jobisco', label: 'What is your occupation (based on the main job)?', hint: 'ISCO-08 coded at 2-digit level (i.e. a 4-digit code – OC00); see annex 8' });
row({ type: 'text', name: 'locnace', label: 'What is the economic activity of the local unit where you are employed?', hint: 'NACE Rev. 2 coded at 2-digit level (i.e. a 3-digit code – X00); see annex 7' });
row({ type: 'select_one locsizefirm', name: 'locsizefirm', label: 'How many people are working at the local unit where you are employed?' });
row({
  type: 'integer',
  name: 'locsizefirm_exact',
  label: 'Exact number of persons, if between 1 and 9',
  constraint: '. >= 1 and . <= 9',
  relevant: eq('locsizefirm', '1_9'),
});
row({
  type: 'integer',
  name: 'jobtime',
  label: 'What year did you start working in your current main job?',
  hint: 'Year in which the person started working for current employer or as self-employed in main job',
  constraint: '. >= 1000 and . <= 9999',
});
end();
end();

// ---- Educational background / migrant background of the parents -------------
begin('parents_edu', 'Educational background');
row({ type: 'select_one hatfather', name: 'hatfather', label: 'What is the highest level of education successfully completed by your father (male guardian)?' });
row({ type: 'select_one hatmother', name: 'hatmother', label: 'What is the highest level of education successfully completed by your mother (female guardian)?' });
end();
begin('parents_birth', 'Citizenship and migrant background');
row({
  type: 'select_one birthfather',
  name: 'birthfather',
  label: 'In which country was your father born (male guardian)?',
  hint:
    "In situations (e.g. specific countries or regions) where this question may not capture appropriately the information on the place of usual residence of the mother of the person's father at the time of delivery, the following question should be asked: \"Which was the country of usual residence of your father's mother at the time when he was born?\"",
});
row({ type: 'text', name: 'birthfather_country', label: 'Country of birth of the father (SCL GEO alpha-2 code); see annex 6', relevant: eq('birthfather', 'geo') });
row({
  type: 'select_one birthmother',
  name: 'birthmother',
  label: 'In which country was your mother born (female guardian)?',
  hint:
    "In situations (e.g. specific countries or regions) where this question may not capture appropriately the information on the place of usual residence of the mother of the person's mother at the time of delivery, the following question should be asked: \"Which was the country of usual residence of your mother's mother at the time when she was born?\"",
});
row({ type: 'text', name: 'birthmother_country', label: 'Country of birth of the mother (SCL GEO alpha-2 code); see annex 6', relevant: eq('birthmother', 'geo') });
end();

// ---- Access to information on learning possibilities and guidance -----------
begin('guidance', 'Access to information on learning possibilities and guidance (12 months)');
row({
  type: 'select_one yes_no',
  name: 'seekinfo',
  label: 'During the last 12 months, have you looked for any information concerning learning possibilities (either on formal or non-formal education and training)?',
});
row({
  type: 'select_one selected',
  name: 'guide_1',
  label: 'During the last 12 months, have you received any counselling or advice on learning possibilities, including support to search for information and applying to learning possibilities?',
});
row({
  type: 'select_one selected',
  name: 'guide_2',
  label: 'During the last 12 months, have you done any screening or evaluation of your individual learning needs based on professional tests and/or interviews (i.e. test/interviews developed by a professional counsellor or advisor)?',
});
row({
  type: 'select_one selected',
  name: 'guide_3',
  label: 'During the last 12 months, have you received any advice or help on the procedure for validation or recognition of skills, competences or prior learning?',
});
const GUIDE = any(eq('guide_1', '1'), eq('guide_2', '1'), eq('guide_3', '1'));
row({
  type: 'select_multiple guidesource',
  name: 'guidesource',
  label: 'From which of the following institutions/organisations did you receive guidance on learning in the last 12 months? (mark all that apply)',
  relevant: GUIDE,
});
row({
  type: 'select_multiple guideinter',
  name: 'guideinter',
  label: 'With whom did you interact for the guidance received in the last 12 months? (mark all that apply)',
  relevant: GUIDE,
});
end();

// ---- Formal education --------------------------------------------------------
begin('fed_g', 'Participation in formal education activities (12 months)');
row({
  type: 'select_one yes_no',
  name: 'fed',
  label: 'During the last 12 months, that is since <<month, year>> have you participated (as a student or apprentice) in formal education or training (<any of the following formal education or training programmes>)?',
  hint: 'The question(s) for this variable should be phrased by countries in a way that the concept of formal education and training – designed to lead to a qualification/certificate recognised by the relevant national education authorities – is described as fully as possible. This can be achieved by listing the formal education programmes/qualifications (or categories thereof).',
});
row({
  type: 'integer',
  name: 'fednum',
  label: 'In how many formal education or training activities did you participate in during the last 12 months?',
  hint: 'Number of activities',
  constraint: '. >= 1 and . <= 99',
  relevant: eq('fed', '1'),
});
end();
const MRF = 'the most recent formal education or training activity (<FEDNAME>)';
begin('fed_recent', 'Most recent formal education activity - details (12 months)', { relevant: '${fednum} >= 1' });
row({ type: 'integer', name: 'fedstartyear', label: 'In which year did your most recent formal education or training activity (<FEDNAME>) start?', hint: 'The 4 digits of the year when the most recent formal education or training activity started', constraint: '. >= 1000 and . <= 9999' });
row({ type: 'integer', name: 'fedstartmonth', label: 'In which month did your most recent formal education or training activity (<FEDNAME>) start?', hint: 'Number of the month when the most recent formal education or training activity started (2 digits)', constraint: '. >= 1 and . <= 12' });
row({
  type: 'select_one mainstat_start',
  name: 'fedmainstat',
  label: 'What was your main labour status at the start of your most recent formal education or training activity (<FEDNAME>)?\nor\nThinking about the situation at the moment when you started your most recent formal education or training activity (<FEDNAME>), to which of the following groups would you assign yourself?',
});
row({ type: 'select_one isced_prog', name: 'fedlevel', label: `What is/was the level of ${MRF} <level of the corresponding formal education or training programme>?` });
row({ type: 'select_one fedcomp', name: 'fedcomp', label: `Have you successfully completed ${MRF}?` });
row({ type: 'select_one fedfield', name: 'fedfield', label: `What was the field of ${MRF}?`, relevant: '${fedlevel} >= 34' });
row({
  type: 'select_one place',
  name: 'fedplace',
  label: 'How was the instruction of the most recent formal education activity (<FEDNAME>) organised?\nOr\nWhere did the instruction hours of the most recent formal education activity (<FEDNAME>) take place?',
});
row({ type: 'select_one freq', name: 'fedonmat', label: `During ${MRF}, how frequently did the teaching staff make teaching materials available via the internet to participants, or point to such material?` });
row({ type: 'select_one freq', name: 'fedontea', label: `During ${MRF}, how frequently did you interact online with the teaching staff regarding training contents? (E.g. via the website of the training provider, e-mail, WhatsApp, social media, etc.)` });
row({ type: 'select_one freq', name: 'fedonpar', label: `During ${MRF}, how frequently did you interact online with other participants regarding training contents? (E.g. via the website of the training provider, e-mail, WhatsApp, social media, etc.)` });
row({ type: 'select_multiple fedreason', name: 'fedreason', label: `What were the reasons for participating in ${MRF}? (mark all that apply)` });
row({
  type: 'select_one fedreason',
  name: 'fedreasonmain',
  label: 'Among the reasons you previously selected, which one was the most important?',
  hint: 'Code of the reason from 01a to 09 as in variable FEDREASON',
  relevant: 'count-selected(${fedreason}) > 0',
});
row({
  type: 'select_one worktime',
  name: 'fedworktime',
  label: `Did ${MRF} take place during paid working hours? Please also take into account instruction hours that have been compensated for by free time.`,
  hint: 'FEDWORKTIME = 5 should only be asked if the respondent was not employed at the start of the activity (FEDMAINSTAT ≠ 1)',
});
row({ type: 'integer', name: 'fednbhours', label: `During the last 12 months, how many instruction hours did you attend for ${MRF}?`, hint: 'Total number of instruction hours', constraint: '. >= 0 and . <= 9999' });
row({ type: 'select_one paid', name: 'fedpaid', label: `Which of the following best describes the payment for tuition, registration, exam fees, and expenses for books or technical study means, regarding your studies for ${MRF}?` });
row({
  type: 'select_multiple paidby',
  name: 'fedpaidby',
  label: `Which of the following entities paid (partly/fully) for tuition, registration, exam fees, and expenses for books or technical study means, regarding your studies for ${MRF}? (mark all that apply)`,
  relevant: any(eq('fedpaid', '2'), eq('fedpaid', '3')),
});
row({ type: 'select_one usea', name: 'fedusea', label: `How much are you currently using the skills or knowledge that you acquired through ${MRF}?` });
row({
  type: 'select_multiple outcome',
  name: 'fedoutcome',
  label: `Have the new skills or knowledge acquired through ${MRF} helped you in any of the following ways? (mark all that apply)`,
  hint: 'FEDOUTCOME_7 (‘no outcome yet’) should only be selected if none of the other variables (FEDOUTCOME_1 to FEDOUTCOME_6) are selected.',
  constraint: "not(selected(., '7') and count-selected(.) > 1)",
});
row({
  type: 'select_one outcome',
  name: 'fedoutcomemain',
  label: 'Among the outcomes you previously selected, which one was the most important?',
  hint: 'Code of the outcome from 1 to 7 as in variable FEDOUTCOME',
  relevant: 'count-selected(${fedoutcome}) > 0',
});
end();

// ---- Non-formal education ----------------------------------------------------
begin('nfe_g', 'Participation in non-formal education activities (12 months)');
begin('nfe', 'During the last 12 months have you participated in any of the following activities - completed or ongoing - with the intention to improve knowledge or skills in any area (including hobbies) either during your free time or your working time?', { appearance: 'table-list' });
row({ type: 'select_one yes_no', name: 'nfecourse', label: 'a) Courses' });
row({ type: 'select_one yes_no', name: 'nfeworkshop', label: 'b) Workshops and seminars' });
row({ type: 'select_one yes_no', name: 'nfeguidejt', label: 'c) Guided on-the-job training', relevant: any(eq('mainstat', '10'), eq('emp12m', '1')) });
row({ type: 'select_one yes_no', name: 'nfelesson', label: 'd) Private lessons' });
end();
const NFE = any(eq('nfecourse', '1'), eq('nfeworkshop', '1'), eq('nfeguidejt', '1'), eq('nfelesson', '1'));
row({
  type: 'integer',
  name: 'nfenum',
  label: 'In how many such non-formal learning activities have you participated during the last 12 months?',
  hint: 'Number of activities',
  constraint: '. >= 1 and . <= 99',
  relevant: NFE,
});
end();
const NFEN = '${nfenum} >= 1';

// Per-activity block NFEACTxx, transcribed once (xx = 01).
begin('nfeact01', 'Non-formal education activities - details (12 months)', { relevant: NFEN });
row({
  type: 'select_one nfetype',
  name: 'nfeact01_type',
  label: 'What was the type of the non-formal learning activity (<NFEACTxx_NAME>)?',
  hint: '(NFEACTxx_TYPE = 3 might only be asked if MAINSTAT = 10 or EMP12M = 1, see implementation guidelines below)',
});
row({
  type: 'select_one mainstat_start',
  name: 'nfeact01_mainstat',
  label: 'What was your main labour status at the start of the non-formal learning activity (<NFEACTxx_NAME>)?\nor\nThinking about the situation at the moment when you started the non-formal learning activity (<NFEACTxx_NAME>), to which of the following groups would you assign yourself?',
});
row({ type: 'select_one purp', name: 'nfeact01_purp', label: 'Was the non-formal learning activity (<NFEACTxx_NAME>) mainly job-related?' });
row({
  type: 'select_one worktime',
  name: 'nfeact01_worktime',
  label: 'Did the non-formal learning activity (<NFEACTxx_NAME>) take place during paid working hours? Please also take into account instruction hours that have been compensated for by free time.',
  hint: 'NFEACTxx_WORKTIME = 5 should only be asked if the respondent was not employed at the start of the activity (NFEACTxx_MAINSTAT ≠ 1)',
});
row({ type: 'select_one yes_no', name: 'nfeact01_paidby', label: 'Did your employer or prospective employer pay partially or fully for the non-formal learning activity (<NFEACTxx_NAME>)?' });
end();

// Details on the 1st randomly-selected activity (NFERAND1).
const R1 = 'the 1st non-formal learning activity (<NFERAND1_NAME>)';
begin('nferand1', 'Non-formal education activities - details (12 months)', { relevant: NFEN });
row({ type: 'select_one nfefield', name: 'nfefield1', label: `What was the field of ${R1}?` });
row({ type: 'select_one skills', name: 'nfeskillsmain1', label: 'What were the main skills acquired through the non-formal learning activity (<NFERAND1_NAME>)? (select one item from the list only)' });
row({
  type: 'select_one place',
  name: 'nfeplace1',
  label: `How was (the instruction of) ${R1} organised?\nOr\nWhere did (the instruction hours of) ${R1} take place?`,
});
row({ type: 'select_one freq', name: 'nfeonmat1', label: `During ${R1}, how frequently did the teaching staff make teaching materials available via the internet to participants, or point to such material?` });
row({ type: 'select_one freq', name: 'nfeontea1', label: `During ${R1}, how frequently did you interact online with the teaching staff regarding training contents? (E.g. via the website of the training provider, e-mail, WhatsApp, social media, etc.)` });
row({ type: 'select_one freq_part', name: 'nfeonpar1', label: `During ${R1}, how frequently did you interact online with other participants regarding training contents? (E.g. via the website of the training provider, e-mail, WhatsApp, social media, etc.)` });
row({ type: 'select_one initia', name: 'nfeinitia1', label: `Who initiated ${R1}?`, relevant: eq('nfeact01_purp', '1') });
row({
  type: 'select_multiple nfereason',
  name: 'nfereason1',
  label: `What were the reasons for participating in ${R1}? (mark all that apply)`,
  hint: 'Reasons 01a, 01b, 02, 03, 04 and 13 should only be asked if the activity is job-related (NFEACTxx_PURP = 1).',
});
row({
  type: 'select_one nfereason',
  name: 'nfereasonmain1',
  label: 'Among the reasons you previously selected, which one was the most important?',
  hint: 'Code of the reason from 01a to 13 as in variable NFEREASON1',
  relevant: 'count-selected(${nfereason1}) > 0',
});
row({ type: 'integer', name: 'nfenbhours1', label: `During the last 12 months, how many instruction hours did you attend for ${R1}?`, hint: 'Total number of instruction hours', constraint: '. >= 0 and . <= 9999' });
row({ type: 'select_one paid', name: 'nfepaid1', label: `Which of the following best describes the payment for tuition, registration, exam fees, and expenses for books or technical study means, regarding your studies for ${R1}?` });
row({
  type: 'select_multiple paidby',
  name: 'nfepaidby1',
  label: `Which one of the following entities paid (partly/fully) for tuition, registration, exam fees, and expenses for books or technical study means, regarding your studies for ${R1}? (mark all that apply)`,
  relevant: any(eq('nfepaid1', '2'), eq('nfepaid1', '3')),
});
row({
  type: 'decimal',
  name: 'nfepaidval1',
  label: `How much did you personally or any member of your household or any relative pay for tuition, registration, exam fees, and expenses for books or technical study means, regarding your studies for ${R1}?`,
  hint: 'In euros',
  relevant: `${eq('nfepaid1', '1')} or ${eq('nfepaid1', '2')} or (${eq('nfepaid1', '3')} and selected(\${nfepaidby1}, '4'))`,
});
row({ type: 'select_one provider', name: 'nfeprovider1', label: `Who was the provider of ${R1}?` });
row({
  type: 'select_one cert',
  name: 'nfecert1',
  label: `Does/did ${R1} lead to a certificate?\nIf yes, is/was this certificate required by your employer or a professional body for the execution of your current or planned activity as employer or employee?`,
});
row({ type: 'select_one usea', name: 'nfeusea1', label: `How much are you currently using the skills or knowledge that you acquired through ${R1}?` });
row({
  type: 'select_multiple outcome',
  name: 'nfeoutcome1',
  label: `Have the new skills or knowledge acquired through ${R1} helped you in any of the following ways? (mark all that apply)`,
  hint: 'NFEOUTCOME1_7 (‘no outcome yet’) should only be selected if none of the other variables (NFEOUTCOME1_1 to NFEOUTCOME1_6) are selected.',
  constraint: "not(selected(., '7') and count-selected(.) > 1)",
});
row({
  type: 'select_one outcome',
  name: 'nfeoutcomemain1',
  label: 'Among the outcomes you previously selected, which one was the most important?',
  hint: 'Code of the outcome from 1 to 7 as in variable NFEOUTCOME1',
  relevant: 'count-selected(${nfeoutcome1}) > 0',
});
end();

// ---- Obstacles ---------------------------------------------------------------
begin('obstacles', 'Obstacles to participation in education and training (12 months)');
const PART = `${eq('fed', '1')} or ${NFE}`;
const NOPART = `${eq('fed', '2')} and ${eq('nfecourse', '2')} and ${eq('nfeworkshop', '2')} and not(${eq('nfeguidejt', '1')}) and ${eq('nfelesson', '2')}`;
row({
  type: 'select_one want_a',
  name: 'want_a',
  label: 'You have previously described the formal and/or non-formal education and training activities in which you have participated or are currently participating. Would you have liked to participate even more in such education and training activities in the past 12 months?',
  hint: 'WANT_A: for persons who participated in formal and/or non-formal education and training (FEDNUM ≥ 1 or NFENUM ≥ 1)',
  relevant: PART,
});
row({
  type: 'select_one yes_no',
  name: 'noneed_a',
  label: 'Is this because you did not need additional education or training?',
  hint: 'NONEED_A: for persons who participated in education and training and did not want to participate more (WANT_A = no / WANT = 1)',
  relevant: eq('want_a', '1'),
});
row({
  type: 'select_one want_b',
  name: 'want_b',
  label: 'You have previously indicated that you did not participate in any formal or non-formal education and training activities. Would you have liked to participate in such education and training activities in the past 12 months?',
  hint: 'WANT_B: for persons who did not participate in any formal or non-formal education and training (FEDNUM = 0 and NFENUM = 0)',
  relevant: NOPART,
});
row({
  type: 'select_one yes_no',
  name: 'noneed_b',
  label: 'Is this because you did not need education or training?',
  hint: 'NONEED_B: for persons who did not participate in education and training and did not want to participate (WANT_B = no / WANT = 3)',
  relevant: eq('want_b', '3'),
});
row({
  type: 'select_multiple difftype',
  name: 'difftype',
  label:
    'From the following list, what were the reasons for you not to participate in education and training? (mark all that apply)\nOr\nWhich of the following obstacles prevented you from participation in education and training? (mark all that apply)',
  hint: 'WANT = 2 or 4',
  relevant: any(eq('want_a', '2'), eq('want_b', '4')),
});
row({
  type: 'select_multiple difftype',
  name: 'difftype_noneed',
  label:
    'From the following list, what were the reasons for you not to want to participate in education and training? (mark all that apply)\nOr\nWhich of the following obstacles prevented you from wanting to participate in education and training? (mark all that apply)',
  hint: 'NONEED = 2',
  relevant: any(eq('noneed_a', '2'), eq('noneed_b', '2')),
});
row({
  type: 'select_one difftype',
  name: 'diffmain',
  label: 'Among the reasons/obstacles that you previously selected, which one was the most important?',
  hint: 'Code of the reason from 01 to 14 as in variable DIFFTYPE',
  relevant: 'count-selected(${difftype}) > 0 or count-selected(${difftype_noneed}) > 0',
});
end();

// ---- Informal learning ---------------------------------------------------------
begin('informal', 'Informal learning');
row({
  type: 'note',
  name: 'inf_intro',
  label:
    'Apart from the learning activities mentioned earlier, you can also deliberately acquire or improve knowledge and skills in a planned manner by self-learning or teaching yourself, whether at work or during your free time, alone or with others. Below is a list of various forms how this can be done.\n' +
    'Or\n' +
    'Apart from the learning activities mentioned earlier, you can also acquire or improve knowledge and skills on your own, outside of institutionalised education and training activities, either at work or during your free time. Yet, this learning must be planned beforehand and you must have goals for yourself regarding own efforts and outcomes. In being intentional and planned, such learning is distinguished from incidental or random learning, i.e. knowledge and skills that we acquire every day, without thinking much about it.',
  hint: 'Suggested introduction to the model question:',
});
row({
  type: 'select_multiple inf',
  name: 'inf',
  label: 'During the past 12 months, have you deliberately tried to learn anything on a particular topic or area, or are you currently doing it, through one of the forms mentioned below? (mark all that apply)',
});
row({
  type: 'select_one infpurp',
  name: 'infpurp',
  label: 'Were any of these informal learning activities mainly job-related?',
  relevant: 'count-selected(${inf}) > 0',
});
end();

// ---- Self-reported skills: languages ---------------------------------------------
begin('languages', 'Self-reported skills');
begin('langmother', 'Which language(s) is (are) your mother tongue(s)?', { appearance: 'field-list' });
row({ type: 'text', name: 'langmoth1', label: 'Code of the first language (ISO 639 alpha-3); see annex 9' });
row({ type: 'text', name: 'langmoth2', label: 'Code of the second language (ISO 639 alpha-3) or 000 (none); see annex 9' });
end();
begin(
  'langused',
  'Apart from your mother tongue(s), are there any other languages you speak or understand, even if just a little?\nIf the answer is yes:\nWhat other languages are they? List all languages.\nor\nWhat other languages are they? If there are several, list all languages ranked by your language proficiency.',
  { appearance: 'field-list' },
);
const nth = ['first', 'second', 'third', 'fourth', 'fifth', 'sixth', 'seventh'];
nth.forEach((o, i) => row({ type: 'text', name: `langused_${i + 1}`, label: `Code of the ${o} language or 000 (none) or -1` }));
end();
row({
  type: 'text',
  name: 'langbest1',
  label: 'Which of the languages you mentioned before (excluding your mother tongue(s)) do you know best?',
  hint: 'Code of the language (ISO 639 alpha-3)\nNo question is needed if the languages used were already ranked by proficiency level in the questions asked for LANGUSED, i.e. LANGBEST1 is already identified.',
  relevant: "${langused_1} != '' and ${langused_1} != '000' and ${langused_1} != '-1'",
});
row({
  type: 'select_one langlevel',
  name: 'langlevel1',
  label: 'Which one of the following descriptions best illustrates your knowledge about the first best known language you mentioned before (excluding mother tongue)?',
  hint: '(Descriptions to be taken from https://www.coe.int/en/web/portfolio/self-assessment-grid)',
  relevant: "${langbest1} != ''",
});
end();

// ---- Health ------------------------------------------------------------------
begin('health', 'Disability and other elements of Minimum European Health Module');
row({ type: 'select_one genhealth', name: 'genhealth', label: 'How is your health in general? Is it… very good, good, fair, bad, very bad.' });
row({
  type: 'select_one gali',
  name: 'gali',
  label: 'Are you limited because of a health problem in activities people usually do? Would you say you are… severely limited, limited but not severely, or not limited at all?',
});
row({
  type: 'select_one yes_no',
  name: 'gali_q2',
  label: 'Have you been limited for at least the past 6 months?',
  hint: 'If answer to QUESTION_1 is ‘severely limited’ or ‘limited but not severely’ ask QUESTION_2:',
  relevant: any(eq('gali', '1'), eq('gali', '2')),
});
end();

// ---- Income (asked last, as the Excel recommends) -------------------------------
begin('income', 'Total monthly household income');
row({
  type: 'decimal',
  name: 'hhincome',
  label:
    'Please take into account all sources of cash (money) income of all household members: Income from work (employment and self-employment), social benefits, and other sources of cash income (e.g. from capital and investments or inter-household cash transfers received).\n' +
    'What is the [net] monthly [cash/money] income of your household after deduction of taxes and contributions to social security and pensions (deducting cash transfers paid to other households, e.g. alimony or child support payments)? If income varies between months, please give a monthly average.',
  hint: '[Exact figure to be provided]',
});
row({
  type: 'select_one hhincome_band',
  name: 'hhincome_band',
  label: 'If you do not know the exact amount, could you estimate which of the following intervals best represents the total net monthly income of your household?',
  hint: 'The cut-off values between income bands should be presented in the national currency.',
});
end();

const study = {
  title: 'Adult Education Survey 2022 (AES): Model questions',
  shortTitle: '2022 AES',
  id: 'q19-eu-aes2022',
  authors: [{ name: 'Eurostat', affiliation: 'European Commission, Eurostat, Unit F-3: Labour market and lifelong learning' }],
  producer: { name: 'Eurostat' },
  citation:
    'Eurostat. (2023). 2022 AES manual: Methodological guidelines for the adult education survey (Version 1.3, 30 March 2023). European Commission, Eurostat, Unit F-3. https://circabc.europa.eu/w/browse/a81f03f5-b8e6-4b84-a93d-ef847ef85db0',
  url: 'https://circabc.europa.eu/w/browse/a81f03f5-b8e6-4b84-a93d-ef847ef85db0',
  sourceLabel: '2022 AES MANUAL V1.3, Section 3 "Explanatory notes" (Eurostat, CIRCABC)',
  abstract:
    'Model questions proposed by Eurostat for the 2022 wave of the EU Adult Education Survey, which national statistical institutes adapt into their own questionnaires. They ask adults about their household and background, labour status and job, access to guidance, participation in formal and non-formal education and training in the last 12 months (with details on the most recent formal activity and a randomly selected non-formal activity), obstacles to participation, informal learning, languages, health and household income. The wording follows the 2022 AES manual, version 1.3 (2023); Eurostat’s earlier model questionnaire for the 2022 AES (2021) is at https://ec.europa.eu/eurostat/cache/metadata/Annexes/trng_aes_12m_sims_hr_an_3.pdf.',
  keywords: ['adult education', 'lifelong learning', 'non-formal education', 'informal learning', 'formal education', 'obstacles to participation', 'AES', 'official statistics'],
  topics: ['Education', 'Adult learning', 'Continuing education'],
  nation: 'European Union',
  universe: 'Residents aged 18–69 in private households',
  anlyUnit: 'Individuals',
  timePrd: '2022',
  dataKind: 'Model questionnaire',
  license:
    '© European Union. Eurostat copyright notice (https://ec.europa.eu/eurostat/help/copyright-notice): "The copyright for the editorial content of this website, which is owned by the EU, is licensed under the Creative Commons Attribution 4.0 International licence. This means that you can re-use the content provided you acknowledge the source and indicate any changes you have made." "Reuse of statistical data, metadata, publications, and other dissemination tools published on this website for commercial or non-commercial purposes is authorised provided the source is acknowledged. The reuse policy of the European Commission is implemented by the Decision of 12 December 2011." (Commission Decision 2011/833/EU). "When reuse involves translations of publications or modifications to the data or text, this must be stated clearly to the end user of the information. A disclaimer regarding the non-responsibility of Eurostat shall be included." ' +
    'Changes: format conversion of the model questions in Section 3 of the 2022 AES manual (v1.3) to DDI; wording unchanged; technical and derived variables, transmission codes and repeated activity blocks omitted. Disclaimer: this DDI version was made by the qwac project (CorrelAid), not by Eurostat; Eurostat bears no responsibility for it.',
  copyright: '© European Union, 2023 (Eurostat)',
};

writeFileSync(
  out,
  JSON.stringify({ study, survey: S, choices: C, settings: { default_language: 'English (en)', form_title: study.title, form_id: study.id } }, null, 2) + '\n',
);
console.log(`rows: ${S.length}, choices: ${C.length}`);
