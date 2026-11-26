import {resolve} from 'path'
import {defineConfig} from 'vite'
import vue from '@vitejs/plugin-vue'
import {execSync} from 'child_process'
import {micromark} from 'micromark'

// Get commit SHA at build time
function getCommitSHA() {
    try {
        return execSync('git rev-parse HEAD', {encoding: 'utf-8'}).trim()
    } catch (error) {
        console.warn('Warning: Could not get git commit SHA:', error instanceof Error ? error.message : String(error))
        return process.env.VITE_COMMIT_SHA || ''
    }
}

const commitSHA = process.env.VITE_COMMIT_SHA || getCommitSHA()

// https://vite.dev/config/
export default defineConfig({
    plugins: [
        vue(),
        {
            name: 'compile-markdown',
            transform(source, id) {
                if (id.endsWith('.md')) {
                    return `export default ${JSON.stringify(micromark(source))}`
                }
            },
        },
        {
            name: 'inject-commit-sha',
            transformIndexHtml: () => [
                {tag: 'meta', attrs: {name: 'commit-sha', content: commitSHA}, injectTo: 'head'},
            ],
        },
    ],

    resolve: {
        alias: {
            '@': resolve(__dirname, './src'),
        },
    },

    // Define environment variables at build time
    define: {
        'import.meta.env.VITE_COMMIT_SHA': JSON.stringify(commitSHA),
    },

    server: {
        port: 3000,
        host: '0.0.0.0',
        proxy: {
            '^/(api|service)(/|$)': {
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
