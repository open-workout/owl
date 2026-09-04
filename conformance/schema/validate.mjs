#!/usr/bin/env node
// Validates every fixture under conformance/{parse,resolve,progress}/
// against the JSON Schemas in spec/ and conformance/schema/. Run via
// `npm run validate` (see package.json) or directly with Node.
//
// This only checks structural shape — it cannot check that a fixture's
// `expected.json` is the *correct* output for its input, since that
// requires an actual implementation of resolve()/progress() to compare
// against (see reference/README.md). It exists so a malformed fixture
// (wrong field name, missing required key) fails fast in CI rather than
// silently sitting in the corpus.

import { readFileSync, readdirSync, statSync } from 'node:fs';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import Ajv2020 from 'ajv/dist/2020.js';

const ROOT = join(dirname(fileURLToPath(import.meta.url)), '..', '..');
const ajv = new Ajv2020({ strict: false, allErrors: true });

const readJson = (p) => JSON.parse(readFileSync(p, 'utf8'));
const compile = (relPath) => ajv.compile(readJson(join(ROOT, relPath)));

const schemas = {
  canonicalForm: compile('spec/canonical-form.schema.json'),
  state: compile('conformance/schema/state.schema.json'),
  cursor: compile('conformance/schema/cursor.schema.json'),
  sessionLog: compile('conformance/schema/session-log.schema.json'),
  parseError: compile('conformance/schema/parse-error.schema.json'),
  session: compile('conformance/schema/session.schema.json'),
};

let failures = 0;
let checked = 0;

function validate(label, schemaKey, absPath) {
  checked++;
  const data = readJson(absPath);
  if (!schemas[schemaKey](data)) {
    failures++;
    console.log(`FAIL [${schemaKey}] ${absPath.replace(ROOT + '/', '')}`);
    console.log(JSON.stringify(schemas[schemaKey].errors, null, 2));
  }
}

function subdirs(dir) {
  if (!statSync(dir, { throwIfNoEntry: false })) return [];
  return readdirSync(dir, { withFileTypes: true })
    .filter((e) => e.isDirectory())
    .map((e) => join(dir, e.name));
}

function fileExists(p) {
  return statSync(p, { throwIfNoEntry: false }) !== undefined;
}

// parse/<case>/: source.owl + exactly one of expected.json | expected-error.json
for (const dir of subdirs(join(ROOT, 'conformance/parse'))) {
  const expected = join(dir, 'expected.json');
  const expectedError = join(dir, 'expected-error.json');
  const hasExpected = fileExists(expected);
  const hasError = fileExists(expectedError);
  if (hasExpected === hasError) {
    failures++;
    console.log(`FAIL [parse] ${dir}: expected exactly one of expected.json / expected-error.json`);
    continue;
  }
  if (hasExpected) validate('parse', 'canonicalForm', expected);
  else validate('parse', 'parseError', expectedError);
}

// resolve/<case>/: canonical.json, state.json, cursor.json, expected.json (session)
for (const dir of subdirs(join(ROOT, 'conformance/resolve'))) {
  validate('resolve', 'canonicalForm', join(dir, 'canonical.json'));
  validate('resolve', 'state', join(dir, 'state.json'));
  validate('resolve', 'cursor', join(dir, 'cursor.json'));
  validate('resolve', 'session', join(dir, 'expected.json'));
}

// progress/<case>/: canonical.json, state.json (before), log.json, expected.json (state after)
for (const dir of subdirs(join(ROOT, 'conformance/progress'))) {
  validate('progress', 'canonicalForm', join(dir, 'canonical.json'));
  validate('progress', 'state', join(dir, 'state.json'));
  validate('progress', 'sessionLog', join(dir, 'log.json'));
  validate('progress', 'state', join(dir, 'expected.json'));
}

console.log(`\n${checked} file(s) checked, ${failures} failure(s).`);
process.exit(failures === 0 ? 0 : 1);
