import tailwindcss from '@tailwindcss/vite'
import react from '@vitejs/plugin-react'
import { defineConfig } from 'vite'

export default defineConfig({
  plugins: [react(), tailwindcss()],
  server: {
    proxy: {
      '/api': process.env.AUTOVOLUME_BACKEND ?? 'http://localhost:8080',
    },
  },
  build: {
    // Inside the Go module so the backend can //go:embed it (embed can't reach ../).
    outDir: '../backend/web/dist',
    emptyOutDir: true,
  },
})
