import {resolve} from 'path'
import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import {execSync} from 'child_process'

// Get commit SHA at build time
function getCommitSHA() {
    try {
        return execSync('git rev-parse HEAD', {encoding: 'utf-8'}).trim()
    } catch (error) {
        console.warn('Warning: Could not get git commit SHA:', error.message)
        return process.env.VITE_COMMIT_SHA || ''
    }
}

// https://vite.dev/config/
export default defineConfig({
    plugins: [
        vue()
    ],

    resolve: {
        alias: {
            '@': resolve(__dirname, './src'),
        },
    },

    // Define environment variables at build time
    define: {
        'import.meta.env.VITE_COMMIT_SHA': JSON.stringify(process.env.VITE_COMMIT_SHA || getCommitSHA()),
    },

    server: {
        port: 3000,
        host: '0.0.0.0',
        proxy: {
            '/api': {
                target: 'http://127.0.0.1:8080',
                secure: false,
                changeOrigin: true,
            },
        }
    },

    build: {
        sourcemap: false,
        rollupOptions: {
            output: {
                entryFileNames: 'assets/[hash].js',
                chunkFileNames: 'assets/[hash].js',
                assetFileNames: 'assets/[hash].[ext]',
            },
        },
    }
})