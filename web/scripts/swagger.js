import * as path from 'node:path'
import {generateApi} from 'swagger-typescript-api'

await generateApi({
    input: path.resolve(import.meta.dirname, '../../docs/swagger.json'),
    output: path.resolve(import.meta.dirname, '../src/api'),
    fileName: 'api.ts',
    generateClient: true,
    httpClientType: 'axios',
    extractRequestBody: true,
    extractResponseBody: true,
})
