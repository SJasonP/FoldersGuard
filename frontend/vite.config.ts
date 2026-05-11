import {defineConfig} from 'vite';
import react from '@vitejs/plugin-react';

export default defineConfig({
    plugins: [react()],
    clearScreen: false,
    build: {
        emptyOutDir: false,
    },
    server: {
        strictPort: true,
    },
});
