import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

// https://vite.dev/config/
export default defineConfig({
  plugins: [react()],
  server: {
    proxy: { '/api': 'http://127.0.0.1:8642' },
  },
  build: {
    // Keep the committed placeholder.html in dist/ alive across builds.
    emptyOutDir: false,
  },
})
