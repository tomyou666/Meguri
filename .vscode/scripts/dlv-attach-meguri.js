#!/usr/bin/env node
/**
 * Waits for meguri (wails3 dev) and starts legacy RPC headless delve for VS Code.
 * Use launch.json: "debugAdapter": "legacy", "apiVersion": 2, "mode": "remote"
 */
const { spawn, execSync } = require('node:child_process');
const { existsSync } = require('node:fs');
const { homedir } = require('node:os');
const { join } = require('node:path');

const PROCESS_NAME = 'meguri';
const PORT = 2345;
const TIMEOUT_SEC = 180;
const POLL_MS = 500;

function sleep(ms) {
	return new Promise((resolve) => setTimeout(resolve, ms));
}

function getDlvPath() {
	const frontDir = join(__dirname, '..', '..', 'front');
	const execOpts = { encoding: 'utf8', stdio: ['ignore', 'pipe', 'ignore'] };

	try {
		const path = execSync('go tool -n dlv', { ...execOpts, cwd: frontDir })
			.trim()
			.split(/\r?\n/)[0];
		if (path && existsSync(path)) {
			return path;
		}
	} catch {
		// fall through
	}

	try {
		const which = process.platform === 'win32' ? 'where' : 'which';
		const path = execSync(`${which} dlv`, execOpts).trim().split(/\r?\n/)[0];
		if (path && existsSync(path)) {
			return path;
		}
	} catch {
		// fall through
	}

	const candidates = [
		join(
			homedir(),
			'go',
			'bin',
			process.platform === 'win32' ? 'dlv.exe' : 'dlv',
		),
	];
	for (const p of candidates) {
		if (existsSync(p)) {
			return p;
		}
	}

	throw new Error(
		'dlv not found. Run: make tools  (or: cd front && go mod download)',
	);
}

function findPid() {
	if (process.platform === 'win32') {
		try {
			const out = execSync(
				`tasklist /FI "IMAGENAME eq ${PROCESS_NAME}.exe" /FO CSV /NH`,
				{ encoding: 'utf8' },
			);
			const line = out
				.split(/\r?\n/)
				.map((l) => l.trim())
				.find((l) => l.length > 0);
			if (!line) {
				return null;
			}
			// "meguri.exe","12345",...
			const match = line.match(/"[^"]+","(\d+)"/);
			return match ? Number.parseInt(match[1], 10) : null;
		} catch {
			return null;
		}
	}

	try {
		const out = execSync(`pgrep -x ${PROCESS_NAME}`, {
			encoding: 'utf8',
		}).trim();
		const pid = out.split(/\r?\n/)[0];
		return pid ? Number.parseInt(pid, 10) : null;
	} catch {
		return null;
	}
}

async function waitForPid() {
	const deadline = Date.now() + TIMEOUT_SEC * 1000;
	console.log(
		`Waiting for process '${PROCESS_NAME}' (timeout ${TIMEOUT_SEC}s)...`,
	);
	while (Date.now() < deadline) {
		const pid = findPid();
		if (pid) {
			return pid;
		}
		await sleep(POLL_MS);
	}
	throw new Error(
		`Process '${PROCESS_NAME}' not found. Ensure wails3 dev has started the app.`,
	);
}

async function main() {
	const pid = await waitForPid();
	const dlv = getDlvPath();
	console.log(
		`Attaching to ${PROCESS_NAME} (PID ${pid}) on 127.0.0.1:${PORT} (legacy RPC headless)...`,
	);

	const child = spawn(
		dlv,
		[
			'attach',
			String(pid),
			'--headless',
			'--listen',
			`127.0.0.1:${PORT}`,
			'--api-version',
			'2',
			'--accept-multiclient',
			'--continue',
		],
		{ stdio: 'inherit' },
	);

	child.on('exit', (code, signal) => {
		if (signal) {
			process.kill(process.pid, signal);
		} else {
			process.exit(code ?? 1);
		}
	});
}

main().catch((err) => {
	console.error(err.message ?? err);
	process.exit(1);
});
