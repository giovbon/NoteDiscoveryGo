// Boots a throwaway Basalt instance for the Playwright smoke tests.
//
// The vault is copied to a fresh temp directory on every run, so tests can
// create/edit notes without touching the fixtures committed in this folder.
import { execFileSync, spawn } from 'node:child_process';
import { cpSync, mkdtempSync, rmSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { dirname, join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

const here = dirname(fileURLToPath(import.meta.url));
const repoRoot = resolve(here, '../../..');
const port = process.env.E2E_PORT || '8123';

const workDir = mkdtempSync(join(tmpdir(), 'basalt-e2e-'));
const vaultDir = join(workDir, 'vault');
cpSync(join(here, 'fixtures/vault'), vaultDir, { recursive: true });

const binary = join(workDir, 'basalt');
execFileSync('go', ['build', '-o', binary, './cmd/basalt'], {
  cwd: repoRoot,
  stdio: 'inherit',
});

const child = spawn(binary, [], {
  cwd: repoRoot,
  env: {
    ...process.env,
    PORT: port,
    NOTES_DIR: vaultDir,
    INDEX_FILE: join(workDir, 'index.gob'),
    APP_NAME: 'Basalt E2E',
    DEFAULT_THEME: 'light',
    AUTHENTICATION_ENABLED: 'false',
    // Keep autosave fast so the edit test does not wait a full second.
    AUTOSAVE_DELAY_MS: '300',
  },
  stdio: 'inherit',
});

let shuttingDown = false;
function shutdown(code) {
  if (shuttingDown) return;
  shuttingDown = true;
  if (!child.killed) child.kill('SIGTERM');
  rmSync(workDir, { recursive: true, force: true });
  process.exit(code);
}

process.on('SIGTERM', () => shutdown(0));
process.on('SIGINT', () => shutdown(0));
child.on('exit', (code, signal) => shutdown(code ?? (signal ? 1 : 0)));
