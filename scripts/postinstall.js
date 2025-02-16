const { exec } = require('child_process');
const os = require('os');

if (os.platform() === 'linux') {
    exec('npm install esbuild-linux-64@^0.15.18', (error, stdout, stderr) => {
        if (error) {
            console.error(`Error installing esbuild-linux-64: ${error}`);
            return;
        }
        console.log(`esbuild-linux-64 installed successfully: ${stdout}`);
    });
}