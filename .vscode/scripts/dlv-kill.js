#!/usr/bin/env node
const { execSync } = require('node:child_process');

try {
	if (process.platform === 'win32') {
		execSync('taskkill /F /IM dlv.exe', { stdio: 'ignore' });
	} else {
		execSync('pkill -x dlv', { stdio: 'ignore' });
	}
} catch {
	// no running dlv — ok
}
